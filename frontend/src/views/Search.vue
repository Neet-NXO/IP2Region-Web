<template>
  <div class="container">
    <div class="section">
      <h2 class="section-title">IP地址查询</h2>
      
      <!-- 状态信息 -->
      <div v-if="status" class="status-info mb-20">
        <el-tag v-if="status.loaded" type="success">
          数据库已加载 - {{ getModeText(status.searchMode) }}
        </el-tag>
        <el-tag v-else type="info">未加载数据库</el-tag>
        
        <!-- 智能模式选择提示 -->
        <div v-if="status.loaded" class="mode-tip">
          <el-alert
            :title="`已自动选择${getModeText(status.searchMode)}`"
            type="success"
            :description="`检测到数据库已加载为${getModeText(status.searchMode)}，已自动为您选择此模式以获得最佳性能。`"
            show-icon
            :closable="false"
            size="small"
          />
        </div>
      </div>
      
      <el-form :model="searchForm" label-width="140px" @submit.prevent="handleSearch">
        <el-form-item label="搜索模式">
          <el-radio-group v-model="searchForm.searchMode">
            <el-radio value="file">
              <div class="mode-option">
                <div class="mode-name">文件模式</div>
                <div class="mode-desc">每次查询从文件读取，内存占用最低</div>
              </div>
            </el-radio>
            <el-radio value="vector">
              <div class="mode-option">
                <div class="mode-name">向量模式</div>
                <div class="mode-desc">缓存向量索引，平衡内存和性能</div>
              </div>
            </el-radio>
            <el-radio value="memory">
              <div class="mode-option">
                <div class="mode-name">内存模式</div>
                <div class="mode-desc">全部数据加载到内存，性能最佳</div>
              </div>
            </el-radio>
          </el-radio-group>
        </el-form-item>
        
        <el-form-item label="数据库文件" v-if="searchForm.searchMode === 'file'">
          <el-input v-model="searchForm.dbPath" placeholder="请输入.xdb数据库文件路径，如: ip2region.xdb"></el-input>
        </el-form-item>
        
        <el-form-item label="IP地址">
          <div class="flex-row">
            <el-input 
              v-model="searchForm.ip" 
              placeholder="请输入IP地址，如: 114.114.114.114"
              @keyup.enter="handleSearch"
            ></el-input>
            <el-button type="primary" class="ml-10" @click="handleSearch" :loading="loading">查询</el-button>
          </div>
        </el-form-item>
      </el-form>
      
      <!-- 模式提示 -->
      <div v-if="searchForm.searchMode !== 'file' && !status?.loaded" class="xdb-warning mt-20">
        <el-alert
          :title="`${getModeText(searchForm.searchMode)}需要先加载数据库`"
          type="warning"
          :description="`${getModeText(searchForm.searchMode)}需要先在首页加载XDB文件。点击按钮前往首页加载，或切换到文件模式进行查询。`"
          show-icon
          :closable="false"
        >
          <template #default>
            <div class="alert-actions">
              <el-button type="primary" size="small" @click="goToHome">
                前往首页加载
              </el-button>
            </div>
          </template>
        </el-alert>
      </div>
      
      <!-- 文件模式性能说明 -->
      <div v-if="searchForm.searchMode === 'file'" class="file-mode-tip mt-20">
        <el-alert
          title="文件模式性能说明"
          type="info"
          description="文件模式下，初次查询会从硬盘读取数据，后续查询相同或相近IP地址时，由于操作系统的文件缓存机制，查询速度会显著提升，IO操作次数可能为0。这是正常现象，不是程序bug。"
          show-icon
          :closable="true"
        />
      </div>
      
      <el-divider />
      
      <div v-if="searchResult" class="result-section">
        <h3>查询结果</h3>
        
        <el-descriptions :column="2" border>
          <el-descriptions-item label="查询IP">
            <el-tag type="primary">{{ searchForm.ip }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="归属地信息">
            <span class="region-text">{{ searchResult.region }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="搜索模式">
            <el-tag :type="getModeTagType(searchResult.searchMode)">
              {{ getModeText(searchResult.searchMode) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="IO操作次数">
            <el-tag :type="searchResult.ioCount === 0 ? 'success' : 'warning'">
              {{ searchResult.ioCount }} 次
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="查询耗时">
            <el-tag type="info">{{ formatTime(searchResult.tookNanoseconds) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="查询时间">
            <span>{{ searchResult.queryTime }}</span>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </div>
  </div>
</template>

<script>
import { searchIP, getStatus } from '@/api'
import { ElMessage } from 'element-plus'

export default {
  name: 'Search',
  data() {
    return {
      loading: false,
      searchForm: {
        ip: '',
        searchMode: 'file',
        dbPath: 'ip2region.xdb'
      },
      searchResult: null,
      status: null
    }
  },
  mounted() {
    this.fetchStatus()
  },
  watch: {
    // 监听搜索模式的变化
    'searchForm.searchMode'(newMode, oldMode) {
      if (oldMode && newMode !== oldMode) {
        // 当用户主动切换模式时给出提示
        if (newMode === 'file') {
          this.$nextTick(() => {
            ElMessage.info('已切换到文件模式，每次查询会根据操作系统缓存情况产生不同的IO次数')
          })
        } else if (this.status?.loaded && this.status.searchMode !== newMode) {
          ElMessage.warning(`当前数据库加载为${this.getModeText(this.status.searchMode)}，建议使用已加载的模式以获得最佳性能`)
        }
      }
    }
  },
  methods: {
    async handleSearch() {
      if (!this.searchForm.ip.trim()) {
        ElMessage.warning('请输入IP地址')
        return
      }

      // 检查模式要求
      if (this.searchForm.searchMode !== 'file' && !this.status?.loaded) {
        ElMessage.warning(`${this.getModeText(this.searchForm.searchMode)}需要先加载数据库`)
        return
      }

      this.loading = true
      try {
        const result = await searchIP(
          this.searchForm.ip, 
          this.searchForm.dbPath, 
          this.searchForm.searchMode
        )
        this.searchResult = result.data
        ElMessage.success('查询成功')
      } catch (error) {
        console.error('查询失败:', error)
        ElMessage.error(error.message || '查询失败')
        this.searchResult = null
      } finally {
        this.loading = false
      }
    },

    async fetchStatus() {
      try {
        const result = await getStatus()
        this.status = result.data
        // 根据数据库状态自动选择搜索模式
        this.autoSelectSearchMode()
      } catch (error) {
        console.error('获取状态失败:', error)
      }
    },

    // 根据数据库加载状态自动选择搜索模式
    autoSelectSearchMode() {
      if (this.status?.loaded && this.status?.searchMode) {
        // 如果数据库已加载，默认使用已加载的模式
        this.searchForm.searchMode = this.status.searchMode
        // 如果是内存模式或向量模式，清空数据库路径（因为已经加载到内存中）
        if (this.status.searchMode === 'memory' || this.status.searchMode === 'vector') {
          this.searchForm.dbPath = ''
        }
      }
    },

    goToHome() {
      this.$router.push('/')
    },

    // 格式化时间
    formatTime(nanoseconds) {
      if (nanoseconds < 1000) {
        return `${nanoseconds} ns`
      } else if (nanoseconds < 1000000) {
        return `${(nanoseconds / 1000).toFixed(2)} μs`
      } else if (nanoseconds < 1000000000) {
        return `${(nanoseconds / 1000000).toFixed(2)} ms`
      } else {
        return `${(nanoseconds / 1000000000).toFixed(2)} s`
      }
    },

    // 获取模式文本
    getModeText(mode) {
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
    },

    // 获取模式标签类型
    getModeTagType(mode) {
      switch (mode) {
        case 'file':
          return 'info'
        case 'vector':
          return 'warning'
        case 'memory':
          return 'success'
        default:
          return 'info'
      }
    }
  }
}
</script>

<style scoped>
.container {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.section {
  background: white;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.section-title {
  margin: 0 0 30px 0;
  color: #303133;
  font-size: 24px;
  font-weight: 600;
}

.status-info {
  margin-bottom: 20px;
}

.mode-tip {
  margin-top: 10px;
}

.mb-20 {
  margin-bottom: 20px;
}

.mt-20 {
  margin-top: 20px;
}

.ml-10 {
  margin-left: 10px;
}

.flex-row {
  display: flex;
  align-items: center;
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

.xdb-warning {
  margin: 20px 0;
}

.alert-actions {
  margin-top: 10px;
}

.result-section {
  margin-top: 30px;
}

.result-section h3 {
  margin-bottom: 20px;
  color: #606266;
}

.region-text {
  font-weight: bold;
  color: #409eff;
  font-size: 16px;
}

:deep(.el-radio) {
  display: block;
  margin-bottom: 15px;
  height: auto;
}

:deep(.el-radio__label) {
  white-space: normal;
}
</style>