<template>
  <div class="container">
    <div class="section">
      <h2 class="section-title">生成数据库</h2>
      
      <el-form :model="genForm" label-width="120px" @submit.prevent="handleGenerate">
        <el-form-item label="源文件路径" required>
          <el-input 
            v-model="genForm.srcFile" 
            placeholder="请输入源文件路径，如：data/ip.merge.txt"
          ></el-input>
          <div class="form-item-help">源文件需要是IP2Region支持的文本格式，每行格式为：startIP|endIP|国家|区域|省份|城市|ISP</div>
        </el-form-item>
        
        <el-form-item label="目标文件路径" required>
          <el-input 
            v-model="genForm.dstFile" 
            placeholder="请输入目标文件路径，如：data/ip2region.xdb"
          ></el-input>
          <div class="form-item-help">目标文件是生成的二进制数据库文件，扩展名通常为.xdb</div>
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="handleGenerate" :loading="loading">生成数据库</el-button>
        </el-form-item>
      </el-form>
      
      <el-divider />
      
      <!-- 任务状态显示（仅显示状态文本，不显示进度条） -->
      <div v-if="taskStatus && (taskStatus.status === 'pending' || taskStatus.status === 'processing')" class="status-section">
        <h3>生成状态</h3>
        <p class="status-message">{{ getStatusMessage() }}</p>
        <p class="time-info">
          开始时间: {{ formatTime(taskStatus.startTime) }}<br>
          已耗时: {{ getElapsedTime() }}
        </p>
      </div>
      
      <!-- 生成结果显示 -->
      <div v-if="taskStatus && taskStatus.status === 'completed'" class="result-section">
        <h3>生成结果</h3>
        
        <el-result
          icon="success"
          title="数据库生成成功"
          sub-title="数据库文件已成功生成"
        >
          <template #extra>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="源文件">{{ genForm.srcFile }}</el-descriptions-item>
              <el-descriptions-item label="目标文件">{{ genForm.dstFile }}</el-descriptions-item>
              <el-descriptions-item label="耗时">{{ getElapsedTime() }}</el-descriptions-item>
            </el-descriptions>
          </template>
        </el-result>
      </div>
      
      <!-- 错误显示 -->
      <div v-if="error || (taskStatus && taskStatus.status === 'failed')" class="error-section">
        <el-alert
          :title="error || taskStatus.errorMessage"
          type="error"
          description="请检查文件路径是否正确，源文件是否存在且格式正确"
          show-icon
          :closable="false"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { saveAndGenerateDbWithProgress, getGenerateTaskStatus } from '@/api'

const genForm = ref({
  srcFile: '',
  dstFile: ''
})

const loading = ref(false)
const error = ref('')
const taskId = ref('')
const taskStatus = ref(null)
const taskTimer = ref(null)

// 获取状态消息
const getStatusMessage = () => {
  if (!taskStatus.value) return ''
  
  switch (taskStatus.value.status) {
    case 'pending': return '准备生成数据库...'
    case 'processing': return '正在生成数据库，请稍候...'
    case 'completed': return '数据库生成完成'
    case 'failed': return '生成失败: ' + (taskStatus.value.errorMessage || '未知错误')
    default: return taskStatus.value.status
  }
}

// 格式化时间戳为可读时间
const formatTime = (timestamp) => {
  if (!timestamp) return '未知'
  const date = new Date(timestamp)
  return date.toLocaleString()
}

// 计算已耗时
const getElapsedTime = () => {
  if (!taskStatus.value || !taskStatus.value.durationSeconds) return '0秒'
  
  const duration = Math.floor(taskStatus.value.durationSeconds)
  const hours = Math.floor(duration / 3600)
  const minutes = Math.floor((duration % 3600) / 60)
  const seconds = duration % 60
  
  return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`
}

// 查询任务状态
const queryTaskStatus = async () => {
  if (!taskId.value) return
  
  try {
    const res = await getGenerateTaskStatus(taskId.value)
    taskStatus.value = res.data
    
    // 如果任务已完成，停止查询
    if (taskStatus.value.status === 'completed' || taskStatus.value.status === 'failed') {
      clearInterval(taskTimer.value)
      loading.value = false
    }
  } catch (err) {
    console.error('查询任务状态失败:', err)
    clearInterval(taskTimer.value)
    loading.value = false
  }
}

// 生成数据库
const handleGenerate = async () => {
  if (!genForm.value.srcFile || !genForm.value.dstFile) {
    error.value = '请输入源文件路径和目标文件路径'
    return
  }
  
  loading.value = true
  error.value = ''
  taskStatus.value = null
  
  try {
    // 清除之前的定时器
    if (taskTimer.value) {
      clearInterval(taskTimer.value)
    }
    
    // 调用异步生成接口（使用正确的API）
    const res = await saveAndGenerateDbWithProgress(genForm.value.srcFile, genForm.value.dstFile)
    taskId.value = res.data.taskId
    
    // 开始定时查询任务状态
    taskTimer.value = setInterval(queryTaskStatus, 1000)
  } catch (err) {
    error.value = err.message || '生成数据库失败'
    loading.value = false
  }
}

// 组件销毁时清除定时器
onUnmounted(() => {
  if (taskTimer.value) {
    clearInterval(taskTimer.value)
  }
})
</script>

<style scoped>
.form-item-help {
  font-size: 12px;
  color: #606266;
  margin-top: 5px;
}

.result-section, .error-section, .status-section {
  margin-top: 20px;
}

.status-section h3, .result-section h3 {
  margin-bottom: 15px;
  font-weight: 500;
}

.status-message {
  margin-top: 10px;
  font-size: 14px;
  color: #303133;
}

.time-info {
  margin-top: 10px;
  font-size: 13px;
  color: #606266;
}
</style> 