<template>
  <div class="welcome-view">
    <el-card class="welcome-card">
      <el-steps :active="currentStep" align-center class="welcome-steps">
        <el-step title="欢迎" />
        <el-step title="功能介绍" />
        <el-step title="安全设置" />
        <el-step title="开始使用" />
      </el-steps>

      <div class="step-content">
        <div v-if="currentStep === 0" class="step-welcome">
          <div class="welcome-icon">
            <el-icon :size="64" color="#1890ff"><Monitor /></el-icon>
          </div>
          <h1>欢迎使用 Catch</h1>
          <p class="welcome-desc">一款面向非技术用户的文件整理工具</p>
          <p class="welcome-sub">通过直观的网页界面，轻松完成文件查找、删除、重命名、移动等操作</p>
        </div>

        <div v-if="currentStep === 1" class="step-features">
          <h2>核心功能</h2>
          <div class="feature-grid">
            <div class="feature-item">
              <el-icon :size="32" color="#1890ff"><Search /></el-icon>
              <h3>文件查找</h3>
              <p>按名称、类型、大小、日期多条件搜索</p>
            </div>
            <div class="feature-item">
              <el-icon :size="32" color="#f5222d"><Delete /></el-icon>
              <h3>文件删除</h3>
              <p>支持回收站、过期清理、永久删除三种模式</p>
            </div>
            <div class="feature-item">
              <el-icon :size="32" color="#52c41a"><Edit /></el-icon>
              <h3>文件重命名</h3>
              <p>批量重命名，支持前缀、后缀、序号等规则</p>
            </div>
            <div class="feature-item">
              <el-icon :size="32" color="#faad14"><Rank /></el-icon>
              <h3>文件移动/复制</h3>
              <p>快速移动或复制文件到指定目录</p>
            </div>
          </div>
        </div>

        <div v-if="currentStep === 2" class="step-security">
          <h2>安全设置（可选）</h2>
          <p class="security-desc">设置安全密码后，永久删除操作需要密码验证，防止误操作</p>
          <el-form :model="passwordForm" label-width="100px" label-position="left" class="password-form">
            <el-form-item label="设置密码">
              <el-input v-model="passwordForm.password" type="password" placeholder="输入安全密码（可跳过）" show-password />
            </el-form-item>
            <el-form-item label="确认密码">
              <el-input v-model="passwordForm.confirm" type="password" placeholder="再次输入密码" show-password />
            </el-form-item>
            <el-form-item label="密码提示">
              <el-input v-model="passwordForm.hint" placeholder="忘记密码时的提示信息" />
            </el-form-item>
            <el-alert v-if="passwordError" :title="passwordError" type="error" :closable="false" show-icon class="password-alert" />
          </el-form>
        </div>

        <div v-if="currentStep === 3" class="step-start">
          <div class="start-icon">
            <el-icon :size="64" color="#52c41a"><CircleCheckFilled /></el-icon>
          </div>
          <h2>一切就绪！</h2>
          <p>您已准备好使用 Catch 管理您的文件</p>
          <div class="tips">
            <el-alert title="提示：您随时可以在设置中修改安全密码和其他配置" type="info" :closable="false" show-icon />
          </div>
        </div>
      </div>

      <div class="step-actions">
        <el-button v-if="currentStep > 0" @click="currentStep--">上一步</el-button>
        <el-button v-if="currentStep < 3" type="primary" @click="handleNext">
          {{ currentStep === 2 && !passwordForm.password ? '跳过' : '下一步' }}
        </el-button>
        <el-button v-if="currentStep === 3" type="success" size="large" @click="handleFinish" :loading="finishing">
          开始使用
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { Monitor, Search, Delete, Edit, Rank, CircleCheckFilled } from '@element-plus/icons-vue'
import { setPassword, updateConfig } from '../api/config'
import { ElMessage } from 'element-plus'

const router = useRouter()
const currentStep = ref(0)
const finishing = ref(false)
const passwordError = ref('')

const passwordForm = reactive({
  password: '',
  confirm: '',
  hint: '',
})

const handleNext = async () => {
  if (currentStep.value === 2 && passwordForm.password) {
    passwordError.value = ''
    if (passwordForm.password.length < 4) {
      passwordError.value = '密码长度至少4位'
      return
    }
    if (passwordForm.password !== passwordForm.confirm) {
      passwordError.value = '两次输入的密码不一致'
      return
    }
    try {
      await setPassword({
        new_password: passwordForm.password,
        password_hint: passwordForm.hint,
      })
      ElMessage.success('安全密码设置成功')
    } catch (err) {
      ElMessage.error(err.message || '密码设置失败')
      return
    }
  }
  currentStep.value++
}

const handleFinish = async () => {
  finishing.value = true
  try {
    await updateConfig({ first_launch: false })
    router.replace('/')
  } catch (err) {
    ElMessage.error(err.message || '初始化失败')
  } finally {
    finishing.value = false
  }
}
</script>

<style scoped>
.welcome-view {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  min-height: calc(100vh - 48px);
  padding-top: 40px;
}

.welcome-card {
  width: 700px;
  max-width: 90vw;
}

.welcome-steps {
  margin-bottom: 32px;
}

.step-content {
  min-height: 280px;
  display: flex;
  justify-content: center;
  align-items: center;
}

.step-welcome,
.step-start {
  text-align: center;
}

.welcome-icon,
.start-icon {
  margin-bottom: 16px;
}

.step-welcome h1 {
  font-size: 28px;
  color: #1890ff;
  margin-bottom: 12px;
}

.welcome-desc {
  font-size: 18px;
  color: #333;
  margin-bottom: 8px;
}

.welcome-sub {
  font-size: 14px;
  color: #999;
  line-height: 1.6;
}

.step-features {
  width: 100%;
}

.step-features h2 {
  text-align: center;
  margin-bottom: 24px;
  font-size: 20px;
  color: #333;
}

.feature-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.feature-item {
  text-align: center;
  padding: 20px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  transition: box-shadow 0.3s;
}

.feature-item:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.feature-item h3 {
  margin: 8px 0 4px;
  font-size: 15px;
  color: #333;
}

.feature-item p {
  font-size: 13px;
  color: #999;
  margin: 0;
}

.step-security {
  width: 100%;
}

.step-security h2 {
  text-align: center;
  margin-bottom: 8px;
  font-size: 20px;
  color: #333;
}

.security-desc {
  text-align: center;
  color: #666;
  font-size: 14px;
  margin-bottom: 20px;
}

.password-form {
  max-width: 400px;
  margin: 0 auto;
}

.password-alert {
  margin-top: 8px;
}

.step-start h2 {
  font-size: 24px;
  color: #52c41a;
  margin-bottom: 8px;
}

.step-start p {
  color: #666;
  font-size: 14px;
  margin-bottom: 16px;
}

.tips {
  max-width: 400px;
  margin: 0 auto;
}

.step-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #e4e7ed;
}
</style>
