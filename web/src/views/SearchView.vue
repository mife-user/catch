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
          <el-input v-model="searchForm.path" placeholder="输入搜索路径，如 C:\Users" clearable>
            <template #append>
              <el-button @click="selectPath">浏览</el-button>
            </template>
          </el-input>
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

      <el-table :data="results" @selection-change="handleSelectionChange" stripe>
        <el-table-column type="selection" width="55" />
        <el-table-column prop="name" label="文件名" min-width="200" show-overflow-tooltip />
        <el-table-column prop="path" label="路径" min-width="300" show-overflow-tooltip />
        <el-table-column prop="size" label="大小" width="120">
          <template #default="{ row }">
            {{ formatSize(row.size) }}
          </template>
        </el-table-column>
        <el-table-column prop="mod_time" label="修改时间" width="180" />
        <el-table-column prop="extension" label="扩展名" width="100" />
      </el-table>

      <div v-if="skipped.length > 0" class="skipped-info">
        <el-alert :title="`已跳过 ${skipped.length} 个无权限文件`" type="warning" :closable="false" show-icon />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { searchFiles, deleteFiles, moveFiles, copyFiles } from '../api/files'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'

const router = useRouter()
const loading = ref(false)
const results = ref([])
const skipped = ref([])
const selectedFiles = ref([])
const customExtsInput = ref('')

const searchForm = reactive({
  path: '',
  pattern: '',
  file_type: 'all',
  min_size: 0,
  max_size: 0,
  mod_after: '',
  mod_before: '',
})

const handleSearch = async () => {
  loading.value = true
  try {
    const params = { ...searchForm }
    if (searchForm.file_type === 'custom' && customExtsInput.value) {
      params.custom_exts = customExtsInput.value.split(',').map(e => e.trim())
    }
    if (!params.min_size) delete params.min_size
    if (!params.max_size) delete params.max_size

    const data = await searchFiles(params)
    results.value = data.files || []
    skipped.value = data.skipped || []
    ElMessage.success(`找到 ${data.total || 0} 个文件`)
  } catch (err) {
    ElMessage.error(err.message || '查找失败')
  } finally {
    loading.value = false
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
  // el-table handles this via checkbox
}

const invertSelection = () => {
  // This is handled by the table component
}

const handleBatchDelete = () => {
  if (selectedFiles.value.length === 0) {
    ElMessage.warning('请先选择文件')
    return
  }
  router.push({ path: '/delete', query: { files: selectedFiles.value.map(f => f.path) } })
}

const handleBatchMove = () => {
  if (selectedFiles.value.length === 0) {
    ElMessage.warning('请先选择文件')
    return
  }
  router.push({ path: '/move', query: { files: selectedFiles.value.map(f => f.path), mode: 'move' } })
}

const handleBatchCopy = () => {
  if (selectedFiles.value.length === 0) {
    ElMessage.warning('请先选择文件')
    return
  }
  router.push({ path: '/move', query: { files: selectedFiles.value.map(f => f.path), mode: 'copy' } })
}

const handleBatchRename = () => {
  if (selectedFiles.value.length === 0) {
    ElMessage.warning('请先选择文件')
    return
  }
  router.push({ path: '/rename', query: { files: selectedFiles.value.map(f => f.path) } })
}

const selectPath = () => {
  ElMessage.info('请直接在输入框中输入路径')
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
</style>
