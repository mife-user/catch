<template>
  <div class="uninstall-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <el-icon><Monitor /></el-icon>
          <span>软件彻底卸载</span>
        </div>
      </template>

      <el-alert
        title="此功能将深度清理软件的所有残留文件、注册表项和系统服务，请谨慎操作"
        type="warning"
        :closable="false"
        show-icon
        class="warning-alert"
      />

      <div class="scan-section">
        <el-button type="primary" @click="handleScan" :loading="scanning">
          <el-icon><Search /></el-icon>
          扫描已安装软件
        </el-button>
      </div>
    </el-card>

    <ProgressBar
      v-if="analyzing"
      :visible="analyzing"
      title="正在分析软件残留..."
      position="inline"
      @cancel="analyzing = false"
    />

    <el-card v-if="apps.length > 0" class="apps-card">
      <template #header>
        <div class="card-header">
          <span>已安装软件 ({{ apps.length }})</span>
          <el-input
            v-model="appSearch"
            placeholder="搜索软件..."
            clearable
            size="small"
            prefix-icon="Search"
            style="width: 240px"
          />
        </div>
      </template>

      <el-table :data="filteredApps" stripe highlight-current-row @current-change="handleAppSelect">
        <el-table-column prop="name" label="软件名称" min-width="200" show-overflow-tooltip />
        <el-table-column prop="publisher" label="发布者" min-width="150" show-overflow-tooltip />
        <el-table-column prop="version" label="版本" width="120" />
        <el-table-column prop="install_date" label="安装日期" width="120" />
        <el-table-column prop="size" label="大小" width="120">
          <template #default="{ row }">
            {{ row.size > 0 ? formatSize(row.size) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" align="center">
          <template #default="{ row }">
            <el-button type="danger" size="small" @click.stop="handleAnalyze(row)">
              深度分析
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card v-if="analyzeResult" class="analyze-card">
      <template #header>
        <div class="card-header">
          <span>残留分析 - {{ selectedApp?.name }}</span>
          <el-tag type="info">总大小: {{ formatSize(analyzeResult.total_size) }}</el-tag>
        </div>
      </template>

      <div class="analyze-sections">
        <div class="analyze-section">
          <div class="section-header">
            <el-checkbox v-model="cleanOptions.programFiles" :indeterminate="false">
              <span class="section-title">程序文件</span>
            </el-checkbox>
            <el-tag size="small" type="info">{{ analyzeResult.program_files?.length || 0 }} 项</el-tag>
          </div>
          <div class="section-list">
            <div v-for="(item, idx) in analyzeResult.program_files?.slice(0, 10)" :key="idx" class="list-item" :class="{ important: analyzeResult.important?.[idx] }">
              <el-icon v-if="analyzeResult.important?.[idx]" color="#f5222d"><WarningFilled /></el-icon>
              <span>{{ item }}</span>
            </div>
            <div v-if="(analyzeResult.program_files?.length || 0) > 10" class="more-hint">
              ... 还有 {{ analyzeResult.program_files.length - 10 }} 项
            </div>
          </div>
        </div>

        <div class="analyze-section">
          <div class="section-header">
            <el-checkbox v-model="cleanOptions.registry">
              <span class="section-title">注册表项</span>
            </el-checkbox>
            <el-tag size="small" type="info">{{ analyzeResult.registry_items?.length || 0 }} 项</el-tag>
          </div>
          <div class="section-list">
            <div v-for="(item, idx) in analyzeResult.registry_items?.slice(0, 10)" :key="idx" class="list-item" :class="{ important: analyzeResult.important?.[analyzeResult.program_files?.length + idx] }">
              <el-icon v-if="analyzeResult.important?.[analyzeResult.program_files?.length + idx]" color="#f5222d"><WarningFilled /></el-icon>
              <span>{{ item }}</span>
            </div>
            <div v-if="(analyzeResult.registry_items?.length || 0) > 10" class="more-hint">
              ... 还有 {{ analyzeResult.registry_items.length - 10 }} 项
            </div>
          </div>
        </div>

        <div class="analyze-section">
          <div class="section-header">
            <el-checkbox v-model="cleanOptions.config">
              <span class="section-title">配置文件</span>
            </el-checkbox>
            <el-tag size="small" type="info">{{ analyzeResult.config_files?.length || 0 }} 项</el-tag>
          </div>
          <div class="section-list">
            <div v-for="item in analyzeResult.config_files" :key="item" class="list-item">
              <span>{{ item }}</span>
            </div>
            <div v-if="(analyzeResult.config_files?.length || 0) === 0" class="empty-hint">未发现配置文件</div>
          </div>
        </div>

        <div class="analyze-section">
          <div class="section-header">
            <el-checkbox v-model="cleanOptions.services">
              <span class="section-title">系统服务</span>
            </el-checkbox>
            <el-tag size="small" type="info">{{ analyzeResult.service_names?.length || 0 }} 项</el-tag>
          </div>
          <div class="section-list">
            <div v-for="item in analyzeResult.service_names" :key="item" class="list-item important">
              <el-icon color="#f5222d"><WarningFilled /></el-icon>
              <span>{{ item }}</span>
            </div>
            <div v-if="(analyzeResult.service_names?.length || 0) === 0" class="empty-hint">未发现系统服务</div>
          </div>
        </div>
      </div>

      <div class="execute-section">
        <el-button type="danger" size="large" @click="handleUninstall" :loading="uninstalling">
          <el-icon><Delete /></el-icon>
          执行彻底卸载
        </el-button>
        <el-alert
          v-if="hasImportantItems"
          title="检测到重要项目（红色标记），删除后可能影响系统稳定性"
          type="error"
          :closable="false"
          show-icon
          class="important-alert"
        />
      </div>
    </el-card>

    <ProgressBar
      v-if="uninstalling"
      :visible="uninstalling"
      title="正在执行卸载..."
      :done="uninstallProgress?.done"
      :total="uninstallProgress?.total"
      :start-time="uninstallStartTime"
      position="inline"
      @cancel="uninstalling = false"
    />

    <el-card v-if="uninstallResult" class="result-card">
      <template #header>
        <span>卸载结果</span>
      </template>
      <el-result
        v-if="uninstallResult.failed.length === 0"
        icon="success"
        title="卸载完成"
        :sub-title="`清理文件 ${uninstallResult.cleaned_files} 个，注册表 ${uninstallResult.cleaned_registry} 项，释放 ${formatSize(uninstallResult.freed)}`"
      />
      <el-result
        v-else
        icon="warning"
        :title="`卸载完成，${uninstallResult.failed.length} 项失败`"
      >
        <template #extra>
          <div class="failed-list">
            <p v-for="item in uninstallResult.failed" :key="item" class="failed-item">{{ item }}</p>
          </div>
        </template>
      </el-result>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { Monitor, Search, Delete, WarningFilled } from '@element-plus/icons-vue'
import { scanUninstallApps, analyzeUninstall, executeUninstall } from '../api/cleanup'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProgressBar from '../components/ProgressBar.vue'

const scanning = ref(false)
const analyzing = ref(false)
const uninstalling = ref(false)
const apps = ref([])
const appSearch = ref('')
const selectedApp = ref(null)
const analyzeResult = ref(null)
const uninstallResult = ref(null)
const uninstallProgress = ref(null)
const uninstallStartTime = ref(0)

let ws = null
let clientId = ''

const cleanOptions = reactive({
  programFiles: true,
  registry: true,
  config: true,
  services: false,
})

const filteredApps = computed(() => {
  if (!appSearch.value) return apps.value
  const kw = appSearch.value.toLowerCase()
  return apps.value.filter(a => a.name.toLowerCase().includes(kw) || (a.publisher && a.publisher.toLowerCase().includes(kw)))
})

const hasImportantItems = computed(() => {
  if (!analyzeResult.value) return false
  return (analyzeResult.value.important || []).some(i => i)
})

onMounted(() => {
  clientId = 'client_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9)
  connectWebSocket()
})

onUnmounted(() => {
  if (ws) {
    ws.close()
    ws = null
  }
})

const connectWebSocket = () => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/api/ws?client_id=${clientId}`
  try {
    ws = new WebSocket(wsUrl)
    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'operation_progress' && msg.payload.operation === 'uninstall') {
          uninstallProgress.value = msg.payload
        }
      } catch {}
    }
    ws.onerror = () => {}
    ws.onclose = () => { ws = null }
  } catch {}
}

const handleScan = async () => {
  scanning.value = true
  apps.value = []
  analyzeResult.value = null
  uninstallResult.value = null

  try {
    const data = await scanUninstallApps()
    apps.value = data.apps || []
    ElMessage.success(`发现 ${apps.value.length} 个已安装软件`)
  } catch (err) {
    ElMessage.error(err.message || '扫描失败')
  } finally {
    scanning.value = false
  }
}

const handleAppSelect = (row) => {
  selectedApp.value = row
}

const handleAnalyze = async (app) => {
  selectedApp.value = app
  analyzing.value = true
  analyzeResult.value = null
  uninstallResult.value = null

  try {
    const data = await analyzeUninstall({ registry_key: app.registry_key })
    analyzeResult.value = data
    ElMessage.success('分析完成')
  } catch (err) {
    ElMessage.error(err.message || '分析失败')
  } finally {
    analyzing.value = false
  }
}

const handleUninstall = async () => {
  if (!selectedApp.value || !analyzeResult.value) return

  try {
    await ElMessageBox.confirm(
      `确定要彻底卸载 "${selectedApp.value.name}" 吗？此操作不可逆，将删除所有相关文件、注册表项和配置！`,
      '确认卸载',
      { confirmButtonText: '确定卸载', cancelButtonText: '取消', type: 'error' }
    )
  } catch {
    return
  }

  uninstalling.value = true
  uninstallResult.value = null
  uninstallStartTime.value = Date.now()
  uninstallProgress.value = null

  try {
    const req = {
      registry_key: selectedApp.value.registry_key,
      clean_files: cleanOptions.programFiles ? analyzeResult.value.program_files : [],
      clean_registry: cleanOptions.registry ? analyzeResult.value.registry_items : [],
      clean_config: cleanOptions.config ? analyzeResult.value.config_files : [],
      clean_services: cleanOptions.services ? analyzeResult.value.service_names : [],
    }

    const data = await executeUninstall(req)
    uninstallResult.value = data
    ElMessage.success('卸载完成')
  } catch (err) {
    ElMessage.error(err.message || '卸载失败')
  } finally {
    uninstalling.value = false
    uninstallProgress.value = null
  }
}

const formatSize = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
</script>

<style scoped>
.uninstall-view {
  max-width: 1200px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.apps-card .card-header,
.analyze-card .card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.warning-alert {
  margin-bottom: 16px;
}

.scan-section {
  display: flex;
  justify-content: center;
  padding: 20px 0;
}

.apps-card,
.analyze-card,
.result-card {
  margin-top: 20px;
}

.analyze-sections {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.analyze-section {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 12px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.section-title {
  font-weight: 500;
  font-size: 14px;
}

.section-list {
  max-height: 200px;
  overflow-y: auto;
}

.list-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 0;
  font-size: 12px;
  color: #606266;
}

.list-item.important {
  color: #f5222d;
  font-weight: 500;
}

.more-hint {
  font-size: 12px;
  color: #909399;
  padding: 4px 0;
}

.empty-hint {
  font-size: 12px;
  color: #909399;
  padding: 8px 0;
  text-align: center;
}

.execute-section {
  margin-top: 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.important-alert {
  width: 100%;
}

.failed-list {
  text-align: left;
  max-height: 200px;
  overflow-y: auto;
}

.failed-item {
  color: #f5222d;
  font-size: 13px;
  margin: 4px 0;
}
</style>
