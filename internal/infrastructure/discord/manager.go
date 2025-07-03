package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
	
	"github.com/gorilla/websocket"
	"midjourney-proxy-go/internal/infrastructure/config"
	"midjourney-proxy-go/pkg/logger"
)

// Manager Discord连接管理器
type Manager struct {
	config     config.DiscordConfig
	logger     logger.Logger
	instances  map[string]*Instance
	selector   *AccountSelector
	mutex      sync.RWMutex
	started    bool
	stopCh     chan struct{}
}

// Instance Discord实例
type Instance struct {
	ID        string
	Account   config.DiscordAccount
	Connected bool
	LastPing  time.Time
	conn      *websocket.Conn
	ctx       context.Context
	cancel    context.CancelFunc
	heartbeat chan struct{}
	messages  chan DiscordMessage
}

// DiscordMessage Discord消息结构
type DiscordMessage struct {
	Op   int             `json:"op"`
	D    json.RawMessage `json:"d"`
	S    int             `json:"s"`
	T    string          `json:"t"`
}

// HelloPayload Discord Hello消息
type HelloPayload struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

// IdentifyPayload Discord Identify消息
type IdentifyPayload struct {
	Token      string `json:"token"`
	Properties struct {
		OS      string `json:"$os"`
		Browser string `json:"$browser"`
		Device  string `json:"$device"`
	} `json:"properties"`
	Intents int `json:"intents"`
}

// NewManager 创建Discord管理器
func NewManager(config config.DiscordConfig, logger logger.Logger) *Manager {
	return &Manager{
		config:    config,
		logger:    logger,
		instances: make(map[string]*Instance),
		selector:  NewAccountSelector(AccountSelectBestWaitIdle, logger),
		stopCh:    make(chan struct{}),
	}
}

// Start 启动Discord管理器
func (m *Manager) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.started {
		return nil
	}
	
	m.logger.Info("Starting Discord manager...")
	
	// 初始化Discord实例
	for _, account := range m.config.Accounts {
		if account.Enabled {
			instance := &Instance{
				ID:        account.ID,
				Account:   account,
				Connected: false,
				LastPing:  time.Now(),
				heartbeat: make(chan struct{}),
				messages:  make(chan DiscordMessage, 100),
			}
			
			m.instances[account.ID] = instance
			
			// 启动WebSocket连接
			go m.startInstance(instance)
		}
	}
	
	m.started = true
	m.logger.Info("Discord manager started")
	
	return nil
}

// Stop 停止Discord管理器
func (m *Manager) Stop() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if !m.started {
		return
	}
	
	m.logger.Info("Stopping Discord manager...")
	
	close(m.stopCh)
	
	// 停止所有实例
	for _, instance := range m.instances {
		m.stopInstance(instance)
	}
	
	m.started = false
	m.logger.Info("Discord manager stopped")
}

// GetInstance 获取Discord实例
func (m *Manager) GetInstance(id string) *Instance {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return m.instances[id]
}

// GetAvailableInstance 获取可用的Discord实例
func (m *Manager) GetAvailableInstance() *Instance {
	return m.GetAvailableInstanceWithFilter(nil)
}

// GetAvailableInstanceWithFilter 根据过滤器获取可用的Discord实例
func (m *Manager) GetAvailableInstanceWithFilter(filter *entity.AccountFilter) *Instance {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return m.selector.SelectAccount(m.instances, filter)
}

// GetAllInstances 获取所有Discord实例
func (m *Manager) GetAllInstances() map[string]*Instance {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	result := make(map[string]*Instance)
	for k, v := range m.instances {
		result[k] = v
	}
	
	return result
}

// AddAccount 添加Discord账号
func (m *Manager) AddAccount(account config.DiscordAccount) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if account.Enabled {
		instance := &Instance{
			ID:        account.ID,
			Account:   account,
			Connected: false,
			LastPing:  time.Now(),
			heartbeat: make(chan struct{}),
			messages:  make(chan DiscordMessage, 100),
		}
		
		m.instances[account.ID] = instance
		
		// 如果管理器已启动，立即启动新实例
		if m.started {
			go m.startInstance(instance)
		}
	}
	
	return nil
}

// RemoveAccount 移除Discord账号
func (m *Manager) RemoveAccount(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if instance, exists := m.instances[id]; exists {
		m.stopInstance(instance)
		delete(m.instances, id)
	}
	
	return nil
}

// startInstance 启动Discord实例
func (m *Manager) startInstance(instance *Instance) {
	m.logger.Infof("Starting Discord instance: %s", instance.ID)
	
	instance.ctx, instance.cancel = context.WithCancel(context.Background())
	
	// 连接到Discord Gateway
	if err := m.connectWebSocket(instance); err != nil {
		m.logger.Errorf("Failed to connect to Discord WebSocket: %v", err)
		return
	}
	
	// 启动消息处理循环
	go m.handleMessages(instance)
	
	// 启动心跳循环
	go m.heartbeatLoop(instance)
	
	instance.Connected = true
	instance.LastPing = time.Now()
	
	m.logger.Infof("Discord instance connected: %s", instance.ID)
}

// connectWebSocket 连接WebSocket
func (m *Manager) connectWebSocket(instance *Instance) error {
	gatewayURL := "wss://gateway.discord.gg/?v=10&encoding=json"
	
	headers := http.Header{}
	headers.Set("User-Agent", "midjourney-proxy-go/1.0")
	
	conn, _, err := websocket.DefaultDialer.Dial(gatewayURL, headers)
	if err != nil {
		return fmt.Errorf("failed to dial WebSocket: %w", err)
	}
	
	instance.conn = conn
	
	// 读取Hello消息
	var hello DiscordMessage
	if err := conn.ReadJSON(&hello); err != nil {
		return fmt.Errorf("failed to read hello message: %w", err)
	}
	
	if hello.Op != 10 { // Hello opcode
		return fmt.Errorf("expected hello message, got opcode %d", hello.Op)
	}
	
	// 解析心跳间隔
	var helloPayload HelloPayload
	if err := json.Unmarshal(hello.D, &helloPayload); err != nil {
		return fmt.Errorf("failed to parse hello payload: %w", err)
	}
	
	// 发送Identify消息
	identify := DiscordMessage{
		Op: 2, // Identify opcode
		D: json.RawMessage(fmt.Sprintf(`{
			"token": "%s",
			"properties": {
				"$os": "linux",
				"$browser": "midjourney-proxy-go",
				"$device": "midjourney-proxy-go"
			},
			"intents": 513
		}`, instance.Account.UserToken)),
	}
	
	if err := conn.WriteJSON(identify); err != nil {
		return fmt.Errorf("failed to send identify: %w", err)
	}
	
	return nil
}

// handleMessages 处理消息
func (m *Manager) handleMessages(instance *Instance) {
	defer func() {
		if instance.conn != nil {
			instance.conn.Close()
		}
	}()
	
	for {
		select {
		case <-instance.ctx.Done():
			return
		default:
			var msg DiscordMessage
			if err := instance.conn.ReadJSON(&msg); err != nil {
				m.logger.Errorf("Failed to read message from instance %s: %v", instance.ID, err)
				instance.Connected = false
				return
			}
			
			// 处理不同类型的消息
			switch msg.Op {
			case 0: // Dispatch
				m.handleDispatch(instance, msg)
			case 1: // Heartbeat
				m.sendHeartbeat(instance)
			case 7: // Reconnect
				m.logger.Infof("Discord requested reconnect for instance %s", instance.ID)
				// TODO: 实现重连逻辑
			case 9: // Invalid Session
				m.logger.Errorf("Invalid session for instance %s", instance.ID)
				instance.Connected = false
				return
			case 11: // Heartbeat ACK
				instance.LastPing = time.Now()
			}
		}
	}
}

// handleDispatch 处理Dispatch消息
func (m *Manager) handleDispatch(instance *Instance, msg DiscordMessage) {
	switch msg.T {
	case "READY":
		m.logger.Infof("Instance %s is ready", instance.ID)
		instance.Connected = true
	case "MESSAGE_CREATE", "MESSAGE_UPDATE":
		// 处理Midjourney机器人消息
		m.handleMidjourneyMessage(instance, msg)
	case "INTERACTION_CREATE":
		// 处理交互事件
		m.handleInteraction(instance, msg)
	}
}

// handleMidjourneyMessage 处理Midjourney消息
func (m *Manager) handleMidjourneyMessage(instance *Instance, msg DiscordMessage) {
	// TODO: 解析Midjourney机器人消息，更新任务状态
	m.logger.Debugf("Received Midjourney message in instance %s", instance.ID)
}

// handleInteraction 处理交互
func (m *Manager) handleInteraction(instance *Instance, msg DiscordMessage) {
	// TODO: 处理按钮点击等交互事件
	m.logger.Debugf("Received interaction in instance %s", instance.ID)
}

// heartbeatLoop 心跳循环
func (m *Manager) heartbeatLoop(instance *Instance) {
	ticker := time.NewTicker(41250 * time.Millisecond) // Discord心跳间隔
	defer ticker.Stop()
	
	for {
		select {
		case <-instance.ctx.Done():
			return
		case <-ticker.C:
			m.sendHeartbeat(instance)
		}
	}
}

// sendHeartbeat 发送心跳
func (m *Manager) sendHeartbeat(instance *Instance) {
	heartbeat := DiscordMessage{
		Op: 1, // Heartbeat opcode
		D:  json.RawMessage("null"),
	}
	
	if err := instance.conn.WriteJSON(heartbeat); err != nil {
		m.logger.Errorf("Failed to send heartbeat for instance %s: %v", instance.ID, err)
		instance.Connected = false
	}
}

// stopInstance 停止Discord实例
func (m *Manager) stopInstance(instance *Instance) {
	m.logger.Infof("Stopping Discord instance: %s", instance.ID)
	
	if instance.cancel != nil {
		instance.cancel()
	}
	
	if instance.conn != nil {
		instance.conn.Close()
	}
	
	instance.Connected = false
	
	m.logger.Infof("Discord instance stopped: %s", instance.ID)
}

// IsConnected 检查实例是否已连接
func (i *Instance) IsConnected() bool {
	return i.Connected
}

// GetAccount 获取账号信息
func (i *Instance) GetAccount() config.DiscordAccount {
	return i.Account
}

// SendMessage 发送消息
func (i *Instance) SendMessage(channelID, content string) error {
	if !i.Connected || i.conn == nil {
		return fmt.Errorf("instance not connected")
	}
	
	// TODO: 实现通过HTTP API发送消息
	// 这需要使用Discord的REST API，不是WebSocket
	return nil
}

// SubmitImagine 提交Imagine任务
func (i *Instance) SubmitImagine(prompt string) error {
	if !i.Connected {
		return fmt.Errorf("instance not connected")
	}
	
	// TODO: 实现提交Imagine任务逻辑
	// 这需要构造Discord交互负载并通过HTTP API发送
	return nil
}

// SetAccountSelectMode 设置账号选择模式
func (m *Manager) SetAccountSelectMode(mode AccountSelectMode) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.selector.SetSelectMode(mode)
	m.logger.Infof("Account select mode changed to: %s", mode)
}

// GetAccountSelectMode 获取账号选择模式
func (m *Manager) GetAccountSelectMode() AccountSelectMode {
	return m.selector.GetSelectMode()
}

// GetAccountSelectStats 获取账号选择器统计信息
func (m *Manager) GetAccountSelectStats() map[string]interface{} {
	return m.selector.GetStats()
}