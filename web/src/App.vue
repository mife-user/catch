<template>
  <div class="app-container">
    <div class="sidebar">
      <div class="sidebar-header">
        <h2>Catch</h2>
        <span class="version">v1.0.0</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        class="sidebar-menu"
        @select="handleMenuSelect"
        background-color="#001529"
        text-color="#ffffffa6"
        active-text-color="#ffffff"
      >
        <el-menu-item index="/search">
          <el-icon><Search /></el-icon>
          <span>文件查找</span>
        </el-menu-item>
        <el-menu-item index="/delete">
          <el-icon><Delete /></el-icon>
          <span>文件删除</span>
        </el-menu-item>
        <el-menu-item index="/rename">
          <el-icon><Edit /></el-icon>
          <span>文件重命名</span>
        </el-menu-item>
        <el-menu-item index="/move">
          <el-icon><Rank /></el-icon>
          <span>文件移动</span>
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

    <div class="main-content">
      <router-view />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Search, Delete, Edit, Rank, Setting } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()

const activeMenu = computed(() => route.path)

const handleMenuSelect = (index) => {
  router.push(index)
}
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

.sidebar {
  width: 200px;
  min-width: 200px;
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

.main-content {
  flex: 1;
  background-color: #f0f2f5;
  overflow-y: auto;
  padding: 24px;
}
</style>
