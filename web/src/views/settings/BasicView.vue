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
          <el-input-number v-model="form.port" :min="3000" :max="3100" />
        </el-form-item>

        <el-form-item label="过期天数">
          <el-input-number v-model="form.expire_days" :min="1" :max="365" />
        </el-form-item>

        <el-form-item label="回收站路径">
          <el-input v-model="form.trash_path" placeholder="默认 ~/.catch-trash" />
        </el-form-item>

        <el-form-item label="默认搜索路径">
          <el-input v-model="form.default_path" placeholder="默认 ~" />
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
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Setting, Delete } from '@element-plus/icons-vue'
import { getConfig, updateConfig } from '../../api/config'
import { ElMessage } from 'element-plus'

const loading = ref(false)
const saving = ref(false)

const form = reactive({
  port: 3000,
  expire_days: 7,
  trash_path: '~/.catch-trash',
  default_path: '~',
  favorites: [],
})

onMounted(async () => {
  loading.value = true
  try {
    const config = await getConfig()
    form.port = config.server?.port || 3000
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

const addFavorite = () => {
  form.favorites.push('')
}

const removeFavorite = (index) => {
  form.favorites.splice(index, 1)
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
</style>
