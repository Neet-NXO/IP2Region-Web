import axios from 'axios'
import { ElMessage } from 'element-plus'

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

export function loadXdbToMemory(dbPath, searchMode = 'vector') {
  return http.post('/load-xdb', {
    dbPath,
    searchMode
  })
}

export function unloadXdb() {
  return http.post('/unload-xdb')
}

export function getXdbStatus() {
  return http.get('/xdb-status')
}

export function exportXdb(xdbPath, exportPath) {
  return http.post('/export-xdb', {
    xdbPath,
    exportPath
  })
}

export function getExportTaskStatus(taskId) {
  return http.get(`/export-task/${taskId}`)
}

export function cancelExportTask(taskId) {
  return http.post(`/export-task/${taskId}/cancel`)
}

export function generateDb(srcFile, dstFile) {
  return http.post('/generate', {
    srcFile,
    dstFile
  })
}

export function getGenerateTaskStatus(taskId) {
  return http.get(`/generate-task/${taskId}`)
}

export function cancelGenerateTask(taskId) {
  return http.post(`/generate-task/${taskId}/cancel`)
}

export function editSegment(segment, srcFile) {
  return http.post('/edit/segment', {
    segment,
    srcFile
  })
}

export function editFromFile(file, srcFile) {
  return http.post('/edit/file', {
    file,
    srcFile
  })
}

export function listSegments(srcFile, offset = 0, size = 10) {
  return http.post('/list/segments', {
    srcFile,
    offset,
    size
  })
}

export function saveEdit(srcFile) {
  return http.post('/edit/save', {
    srcFile
  })
}

export function editSegmentPut(segment, srcFile) {
  return http.put('/edit/segment', {
    segment,
    srcFile
  })
}

export function saveAndGenerateDb(srcFile, dstFile) {
  return http.post('/edit/saveAndGenerate', {
    srcFile,
    dstFile
  })
}

export function getCurrentEditFile() {
  return http.get('/edit/current-file')
}

export function unloadEditFile() {
  return http.post('/edit/unload-file')
}

export function saveAndGenerateDbWithProgress(srcFile, dstFile) {
  return http.post('/generate-with-progress', {
    srcFile,
    dstFile
  })
}

export default http 
