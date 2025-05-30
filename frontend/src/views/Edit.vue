<template>
  <div class="container">
    <div class="section">
      <h2 class="section-title">编辑数据</h2>
      
      <!-- 添加当前加载文件的提示 -->
      <el-alert
        v-if="segmentsLoaded"
        type="info"
        :title="`当前加载的源文件: ${editForm.srcFile}`"
        :closable="false"
        show-icon
        class="mb-20"
      />
      
      <el-form :model="editForm" label-width="100px">
        <el-form-item label="源文件路径" required>
          <div class="flex-row">
            <el-input 
              v-model="editForm.srcFile" 
              placeholder="请输入源文件路径，如：data/ip.merge.txt"
            ></el-input>
            <el-button 
              type="primary" 
              class="ml-10" 
              @click="handleLoadSegments" 
              :loading="loading.list"
            >
              加载数据
            </el-button>
            <el-button 
              v-if="segmentsLoaded"
              type="danger" 
              class="ml-10" 
              @click="handleUnloadFile"
            >
              卸载文件
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      
      <el-divider />
      
      <div v-if="segmentsLoaded" class="edit-section">
        <div class="tool-bar flex-row space-between mb-20">
          <div>
            <el-button type="primary" @click="dialogVisible.add = true">添加IP段</el-button>
            <el-button type="primary" @click="dialogVisible.import = true">从文件导入</el-button>
            <el-button type="primary" @click="dialogVisible.edit = true">编辑IP段(PUT)</el-button>
          </div>
          <div>
            <el-button 
              type="success" 
              @click="dialogVisible.generate = true" 
              :disabled="!segmentsLoaded"
            >
              保存并生成XDB
            </el-button>
          </div>
        </div>
        
        <div class="segment-list">
          <el-table 
            ref="tableRef"
            :data="segments" 
            style="width: 100%" 
            border 
            max-height="500"
            @row-click="handleRowClick"
            highlight-current-row
          >
            <el-table-column label="起始IP" width="140">
              <template #default="scope">
                {{ intToIp(scope.row.rawData.StartIP) }}
              </template>
            </el-table-column>
            <el-table-column label="结束IP" width="140">
              <template #default="scope">
                {{ intToIp(scope.row.rawData.EndIP) }}
              </template>
            </el-table-column>
            <el-table-column label="区域信息" min-width="200">
              <template #default="scope">
                {{ scope.row.rawData.Region }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="180">
              <template #default="scope">
                <el-button 
                  type="primary" 
                  size="small" 
                  @click="handleEditRow(scope.row)"
                  style="margin-right: 5px"
                >
                  编辑
                </el-button>
                <el-button 
                  type="danger" 
                  size="small" 
                  @click="handleDelete(scope.$index)"
                >
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          
          <div class="pagination mt-20 text-right">
            <el-pagination
              v-model:current-page="currentPage"
              v-model:page-size="pageSize"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              :total="totalSegments"
              @size-change="handleSizeChange"
              @current-change="handleCurrentChange"
            />
          </div>
        </div>
      </div>
      
      <div v-if="error" class="error-section mt-20">
        <el-alert
          :title="error"
          type="error"
          show-icon
          :closable="false"
        />
      </div>
      
      <!-- 添加IP段对话框 -->
      <el-dialog
        v-model="dialogVisible.add"
        title="添加IP段"
        width="500px"
      >
        <el-form :model="addForm" label-width="100px">
          <el-form-item label="IP段" required>
            <el-input 
              v-model="addForm.segment" 
              placeholder="格式：startIP|endIP|国家|区域|省份|城市|ISP"
            ></el-input>
            <div class="form-item-help">例如：36.132.128.0|36.132.147.255|中国|0|黑龙江省|哈尔滨市|移动</div>
          </el-form-item>
        </el-form>
        <template #footer>
          <span class="dialog-footer">
            <el-button @click="dialogVisible.add = false">取 消</el-button>
            <el-button type="primary" @click="handleAddSegment" :loading="loading.add">
              确 定
            </el-button>
          </span>
        </template>
      </el-dialog>
      
      <!-- 从文件导入对话框 -->
      <el-dialog
        v-model="dialogVisible.import"
        title="从文件导入"
        width="500px"
      >
        <el-form :model="importForm" label-width="100px">
          <el-form-item label="文件路径" required>
            <el-input 
              v-model="importForm.file" 
              placeholder="请输入文件路径"
            ></el-input>
          </el-form-item>
        </el-form>
        <template #footer>
          <span class="dialog-footer">
            <el-button @click="dialogVisible.import = false">取 消</el-button>
            <el-button type="primary" @click="handleImportFile" :loading="loading.import">
              确 定
            </el-button>
          </span>
        </template>
      </el-dialog>
      
      <!-- 通过PUT编辑IP段对话框 -->
      <el-dialog
        v-model="dialogVisible.edit"
        title="编辑IP段 (PUT方式)"
        width="500px"
      >
        <el-form :model="editSegmentForm" label-width="100px">
          <el-form-item label="IP段" required>
            <el-input 
              v-model="editSegmentForm.segment" 
              placeholder="格式：startIP|endIP|国家|区域|省份|城市|ISP"
            ></el-input>
            <div class="form-item-help">例如：36.132.128.0|36.132.147.255|中国|0|黑龙江省|哈尔滨市|移动</div>
          </el-form-item>
        </el-form>
        <template #footer>
          <span class="dialog-footer">
            <el-button @click="dialogVisible.edit = false">取 消</el-button>
            <el-button type="primary" @click="handleEditSegmentPut" :loading="loading.edit">
              确 定
            </el-button>
          </span>
        </template>
      </el-dialog>
      
      <!-- 保存并生成XDB对话框 -->
      <el-dialog
        v-model="dialogVisible.generate"
        title="保存并生成XDB文件"
        width="500px"
        :close-on-click-modal="false"
        :close-on-press-escape="!generatingTask"
        :show-close="!generatingTask"
        @closed="handleGenerateDialogClosed"
      >
        <el-form :model="generateForm" label-width="100px">
          <el-form-item label="目标文件" required>
            <el-input 
              v-model="generateForm.dstFile" 
              placeholder="请输入目标文件路径，如：data/ip2region.xdb"
              :disabled="generatingTask"
            ></el-input>
          </el-form-item>

          <!-- 生成状态显示 -->
          <el-form-item v-if="generatingTask">
            <div class="export-progress">
              <div class="progress-info">
                <p><strong>当前状态:</strong> {{ generateTaskStatusText }}</p>
                <p><strong>已运行时间:</strong> {{ getFormattedGenerateRunningTime() }}</p>
              </div>
            </div>
          </el-form-item>
        </el-form>
        <template #footer>
          <div class="dialog-footer">
            <el-button @click="dialogVisible.generate = false" :disabled="generatingTask">
              取 消
            </el-button>
            <template v-if="!generatingTask">
              <el-button type="primary" @click="handleSaveAndGenerate" :loading="loading.generate">
                确 定
              </el-button>
            </template>
            <template v-else-if="generateTask.status !== 'completed' && generateTask.status !== 'failed'">
              <el-button type="danger" @click="handleCancelGenerateTask">
                取消生成
              </el-button>
            </template>
          </div>
        </template>
      </el-dialog>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, onUnmounted } from 'vue'
import { 
  listSegments, 
  editSegment, 
  editSegmentPut, 
  editFromFile, 
  saveAndGenerateDb,
  saveAndGenerateDbWithProgress,
  getGenerateTaskStatus,
  cancelGenerateTask,
  getCurrentEditFile,
  unloadEditFile
} from '@/api'
import { ElMessage, ElMessageBox } from 'element-plus'

// 表单数据
const editForm = ref({
  srcFile: ''
})

const addForm = ref({
  segment: ''
})

const importForm = ref({
  file: ''
})

const editSegmentForm = ref({
  segment: ''
})

const generateForm = ref({
  dstFile: ''
})

// 列表数据
const segments = ref([])
const totalSegments = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const segmentsLoaded = ref(false)
const needSave = ref(false)

// 对话框状态
const dialogVisible = reactive({
  add: false,
  import: false,
  edit: false,
  generate: false
})

// 加载状态
const loading = reactive({
  list: false,
  add: false,
  import: false,
  edit: false,
  save: false,
  generate: false
})

// 错误信息
const error = ref('')

// 生成任务相关
const generatingTask = ref(false)
const generateTask = ref({})
const generateTaskId = ref('')
const generateTimerId = ref(null)

// 计算生成任务状态文本
const generateTaskStatusText = computed(() => {
  if (!generateTask.value.status) return '准备中';
  
  switch (generateTask.value.status) {
    case 'pending': return '等待处理';
    case 'processing': return '正在生成XDB';
    case 'completed': return '生成完成';
    case 'failed': return '生成失败: ' + (generateTask.value.errorMessage || '未知错误');
    default: return generateTask.value.status;
  }
})

// 格式化显示运行时间（秒）
const getFormattedGenerateRunningTime = () => {
  if (!generateTask.value.durationSeconds && generateTask.value.durationSeconds !== 0) return '0秒';
  return `${generateTask.value.durationSeconds}秒`;
}

// 组件挂载时尝试从后端获取当前编辑的文件路径
onMounted(async () => {
  try {
    const response = await getCurrentEditFile();
    console.log('获取当前编辑文件信息:', response);
    if (response.data && response.data.fileLoaded && response.data.currentEditFile) {
      editForm.value.srcFile = response.data.currentEditFile;
      ElMessage.info(`已恢复编辑文件: ${response.data.currentEditFile}`);
      // 自动加载当前编辑的文件
      await handleLoadSegments();
    } else {
      console.log('当前没有加载的编辑文件');
    }
  } catch (error) {
    console.error('获取当前编辑文件信息失败:', error);
    // 不显示错误消息，因为可能只是没有加载文件
  }
});

// 开始轮询获取生成任务状态
const startPollingGenerateStatus = (taskId) => {
  // 清理之前的定时器
  if (generateTimerId.value) {
    clearInterval(generateTimerId.value)
    generateTimerId.value = null
  }
  
  console.log('开始轮询生成任务状态，任务ID:', taskId)
  
  // 立即查询一次
  getGenerateTaskProgress(taskId)
  
  // 设置定时轮询
  generateTimerId.value = setInterval(() => {
    getGenerateTaskProgress(taskId)
  }, 1000) // 2秒查询一次
}

// 获取生成任务进度
const getGenerateTaskProgress = async (taskId) => {
  try {
    const response = await getGenerateTaskStatus(taskId)
    console.log('获取生成任务状态成功:', response.data)
    
    // 确保任务数据有效
    if (!response.data) {
      console.error('任务数据为空')
      return
    }
    
    generateTask.value = response.data
    
    // 如果任务已完成或失败，停止轮询
    if (generateTask.value.status === 'completed' || generateTask.value.status === 'failed') {
      if (generateTimerId.value) {
        clearInterval(generateTimerId.value)
        generateTimerId.value = null
      }
      
      // 显示完成或失败消息
      if (generateTask.value.status === 'completed') {
        ElMessage.success(`XDB文件生成成功：${generateTask.value.dstFile}`)
        generatingTask.value = false
        loading.generate = false
        dialogVisible.generate = false
      } else if (generateTask.value.status === 'failed') {
        ElMessage.error(`生成失败：${generateTask.value.errorMessage || '未知错误'}`)
        generatingTask.value = false
        loading.generate = false
      }
    }
  } catch (error) {
    console.error('获取生成任务状态失败', error)
    
    // 失败重试次数限制
    if (!generateTask.value.retryCount) {
      generateTask.value.retryCount = 1
    } else {
      generateTask.value.retryCount++
    }
    
    // 超过5次失败后停止轮询
    if (generateTask.value.retryCount > 5) {
      if (generateTimerId.value) {
        clearInterval(generateTimerId.value)
        generateTimerId.value = null
      }
      ElMessage.error('获取任务状态失败，请检查网络连接')
      generatingTask.value = false
      loading.generate = false
    }
  }
}

// 取消生成任务
const handleCancelGenerateTask = async () => {
  if (!generateTaskId.value) return
  
  try {
    await cancelGenerateTask(generateTaskId.value)
    ElMessage.warning('生成任务已取消')
    
    // 立即更新一次状态
    await getGenerateTaskProgress(generateTaskId.value)
  } catch (error) {
    ElMessage.error(error.message || '取消任务失败')
  }
}

// 加载IP段列表
const handleLoadSegments = async () => {
  if (!editForm.value.srcFile) {
    error.value = '请输入源文件路径'
    ElMessage.warning('请先输入源文件路径')
    return
  }
  
  loading.list = true
  error.value = ''
  
  try {
    const res = await listSegments(
      editForm.value.srcFile,
      (currentPage.value - 1) * pageSize.value,
      pageSize.value
    )
    
    // 将返回的IP段对象转换为可显示的字符串格式
    segments.value = res.data.segments.map(segment => {
      // 将整数IP转换为IP地址字符串
      const startIP = intToIp(segment.StartIP)
      const endIP = intToIp(segment.EndIP)
      const region = segment.Region
      
      // 格式化为字符串格式: startIP|endIP|region
      const segmentStr = `${startIP}|${endIP}|${region}`
      
      return { 
        segment: segmentStr,
        rawData: segment 
      }
    })
    
    totalSegments.value = res.data.total
    segmentsLoaded.value = true
    needSave.value = false
    
    if (segments.value.length > 0) {
      ElMessage.success(`成功加载 ${segments.value.length} 条IP段记录，共 ${totalSegments.value} 条`)
    } else {
      ElMessage.info('文件已加载，但没有找到IP段记录')
    }
  } catch (err) {
    error.value = err.message || '加载数据失败'
    ElMessage.error(`加载IP段失败: ${err.message || '未知错误'}`)
    console.error('加载IP段失败:', err)
  } finally {
    loading.list = false
  }
}

// 整数IP转换为IP地址字符串（如：16777216 -> 1.0.0.0）
const intToIp = (ipInt) => {
  if (typeof ipInt !== 'number') return '无效IP'
  
  const a = (ipInt >> 24) & 0xff
  const b = (ipInt >> 16) & 0xff
  const c = (ipInt >> 8) & 0xff
  const d = ipInt & 0xff
  
  return `${a}.${b}.${c}.${d}`
}

// 构建IP段显示字符串
const buildSegmentString = (segment) => {
  if (!segment) return ''
  if (typeof segment === 'string') return segment
  
  const startIP = segment.startIP || intToIp(segment.StartIP)
  const endIP = segment.endIP || intToIp(segment.EndIP)
  const region = segment.region || segment.Region
  
  return `${startIP}|${endIP}|${region}`
}

// 分页事件处理
const handleSizeChange = () => {
  handleLoadSegments()
}

const handleCurrentChange = () => {
  handleLoadSegments()
}

// 添加IP段
const handleAddSegment = async () => {
  if (!addForm.value.segment) {
    ElMessage.warning('请输入IP段')
    return
  }
  
  loading.add = true
  error.value = ''
  
  try {
    const res = await editSegment(addForm.value.segment, editForm.value.srcFile)
    ElMessage.success('添加成功')
    dialogVisible.add = false
    addForm.value.segment = ''
    needSave.value = true
    
    // 重新加载当前页数据
    await handleLoadSegments()
  } catch (err) {
    error.value = err.message || '添加IP段失败'
  } finally {
    loading.add = false
  }
}

// 删除IP段
const handleDelete = async (index) => {
  try {
    await ElMessageBox.confirm('确定要删除该IP段吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 这里简化处理：前端删除后标记需要保存，但不直接调用API
    // 实际删除在保存时处理
    segments.value.splice(index, 1)
    needSave.value = true
    ElMessage.success('已标记删除，点击保存按钮生效')
  } catch {
    // 用户取消
  }
}

// 从文件导入
const handleImportFile = async () => {
  if (!importForm.value.file) {
    ElMessage.warning('请输入文件路径')
    return
  }
  
  loading.import = true
  error.value = ''
  
  try {
    const res = await editFromFile(importForm.value.file, editForm.value.srcFile)
    ElMessage.success(`导入成功，删除了${res.data.oldCount}条，添加了${res.data.newCount}条`)
    dialogVisible.import = false
    importForm.value.file = ''
    needSave.value = true
    
    // 重新加载当前页数据
    await handleLoadSegments()
  } catch (err) {
    error.value = err.message || '从文件导入失败'
  } finally {
    loading.import = false
  }
}

// 通过PUT方法编辑IP段
const handleEditSegmentPut = async () => {
  if (!editSegmentForm.value.segment) {
    ElMessage.warning('请输入IP段')
    return
  }
  
  loading.edit = true
  error.value = ''
  
  try {
    const res = await editSegmentPut(editSegmentForm.value.segment, editForm.value.srcFile)
    ElMessage.success('编辑成功')
    dialogVisible.edit = false
    editSegmentForm.value.segment = ''
    needSave.value = true
    
    // 重新加载当前页数据
    await handleLoadSegments()
  } catch (err) {
    error.value = err.message || '编辑IP段失败'
  } finally {
    loading.edit = false
  }
}

// 保存并生成
const handleSaveAndGenerate = async () => {
  if (!editForm.value.srcFile) {
    ElMessage.warning('请先编辑一个文件')
    return
  }
  
  if (!generateForm.value.dstFile) {
    ElMessage.warning('请输入目标文件路径')
    return
  }
  
  loading.generate = true
  
  // 重置任务状态
  generatingTask.value = false
  generateTask.value = {}
  generateTaskId.value = ''
  
  try {
    // 使用新的异步生成接口
    const response = await saveAndGenerateDbWithProgress(editForm.value.srcFile, generateForm.value.dstFile)
    
    // 保存任务ID
    generateTaskId.value = response.data.taskId
    
    // 显示进度
    generatingTask.value = true
    generateTask.value = {
      status: 'pending',
      startTime: new Date().getTime()
    }
    
    // 开始轮询任务状态
    startPollingGenerateStatus(generateTaskId.value)
    
    ElMessage.info(`生成任务已创建，正在处理中...`)
  } catch (error) {
    ElMessage.error(error.message || '保存并生成数据库失败')
    loading.generate = false
    generatingTask.value = false
    generateTask.value = {}
    generateTaskId.value = ''
  }
}

// 处理编辑行
const handleEditRow = (row) => {
  // 使用buildSegmentString构建IP段字符串
  editSegmentForm.value.segment = buildSegmentString(row.rawData)
  dialogVisible.edit = true
}

// 添加表格ref
const tableRef = ref(null)

// 添加行点击事件
const handleRowClick = (row) => {
  // 点击行时，打开编辑对话框
  handleEditRow(row)
}

// 添加对话框关闭事件处理，确保任务相关状态被重置
const handleGenerateDialogClosed = () => {
  // 重置生成任务相关状态
  generatingTask.value = false
  generateTask.value = {}
  generateTaskId.value = ''
}

// 卸载文件
const handleUnloadFile = async () => {
  if (!editForm.value.srcFile) {
    ElMessage.warning('当前没有加载的文件');
    return;
  }

  try {
    // 显示确认对话框
    await ElMessageBox.confirm(
      '确定要卸载当前编辑的文件吗？所有未保存的更改将会丢失。',
      '确认卸载',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    );

    const currentFile = editForm.value.srcFile;
    
    // 调用后端API卸载文件
    const response = await unloadEditFile();
    console.log('卸载文件响应:', response);
    
    // 重置所有状态
    segments.value = []
    totalSegments.value = 0
    currentPage.value = 1
    pageSize.value = 20
    segmentsLoaded.value = false
    needSave.value = false
    dialogVisible.add = false
    dialogVisible.import = false
    dialogVisible.edit = false
    dialogVisible.generate = false
    editForm.value.srcFile = ''
    error.value = ''
    generatingTask.value = false
    generateTask.value = {}
    generateTaskId.value = ''
    
    ElMessage.success(`文件 ${currentFile} 已成功卸载`);
  } catch (error) {
    if (error.message !== 'cancel') {
      console.error('卸载文件失败:', error);
      ElMessage.error(error.message || '卸载文件失败');
    }
  }
}

// 组件卸载时确保清理定时器
onUnmounted(() => {
  if (generateTimerId.value) {
    clearInterval(generateTimerId.value);
    generateTimerId.value = null;
  }
});
</script>

<style scoped>
.form-item-help {
  font-size: 12px;
  color: #606266;
  margin-top: 5px;
}

.edit-section {
  margin-top: 20px;
}

.pagination {
  display: flex;
  justify-content: flex-end;
}

.flex-row {
  display: flex;
  align-items: center;
}

.space-between {
  justify-content: space-between;
}

.ml-10 {
  margin-left: 10px;
}

.mb-20 {
  margin-bottom: 20px;
}

.mt-20 {
  margin-top: 20px;
}

.text-right {
  text-align: right;
}

/* 高亮搜索结果 */
.el-table .current-row {
  background-color: #ecf5ff !important;
}

.export-progress {
  margin-top: 20px;
  padding: 10px;
  background-color: #fff;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.progress-info {
  margin-bottom: 10px;
}

.progress-info p {
  margin: 5px 0;
}

.progress-info strong {
  font-weight: bold;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
}
</style>