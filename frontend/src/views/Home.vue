<template>
  <div class="container">
    <div class="section">
      <h1 class="text-center">欢迎使用 IP2Region Web</h1>
      <p class="text-center">高效的IP地址查询和管理工具</p>

      <el-divider />

      <div class="load-xdb-section">
        <el-card shadow="hover">
          <div class="xdb-loader">
            <h3>XDB文件加载</h3>
            <p>选择合适的加载模式来优化IP查询性能。</p>
            <p class="xdb-tip">
              <el-tag type="info" effect="plain">提示</el-tag>
              请确保XDB文件有效且完整，文件大小通常为几MB。如果遇到加载错误，可能是文件不完整或格式错误。
            </p>

            <el-form>
              <el-form-item label="数据库文件">
                <el-input v-model="xdbPath" placeholder="请输入XDB文件路径，例如：./ip2region.xdb" clearable />
              </el-form-item>

              <el-form-item label="加载模式">
                <el-radio-group v-model="loadMode">
                  <el-radio value="vector">
                    <div class="mode-option">
                      <div class="mode-name">向量模式</div>
                      <div class="mode-desc">缓存向量索引到内存，平衡内存占用和查询性能（推荐）</div>
                    </div>
                  </el-radio>
                  <el-radio value="memory">
                    <div class="mode-option">
                      <div class="mode-name">内存模式</div>
                      <div class="mode-desc">将整个XDB文件加载到内存，查询性能最佳</div>
                    </div>
                  </el-radio>
                </el-radio-group>
              </el-form-item>

              <el-form-item>
                <el-button v-if="!loadSuccess" @click="loadXdbToMemory" :loading="loading" type="primary" size="large">
                  加载数据库
                </el-button>
                <el-button v-else @click="unloadXdb" :loading="unloading" type="danger" size="large">
                  卸载数据库
                </el-button>
              </el-form-item>

              <el-form-item v-if="loadSuccess">
                <el-button type="success" @click="showExportDialog" :loading="exporting">
                  <el-icon><Download /></el-icon> 导出XDB
                </el-button>
                <span class="export-tip">导出XDB文件内容到文本文件</span>
              </el-form-item>
            </el-form>

            <!-- 导出对话框 -->
            <el-dialog
              v-model="exportDialogVisible"
              title="导出XDB文件"
              width="500px"
              :close-on-click-modal="false"
              :close-on-press-escape="!exportingTask"
              :show-close="!exportingTask"
            >
              <el-form>
                <el-form-item label="XDB文件路径">
                  <el-input v-model="exportForm.xdbPath" disabled></el-input>
                </el-form-item>
                <el-form-item label="导出文件路径">
                  <el-input
                    v-model="exportForm.exportPath"
                    placeholder="请输入导出文件路径，例如：ip2region.xdb.export.txt"
                    :disabled="exportingTask">
                  </el-input>
                </el-form-item>

                <!-- 导出进度显示 -->
                <el-form-item v-if="exportingTask">
                  <div class="export-progress">
                    <div class="progress-info">
                      <p><strong>状态:</strong> {{ exportTaskStatusText }}</p>
                      <p><strong>已发现IP段:</strong> {{ exportTask.segmentCount || 0 }}</p>
                      <p><strong>当前处理IP:</strong> {{ formatIPAddress(exportTask.recordCount || 0) }}</p>
                      <el-progress :percentage="exportTask.progress || 0" :stroke-width="10" striped />
                      <p><strong>已运行时间:</strong> {{ getFormattedRunningTime() }}</p>
                    </div>
                  </div>
                </el-form-item>
              </el-form>
              <template #footer>
                <div class="dialog-footer">
                  <el-button @click="cancelExport">
                    {{ exportingTask ? '关闭' : '取消' }}
                  </el-button>
                  <template v-if="!exportingTask">
                    <el-button type="primary" @click="handleExport" :loading="exporting">
                      导出
                    </el-button>
                  </template>
                  <template v-else-if="exportTask.status !== 'completed' && exportTask.status !== 'failed'">
                    <el-button type="danger" @click="handleCancelTask">
                      取消导出
                    </el-button>
                  </template>
                </div>
              </template>
            </el-dialog>

            <div v-if="loadResult" class="load-result">
              <el-alert
                v-if="loadSuccess"
                type="success"
                :title="`XDB文件加载成功: ${loadResult.dbPath}`"
                show-icon
              >
                <template #default>
                  <div class="result-details">
                    <p><strong>内存模式:</strong> {{ loadResult.inMemoryMode ? '是' : '否' }}</p>
                    <p><strong>缓冲区大小:</strong> {{ loadResult.bufferSizeKB }} KB</p>
                    <p><strong>向量索引已加载:</strong> {{ loadResult.vectorLoaded ? '是' : '否' }}</p>
                    <p><strong>向量索引大小:</strong> {{ loadResult.vectorSizeKB }} KB</p>
                    <p><strong>加载耗时:</strong> {{ loadResult.loadTimeTaken }}</p>
                  </div>
                </template>
              </el-alert>
              <el-alert
                v-else
                type="error"
                :title="errorMsg"
                show-icon
              />
            </div>
          </div>
        </el-card>
      </div>

      <el-divider />

      <div class="feature-section">
        <h2>核心功能</h2>
        <el-row :gutter="20">
          <el-col :span="8">
            <el-card shadow="hover" class="feature-card">
              <el-icon size="30" color="#409EFF"><Search /></el-icon>
              <h3>IP地址查询</h3>
              <p>快速查询IP地址的地理位置信息，支持国家、地区、省份、城市和ISP信息。</p>
              <el-button type="primary" @click="router.push('/search')">开始查询</el-button>
            </el-card>
          </el-col>
          <el-col :span="8">
            <el-card shadow="hover" class="feature-card">
              <el-icon size="30" color="#409EFF"><DataLine /></el-icon>
              <h3>生成数据库</h3>
              <p>将文本格式的IP段数据转换为高效的二进制索引文件，实现快速查询。</p>
              <el-button type="primary" @click="router.push('/generate')">数据库生成</el-button>
            </el-card>
          </el-col>
          <el-col :span="8">
            <el-card shadow="hover" class="feature-card">
              <el-icon size="30" color="#409EFF"><Edit /></el-icon>
              <h3>数据编辑</h3>
              <p>提供IP段数据的编辑功能，支持单条编辑和批量导入，方便维护IP数据。</p>
              <el-button type="primary" @click="router.push('/edit')">编辑数据</el-button>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <el-divider />

      <div class="about-section">
        <h2>关于 IP2Region</h2>
        <p>IP2Region是一个离线IP地址定位库和IP定位服务，支持亿级别的数据量，响应时间在毫秒级别。提供了众多主流编程语言的SDK实现，包括：C、C++、Java、PHP、Python、Node、Go等。可广泛应用于大数据分析、物联网、网络安全等领域。</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Search, DataLine, Edit, Download } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import {
  loadXdbToMemory as loadXdbToMemoryApi,
  unloadXdb as unloadXdbApi,
  getXdbStatus as getXdbStatusApi,
  exportXdb as exportXdbApi,
  getExportTaskStatus as getExportTaskStatusApi,
  cancelExportTask as cancelExportTaskApi
} from '@/api'

const router = useRouter()

// XDB文件加载相关数据
const xdbPath = ref('./ip2region.xdb')
const loadMode = ref('vector')
const loading = ref(false)
const unloading = ref(false)
const loadResult = ref(null)
const loadSuccess = ref(false)
const errorMsg = ref('')

// 导出相关
const exportDialogVisible = ref(false)
const exporting = ref(false)
const exportForm = ref({
  xdbPath: '',
  exportPath: 'ip2region.xdb.export.txt'
})

// 导出任务相关
const exportingTask = ref(false)
const exportTask = ref({})
const exportTaskId = ref('')
const exportStatusTimer = ref(null)

// 计算导出任务状态文本
const exportTaskStatusText = computed(() => {
  if (!exportTask.value.status) return '准备中';

  let statusText = '';
  switch (exportTask.value.status) {
    case 'pending':
      statusText = '等待处理';
      break;
    case 'processing':
      statusText = exportTask.value.detailedStatus || '正在导出'; // 使用 detailedStatus
      break;
    case 'completed':
      statusText = '导出完成';
      break;
    case 'failed':
      statusText = '导出失败: ' + (exportTask.value.errorMessage || '未知错误');
      break;
    default:
      statusText = exportTask.value.status;
  }
  // 如果是正在处理，并且有 detailedStatus，就用 detailedStatus
  if (exportTask.value.status === 'processing' && exportTask.value.detailedStatus) {
    return exportTask.value.detailedStatus;
  }
  return statusText;
})

// 组件挂载时检查XDB加载状态
onMounted(async () => {
  try {
    const response = await getXdbStatusApi()
    if (response.data && response.data.loaded) {
      loadResult.value = response.data.status
      loadSuccess.value = true
      xdbPath.value = response.data.status.dbPath
    }
  } catch (error) {
    console.error('获取XDB状态失败', error)
  }
})

// 加载XDB文件到内存
const loadXdbToMemory = async () => {
  if (!xdbPath.value) {
    ElMessage.warning('请输入XDB文件路径')
    return
  }

  loading.value = true
  loadResult.value = null
  loadSuccess.value = false
  errorMsg.value = ''

  try {
    const response = await loadXdbToMemoryApi(xdbPath.value, loadMode.value)
    loadResult.value = response.data
    loadSuccess.value = true
    ElMessage.success(`数据库加载成功 - ${getModeText(loadMode.value)}`)
  } catch (error) {
    errorMsg.value = error.message || '加载失败，请检查文件路径是否正确'
    // 提供更详细的错误提示
    if (error.message && error.message.includes('文件大小不足')) {
      errorMsg.value = `${error.message}。请下载完整的XDB文件或使用"生成数据库"功能创建新的XDB文件。`
    }
    ElMessage.error(errorMsg.value)
  } finally {
    loading.value = false
  }
}

// 获取模式文本
const getModeText = (mode) => {
  switch (mode) {
    case 'file':
      return '文件模式'
    case 'vector':
      return '向量模式'
    case 'memory':
      return '内存模式'
    default:
      return '未知模式'
  }
}

// 卸载内存中的XDB文件
const unloadXdb = async () => {
  if (!loadSuccess.value) {
    ElMessage.warning('当前没有加载的XDB文件')
    return
  }

  unloading.value = true

  try {
    await unloadXdbApi()
    loadResult.value = null
    loadSuccess.value = false
    ElMessage.success('XDB文件已从内存中卸载，现在可以安全地删除或修改文件')

    // 延迟检查加载状态，确保卸载成功
    setTimeout(async () => {
      try {
        const status = await getXdbStatusApi()
        if (status.data && status.data.loaded) {
          ElMessage.warning('文件卸载可能未完全成功，请稍后再试')
        }
      } catch (error) {
        console.error('检查卸载状态失败', error)
      }
    }, 1000)
  } catch (error) {
    ElMessage.error(error.message || '卸载失败')
  } finally {
    unloading.value = false
  }
}

// 显示导出对话框
const showExportDialog = () => {
  exportForm.value.xdbPath = xdbPath.value
  exportDialogVisible.value = true
  exportingTask.value = false
  exportTask.value = {}
  exportTaskId.value = ''

  // 清理之前的定时器
  if (exportStatusTimer.value) {
    clearInterval(exportStatusTimer.value)
    exportStatusTimer.value = null
  }
}

// 取消导出对话框
const cancelExport = () => {
  // 如果正在导出任务中并且已完成，则关闭对话框
  if (exportingTask.value) {
    // 清理定时器
    if (exportStatusTimer.value) {
      clearInterval(exportStatusTimer.value)
      exportStatusTimer.value = null
    }
    exportDialogVisible.value = false
    return
  }

  // 否则直接关闭对话框
  exportDialogVisible.value = false
}

// 开始轮询获取导出任务状态 - 增强版
const startPollingTaskStatus = (taskId) => {
  // 清理之前的定时器
  if (exportStatusTimer.value) {
    clearInterval(exportStatusTimer.value)
    exportStatusTimer.value = null
  }

  console.log('开始轮询任务状态，任务ID:', taskId)

  // 立即查询一次
  getExportTaskStatus(taskId)

  // 设置定时轮询
  exportStatusTimer.value = setInterval(() => {
    getExportTaskStatus(taskId)
  }, 2000) // 增加间隔至2秒，减少请求频率
}

// 获取导出任务状态 - 增强版
const getExportTaskStatus = async (taskId) => {
  console.log('正在查询任务状态:', taskId)

  // 添加前端超时检测
  const now = new Date().getTime()
  const taskStartTime = exportTask.value.startTime || now
  const elapsedSeconds = getRunningTime()

  // 如果任务运行超过900秒(15分钟)，可能存在问题
  if (elapsedSeconds > 900) {
    if (exportStatusTimer.value) {
      clearInterval(exportStatusTimer.value)
      exportStatusTimer.value = null
    }
    ElMessage.error('导出任务执行时间过长，可能存在问题。建议取消并重试。')
    return
  }

  try {
    const response = await getExportTaskStatusApi(taskId)
    console.log('获取任务状态成功:', response.data)

    // 确保任务数据有效
    if (!response.data) {
      console.error('任务数据为空')
      return
    }

    exportTask.value = response.data
    console.log('Updated exportTask in getExportTaskStatus:', JSON.parse(JSON.stringify(exportTask.value))); // 添加日志

    // 如果任务已完成或失败，停止轮询
    if (exportTask.value.status === 'completed' || exportTask.value.status === 'failed') {
      if (exportStatusTimer.value) {
        clearInterval(exportStatusTimer.value)
        exportStatusTimer.value = null
      }

      // 显示完成或失败消息
      if (exportTask.value.status === 'completed') {
        ElMessage.success(`导出成功：${exportTask.value.segmentCount} 个IP段`)
        exporting.value = false
      } else if (exportTask.value.status === 'failed') {
        ElMessage.error(`导出失败：${exportTask.value.errorMessage || '未知错误'}`)
        exporting.value = false
      }
    }
  } catch (error) {
    console.error('获取导出任务状态失败', error)

    // 显示更详细的错误信息
    if (error.response) {
      console.error('错误响应:', error.response)
    }

    // 失败重试次数限制
    if (!exportTask.value.retryCount) {
      exportTask.value.retryCount = 1
    } else {
      exportTask.value.retryCount++
    }

    // 超过5次失败后停止轮询
    if (exportTask.value.retryCount > 5) {
      if (exportStatusTimer.value) {
        clearInterval(exportStatusTimer.value)
        exportStatusTimer.value = null
      }
      ElMessage.error('获取任务状态失败，请检查网络连接')
      exporting.value = false
    }
  }
}

// 取消导出任务
const handleCancelTask = async () => {
  if (!exportTaskId.value) return

  try {
    await cancelExportTaskApi(exportTaskId.value)
    ElMessage.warning('导出任务已取消')

    // 立即更新一次状态
    await getExportTaskStatus(exportTaskId.value)
  } catch (error) {
    ElMessage.error(error.message || '取消任务失败')
  }
}

// 导出XDB文件 - 优化版
const handleExport = async () => {
  if (!exportForm.value.xdbPath || !exportForm.value.exportPath) {
    ElMessage.warning('请填写完整的导出信息')
    return
  }

  // 检查是否已有导出任务在进行中
  if (exportingTask.value) {
    ElMessage.warning('已有导出任务正在进行，请等待完成或取消当前任务')
    return
  }

  exporting.value = true

  // 重置任务状态
  exportingTask.value = false
  exportTask.value = {}
  exportTaskId.value = ''

  try {
    console.log('开始创建导出任务:', exportForm.value)
    const response = await exportXdbApi(exportForm.value.xdbPath, exportForm.value.exportPath)
    console.log('导出任务创建成功:', response.data)

    // 检查响应中是否有taskId
    if (!response.data || !response.data.taskId) {
      throw new Error('服务器未返回任务ID')
    }

    // 保存任务ID
    exportTaskId.value = response.data.taskId

    // 显示进度
    exportingTask.value = true
    exportTask.value = {
      status: 'pending',
      startTime: new Date().getTime(),
      progress: 0, // 初始化 progress
      recordCount: 0, // 初始化 recordCount
      segmentCount: 0 // 初始化 segmentCount
    }

    // 开始轮询任务状态
    startPollingTaskStatus(exportTaskId.value)

    ElMessage.info(`导出任务已创建(ID:${exportTaskId.value})，正在处理中...`)
  } catch (error) {
    console.error('创建导出任务失败', error)
    ElMessage.error(error.message || '创建导出任务失败')
    exporting.value = false
  }
}

// 计算任务已运行时间（秒）
const getRunningTime = () => {
  if (!exportTask.value.startTime) return 0

  // 优先使用服务器返回的运行时间
  if (exportTask.value.durationSeconds !== undefined) {
    return exportTask.value.durationSeconds
  }

  // 处理后端返回的时间字符串或前端生成的时间戳
  let startTime = 0
  if (typeof exportTask.value.startTime === 'string') {
    // 如果是后端返回的ISO格式时间字符串
    startTime = new Date(exportTask.value.startTime).getTime()
  } else if (typeof exportTask.value.startTime === 'number') {
    // 如果是毫秒级时间戳
    startTime = exportTask.value.startTime
  } else {
    return 0
  }

  const now = new Date().getTime()
  // 确保时间差有效，防止返回NaN
  if (isNaN(startTime) || startTime <= 0 || now < startTime) {
    return 0
  }

  const seconds = Math.floor((now - startTime) / 1000)
  return seconds >= 0 ? seconds : 0
}

// 格式化显示运行时间
const getFormattedRunningTime = () => {
  // 获取秒数
  const seconds = getRunningTime()
  return `${seconds}秒`
}

// 格式化数字显示（添加千分位分隔符）
const formatNumber = (num) => {
  if (typeof num !== 'number') {
    return '0'
  }
  return num.toLocaleString()
}

// 将IP地址数值转换为可读的IP地址格式
const formatIPAddress = (ipNum) => {
  if (typeof ipNum !== 'number' || ipNum === 0) {
    return '0.0.0.0'
  }

  // 将32位整数转换为IP地址
  const a = (ipNum >>> 24) & 0xFF
  const b = (ipNum >>> 16) & 0xFF
  const c = (ipNum >>> 8) & 0xFF
  const d = ipNum & 0xFF

  return `${a}.${b}.${c}.${d}`
}
</script>

<style scoped>
.feature-section, .about-section, .load-xdb-section {
  margin: 30px 0;
}

.feature-card {
  height: 240px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  transition: all 0.3s;
}

.feature-card:hover {
  transform: translateY(-5px);
}

.feature-card h3 {
  margin: 15px 0 10px;
  color: #303133;
}

.feature-card p {
  margin-bottom: 15px;
  color: #606266;
  flex-grow: 1;
}

.xdb-loader {
  text-align: center;
}

.xdb-loader h3 {
  margin-bottom: 10px;
  color: #303133;
}

.xdb-loader p {
  margin-bottom: 20px;
  color: #606266;
}

.xdb-tip {
  margin-bottom: 20px;
  text-align: left;
  background-color: #f8f8f8;
  padding: 10px;
  border-radius: 4px;
  font-size: 14px;
  color: #606266;
}

.xdb-tip .el-tag {
  margin-right: 8px;
  vertical-align: middle;
}

.load-result {
  margin-top: 20px;
  text-align: left;
}

.result-details {
  margin: 10px 0;
}

.result-details p {
  margin: 5px 0;
  font-size: 14px;
}

.export-tip {
  margin-left: 10px;
  color: #909399;
  font-size: 12px;
}

.export-progress {
  margin-top: 10px;
}

.mode-option {
  padding: 5px 0;
  margin-left: 10px;
}

.mode-name {
  font-weight: bold;
  color: #303133;
}

.mode-desc {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
  line-height: 1.4;
}

:deep(.el-radio) {
  display: block;
  margin-bottom: 15px;
  height: auto;
}

:deep(.el-radio__label) {
  white-space: normal;
}

.progress-info {
  margin-bottom: 10px;
}

.progress-info p {
  margin: 5px 0;
  font-size: 14px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
}
</style>