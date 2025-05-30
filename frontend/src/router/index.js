import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/Home.vue'),
    meta: {
      title: '首页'
    }
  },
  {
    path: '/search',
    name: 'Search',
    component: () => import('@/views/Search.vue'),
    meta: {
      title: 'IP查询'
    }
  },
  {
    path: '/generate',
    name: 'Generate',
    component: () => import('@/views/Generate.vue'),
    meta: {
      title: '生成数据库'
    }
  },
  {
    path: '/edit',
    name: 'Edit',
    component: () => import('@/views/Edit.vue'),
    meta: {
      title: '编辑数据'
    }
  }
  // IP修改功能已被移除
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

// 路由标题设置
router.beforeEach((to, from, next) => {
  document.title = `${to.meta.title} - IP2Region Web`
  next()
})

export default router 