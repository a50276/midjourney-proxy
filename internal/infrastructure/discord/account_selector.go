package discord

import (
	"math/rand"
	"sort"
	"sync"
	"time"

	"midjourney-proxy-go/internal/domain/entity"
	"midjourney-proxy-go/pkg/logger"
)

// AccountSelectMode 账号选择模式
type AccountSelectMode string

const (
	AccountSelectBestWaitIdle AccountSelectMode = "BestWaitIdle" // 最佳等待空闲
	AccountSelectRandom       AccountSelectMode = "Random"       // 随机选择
	AccountSelectWeight       AccountSelectMode = "Weight"       // 权重选择
	AccountSelectPolling      AccountSelectMode = "Polling"      // 轮询选择
)

// AccountSelector 账号选择器
type AccountSelector struct {
	mode         AccountSelectMode
	pollingIndex int
	mutex        sync.RWMutex
	logger       logger.Logger
}

// NewAccountSelector 创建账号选择器
func NewAccountSelector(mode AccountSelectMode, logger logger.Logger) *AccountSelector {
	return &AccountSelector{
		mode:         mode,
		pollingIndex: 0,
		logger:       logger,
	}
}

// SelectAccount 选择账号
func (s *AccountSelector) SelectAccount(instances map[string]*Instance, filter *entity.AccountFilter) *Instance {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 获取可用实例
	availableInstances := s.getAvailableInstances(instances, filter)
	if len(availableInstances) == 0 {
		return nil
	}

	switch s.mode {
	case AccountSelectBestWaitIdle:
		return s.selectBestWaitIdle(availableInstances)
	case AccountSelectRandom:
		return s.selectRandom(availableInstances)
	case AccountSelectWeight:
		return s.selectByWeight(availableInstances)
	case AccountSelectPolling:
		return s.selectByPolling(availableInstances)
	default:
		return s.selectBestWaitIdle(availableInstances)
	}
}

// getAvailableInstances 获取可用实例
func (s *AccountSelector) getAvailableInstances(instances map[string]*Instance, filter *entity.AccountFilter) []*Instance {
	var available []*Instance

	for _, instance := range instances {
		if !instance.IsConnected() {
			continue
		}

		account := instance.GetAccount()

		// 检查账号是否可接受新任务
		// 这里需要从数据库加载完整的账号信息进行判断
		// 暂时简化处理
		if !account.Enabled {
			continue
		}

		// 应用过滤器
		if filter != nil {
			// 检查实例ID
			if filter.InstanceID != "" && filter.InstanceID != instance.ID {
				continue
			}

			// 检查生成模式
			if filter.Mode != "" {
				// 这里需要检查账号支持的模式
				// 暂时简化处理
			}

			// 检查机器人类型
			if filter.BotType != "" {
				// 这里需要检查账号支持的机器人类型
				// 暂时简化处理
			}

			// 检查Remix模式
			if filter.RemixEnabled {
				// 这里需要检查账号是否开启Remix
				// 暂时简化处理
			}
		}

		available = append(available, instance)
	}

	return available
}

// selectBestWaitIdle 选择最佳等待空闲实例
func (s *AccountSelector) selectBestWaitIdle(instances []*Instance) *Instance {
	if len(instances) == 0 {
		return nil
	}

	// 按队列大小和等待时间排序，选择最优的
	sort.Slice(instances, func(i, j int) bool {
		// 这里需要实际的队列信息，暂时使用简单逻辑
		return instances[i].LastPing.After(instances[j].LastPing)
	})

	return instances[0]
}

// selectRandom 随机选择实例
func (s *AccountSelector) selectRandom(instances []*Instance) *Instance {
	if len(instances) == 0 {
		return nil
	}

	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(instances))
	return instances[index]
}

// selectByWeight 按权重选择实例
func (s *AccountSelector) selectByWeight(instances []*Instance) *Instance {
	if len(instances) == 0 {
		return nil
	}

	// 计算总权重
	totalWeight := 0
	for _, instance := range instances {
		account := instance.GetAccount()
		weight := account.Weight
		if weight <= 0 {
			weight = 1 // 默认权重
		}
		totalWeight += weight
	}

	if totalWeight == 0 {
		return s.selectRandom(instances)
	}

	// 随机选择
	rand.Seed(time.Now().UnixNano())
	randomWeight := rand.Intn(totalWeight)

	currentWeight := 0
	for _, instance := range instances {
		account := instance.GetAccount()
		weight := account.Weight
		if weight <= 0 {
			weight = 1
		}
		currentWeight += weight
		if currentWeight > randomWeight {
			return instance
		}
	}

	return instances[0] // 兜底
}

// selectByPolling 轮询选择实例
func (s *AccountSelector) selectByPolling(instances []*Instance) *Instance {
	if len(instances) == 0 {
		return nil
	}

	// 按排序字段排序，确保轮询顺序一致
	sort.Slice(instances, func(i, j int) bool {
		accountI := instances[i].GetAccount()
		accountJ := instances[j].GetAccount()
		if accountI.Sort != accountJ.Sort {
			return accountI.Sort < accountJ.Sort
		}
		return accountI.ID < accountJ.ID
	})

	// 轮询选择
	selected := instances[s.pollingIndex%len(instances)]
	s.pollingIndex++

	return selected
}

// GetSelectMode 获取选择模式
func (s *AccountSelector) GetSelectMode() AccountSelectMode {
	return s.mode
}

// SetSelectMode 设置选择模式
func (s *AccountSelector) SetSelectMode(mode AccountSelectMode) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.mode = mode
	s.pollingIndex = 0 // 重置轮询索引
}

// GetStats 获取选择器统计信息
func (s *AccountSelector) GetStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return map[string]interface{}{
		"mode":           string(s.mode),
		"polling_index":  s.pollingIndex,
	}
}