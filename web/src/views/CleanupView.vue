<template>
  <div class="cleanup-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <el-icon><Brush /></el-icon>
          <span>文件清理</span>
        </div>
      </template>

      <el-form :model="cleanupForm" label-width="100px" label-position="left">
        <el-form-item label="扫描路径">
          <UnifiedSelect
            v-model="cleanupForm.path"
            mode="path"
            placeholder="选择扫描路径"
            dialog-title="选择扫描路径"
            prefix-icon="Folder"
          />
        </el-form-item>

        <el-form-item label="清理规则">
          <div class="rules-grid">
            <div
              v-for="rule in rules"
              :key="rule.id"
              class="rule-card"
              :class="{
                selected: selectedRules.includes(rule.id),
                important: rule.important
              }"
              @click="toggleRule(rule.id)"
            >
              <div class="rule-header">
                <el-checkbox
                  :model-value="selectedRules.includes(rule.id)"
                  @change="toggleRule(rule.id)"
                  @click.stop
                />
                <span class="rule-name">{{ rule.name }}</span>
                <el-tag v-if="rule.important" type="danger" size="small">重要</el-tag>
              </div>
              <div class="rule-desc">{{ rule.description }}</div>
              <div v-if="rule.file_types && rule.file_types.length > 0" class="rule-types">
                <el-tag v-for="ft in rule.file_types.slice(0, 4)" :key="ft" size="small" type="info" class="type-tag">{{ ft }}</el-tag>
                <el-tag v-if="rule.file_types.length > 4" size="small" type="info">+{{ rule.file_types.length - 4 }}</el-tag>
              </div>
            </div>
          </div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleScan" :loading="scanning">
            <el-icon><Search /></el-icon>
            开始扫描
          </el-button>
          <el-button @click="selectAllRules">全选规则</el-button>
          <el-button @click="clearRules">清空选择</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <ProgressBar
      v-if="scanning"
      :visible="scanning"
      title="正在扫描..."
      :done="scanProgress?.done"
      :total="scanProgress?.total"
      :start-time="scanStartTime"
      position="inline"
      @cancel="scanning = false"
    />

    <el-card v-if="scanResult" class="result-card">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span>扫描结果</span>
            <el-tag type="info">{{ scanResult.total }} 个文件</el-tag>
            <el-tag type="warning">共 {{ formatSize(scanResult.total_size) }}</el-tag>
          </div>
          <div class="header-right">
            <el-button size="small" @click="selectAll">全选</el-button>
            <el-button size="small" @click="invertSelection">反选</el-button>
            <el-button size="small" @click="selectNonImportant">选择非重要</el-button>
          </div>
        </div>
      </template>

      <el-table ref="tableRef" :data="scanResult.items" @selection-change="handleSelectionChange" stripe>
        <el-table-column type="selection" width="55" />
        <el-table-column label="重要" width="60" align="center">
          <template #default="{ row }">
            <el-icon v-if="row.important" color="#f5222d"><WarningFilled /></el-icon>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="文件名" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <span :class="{ 'important-file': row.important }">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路径" min-width="300" show-overflow-tooltip />
        <el-table-column prop="size" label="大小" width="120">
          <template #default="{ row }">
            {{ formatSize(row.size) }}
          </template>
        </el-table-column>
        <el-table-column prop="rule_name" label="规则" width="120">
          <template #default="{ row }">
            <el-tag size="small" :type="row.important ? 'danger' : 'info'">{{ row.rule_name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="mod_time" label="修改时间" width="170" />
      </el-table>

      <div class="batch-actions">
        <el-button
          type="danger"
          @click="handleCleanup"
          :loading="cleaning"
          :disabled="selectedFiles.length === 0"
        >
          <el-icon><Delete /></el-icon>
          清理选中 ({{ selectedFiles.length }})
        </el-button>
        <span v-if="selectedFiles.length > 0" class="selected-info">
          已选 {{ selectedFiles.length }} 个文件
        </span>
      </div>
    </el-card>

    <ProgressBar
      v-if="cleaning"
      :visible="cleaning"
      title="正在清理..."
      :done="cleanProgress?.done"
      :total="cleanProgress?.total"
      :start-time="cleanStartTime"
      position="inline"
      @cancel="cleaning = false"
    />

    <el-card v-if="cleanResult" class="result-card">
      <template #header>
        <span>清理结果</span>
      </template>
      <el-result
        v-if="cleanResult.failed.length === 0"
        icon="success"
        :title="`成功清理 ${cleanResult.cleaned} 个文件`"
        :sub-title="`释放空间: ${formatSize(cleanResult.freed)}`"
      />
      <el-result
        v-else
        icon="warning"
        :title="`清理完成，${cleanResult.failed.length} 个失败`"
        :sub-title="`成功 ${cleanResult.cleaned} 个，释放 ${formatSize(cleanResult.freed)}`"
      >
        <template #extra>
          <div class="failed-list">
            <p v-for="item in cleanResult.failed" :key="item" class="failed-item">{{ item }}</p>
          </div>
        </template>
      </el-result>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { Brush, Search, Delete, WarningFilled } from '@element-plus/icons-vue'
import { getCleanupRules, scanCleanup, executeCleanup } from '../api/cleanup'
import { getConfig } from '../api/config'
import { ElMessage, ElMessageBox } from 'element-plus'
import UnifiedSelect from '../components/UnifiedSelect.vue'
import ProgressBar from '../components/ProgressBar.vue'

const rules = ref([])
const selectedRules = ref([])
const scanning = ref(false)
const cleaning = ref(false)
const scanResult = ref(null)
const cleanResult = ref(null)
const selectedFiles = ref([])
const tableRef = ref(null)
const scanProgress = ref(null)
const cleanProgress = ref(null)
const scanStartTime = ref(0)
const cleanStartTime = ref(0)

let ws = null
let clientId = ''

const cleanupForm = reactive({
  path: '',
})

onMounted(async () => {
  clientId = 'client_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9)
  connectWebSocket()

  try {
    const data = await getCleanupRules()
    rules.value = data.rules || []
    selectedRules.value = rules.value.filter(r => !r.important).map(r => r.id)
  } catch {}

  try {
    const config = await getConfig()
    if (config.search?.default_path) {
      cleanupForm.path = config.search.default_path
    }
  } catch {}
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
        if (msg.type === 'operation_progress') {
          if (msg.payload.operation === 'cleanup_scan') {
            scanProgress.value = msg.payload
          } else if (msg.payload.operation === 'cleanup_execute') {
            cleanProgress.value = msg.payload
          }
        }
      } catch {}
    }
    ws.onerror = () => {}
    ws.onclose = () => { ws = null }
  } catch {}
}

const toggleRule = (id) => {
  const idx = selectedRules.value.indexOf(id)
  if (idx >= 0) {
    selectedRules.value.splice(idx, 1)
  } else {
    selectedRules.value.push(id)
  }
}

const selectAllRules = () => {
  selectedRules.value = rules.value.map(r => r.id)
}

const clearRules = () => {
  selectedRules.value = []
}

const handleScan = async () => {
  if (!cleanupForm.path) {
    ElMessage.warning('请选择扫描路径')
    return
  }
  if (selectedRules.value.length === 0) {
    ElMessage.warning('请至少选择一个清理规则')
    return
  }

  scanning.value = true
  scanResult.value = null
  cleanResult.value = null
  scanStartTime.value = Date.now()
  scanProgress.value = null

  try {
    const data = await scanCleanup({
      path: cleanupForm.path,
      rules: selectedRules.value,
    }, clientId)
    scanResult.value = data
    if (data.total === 0) {
      ElMessage.success('未发现可清理的文件')
    } else {
      ElMessage.success(`发现 ${data.total} 个可清理文件，共 ${formatSize(data.total_size)}`)
    }
  } catch (err) {
    ElMessage.error(err.message || '扫描失败')
  } finally {
    scanning.value = false
    scanProgress.value = null
  }
}

const handleSelectionChange = (selection) => {
  selectedFiles.value = selection
}

const selectAll = () => {
  if (!tableRef.value) return
  scanResult.value.items.forEach(row => {
    tableRef.value.toggleRowSelection(row, true)
  })
}

const invertSelection = () => {
  if (!tableRef.value) return
  scanResult.value.items.forEach(row => {
    tableRef.value.toggleRowSelection(row, !selectedFiles.value.includes(row))
  })
}

const selectNonImportant = () => {
  if (!tableRef.value) return
  scanResult.value.items.forEach(row => {
    tableRef.value.toggleRowSelection(row, !row.important)
  })
}

const handleCleanup = async () => {
  if (selectedFiles.value.length === 0) {
    ElMessage.warning('请选择要清理的文件')
    return
  }

  const hasImportant = selectedFiles.value.some(f => f.important)
  const confirmMsg = hasImportant
    ? `选中的文件中包含重要文件（红色标记），删除后可能影响相关软件正常运行！确定要清理 ${selectedFiles.value.length} 个文件吗？`
    : `确定要清理 ${selectedFiles.value.length} 个文件吗？`

  try {
    await ElMessageBox.confirm(confirmMsg, '确认清理', {
      confirmButtonText: '确定清理',
      cancelButtonText: '取消',
      type: hasImportant ? 'error' : 'warning',
    })
  } catch {
    return
  }

  cleaning.value = true
  cleanResult.value = null
  cleanStartTime.value = Date.now()
  cleanProgress.value = null

  try {
    const paths = selectedFiles.value.map(f => f.path)
    const headers = {}
    if (clientId) headers['X-Client-ID'] = clientId

    const data = await executeCleanup({ paths })
    cleanResult.value = data
    ElMessage.success(`清理完成：成功 ${data.cleaned} 个，释放 ${formatSize(data.freed)}`)
  } catch (err) {
    ElMessage.error(err.message || '清理失败')
  } finally {
    cleaning.value = false
    cleanProgress.value = null
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
.cleanup-view {
  max-width: 1200px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.result-card .card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-right {
  display: flex;
  gap: 8px;
}

.rules-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  width: 100%;
}

.rule-card {
  padding: 12px;
  border: 2px solid #e4e7ed;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.rule-card:hover {
  border-color: #1890ff;
}

.rule-card.selected {
  border-color: #1890ff;
  background-color: #ecf5ff;
}

.rule-card.important {
  border-color: #f5222d33;
}

.rule-card.important.selected {
  border-color: #f5222d;
  background-color: #fff1f0;
}

.rule-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.rule-name {
  font-weight: 500;
  font-size: 14px;
}

.rule-desc {
  font-size: 12px;
  color: #909399;
  margin-bottom: 6px;
}

.rule-types {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.type-tag {
  font-size: 11px;
}

.important-file {
  color: #f5222d;
  font-weight: 500;
}

.batch-actions {
  margin-top: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.selected-info {
  font-size: 13px;
  color: #909399;
}

.result-card {
  margin-top: 20px;
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

@media (max-width: 768px) {
  .rules-grid {
    grid-template-columns: 1fr;
  }

  .header-right {
    flex-wrap: wrap;
  }
}
</style>
