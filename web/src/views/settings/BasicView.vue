<template>
  <div class="basic-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <el-icon><Setting /></el-icon>
          <span>基础设置</span>
        </div>
      </template>

      <el-form :model="form" label-width="120px" label-position="left" v-loading="loading">
        <el-form-item label="服务端口">
          <el-input-number v-model="form.port" :min="3000" :max="3100" @change="handlePortChange" />
          <el-alert v-if="portChanged" title="端口修改需要重启服务才能生效" type="warning" :closable="false" show-icon class="inline-alert" />
        </el-form-item>

        <el-form-item label="过期天数">
          <el-input-number v-model="form.expire_days" :min="1" :max="365" />
          <span class="form-hint">直接删除文件的保留天数</span>
        </el-form-item>

        <el-form-item label="回收站路径">
          <el-input v-model="form.trash_path" placeholder="默认 ~/.catch-trash" clearable>
            <template #append>
              <el-button @click="openBrowseDialog('trash')">浏览</el-button>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="默认搜索路径">
          <el-input v-model="form.default_path" placeholder="默认 ~" clearable>
            <template #append>
              <el-button @click="openBrowseDialog('search')">浏览</el-button>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="收藏目录">
          <div class="favorites-list">
            <div v-for="(fav, index) in form.favorites" :key="index" class="favorite-item">
              <el-input v-model="form.favorites[index]" />
              <el-button type="danger" size="small" @click="removeFavorite(index)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button type="primary" size="small" @click="addFavorite">添加收藏目录</el-button>
          </div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSave" :loading="saving">保存设置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-dialog v-model="showBrowseDialog" title="选择目录" width="600px">
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
import { ref, reactive, onMounted } from 'vue'
import { Setting, Delete, Folder, FolderOpened } from '@element-plus/icons-vue'
import { getConfig, updateConfig } from '../../api/config'
import { browsePath as fetchBrowsePath } from '../../api/files'
import { ElMessage } from 'element-plus'

const loading = ref(false)
const saving = ref(false)
const portChanged = ref(false)
const originalPort = ref(3000)

const form = reactive({
  port: 3000,
  expire_days: 7,
  trash_path: '~/.catch-trash',
  default_path: '~',
  favorites: [],
})

const showBrowseDialog = ref(false)
const browsePathInput = ref('')
const browseItems = ref([])
const browseCurrentPath = ref('')
const browseParentPath = ref('')
const browseTarget = ref('')

onMounted(async () => {
  loading.value = true
  try {
    const config = await getConfig()
    form.port = config.server?.port || 3000
    originalPort.value = form.port
    form.expire_days = config.trash?.expire_days || 7
    form.trash_path = config.trash?.path || '~/.catch-trash'
    form.default_path = config.search?.default_path || '~'
    form.favorites = config.favorites || []
  } catch (err) {
    ElMessage.error('加载配置失败')
  } finally {
    loading.value = false
  }
})

const handlePortChange = () => {
  portChanged.value = form.port !== originalPort.value
}

const addFavorite = () => {
  form.favorites.push('')
}

const removeFavorite = (index) => {
  form.favorites.splice(index, 1)
}

const openBrowseDialog = (target) => {
  browseTarget.value = target
  browsePathInput.value = target === 'trash' ? form.trash_path : form.default_path
  showBrowseDialog.value = true
  loadBrowsePath()
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
  if (browseTarget.value === 'trash') {
    form.trash_path = browseCurrentPath.value
  } else {
    form.default_path = browseCurrentPath.value
  }
  showBrowseDialog.value = false
}

const handleSave = async () => {
  saving.value = true
  try {
    await updateConfig({
      server: { port: form.port },
      trash: { expire_days: form.expire_days, path: form.trash_path },
      search: { default_path: form.default_path },
      favorites: form.favorites.filter(f => f.trim()),
    })
    originalPort.value = form.port
    portChanged.value = false
    ElMessage.success('设置已保存')
  } catch (err) {
    ElMessage.error(err.message || '保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.basic-view {
  max-width: 800px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.favorites-list {
  width: 100%;
}

.favorite-item {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

.favorite-item .el-input {
  flex: 1;
}

.inline-alert {
  margin-left: 12px;
  display: inline-flex;
  width: auto;
}

.form-hint {
  margin-left: 12px;
  font-size: 12px;
  color: #909399;
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
