<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :xs="24" :sm="12" :lg="6">
        <div class="stat-card">
          <div class="stat-icon">
            <el-icon color="#409EFF"><Document /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-number">{{ stats.totalTasks }}</div>
            <div class="stat-label">总任务数</div>
          </div>
        </div>
      </el-col>
      
      <el-col :xs="24" :sm="12" :lg="6">
        <div class="stat-card">
          <div class="stat-icon">
            <el-icon color="#67C23A"><Avatar /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-number">{{ stats.activeAccounts }}</div>
            <div class="stat-label">活跃账号</div>
          </div>
        </div>
      </el-col>
      
      <el-col :xs="24" :sm="12" :lg="6">
        <div class="stat-card">
          <div class="stat-icon">
            <el-icon color="#E6A23C"><UserFilled /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-number">{{ stats.totalUsers }}</div>
            <div class="stat-label">注册用户</div>
          </div>
        </div>
      </el-col>
      
      <el-col :xs="24" :sm="12" :lg="6">
        <div class="stat-card">
          <div class="stat-icon">
            <el-icon color="#F56C6C"><Warning /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-number">{{ stats.failedTasks }}</div>
            <div class="stat-label">失败任务</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="20" class="charts-row">
      <el-col :xs="24" :lg="16">
        <div class="card-container">
          <div class="card-header">
            <h3>任务趋势</h3>
            <el-radio-group v-model="chartTimeRange" size="small">
              <el-radio-button label="7d">最近7天</el-radio-button>
              <el-radio-button label="30d">最近30天</el-radio-button>
              <el-radio-button label="90d">最近90天</el-radio-button>
            </el-radio-group>
          </div>
          <v-chart 
            class="chart" 
            :option="taskTrendOption" 
            :loading="chartLoading"
          />
        </div>
      </el-col>
      
      <el-col :xs="24" :lg="8">
        <div class="card-container">
          <div class="card-header">
            <h3>任务状态分布</h3>
          </div>
          <v-chart 
            class="chart" 
            :option="taskStatusOption"
            :loading="chartLoading"
          />
        </div>
      </el-col>
    </el-row>

    <!-- 最近任务和系统状态 -->
    <el-row :gutter="20" class="content-row">
      <el-col :xs="24" :lg="14">
        <div class="card-container">
          <div class="card-header">
            <h3>最近任务</h3>
            <el-button type="primary" size="small" @click="refreshTasks">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
          
          <el-table 
            :data="recentTasks" 
            v-loading="tasksLoading"
            empty-text="暂无数据"
          >
            <el-table-column prop="id" label="任务ID" width="100" show-overflow-tooltip />
            <el-table-column prop="action" label="类型" width="80">
              <template #default="{ row }">
                <el-tag :type="getActionTagType(row.action)" size="small">
                  {{ row.action }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="getStatusTagType(row.status)" size="small">
                  {{ getStatusText(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="prompt" label="提示词" show-overflow-tooltip />
            <el-table-column prop="created_at" label="创建时间" width="120">
              <template #default="{ row }">
                {{ formatTime(row.created_at) }}
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-col>
      
      <el-col :xs="24" :lg="10">
        <div class="card-container">
          <div class="card-header">
            <h3>系统状态</h3>
            <el-badge :value="systemAlerts.length" :hidden="systemAlerts.length === 0">
              <el-button type="warning" size="small">
                <el-icon><Bell /></el-icon>
                警告
              </el-button>
            </el-badge>
          </div>
          
          <!-- Discord 实例状态 -->
          <div class="system-section">
            <h4>Discord 实例</h4>
            <div class="instance-list">
              <div 
                v-for="instance in discordInstances" 
                :key="instance.id"
                class="instance-item"
              >
                <div class="instance-info">
                  <span class="instance-name">{{ instance.id }}</span>
                  <el-tag 
                    :type="instance.connected ? 'success' : 'danger'" 
                    size="small"
                  >
                    {{ instance.connected ? '在线' : '离线' }}
                  </el-tag>
                </div>
                <div class="instance-stats">
                  <span>延迟: {{ instance.ping }}ms</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 系统资源 -->
          <div class="system-section">
            <h4>系统资源</h4>
            <div class="resource-item">
              <span>CPU 使用率</span>
              <el-progress :percentage="systemResources.cpu" />
            </div>
            <div class="resource-item">
              <span>内存使用率</span>
              <el-progress :percentage="systemResources.memory" />
            </div>
            <div class="resource-item">
              <span>磁盘使用率</span>
              <el-progress :percentage="systemResources.disk" />
            </div>
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Document,
  Avatar,
  UserFilled,
  Warning,
  Refresh,
  Bell
} from '@element-plus/icons-vue'
import dayjs from 'dayjs'

// 响应式数据
const chartLoading = ref(true)
const tasksLoading = ref(true)
const chartTimeRange = ref('7d')

// 统计数据
const stats = ref({
  totalTasks: 1248,
  activeAccounts: 5,
  totalUsers: 89,
  failedTasks: 23
})

// 最近任务
const recentTasks = ref([
  {
    id: 'mj_123456',
    action: 'IMAGINE',
    status: 'SUCCESS',
    prompt: 'a beautiful landscape with mountains',
    created_at: '2024-01-15T10:30:00Z'
  },
  {
    id: 'mj_123457',
    action: 'UPSCALE',
    status: 'IN_PROGRESS',
    prompt: 'upscale previous image',
    created_at: '2024-01-15T10:25:00Z'
  },
  {
    id: 'mj_123458',
    action: 'VARIATION',
    status: 'FAILURE',
    prompt: 'create variations of landscape',
    created_at: '2024-01-15T10:20:00Z'
  }
])

// Discord实例状态
const discordInstances = ref([
  { id: 'instance-1', connected: true, ping: 45 },
  { id: 'instance-2', connected: true, ping: 52 },
  { id: 'instance-3', connected: false, ping: 0 }
])

// 系统资源
const systemResources = ref({
  cpu: 25,
  memory: 68,
  disk: 42
})

// 系统警告
const systemAlerts = ref([
  { type: 'warning', message: 'Discord实例3离线' }
])

// 任务趋势图表配置
const taskTrendOption = computed(() => ({
  title: {
    text: ''
  },
  tooltip: {
    trigger: 'axis'
  },
  legend: {
    data: ['成功', '失败', '进行中']
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true
  },
  xAxis: {
    type: 'category',
    data: ['01-09', '01-10', '01-11', '01-12', '01-13', '01-14', '01-15']
  },
  yAxis: {
    type: 'value'
  },
  series: [
    {
      name: '成功',
      type: 'line',
      data: [120, 132, 101, 134, 90, 230, 210],
      itemStyle: { color: '#67C23A' }
    },
    {
      name: '失败',
      type: 'line',
      data: [10, 15, 8, 12, 5, 18, 20],
      itemStyle: { color: '#F56C6C' }
    },
    {
      name: '进行中',
      type: 'line',
      data: [5, 8, 12, 15, 10, 12, 8],
      itemStyle: { color: '#E6A23C' }
    }
  ]
}))

// 任务状态饼图配置
const taskStatusOption = computed(() => ({
  tooltip: {
    trigger: 'item'
  },
  series: [
    {
      type: 'pie',
      radius: '70%',
      data: [
        { value: 1035, name: '成功', itemStyle: { color: '#67C23A' } },
        { value: 180, name: '失败', itemStyle: { color: '#F56C6C' } },
        { value: 33, name: '进行中', itemStyle: { color: '#E6A23C' } }
      ],
      emphasis: {
        itemStyle: {
          shadowBlur: 10,
          shadowOffsetX: 0,
          shadowColor: 'rgba(0, 0, 0, 0.5)'
        }
      }
    }
  ]
}))

// 工具方法
const getActionTagType = (action: string) => {
  const types = {
    'IMAGINE': 'primary',
    'UPSCALE': 'success',
    'VARIATION': 'warning',
    'DESCRIBE': 'info'
  }
  return types[action] || 'default'
}

const getStatusTagType = (status: string) => {
  const types = {
    'SUCCESS': 'success',
    'FAILURE': 'danger',
    'IN_PROGRESS': 'warning',
    'SUBMITTED': 'info'
  }
  return types[status] || 'default'
}

const getStatusText = (status: string) => {
  const texts = {
    'SUCCESS': '成功',
    'FAILURE': '失败',
    'IN_PROGRESS': '进行中',
    'SUBMITTED': '已提交'
  }
  return texts[status] || status
}

const formatTime = (time: string) => {
  return dayjs(time).format('MM-DD HH:mm')
}

const refreshTasks = async () => {
  tasksLoading.value = true
  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))
    ElMessage.success('刷新成功')
  } catch (error) {
    ElMessage.error('刷新失败')
  } finally {
    tasksLoading.value = false
  }
}

// 生命周期
onMounted(async () => {
  // 模拟数据加载
  setTimeout(() => {
    chartLoading.value = false
    tasksLoading.value = false
  }, 1500)
})
</script>

<style scoped>
.dashboard {
  min-height: 100%;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  background: var(--el-bg-color);
  border-radius: 8px;
  padding: 20px;
  box-shadow: var(--el-box-shadow-light);
  display: flex;
  align-items: center;
  gap: 15px;
  height: 100px;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: var(--el-fill-color-light);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

.stat-info {
  flex: 1;
}

.stat-number {
  font-size: 28px;
  font-weight: bold;
  color: var(--el-text-color-primary);
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: var(--el-text-color-secondary);
  margin-top: 5px;
}

.charts-row {
  margin-bottom: 20px;
}

.content-row {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.card-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.card-header h4 {
  margin: 0 0 15px 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.chart {
  height: 300px;
  width: 100%;
}

.system-section {
  margin-bottom: 25px;
}

.instance-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.instance-item {
  padding: 12px;
  background: var(--el-fill-color-lighter);
  border-radius: 6px;
}

.instance-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 5px;
}

.instance-name {
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.instance-stats {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.resource-item {
  margin-bottom: 15px;
}

.resource-item span {
  display: block;
  margin-bottom: 5px;
  font-size: 14px;
  color: var(--el-text-color-primary);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .stat-card {
    height: auto;
    padding: 15px;
  }
  
  .stat-number {
    font-size: 24px;
  }
  
  .chart {
    height: 250px;
  }
}
</style>