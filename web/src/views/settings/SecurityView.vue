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
          <span v-if="hasPassword && form.password_hint" class="hint-display">
            提示：{{ form.password_hint }}
          </span>
        </el-form-item>

        <template v-if="!hasPassword">
          <el-form-item label="设置密码">
            <el-input v-model="form.new_password" type="password" placeholder="输入新密码" show-password @input="checkStrength" />
          </el-form-item>
          <el-form-item v-if="form.new_password" label="密码强度">
            <div class="strength-bar">
              <div class="strength-fill" :style="{ width: strengthPercent + '%' }" :class="strengthClass"></div>
            </div>
            <span class="strength-text" :class="strengthClass">{{ strengthLabel }}</span>
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
            <el-input v-model="form.new_password" type="password" placeholder="输入新密码（留空则不修改）" show-password @input="checkStrength" />
          </el-form-item>
          <el-form-item v-if="form.new_password" label="密码强度">
            <div class="strength-bar">
              <div class="strength-fill" :style="{ width: strengthPercent + '%' }" :class="strengthClass"></div>
            </div>
            <span class="strength-text" :class="strengthClass">{{ strengthLabel }}</span>
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
          <el-button v-if="hasPassword" type="danger" @click="handleRemove" :loading="removing">
            删除密码
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { Lock } from '@element-plus/icons-vue'
import { getConfig, setPassword, removePassword } from '../../api/config'
import { ElMessage, ElMessageBox } from 'element-plus'

const hasPassword = ref(false)
const saving = ref(false)
const removing = ref(false)

const form = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
  password_hint: '',
})

const strengthPercent = ref(0)
const strengthLabel = ref('')

const strengthClass = computed(() => {
  if (strengthPercent.value <= 25) return 'weak'
  if (strengthPercent.value <= 50) return 'fair'
  if (strengthPercent.value <= 75) return 'good'
  return 'strong'
})

const checkStrength = () => {
  const pwd = form.new_password
  if (!pwd) {
    strengthPercent.value = 0
    strengthLabel.value = ''
    return
  }
  let score = 0
  if (pwd.length >= 4) score += 25
  if (pwd.length >= 8) score += 25
  if (/[A-Z]/.test(pwd) && /[a-z]/.test(pwd)) score += 25
  if (/[0-9]/.test(pwd) && /[^A-Za-z0-9]/.test(pwd)) score += 25
  strengthPercent.value = score
  const labels = { 25: '弱', 50: '一般', 75: '良好', 100: '强' }
  strengthLabel.value = labels[score] || '弱'
}

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
    strengthPercent.value = 0
    strengthLabel.value = ''
  } catch (err) {
    ElMessage.error(err.message || '操作失败')
  } finally {
    saving.value = false
  }
}

const handleRemove = async () => {
  try {
    await ElMessageBox.prompt('请输入当前密码以确认删除', '删除安全密码', {
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
      inputType: 'password',
      inputPlaceholder: '输入当前密码',
      type: 'warning',
    }).then(async ({ value }) => {
      removing.value = true
      try {
        await removePassword({ old_password: value })
        hasPassword.value = false
        form.old_password = ''
        form.new_password = ''
        form.confirm_password = ''
        form.password_hint = ''
        strengthPercent.value = 0
        strengthLabel.value = ''
        ElMessage.success('安全密码已删除，永久删除功能将不可用')
      } catch (err) {
        ElMessage.error(err.message || '密码验证失败')
      } finally {
        removing.value = false
      }
    })
  } catch {}
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

.hint-display {
  margin-left: 12px;
  font-size: 13px;
  color: #909399;
}

.strength-bar {
  width: 200px;
  height: 8px;
  background-color: #e4e7ed;
  border-radius: 4px;
  overflow: hidden;
}

.strength-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.3s, background-color 0.3s;
}

.strength-fill.weak {
  background-color: #f5222d;
}

.strength-fill.fair {
  background-color: #faad14;
}

.strength-fill.good {
  background-color: #1890ff;
}

.strength-fill.strong {
  background-color: #52c41a;
}

.strength-text {
  margin-left: 8px;
  font-size: 13px;
  font-weight: 500;
}

.strength-text.weak {
  color: #f5222d;
}

.strength-text.fair {
  color: #faad14;
}

.strength-text.good {
  color: #1890ff;
}

.strength-text.strong {
  color: #52c41a;
}
</style>
