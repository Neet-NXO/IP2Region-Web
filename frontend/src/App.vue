<template>
  <div class="app-container">
    <el-config-provider :locale="zhCn">
      <el-container>
        <el-header height="60px">
          <div class="header-content">
            <div class="logo">IP2Region Web</div>
            <div class="menu">
              <el-menu 
                :default-active="activeIndex" 
                mode="horizontal" 
                router 
                @select="handleSelect"
              >
                <el-menu-item index="/">首页</el-menu-item>
                <el-menu-item index="/search">IP查询</el-menu-item>
                <el-menu-item index="/generate">生成数据库</el-menu-item>
                <el-menu-item index="/edit">编辑数据</el-menu-item>
              </el-menu>
            </div>
          </div>
        </el-header>
        <el-main>
          <router-view />
        </el-main>
        <el-footer height="40px">
          <div class="footer-content">
            IP2Region Web © {{ new Date().getFullYear() }}
          </div>
        </el-footer>
      </el-container>
    </el-config-provider>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElConfigProvider } from 'element-plus'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'

const route = useRoute()
const router = useRouter()

const activeIndex = computed(() => {
  return route.path
})

const handleSelect = (key) => {
  router.push(key)
}

onMounted(() => {
  // 页面加载完成
})
</script>

<style>
html, body, #app {
  height: 100%;
  margin: 0;
  padding: 0;
}

.app-container {
  height: 100%;
}

.el-container {
  height: 100%;
}

.el-header {
  border-bottom: 1px solid #e6e6e6;
  background-color: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.logo {
  font-size: 20px;
  font-weight: bold;
  color: #409EFF;
}

.menu {
  flex-grow: 1;
  margin-left: 30px;
}

.el-menu {
  width: 100%;
  display: flex;
  justify-content: flex-start;
}

.el-menu--horizontal .el-menu-item {
  flex-shrink: 0;
}

.el-main {
  background-color: #f5f7fa;
  padding: 20px;
}

.el-footer {
  background-color: #fff;
  color: #888;
  border-top: 1px solid #e6e6e6;
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
}

.footer-content {
  text-align: center;
}
</style> 