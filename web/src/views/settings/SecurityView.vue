<template>
  <div class="security-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <el-icon><Lock /></el-icon>
          <span>安全设置</span>
        </div>
      </template>

      <el-form :model="form" label-width="120px" label-position="left">
        <el-form-item label="当前状态">
          <el-tag :type="hasPassword ? 'success' : 'warning'">
            {{ hasPassword ? '已设置密码' : '未设置密码' }}
          </el-tag>
        </el-form-item>

        <template v-if="!hasPassword">
          <el-form-item label="设置密码">
            <el-input v-model="form.new_password" type="password" placeholder="输入新密码" show-password />
          </el-form-item>
          <el-form-item label="确认密码">
            <el-input v-model="form.confirm_password" type="password" placeholder="再次输入密码" show-password />
          </el-form-item>
        </template>

        <template v-else>
          <el-form-item label="旧密码">
            <el-input v-model="form.old_password" type="password" placeholder="输入旧密码" show-password />
          </el-form-item>
          <el-form-item label="新密码">
            <el-input v-model="form.new_password" type="password" placeholder="输入新密码（留空则不修改）" show-password />
          </el-form-item>
          <el-form-item label="确认新密码">
            <el-input v-model="form.confirm_password" type="password" placeholder="再次输入新密码" show-password />
          </el-form-item>
        </template>

        <el-form-item label="密码提示">
          <el-input v-model="form.password_hint" placeholder="输入密码提示信息" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSave" :loading="saving">
            {{ hasPassword ? '修改密码' : '设置密码' }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Lock } from '@element-plus/icons-vue'
import { getConfig, setPassword } from '../../api/config'
import { ElMessage } from 'element-plus'

const hasPassword = ref(false)
const saving = ref(false)

const form = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
  password_hint: '',
})

onMounted(async () => {
  try {
    const config = await getConfig()
    hasPassword.value = config.has_password || false
    form.password_hint = config.security?.password_hint || ''
  } catch {}
})

const handleSave = async () => {
  if (!hasPassword.value && !form.new_password) {
    ElMessage.warning('请输入密码')
    return
  }

  if (form.new_password && form.new_password !== form.confirm_password) {
    ElMessage.warning('两次输入的密码不一致')
    return
  }

  saving.value = true
  try {
    await setPassword({
      old_password: form.old_password,
      new_password: form.new_password,
      password_hint: form.password_hint,
    })
    ElMessage.success(hasPassword.value ? '密码修改成功' : '密码设置成功')
    hasPassword.value = true
    form.old_password = ''
    form.new_password = ''
    form.confirm_password = ''
  } catch (err) {
    ElMessage.error(err.message || '操作失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.security-view {
  max-width: 600px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}
</style>
