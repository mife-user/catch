<template>
  <div class="trash-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <el-icon><Delete /></el-icon>
            <span>Catch 回收站</span>
            <el-tag v-if="items.length > 0" type="info" size="small">{{ items.length }} 个文件</el-tag>
          </div>
          <div class="header-right">
            <el-button type="danger" size="small" @click="handleCleanExpired" :loading="cleaning">
              清理过期文件
            </el-button>
            <el-button size="small" @click="loadTrashList" :loading="loading">
              刷新
            </el-button>
          </div>
        </div>
      </template>

      <div v-if="loading" class="loading-container">
        <el-skeleton :rows="5" animated />
      </div>

      <div v-else-if="items.length === 0" class="empty-container">
        <el-empty description="回收站为空" />
      </div>

      <div v-else>
        <el-table ref="tableRef" :data="items" @selection-change="handleSelectionChange" stripe>
          <el-table-column type="selection" width="55" />
          <el-table-column prop="file_name" label="文件名" min-width="180" show-overflow-tooltip />
          <el-table-column prop="original_path" label="原路径" min-width="300" show-overflow-tooltip />
          <el-table-column prop="file_size" label="大小" width="100">
            <template #default="{ row }">
              {{ formatSize(row.file_size) }}
            </template>
          </el-table-column>
          <el-table-column prop="deleted_at" label="删除时间" width="170" />
          <el-table-column prop="expires_at" label="过期时间" width="170" />
          <el-table-column prop="remaining_days" label="剩余天数" width="100">
            <template #default="{ row }">
              <el-tag :type="row.remaining_days <= 1 ? 'danger' : row.remaining_days <= 3 ? 'warning' : 'success'" size="small">
                {{ row.remaining_days }} 天
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.is_expired ? 'danger' : 'success'" size="small">
                {{ row.is_expired ? '已过期' : '有效' }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>

        <div class="batch-actions">
          <el-button type="primary" @click="handleBatchRestore" :disabled="selectedItems.length === 0">
            恢复选中 ({{ selectedItems.length }})
          </el-button>
        </div>
      </div>
    </el-card>

    <el-card v-if="cleanResult" class="result-card">
      <template #header>
        <span>清理结果</span>
      </template>
      <el-result
        :icon="cleanResult.cleaned > 0 ? 'success' : 'info'"
        :title="cleanResult.cleaned > 0 ? `已清理 ${cleanResult.cleaned} 个过期文件` : '没有需要清理的过期文件'"
      />
    </el-card>

    <el-card v-if="restoreResult" class="result-card">
      <template #header>
        <span>恢复结果</span>
      </template>
      <el-result
        v-if="restoreResult.failed.length === 0"
        icon="success"
        :title="`成功恢复 ${restoreResult.success.length} 个文件`"
      />
      <el-result
        v-else
        icon="warning"
        :title="`成功 ${restoreResult.success.length} 个，失败 ${restoreResult.failed.length} 个`"
      >
        <template #extra>
          <div class="failed-list">
            <p v-for="item in restoreResult.failed" :key="item" class="failed-item">{{ item }}</p>
          </div>
        </template>
      </el-result>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Delete } from '@element-plus/icons-vue'
import { getTrashList, restoreFiles, cleanExpired } from '../api/trash'
import { ElMessage, ElMessageBox } from 'element-plus'

const loading = ref(false)
const cleaning = ref(false)
const items = ref([])
const selectedItems = ref([])
const cleanResult = ref(null)
const restoreResult = ref(null)
const tableRef = ref(null)

onMounted(() => {
  loadTrashList()
})

const loadTrashList = async () => {
  loading.value = true
  try {
    const data = await getTrashList()
    items.value = data.items || []
  } catch (err) {
    ElMessage.error(err.message || '获取回收站列表失败')
  } finally {
    loading.value = false
  }
}

const handleSelectionChange = (selection) => {
  selectedItems.value = selection
}

const handleBatchRestore = async () => {
  if (selectedItems.value.length === 0) {
    ElMessage.warning('请先选择要恢复的文件')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要恢复 ${selectedItems.value.length} 个文件吗？`,
      '确认恢复',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' }
    )
  } catch {
    return
  }

  try {
    const paths = selectedItems.value.map(item => item.original_path)
    const data = await restoreFiles({ original_paths: paths })
    restoreResult.value = data
    cleanResult.value = null
    ElMessage.success(`恢复完成：成功 ${data.success?.length || 0} 个`)
    loadTrashList()
  } catch (err) {
    ElMessage.error(err.message || '恢复失败')
  }
}

const handleCleanExpired = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清理所有过期文件吗？过期文件将被永久删除。',
      '确认清理',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
  } catch {
    return
  }

  cleaning.value = true
  try {
    const data = await cleanExpired()
    cleanResult.value = data
    restoreResult.value = null
    ElMessage.success(`已清理 ${data.cleaned || 0} 个过期文件`)
    loadTrashList()
  } catch (err) {
    ElMessage.error(err.message || '清理失败')
  } finally {
    cleaning.value = false
  }
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
.trash-view {
  max-width: 1200px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.header-right {
  display: flex;
  gap: 8px;
}

.loading-container {
  padding: 20px;
}

.empty-container {
  padding: 40px 0;
}

.batch-actions {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
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
</style>
