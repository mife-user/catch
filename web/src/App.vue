<template>
  <div class="app-container">
    <div v-if="!isWelcome" class="mobile-header">
      <el-button class="hamburger-btn" :icon="Fold" @click="drawerVisible = true" text />
      <h2 class="mobile-title">Catch</h2>
      <span class="mobile-version">v1.1.0</span>
    </div>

    <el-drawer
      v-model="drawerVisible"
      direction="ltr"
      :size="240"
      :show-close="false"
      :with-header="false"
      class="sidebar-drawer"
    >
      <div class="sidebar">
        <div class="sidebar-header">
          <h2>Catch</h2>
          <span class="version">v1.1.0</span>
        </div>
        <div class="sidebar-search">
          <el-input
            v-model="menuSearch"
            placeholder="搜索功能..."
            clearable
            size="small"
            prefix-icon="Search"
          />
        </div>
        <el-menu
          :default-active="activeMenu"
          :default-openeds="defaultOpeneds"
          class="sidebar-menu"
          @select="handleMenuSelect"
          background-color="#001529"
          text-color="#ffffffa6"
          active-text-color="#ffffff"
        >
          <el-sub-menu index="file-manager">
            <template #title>
              <el-icon><Folder /></el-icon>
              <span>文件管理</span>
            </template>
            <el-menu-item v-for="item in filteredFileMenuItems" :key="item.index" :index="item.index">
              <el-icon><component :is="item.icon" /></el-icon>
              <span>{{ item.label }}</span>
              <el-icon
                v-if="dragEnabled"
                class="drag-handle"
                @click.stop
              ><Rank /></el-icon>
            </el-menu-item>
          </el-sub-menu>

          <el-menu-item index="/trash">
            <el-icon><Delete /></el-icon>
            <span>回收站</span>
          </el-menu-item>

          <el-sub-menu index="settings">
            <template #title>
              <el-icon><Setting /></el-icon>
              <span>设置</span>
            </template>
            <el-menu-item index="/settings/basic">基础</el-menu-item>
            <el-menu-item index="/settings/security">安全</el-menu-item>
            <el-menu-item index="/settings/smtp">SMTP</el-menu-item>
            <el-menu-item index="/settings/about">关于</el-menu-item>
          </el-sub-menu>
        </el-menu>
      </div>
    </el-drawer>

    <div v-if="!isWelcome" class="sidebar desktop-sidebar">
      <div class="sidebar-header">
        <h2>Catch</h2>
        <span class="version">v1.1.0</span>
      </div>
      <div class="sidebar-search">
        <el-input
          v-model="menuSearch"
          placeholder="搜索功能..."
          clearable
          size="small"
          prefix-icon="Search"
        />
      </div>
      <el-menu
        :default-active="activeMenu"
        :default-openeds="defaultOpeneds"
        class="sidebar-menu"
        @select="handleMenuSelect"
        background-color="#001529"
        text-color="#ffffffa6"
        active-text-color="#ffffff"
      >
        <el-sub-menu index="file-manager">
          <template #title>
            <el-icon><Folder /></el-icon>
            <span>文件管理</span>
          </template>
          <el-menu-item v-for="item in filteredFileMenuItems" :key="item.index" :index="item.index">
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
            <el-icon
              v-if="dragEnabled"
              class="drag-handle"
              @click.stop
            ><Rank /></el-icon>
          </el-menu-item>
        </el-sub-menu>

        <el-menu-item index="/trash">
          <el-icon><Delete /></el-icon>
          <span>回收站</span>
        </el-menu-item>

        <el-sub-menu index="settings">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>设置</span>
          </template>
          <el-menu-item index="/settings/basic">基础</el-menu-item>
          <el-menu-item index="/settings/security">安全</el-menu-item>
          <el-menu-item index="/settings/smtp">SMTP</el-menu-item>
          <el-menu-item index="/settings/about">关于</el-menu-item>
        </el-sub-menu>
      </el-menu>
    </div>

    <div class="main-content" :class="{ 'full-width': isWelcome }">
      <router-view />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  Search, Delete, Edit, Rank, CopyDocument, Setting, Fold, Folder,
  Brush, Monitor
} from '@element-plus/icons-vue'
import { getConfig } from './api/config'

const router = useRouter()
const route = useRoute()
const drawerVisible = ref(false)
const menuSearch = ref('')
const dragEnabled = ref(true)

const isWelcome = computed(() => route.path === '/welcome')

const defaultFileMenuItems = [
  { index: '/search', label: '文件查找', icon: Search },
  { index: '/delete', label: '文件删除', icon: Delete },
  { index: '/rename', label: '文件重命名', icon: Edit },
  { index: '/move', label: '文件移动', icon: Rank },
  { index: '/copy', label: '文件复制', icon: CopyDocument },
  { index: '/cleanup', label: '文件清理', icon: Brush },
  { index: '/uninstall', label: '软件卸载', icon: Monitor },
]

const fileMenuItems = ref([...defaultFileMenuItems])

const filteredFileMenuItems = computed(() => {
  if (!menuSearch.value) return fileMenuItems.value
  const kw = menuSearch.value.toLowerCase()
  return fileMenuItems.value.filter(item => item.label.toLowerCase().includes(kw))
})

const defaultOpeneds = computed(() => {
  const filePaths = ['/search', '/delete', '/rename', '/move', '/copy', '/cleanup', '/uninstall']
  if (filePaths.includes(route.path)) return ['file-manager']
  if (route.path.startsWith('/settings')) return ['settings']
  return []
})

const activeMenu = computed(() => {
  if (route.path === '/move' && route.query.mode === 'copy') {
    return '/copy'
  }
  return route.path
})

const handleMenuSelect = (index) => {
  drawerVisible.value = false
  if (index === '/copy') {
    router.push({ path: '/move', query: { mode: 'copy' } })
  } else {
    router.push(index)
  }
}

onMounted(async () => {
  try {
    const config = await getConfig()
    if (config.first_launch) {
      router.replace('/welcome')
    }
    if (config.menu_order && Array.isArray(config.menu_order)) {
      const orderMap = {}
      config.menu_order.forEach((idx, i) => { orderMap[idx] = i })
      fileMenuItems.value.sort((a, b) => {
        const oa = orderMap[a.index] !== undefined ? orderMap[a.index] : 999
        const ob = orderMap[b.index] !== undefined ? orderMap[b.index] : 999
        return oa - ob
      })
    }
  } catch {}
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  height: 100%;
  width: 100%;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

.app-container {
  display: flex;
  height: 100vh;
  width: 100vw;
}

.mobile-header {
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 56px;
  background-color: #001529;
  align-items: center;
  padding: 0 16px;
  z-index: 100;
}

.hamburger-btn {
  color: #ffffff !important;
  font-size: 20px;
}

.mobile-title {
  color: #ffffff;
  font-size: 18px;
  font-weight: 600;
  margin-left: 12px;
}

.mobile-version {
  color: #ffffffa6;
  font-size: 12px;
  margin-left: 8px;
}

.sidebar {
  width: 220px;
  min-width: 220px;
  background-color: #001529;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.sidebar-header {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border-bottom: 1px solid #ffffff1a;
}

.sidebar-header h2 {
  color: #ffffff;
  font-size: 20px;
  font-weight: 600;
}

.sidebar-header .version {
  color: #ffffffa6;
  font-size: 12px;
}

.sidebar-search {
  padding: 8px 12px;
  border-bottom: 1px solid #ffffff1a;
}

.sidebar-search .el-input__wrapper {
  background-color: #ffffff1a;
  box-shadow: none;
}

.sidebar-search .el-input__inner {
  color: #ffffffd9;
}

.sidebar-search .el-input__inner::placeholder {
  color: #ffffff73;
}

.sidebar-search .el-input__prefix .el-icon {
  color: #ffffff73;
}

.sidebar-menu {
  border-right: none;
  flex: 1;
  overflow-y: auto;
}

.sidebar-menu .el-menu-item,
.sidebar-menu .el-sub-menu__title {
  height: 48px;
  line-height: 48px;
}

.drag-handle {
  margin-left: auto;
  color: #ffffff4d;
  font-size: 12px;
  cursor: grab;
  transition: color 0.2s;
}

.drag-handle:hover {
  color: #ffffffa6;
}

.main-content {
  flex: 1;
  background-color: #f0f2f5;
  overflow-y: auto;
  padding: 24px;
}

.main-content.full-width {
  padding: 0;
}

.sidebar-drawer .el-drawer__body {
  padding: 0;
  background-color: #001529;
}

@media (max-width: 1200px) and (min-width: 769px) {
  .desktop-sidebar {
    width: 64px;
    min-width: 64px;
  }

  .desktop-sidebar .sidebar-header h2,
  .desktop-sidebar .sidebar-header .version {
    display: none;
  }

  .desktop-sidebar .sidebar-header {
    height: 48px;
  }

  .desktop-sidebar .sidebar-search {
    display: none;
  }

  .desktop-sidebar .el-menu-item span,
  .desktop-sidebar .el-sub-menu__title span {
    display: none;
  }

  .desktop-sidebar .el-menu-item,
  .desktop-sidebar .el-sub-menu__title {
    padding: 0 !important;
    justify-content: center;
  }

  .desktop-sidebar .el-sub-menu .el-menu-item {
    padding: 0 !important;
    justify-content: center;
    min-width: auto;
  }

  .desktop-sidebar .drag-handle {
    display: none;
  }
}

@media (max-width: 768px) {
  .mobile-header {
    display: flex;
  }

  .desktop-sidebar {
    display: none;
  }

  .main-content {
    padding: 16px;
    padding-top: 72px;
  }
}
</style>
