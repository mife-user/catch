<template>
  <div class="tutorial-view">
    <div class="tutorial-container">
      <div class="tutorial-header">
        <h1>Catch 文件整理工具</h1>
        <p class="subtitle">欢迎使用！让我们快速了解如何使用 Catch</p>
      </div>

      <el-steps :active="currentStep" align-center class="tutorial-steps">
        <el-step title="文件查找" />
        <el-step title="文件操作" />
        <el-step title="文件清理" />
        <el-step title="开始使用" />
      </el-steps>

      <div class="tutorial-content">
        <transition name="step-fade" mode="out-in">
          <div v-if="currentStep === 0" key="step0" class="step-panel">
            <div class="step-icon">
              <el-icon :size="64" color="#1890ff"><Search /></el-icon>
            </div>
            <h2>文件查找</h2>
            <p class="step-desc">通过多种条件快速查找文件</p>
            <div class="feature-list">
              <div class="feature-item">
                <el-icon color="#1890ff"><Folder /></el-icon>
                <div>
                  <strong>选择式路径</strong>
                  <p>点击浏览按钮选择搜索路径，无需手动输入</p>
                </div>
              </div>
              <div class="feature-item">
                <el-icon color="#52c41a"><Document /></el-icon>
                <div>
                  <strong>文件类型筛选</strong>
                  <p>按文档、图片、视频、音频等类型快速筛选</p>
                </div>
              </div>
              <div class="feature-item">
                <el-icon color="#faad14"><Calendar /></el-icon>
                <div>
                  <strong>日期范围</strong>
                  <p>选择常用日期范围或自定义日期</p>
                </div>
              </div>
            </div>
          </div>

          <div v-if="currentStep === 1" key="step1" class="step-panel">
            <div class="step-icon">
              <el-icon :size="64" color="#52c41a"><Edit /></el-icon>
            </div>
            <h2>文件操作</h2>
            <p class="step-desc">对查找结果进行批量操作</p>
            <div class="feature-list">
              <div class="feature-item">
                <el-icon color="#f5222d"><Delete /></el-icon>
                <div>
                  <strong>安全删除</strong>
                  <p>支持移至回收站、过期清理和永久删除三种模式</p>
                </div>
              </div>
              <div class="feature-item">
                <el-icon color="#1890ff"><Edit /></el-icon>
                <div>
                  <strong>批量重命名</strong>
                  <p>支持前缀、后缀、序号、替换等多种规则</p>
                </div>
              </div>
              <div class="feature-item">
                <el-icon color="#722ed1"><Rank /></el-icon>
                <div>
                  <strong>移动/复制</strong>
                  <p>批量移动或复制文件到指定目录</p>
                </div>
              </div>
            </div>
          </div>

          <div v-if="currentStep === 2" key="step2" class="step-panel">
            <div class="step-icon">
              <el-icon :size="64" color="#faad14"><Brush /></el-icon>
            </div>
            <h2>文件清理</h2>
            <p class="step-desc">智能清理系统垃圾和软件缓存</p>
            <div class="feature-list">
              <div class="feature-item">
                <el-icon color="#faad14"><Delete /></el-icon>
                <div>
                  <strong>预设规则</strong>
                  <p>临时文件、日志文件、大文件等一键清理</p>
                </div>
              </div>
              <div class="feature-item">
                <el-icon color="#722ed1"><ChatDotRound /></el-icon>
                <div>
                  <strong>聊天软件缓存</strong>
                  <p>清理QQ、微信等聊天软件的缓存文件</p>
                </div>
              </div>
              <div class="feature-item">
                <el-icon color="#f5222d"><WarningFilled /></el-icon>
                <div>
                  <strong>重要文件标记</strong>
                  <p>重要文件红色标记，防止误删</p>
                </div>
              </div>
            </div>
          </div>

          <div v-if="currentStep === 3" key="step3" class="step-panel">
            <div class="step-icon">
              <el-icon :size="64" color="#52c41a"><CircleCheck /></el-icon>
            </div>
            <h2>准备就绪！</h2>
            <p class="step-desc">您已了解 Catch 的基本功能</p>
            <div class="tips">
              <div class="tip-item">
                <el-icon><InfoFilled /></el-icon>
                <span>左侧导航栏可以搜索和拖拽调整功能顺序</span>
              </div>
              <div class="tip-item">
                <el-icon><InfoFilled /></el-icon>
                <span>耗时操作会显示实时进度条，支持取消和暂停</span>
              </div>
              <div class="tip-item">
                <el-icon><InfoFilled /></el-icon>
                <span>红色标记的文件为重要文件，删除前请仔细确认</span>
              </div>
            </div>
          </div>
        </transition>
      </div>

      <div class="tutorial-footer">
        <el-button v-if="currentStep > 0" @click="currentStep--">上一步</el-button>
        <el-button v-if="currentStep < 3" type="primary" @click="currentStep++">下一步</el-button>
        <el-button v-if="currentStep === 3" type="success" @click="handleFinish">开始使用</el-button>
        <el-checkbox v-model="dontShowAgain" class="dont-show">以后不再展示</el-checkbox>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  Search, Folder, Document, Calendar, Edit, Delete, Rank,
  Brush, ChatDotRound, WarningFilled, CircleCheck, InfoFilled
} from '@element-plus/icons-vue'

const router = useRouter()
const currentStep = ref(0)
const dontShowAgain = ref(false)

const handleFinish = () => {
  if (dontShowAgain.value) {
    localStorage.setItem('catch_tutorial_dismissed', 'true')
  }
  router.replace('/search')
}
</script>

<style scoped>
.tutorial-view {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 24px;
}

.tutorial-container {
  background: #ffffff;
  border-radius: 16px;
  padding: 40px;
  max-width: 640px;
  width: 100%;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}

.tutorial-header {
  text-align: center;
  margin-bottom: 32px;
}

.tutorial-header h1 {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
  margin-bottom: 8px;
}

.subtitle {
  font-size: 16px;
  color: #909399;
}

.tutorial-steps {
  margin-bottom: 32px;
}

.tutorial-content {
  min-height: 280px;
}

.step-panel {
  text-align: center;
}

.step-icon {
  margin-bottom: 16px;
}

.step-panel h2 {
  font-size: 22px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
}

.step-desc {
  font-size: 14px;
  color: #909399;
  margin-bottom: 24px;
}

.feature-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  text-align: left;
}

.feature-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  border-radius: 8px;
  background: #f5f7fa;
}

.feature-item .el-icon {
  font-size: 24px;
  margin-top: 2px;
  flex-shrink: 0;
}

.feature-item strong {
  font-size: 14px;
  color: #303133;
  display: block;
  margin-bottom: 4px;
}

.feature-item p {
  font-size: 12px;
  color: #909399;
  margin: 0;
}

.tips {
  display: flex;
  flex-direction: column;
  gap: 12px;
  text-align: left;
}

.tip-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: #f0f9ff;
  border-radius: 8px;
  font-size: 13px;
  color: #1890ff;
}

.tutorial-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid #e4e7ed;
}

.dont-show {
  margin-left: auto;
}

.step-fade-enter-active,
.step-fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.step-fade-enter-from {
  opacity: 0;
  transform: translateX(20px);
}

.step-fade-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

@media (max-width: 768px) {
  .tutorial-container {
    padding: 24px;
  }

  .tutorial-header h1 {
    font-size: 22px;
  }

  .tutorial-footer {
    flex-wrap: wrap;
  }

  .dont-show {
    margin-left: 0;
    width: 100%;
    justify-content: center;
  }
}
</style>
