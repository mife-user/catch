<template>
  <div class="search-view">
    <el-card class="search-card">
      <template #header>
        <div class="card-header">
          <el-icon><Search /></el-icon>
          <span>文件查找</span>
        </div>
      </template>

      <el-form :model="searchForm" label-width="100px" label-position="left">
        <el-form-item label="搜索路径">
          <UnifiedSelect
            v-model="searchForm.path"
            mode="path"
            placeholder="选择搜索路径"
            dialog-title="选择搜索路径"
            :prefix-icon="Folder"
          />
        </el-form-item>

        <el-form-item label="文件名">
          <el-input v-model="searchForm.pattern" placeholder="输入文件名关键词" clearable />
        </el-form-item>

        <el-form-item label="文件类型">
          <UnifiedSelect
            v-model="searchForm.file_type"
            mode="fileType"
            placeholder="选择文件类型"
            dialog-title="选择文件类型"
          />
        </el-form-item>

        <el-form-item v-if="searchForm.file_type === 'custom'" label="自定义扩展名">
          <el-input v-model="customExtsInput" placeholder="输入扩展名，用逗号分隔，如 .txt,.pdf" />
        </el-form-item>

        <el-form-item label="文件大小">
          <UnifiedSelect
            mode="sizeRange"
            :min-size="searchForm.min_size"
            :max-size="searchForm.max_size"
            placeholder="选择文件大小范围"
            dialog-title="选择文件大小范围"
            @update:min-size="searchForm.min_size = $event"
            @update:max-size="searchForm.max_size = $event"
          />
        </el-form-item>

        <el-form-item label="修改日期">
          <UnifiedSelect
            mode="dateRange"
            :mod-after="searchForm.mod_after"
            :mod-before="searchForm.mod_before"
            placeholder="选择日期范围"
            dialog-title="选择修改日期范围"
            @update:mod-after="searchForm.mod_after = $event"
            @update:mod-before="searchForm.mod_before = $event"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSearch" :loading="loading">
            <el-icon><Search /></el-icon>
            开始查找
          </el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <ProgressBar
      v-if="loading && searchProgress"
      :visible="true"
      title="正在扫描..."
      :done="searchProgress.scanned"
      :total="searchProgress.scanned + searchProgress.found"
      :current-item="searchProgress.current_dir"
      :start-time="searchStartTime"
      position="inline"
      :show-pause="false"
      @cancel="loading = false"
    />

    <el-card v-if="results.length > 0" class="results-card">
      <template #header>
        <div class="card-header">
          <span>查找结果 ({{ results.length }} 个文件)</span>
          <div class="actions">
            <el-button size="small" @click="selectAll">全选</el-button>
            <el-button size="small" @click="invertSelection">反选</el-button>
            <el-button size="small" type="danger" @click="handleBatchDelete">删除选中</el-button>
            <el-button size="small" @click="handleBatchMove">移动选中</el-button>
            <el-button size="small" @click="handleBatchCopy">复制选中</el-button>
            <el-button size="small" @click="handleBatchRename">重命名选中</el-button>
          </div>
        </div>
      </template>

      <el-table ref="tableRef" :data="results" @selection-change="handleSelectionChange" stripe>
        <el-table-column type="selection" width="55" />
        <el-table-column label="类型" width="50" align="center">
          <template #default="{ row }">
            <el-icon :size="20" :color="getFileIconColor(row.extension)">
              <component :is="getFileIcon(row.extension)" />
            </el-icon>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="文件名" min-width="200" show-overflow-tooltip />
        <el-table-column prop="path" label="路径" min-width="300" show-overflow-tooltip />
        <el-table-column prop="size" label="大小" width="120">
          <template #default="{ row }">
            {{ formatSize(row.size) }}
          </template>
        </el-table-column>
        <el-table-column prop="mod_time" label="修改时间" width="180" />
        <el-table-column prop="extension" label="扩展名" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="getExtensionTagType(row.extension)">{{ row.extension || '无' }}</el-tag>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="skipped.length > 0" class="skipped-info">
        <el-alert :title="`已跳过 ${skipped.length} 个无权限文件`" type="warning" :closable="false" show-icon />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { Search, Folder, Document, Picture, VideoCamera, Headset } from '@element-plus/icons-vue'
import { searchFiles } from '../api/files'
import { getConfig } from '../api/config'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import UnifiedSelect from '../components/UnifiedSelect.vue'
import ProgressBar from '../components/ProgressBar.vue'

const router = useRouter()
const loading = ref(false)
const results = ref([])
const skipped = ref([])
const selectedFiles = ref([])
const customExtsInput = ref('')
const tableRef = ref(null)
const searchProgress = ref(null)
const searchStartTime = ref(0)

let ws = null
let clientId = ''

const searchForm = reactive({
  path: '',
  pattern: '',
  file_type: 'all',
  min_size: 0,
  max_size: 0,
  mod_after: '',
  mod_before: '',
})

const documentExts = ['.txt', '.doc', '.docx', '.pdf', '.xls', '.xlsx', '.ppt', '.pptx', '.csv', '.md', '.rtf']
const imageExts = ['.jpg', '.jpeg', '.png', '.gif', '.bmp', '.svg', '.webp', '.ico', '.tiff']
const videoExts = ['.mp4', '.avi', '.mkv', '.mov', '.wmv', '.flv', '.webm']
const audioExts = ['.mp3', '.wav', '.flac', '.aac', '.ogg', '.wma', '.m4a']

const getFileIcon = (ext) => {
  if (!ext) return Document
  const lower = ext.toLowerCase()
  if (documentExts.includes(lower)) return Document
  if (imageExts.includes(lower)) return Picture
  if (videoExts.includes(lower)) return VideoCamera
  if (audioExts.includes(lower)) return Headset
  return Document
}

const getFileIconColor = (ext) => {
  if (!ext) return '#909399'
  const lower = ext.toLowerCase()
  if (documentExts.includes(lower)) return '#1890ff'
  if (imageExts.includes(lower)) return '#52c41a'
  if (videoExts.includes(lower)) return '#faad14'
  if (audioExts.includes(lower)) return '#722ed1'
  return '#909399'
}

const getExtensionTagType = (ext) => {
  if (!ext) return 'info'
  const lower = ext.toLowerCase()
  if (documentExts.includes(lower)) return 'primary'
  if (imageExts.includes(lower)) return 'success'
  if (videoExts.includes(lower)) return 'warning'
  if (audioExts.includes(lower)) return 'danger'
  return 'info'
}

const connectWebSocket = () => {
  clientId = 'client_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9)
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/api/ws?client_id=${clientId}`
  try {
    ws = new WebSocket(wsUrl)
    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'search_progress') {
          searchProgress.value = msg.payload
        }
      } catch {}
    }
    ws.onerror = () => {}
    ws.onclose = () => { ws = null }
  } catch {}
}

onMounted(async () => {
  connectWebSocket()

  try {
    const config = await getConfig()
    if (config.search?.default_path) {
      searchForm.path = config.search.default_path
    }
  } catch {}
})

onUnmounted(() => {
  if (ws) {
    ws.close()
    ws = null
  }
})

const handleSearch = async () => {
  if (!searchForm.path) {
    ElMessage.warning('请选择搜索路径')
    return
  }
  loading.value = true
  searchProgress.value = { scanned: 0, found: 0, current_dir: '' }
  searchStartTime.value = Date.now()
  try {
    const params = { ...searchForm }
    if (searchForm.file_type === 'custom' && customExtsInput.value) {
      params.custom_exts = customExtsInput.value.split(',').map(e => e.trim())
    }
    if (!params.min_size) delete params.min_size
    if (!params.max_size) delete params.max_size

    if (clientId) {
      params.client_id = clientId
    }

    const data = await searchFiles(params)
    results.value = data.files || []
    skipped.value = data.skipped || []
    ElMessage.success(`找到 ${data.total || 0} 个文件`)
  } catch (err) {
    ElMessage.error(err.message || '查找失败')
  } finally {
    loading.value = false
    searchProgress.value = null
  }
}

const resetForm = () => {
  Object.assign(searchForm, {
    path: '',
    pattern: '',
    file_type: 'all',
    min_size: 0,
    max_size: 0,
    mod_after: '',
    mod_before: '',
  })
  customExtsInput.value = ''
  results.value = []
  skipped.value = []
}

const handleSelectionChange = (selection) => {
  selectedFiles.value = selection
}

const selectAll = () => {
  if (!tableRef.value) return
  results.value.forEach(row => {
    tableRef.value.toggleRowSelection(row, true)
  })
}

const invertSelection = () => {
  if (!tableRef.value) return
  results.value.forEach(row => {
    tableRef.value.toggleRowSelection(row, !selectedFiles.value.includes(row))
  })
}

const handleBatchDelete = () => {
  if (selectedFiles.value.length === 0) {
    ElMessage.warning('请先选择文件')
    return
  }
  sessionStorage.setItem('catch_selected_files', JSON.stringify(selectedFiles.value.map(f => f.path)))
  router.push('/delete')
}

const handleBatchMove = () => {
  if (selectedFiles.value.length === 0) {
    ElMessage.warning('请先选择文件')
    return
  }
  sessionStorage.setItem('catch_selected_files', JSON.stringify(selectedFiles.value.map(f => f.path)))
  router.push({ path: '/move', query: { mode: 'move' } })
}

const handleBatchCopy = () => {
  if (selectedFiles.value.length === 0) {
    ElMessage.warning('请先选择文件')
    return
  }
  sessionStorage.setItem('catch_selected_files', JSON.stringify(selectedFiles.value.map(f => f.path)))
  router.push({ path: '/move', query: { mode: 'copy' } })
}

const handleBatchRename = () => {
  if (selectedFiles.value.length === 0) {
    ElMessage.warning('请先选择文件')
    return
  }
  sessionStorage.setItem('catch_selected_files', JSON.stringify(selectedFiles.value.map(f => f.path)))
  router.push('/rename')
}

const formatSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
</script>

<style scoped>
.search-view {
  max-width: 1200px;
}

.search-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.results-card .card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.actions {
  display: flex;
  gap: 8px;
}

.skipped-info {
  margin-top: 16px;
}

@media (max-width: 768px) {
  .actions {
    flex-wrap: wrap;
  }

  .results-card .card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
}
</style>
