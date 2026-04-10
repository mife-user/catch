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
              <el-button @click="selectDstPath">浏览</el-button>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="收藏目录">
          <el-select v-model="selectedFavorite" placeholder="从收藏中选择" clearable @change="handleFavoriteSelect">
            <el-option v-for="fav in favorites" :key="fav" :label="fav" :value="fav" />
          </el-select>
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
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Rank } from '@element-plus/icons-vue'
import { moveFiles, copyFiles } from '../api/files'
import { getConfig } from '../api/config'
import { ElMessage } from 'element-plus'
import { useRoute } from 'vue-router'

const route = useRoute()
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

const selectDstPath = () => {
  ElMessage.info('请直接在输入框中输入目标路径')
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
</style>
