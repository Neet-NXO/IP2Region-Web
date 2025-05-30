import axios from 'axios'
import { ElMessage } from 'element-plus'

// 使用绝对路径确保在任何路由下都能正确请求
const baseURL = '/api'

const http = axios.create({
  baseURL,
  timeout: 60000
})

// 请求拦截器
http.interceptors.request.use(
  (config) => {
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
http.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.code !== 0) {
      ElMessage.error(res.msg || '请求失败')
      return Promise.reject(new Error(res.msg || '请求失败'))
    }
    return res
  },
  (error) => {
    // 提取详细错误信息
    let errorMsg = '请求失败'
    if (error.response && error.response.data) {
      // 服务器返回了错误信息
      errorMsg = error.response.data.msg || error.response.data.message || errorMsg
    } else if (error.message) {
      // 有错误消息
      errorMsg = error.message
    }
    ElMessage.error(errorMsg)
    return Promise.reject(new Error(errorMsg))
  }
)

// IP搜索接口
export function searchIP(ip, dbPath, searchMode) {
  const data = { ip }
  if (dbPath) {
    data.dbPath = dbPath
  }
  if (searchMode) {
    data.searchMode = searchMode
  }
  return http.post('/search', data)
}

// 加载XDB文件到内存接口
export function loadXdbToMemory(dbPath, searchMode = 'vector') {
  return http.post('/load-xdb', {
    dbPath,
    searchMode
  })
}

// 获取状态信息接口
export function getStatus() {
  return http.get('/status')
}

// 卸载内存中的XDB文件接口
export function unloadXdb() {
  return http.post('/unload-xdb')
}

// 获取XDB文件加载状态接口
export function getXdbStatus() {
  return http.get('/xdb-status')
}

// 导出XDB文件接口
export function exportXdb(xdbPath, exportPath) {
  return http.post('/export-xdb', {
    xdbPath,
    exportPath
  })
}

// 获取导出任务状态接口
export function getExportTaskStatus(taskId) {
  return http.get(`/export-task/${taskId}`)
}

// 取消导出任务接口
export function cancelExportTask(taskId) {
  return http.post(`/export-task/${taskId}/cancel`)
}

// 生成数据库接口
export function generateDb(srcFile, dstFile) {
  return http.post('/generate', {
    srcFile,
    dstFile
  })
}

// 查询数据库生成任务状态
export function getGenerateTaskStatus(taskId) {
  return http.get(`/generate-task/${taskId}`)
}

// 取消生成任务接口
export function cancelGenerateTask(taskId) {
  return http.post(`/generate-task/${taskId}/cancel`)
}

// 编辑IP段接口
export function editSegment(segment, srcFile) {
  return http.post('/edit/segment', {
    segment,
    srcFile
  })
}

// 从文件批量编辑IP段接口
export function editFromFile(file, srcFile) {
  return http.post('/edit/file', {
    file,
    srcFile
  })
}

// 列出IP段接口
export function listSegments(srcFile, offset = 0, size = 10) {
  return http.post('/list/segments', {
    srcFile,
    offset,
    size
  })
}

// 保存编辑接口
export function saveEdit(srcFile) {
  return http.post('/edit/save', {
    srcFile
  })
}

// 编辑IP段接口（PUT方法）
export function editSegmentPut(segment, srcFile) {
  return http.put('/edit/segment', {
    segment,
    srcFile
  })
}

// 保存编辑并生成数据库接口
export function saveAndGenerateDb(srcFile, dstFile) {
  return http.post('/edit/saveAndGenerate', {
    srcFile,
    dstFile
  })
}

// 获取当前编辑的源文件信息
export function getCurrentEditFile() {
  return http.get('/edit/current-file')
}

// 卸载当前编辑的源文件
export function unloadEditFile() {
  return http.post('/edit/unload-file')
}

// 保存编辑并生成数据库接口（带进度显示）
export function saveAndGenerateDbWithProgress(srcFile, dstFile) {
  return http.post('/generate-with-progress', {
    srcFile,
    dstFile
  })
}

export default http 