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
          <div class="path-row">
            <el-input v-model="searchForm.path" placeholder="输入搜索路径，如 C:\Users" clearable>
              <template #append>
                <el-button @click="showBrowseDialog = true">浏览</el-button>
              </template>
            </el-input>
            <el-select v-if="favorites.length > 0" v-model="selectedFavorite" placeholder="收藏目录" clearable @change="handleFavoriteSelect" class="favorite-select">
              <el-option v-for="fav in favorites" :key="fav" :label="fav" :value="fav" />
            </el-select>
          </div>
        </el-form-item>

        <el-form-item label="文件名">
          <el-input v-model="searchForm.pattern" placeholder="输入文件名关键词" clearable />
        </el-form-item>

        <el-form-item label="文件类型">
          <el-select v-model="searchForm.file_type" placeholder="选择文件类型" clearable>
            <el-option label="全部" value="all" />
            <el-option label="文档" value="document" />
            <el-option label="图片" value="image" />
            <el-option label="视频" value="video" />
            <el-option label="音频" value="audio" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>

        <el-form-item v-if="searchForm.file_type === 'custom'" label="自定义扩展名">
          <el-input v-model="customExtsInput" placeholder="输入扩展名，用逗号分隔，如 .txt,.pdf" />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="最小大小">
              <el-input-number v-model="searchForm.min_size" :min="0" placeholder="字节" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="最大大小">
              <el-input-number v-model="searchForm.max_size" :min="0" placeholder="字节" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="修改日期起">
              <el-date-picker v-model="searchForm.mod_after" type="date" placeholder="选择开始日期" value-format="YYYY-MM-DD" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="修改日期止">
              <el-date-picker v-model="searchForm.mod_before" type="date" placeholder="选择结束日期" value-format="YYYY-MM-DD" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item>
          <el-button type="primary" @click="handleSearch" :loading="loading">
            <el-icon><Search /></el-icon>
            开始查找
          </el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card v-if="searchProgress" class="progress-card">
      <div class="progress-content">
        <el-icon class="progress-icon is-loading"><Loading /></el-icon>
        <span class="progress-text">正在扫描... 已扫描 {{ searchProgress.scanned }} 个文件，找到 {{ searchProgress.found }} 个匹配</span>
      </div>
      <el-progress :percentage="0" :indeterminate="true" :show-text="false" />
    </el-card>

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

    <el-dialog v-model="showBrowseDialog" title="选择目录" width="600px">
      <div class="browse-dialog">
        <div class="browse-path">
          <el-input v-model="browsePath" placeholder="输入路径" @keyup.enter="loadBrowsePath">
            <template #prepend>路径</template>
            <template #append>
              <el-button @click="loadBrowsePath">前往</el-button>
            </template>
          </el-input>
        </div>
        <div class="browse-list">
          <div class="browse-item parent-item" @click="goToParent">
            <el-icon><FolderOpened /></el-icon>
            <span>.. (上级目录)</span>
          </div>
          <div v-for="item in browseItems" :key="item.path" class="browse-item" @click="selectBrowseItem(item)">
            <el-icon><Folder /></el-icon>
            <span>{{ item.name }}</span>
          </div>
          <div v-if="browseItems.length === 0" class="browse-empty">
            该目录下没有子目录
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showBrowseDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmBrowse">选择当前目录</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, nextTick, onMounted, onUnmounted } from 'vue'
import { Search, Folder, FolderOpened, Document, Picture, VideoCamera, Headset, Loading } from '@element-plus/icons-vue'
import { searchFiles, browsePath as fetchBrowsePath } from '../api/files'
import { getConfig } from '../api/config'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'

const router = useRouter()
const loading = ref(false)
const results = ref([])
const skipped = ref([])
const selectedFiles = ref([])
const customExtsInput = ref('')
const tableRef = ref(null)
const favorites = ref([])
const selectedFavorite = ref('')
const searchProgress = ref(null)

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

const showBrowseDialog = ref(false)
const browsePath = ref('')
const browseItems = ref([])
const browseCurrentPath = ref('')
const browseParentPath = ref('')

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
  clientId = 'client_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9)
  connectWebSocket()

  try {
    const config = await getConfig()
    if (config.search?.default_path) {
      searchForm.path = config.search.default_path
    }
    if (config.favorites && config.favorites.length > 0) {
      favorites.value = config.favorites
    }
  } catch {}
})

onUnmounted(() => {
  if (ws) {
    ws.close()
    ws = null
  }
})

const handleFavoriteSelect = (val) => {
  if (val) {
    searchForm.path = val
  }
}

const handleSearch = async () => {
  if (!searchForm.path) {
    ElMessage.warning('请输入搜索路径')
    return
  }
  loading.value = true
  searchProgress.value = { scanned: 0, found: 0, current_dir: '' }
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

const loadBrowsePath = async () => {
  try {
    const data = await fetchBrowsePath(browsePath.value)
    browseItems.value = data.items || []
    browseCurrentPath.value = data.current_path || ''
    browseParentPath.value = data.parent_path || ''
    browsePath.value = data.current_path || ''
  } catch (err) {
    ElMessage.error(err.message || '无法浏览该路径')
  }
}

const goToParent = () => {
  if (browseCurrentPath.value) {
    browsePath.value = browseParentPath.value
    loadBrowsePath()
  }
}

const selectBrowseItem = (item) => {
  browsePath.value = item.path
  loadBrowsePath()
}

const confirmBrowse = () => {
  searchForm.path = browseCurrentPath.value
  showBrowseDialog.value = false
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

.path-row {
  display: flex;
  gap: 8px;
  width: 100%;
}

.path-row .el-input {
  flex: 1;
}

.favorite-select {
  width: 200px;
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

.progress-card {
  margin-bottom: 20px;
}

.progress-content {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.progress-icon {
  font-size: 18px;
  color: #1890ff;
}

.progress-text {
  font-size: 14px;
  color: #606266;
}

.skipped-info {
  margin-top: 16px;
}

.browse-dialog {
  max-height: 400px;
}

.browse-path {
  margin-bottom: 12px;
}

.browse-list {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  max-height: 320px;
  overflow-y: auto;
}

.browse-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.browse-item:hover {
  background-color: #f5f7fa;
}

.browse-item .el-icon {
  color: #e6a23c;
  font-size: 18px;
}

.parent-item {
  color: #909399;
  border-bottom: 1px solid #e4e7ed;
}

.browse-empty {
  padding: 24px;
  text-align: center;
  color: #909399;
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
