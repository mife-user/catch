<template>
  <div class="move-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <el-icon><Rank /></el-icon>
          <span>{{ isCopy ? '文件复制' : '文件移动' }}</span>
          <el-switch
            v-model="isCopy"
            active-text="复制"
            inactive-text="移动"
            class="mode-switch"
          />
        </div>
      </template>

      <el-form :model="moveForm" label-width="100px" label-position="left">
        <el-form-item label="源文件路径">
          <el-input
            v-model="filesInput"
            type="textarea"
            :rows="4"
            placeholder="输入要移动/复制的文件路径，每行一个"
          />
          <div v-if="fileList.length > 0" class="file-count">
            共 {{ fileList.length }} 个文件
          </div>
        </el-form-item>

        <el-form-item label="目标路径">
          <el-input v-model="moveForm.dst_path" placeholder="输入目标目录路径" clearable>
            <template #append>
              <el-button @click="showBrowseDialog = true">浏览</el-button>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="收藏目录">
          <el-select v-model="selectedFavorite" placeholder="从收藏中选择" clearable @change="handleFavoriteSelect">
            <el-option v-for="fav in favorites" :key="fav" :label="fav" :value="fav" />
          </el-select>
          <el-button v-if="favorites.length === 0" type="primary" link @click="router.push('/settings/basic')">
            添加收藏目录
          </el-button>
        </el-form-item>

        <el-form-item label="冲突处理">
          <el-radio-group v-model="moveForm.conflict">
            <el-radio value="skip">跳过</el-radio>
            <el-radio value="rename">重命名</el-radio>
            <el-radio value="overwrite">覆盖</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-alert
          v-if="!isCopy"
          title="移动操作会将文件从原位置转移到目标位置，原位置文件将不存在"
          type="info"
          :closable="false"
          show-icon
          class="mode-alert"
        />
        <el-alert
          v-if="isCopy"
          title="复制操作会在目标位置创建文件副本，原位置文件保持不变"
          type="success"
          :closable="false"
          show-icon
          class="mode-alert"
        />

        <el-form-item>
          <el-button type="primary" @click="handleExecute" :loading="loading">
            {{ isCopy ? '执行复制' : '执行移动' }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card v-if="result" class="result-card">
      <template #header>
        <span>{{ isCopy ? '复制' : '移动' }}结果</span>
      </template>
      <el-result
        v-if="result.failed?.length === 0 && result.skipped?.length === 0"
        icon="success"
        :title="`成功${isCopy ? '复制' : '移动'} ${result.success?.length || 0} 个文件`"
      />
      <el-result
        v-else
        icon="warning"
        :title="`${isCopy ? '复制' : '移动'}完成`"
      >
        <template #extra>
          <div class="result-stats">
            <el-tag v-if="result.success?.length > 0" type="success" size="large">
              成功: {{ result.success.length }}
            </el-tag>
            <el-tag v-if="result.failed?.length > 0" type="danger" size="large">
              失败: {{ result.failed.length }}
            </el-tag>
            <el-tag v-if="result.skipped?.length > 0" type="warning" size="large">
              跳过: {{ result.skipped.length }}
            </el-tag>
          </div>
          <div v-if="result.failed?.length > 0" class="failed-list">
            <p class="failed-title">失败详情：</p>
            <p v-for="item in result.failed" :key="item" class="failed-item">{{ item }}</p>
          </div>
        </template>
      </el-result>
    </el-card>

    <el-dialog v-model="showBrowseDialog" title="选择目标目录" width="600px">
      <div class="browse-dialog">
        <div class="browse-path">
          <el-input v-model="browsePathInput" placeholder="输入路径" @keyup.enter="loadBrowsePath">
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
import { ref, reactive, computed, onMounted } from 'vue'
import { Rank, Folder, FolderOpened } from '@element-plus/icons-vue'
import { moveFiles, copyFiles, browsePath as fetchBrowsePath } from '../api/files'
import { getConfig } from '../api/config'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const result = ref(null)
const filesInput = ref('')
const favorites = ref([])
const selectedFavorite = ref('')

const isCopy = ref(route.query.mode === 'copy')

const fileList = computed(() => {
  return filesInput.value.split('\n').map(p => p.trim()).filter(p => p)
})

const moveForm = reactive({
  dst_path: '',
  conflict: 'skip',
})

const showBrowseDialog = ref(false)
const browsePathInput = ref('')
const browseItems = ref([])
const browseCurrentPath = ref('')
const browseParentPath = ref('')

onMounted(async () => {
  const stored = sessionStorage.getItem('catch_selected_files')
  if (stored) {
    try {
      const files = JSON.parse(stored)
      if (Array.isArray(files) && files.length > 0) {
        filesInput.value = files.join('\n')
      }
    } catch {}
    sessionStorage.removeItem('catch_selected_files')
  }

  try {
    const config = await getConfig()
    favorites.value = config.favorites || []
  } catch {}
})

const handleFavoriteSelect = (val) => {
  if (val) {
    moveForm.dst_path = val
  }
}

const loadBrowsePath = async () => {
  try {
    const data = await fetchBrowsePath(browsePathInput.value)
    browseItems.value = data.items || []
    browseCurrentPath.value = data.current_path || ''
    browseParentPath.value = data.parent_path || ''
    browsePathInput.value = data.current_path || ''
  } catch (err) {
    ElMessage.error(err.message || '无法浏览该路径')
  }
}

const goToParent = () => {
  if (browseCurrentPath.value) {
    browsePathInput.value = browseParentPath.value
    loadBrowsePath()
  }
}

const selectBrowseItem = (item) => {
  browsePathInput.value = item.path
  loadBrowsePath()
}

const confirmBrowse = () => {
  moveForm.dst_path = browseCurrentPath.value
  showBrowseDialog.value = false
}

const handleExecute = async () => {
  const paths = fileList.value
  if (paths.length === 0) {
    ElMessage.warning('请输入文件路径')
    return
  }
  if (!moveForm.dst_path) {
    ElMessage.warning('请输入目标路径')
    return
  }

  const opText = isCopy.value ? '复制' : '移动'
  const conflictText = moveForm.conflict === 'skip' ? '跳过'
    : moveForm.conflict === 'rename' ? '自动重命名'
    : '覆盖'

  try {
    await ElMessageBox.confirm(
      `确定要将 ${paths.length} 个文件${opText}到 "${moveForm.dst_path}" 吗？冲突处理：${conflictText}`,
      `确认${opText}`,
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
  } catch {
    return
  }

  loading.value = true
  result.value = null
  try {
    let data
    if (isCopy.value) {
      data = await copyFiles({
        src_paths: paths,
        dst_path: moveForm.dst_path,
        conflict: moveForm.conflict,
      })
    } else {
      data = await moveFiles({
        src_paths: paths,
        dst_path: moveForm.dst_path,
        conflict: moveForm.conflict,
      })
    }
    result.value = data
    ElMessage.success(`${opText}完成`)
  } catch (err) {
    ElMessage.error(err.message || '操作失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.move-view {
  max-width: 800px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.mode-switch {
  margin-left: auto;
}

.file-count {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}

.mode-alert {
  margin-bottom: 16px;
}

.result-card {
  margin-top: 20px;
}

.result-stats {
  display: flex;
  gap: 12px;
  justify-content: center;
  margin-bottom: 12px;
}

.failed-list {
  text-align: left;
  max-height: 200px;
  overflow-y: auto;
}

.failed-title {
  font-weight: 600;
  color: #f5222d;
  margin-bottom: 4px;
}

.failed-item {
  color: #f5222d;
  font-size: 13px;
  margin: 2px 0;
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
</style>
