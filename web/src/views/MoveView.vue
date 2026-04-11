<template>
  <div class="move-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <el-icon><Rank /></el-icon>
          <span>{{ isCopy ? '文件复制' : '文件移动' }}</span>
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
      <div v-if="result.success?.length > 0">
        <p class="success-text">成功: {{ result.success.length }} 个文件</p>
      </div>
      <div v-if="result.failed?.length > 0">
        <p class="fail-text">失败: {{ result.failed.length }} 个</p>
        <p v-for="item in result.failed" :key="item" class="fail-item">{{ item }}</p>
      </div>
      <div v-if="result.skipped?.length > 0">
        <p class="skip-text">跳过: {{ result.skipped.length }} 个</p>
      </div>
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
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const result = ref(null)
const filesInput = ref('')
const favorites = ref([])
const selectedFavorite = ref('')

const isCopy = computed(() => route.query.mode === 'copy')

const moveForm = reactive({
  dst_path: '',
  conflict: 'skip',
})

const showBrowseDialog = ref(false)
const browsePathInput = ref('')
const browseItems = ref([])
const browseCurrentPath = ref('')

onMounted(async () => {
  if (route.query.files) {
    const files = Array.isArray(route.query.files) ? route.query.files : [route.query.files]
    filesInput.value = files.join('\n')
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
    browsePathInput.value = data.current_path || ''
  } catch (err) {
    ElMessage.error(err.message || '无法浏览该路径')
  }
}

const goToParent = () => {
  browsePathInput.value = ''
  loadBrowsePath()
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
  const paths = filesInput.value.split('\n').map(p => p.trim()).filter(p => p)
  if (paths.length === 0) {
    ElMessage.warning('请输入文件路径')
    return
  }
  if (!moveForm.dst_path) {
    ElMessage.warning('请输入目标路径')
    return
  }

  loading.value = true
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
    ElMessage.success(`${isCopy.value ? '复制' : '移动'}完成`)
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

.result-card {
  margin-top: 20px;
}

.success-text { color: #52c41a; }
.fail-text { color: #f5222d; }
.skip-text { color: #faad14; }
.fail-item { color: #f5222d; font-size: 13px; margin: 2px 0; }

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
