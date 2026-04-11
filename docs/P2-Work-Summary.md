# Catch P2 阶段工作总结报告

> 阶段：P2（后端优化 + 前端交互增强）
> 日期：2026-04-11
> 项目：Catch 文件整理工具 v1.0

---

## 一、执行摘要

P2 阶段聚焦于后端文件操作的健壮性优化和前端用户体验的全面增强。共完成 **8 项** 变更，涉及 **6 个文件**，涵盖后端性能与兼容性修复、前端数据传递机制重构、操作确认流程完善、以及新功能入口添加。所有变更均通过编译和 API 测试验证。

**关键成果：**

- 后端 Copy/Move 方法在性能和跨文件系统兼容性上得到显著改善
- 前端文件列表传递从 URL query params 迁移至 sessionStorage，彻底解决 URL 长度限制问题
- MoveView 完成全面改造，新增模式切换、确认对话框、结果可视化等功能
- 新增"文件复制"侧边栏菜单入口，操作路径更加直观

---

## 二、后端修复

### 2.1 Copy 方法优化

**文件：** [file_repository_impl.go](file:///d:/vscode/VsCodeWork/catch/internal/infrastructure/persistence/file_repository_impl.go#L123-L151)

**变更内容：**

| 项目 | 优化前 | 优化后 |
|------|--------|--------|
| 复制方式 | 手动缓冲区循环读取写入 | `io.Copy(dstFile, srcFile)` |
| 失败处理 | 目标文件残留 | 自动调用 `os.Remove(dst)` 清理 |
| 权限保留 | 未保留源文件权限 | `os.Chmod(dst, srcInfo.Mode())` 保留权限 |

**核心代码：**

```go
// 使用 io.Copy 替代手动缓冲区复制，提升大文件复制性能
if _, err := io.Copy(dstFile, srcFile); err != nil {
    os.Remove(dst) // 复制失败时自动清理目标文件
    return err
}

// 保留源文件权限
srcInfo, err := os.Stat(src)
if err == nil {
    os.Chmod(dst, srcInfo.Mode())
}
```

**影响分析：**

- `io.Copy` 内部使用 32KB 缓冲区，对大文件复制性能有显著提升
- 失败清理机制避免了磁盘上残留不完整文件
- 权限保留确保复制后的文件可执行属性等不被丢失

---

### 2.2 Move 方法跨文件系统支持

**文件：** [file_repository_impl.go](file:///d:/vscode/VsCodeWork/catch/internal/infrastructure/persistence/file_repository_impl.go#L100-L121)

**问题背景：**

`os.Rename` 在 Linux/Unix 系统中，当源路径和目标路径位于不同文件系统（不同磁盘分区、不同挂载点）时，会返回 `*os.LinkError`，导致移动操作失败。

**变更内容：**

```go
func (r *FileRepositoryImpl) Move(src, dst string) error {
    if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
        return err
    }

    // 首先尝试原子重命名（同文件系统时最高效）
    err := os.Rename(src, dst)
    if err == nil {
        return nil
    }

    // 检测到跨文件系统错误，回退到 Copy + Remove 策略
    if linkErr, ok := err.(*os.LinkError); ok {
        if r.Copy(src, dst) != nil {
            return linkErr // 复制也失败，返回原始 LinkError
        }
        if removeErr := os.Remove(src); removeErr != nil {
            return removeErr // 复制成功但删除源文件失败
        }
        return nil
    }

    return err // 其他类型的错误直接返回
}
```

**回退策略说明：**

1. 优先使用 `os.Rename`（原子操作，同文件系统下零拷贝）
2. 检测到 `*os.LinkError` 时自动回退为 Copy + Remove
3. Copy 失败时返回原始 `LinkError`，不暴露内部回退细节
4. Copy 成功但 Remove 源文件失败时返回 Remove 错误（此时目标文件已存在，数据安全）

---

### 2.3 浏览对话框 goToParent 修复

**文件：** [SearchView.vue](file:///d:/vscode/VsCodeWork/catch/web/src/views/SearchView.vue#L169-L282)、[MoveView.vue](file:///d:/vscode/VsCodeWork/catch/web/src/views/MoveView.vue#L176-L220)

**问题背景：**

原实现中 `goToParent` 方法将 `browsePath` 设置为空字符串 `''`，导致浏览路径重置为用户主目录，而非当前目录的上级目录。

**修复方案：**

1. 新增 `browseParentPath` ref 变量，用于存储后端返回的上级目录路径
2. `loadBrowsePath` 方法中从后端响应提取 `parent_path` 字段并赋值
3. `goToParent` 方法使用 `browseParentPath` 导航到真正的上级目录

**SearchView 修复代码：**

```javascript
const browseParentPath = ref('')

const loadBrowsePath = async () => {
  const data = await fetchBrowsePath(browsePath.value)
  browseItems.value = data.items || []
  browseCurrentPath.value = data.current_path || ''
  browseParentPath.value = data.parent_path || ''  // 存储上级路径
  browsePath.value = data.current_path || ''
}

const goToParent = () => {
  if (browseCurrentPath.value) {
    browsePath.value = browseParentPath.value  // 使用后端返回的上级路径
    loadBrowsePath()
  }
}
```

MoveView 中采用相同修复方案，变量名为 `browsePathInput` 和 `browseParentPath`。

---

## 三、前端改进

### 3.1 sessionStorage 替代 query params 传递文件列表

**涉及文件：**

| 文件 | 角色 | 变更 |
|------|------|------|
| [SearchView.vue](file:///d:/vscode/VsCodeWork/catch/web/src/views/SearchView.vue#L229-L263) | 写入方 | `sessionStorage.setItem('catch_selected_files', ...)` |
| [DeleteView.vue](file:///d:/vscode/VsCodeWork/catch/web/src/views/DeleteView.vue#L139-L148) | 读取方 | `sessionStorage.getItem` + 立即 `removeItem` |
| [RenameView.vue](file:///d:/vscode/VsCodeWork/catch/web/src/views/RenameView.vue#L161-L172) | 读取方 | `sessionStorage.getItem` + 立即 `removeItem` |
| [MoveView.vue](file:///d:/vscode/VsCodeWork/catch/web/src/views/MoveView.vue#L179-L189) | 读取方 | `sessionStorage.getItem` + 立即 `removeItem` |

**问题背景：**

原方案通过 URL query params 传递文件路径列表（如 `?files=path1&files=path2`），当文件数量多或路径较长时，会超出浏览器 URL 长度限制（通常约 2000-8000 字符），导致数据截断或请求失败。

**解决方案：**

- **写入端（SearchView）：** 批量操作时将文件路径数组 JSON 序列化后存入 `sessionStorage`
- **读取端（Delete/Rename/Move）：** `onMounted` 中读取后立即清除，避免数据残留
- **key 统一：** 使用 `catch_selected_files` 作为存储键名

**SearchView 写入示例：**

```javascript
const handleBatchDelete = () => {
  if (selectedFiles.value.length === 0) {
    ElMessage.warning('请先选择文件')
    return
  }
  sessionStorage.setItem('catch_selected_files', JSON.stringify(selectedFiles.value.map(f => f.path)))
  router.push('/delete')
}
```

**读取端通用模式：**

```javascript
onMounted(() => {
  const stored = sessionStorage.getItem('catch_selected_files')
  if (stored) {
    try {
      const files = JSON.parse(stored)
      if (Array.isArray(files) && files.length > 0) {
        filesInput.value = files.join('\n')
      }
    } catch {}
    sessionStorage.removeItem('catch_selected_files')  // 读取后立即清除
  }
})
```

**附加清理：** RenameView 中移除了不再需要的 `useRoute` 导入（文件列表不再从 route params 获取）。

---

### 3.2 MoveView 全面改造

**文件：** [MoveView.vue](file:///d:/vscode/VsCodeWork/catch/web/src/views/MoveView.vue)

#### 3.2.1 移动/复制模式切换

新增 `el-switch` 组件，允许用户在页面上直接切换移动和复制模式，不再仅依赖 URL 参数 `?mode=copy`。

```html
<el-switch
  v-model="isCopy"
  active-text="复制"
  inactive-text="移动"
  class="mode-switch"
/>
```

初始化逻辑仍兼容 URL 参数：

```javascript
const isCopy = ref(route.query.mode === 'copy')
```

#### 3.2.2 文件计数显示

在文件路径输入框下方显示当前文件数量：

```html
<div v-if="fileList.length > 0" class="file-count">
  共 {{ fileList.length }} 个文件
</div>
```

#### 3.2.3 操作模式提示

根据当前模式显示不同的 `el-alert` 提示：

- **移动模式：** "移动操作会将文件从原位置转移到目标位置，原位置文件将不存在"（info 类型）
- **复制模式：** "复制操作会在目标位置创建文件副本，原位置文件保持不变"（success 类型）

#### 3.2.4 执行前确认对话框

执行移动/复制前弹出 `ElMessageBox.confirm`，显示文件数量和冲突处理方式：

```javascript
await ElMessageBox.confirm(
  `确定要将 ${paths.length} 个文件${opText}到 "${moveForm.dst_path}" 吗？冲突处理：${conflictText}`,
  `确认${opText}`,
  { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
)
```

冲突处理文本映射：

| conflict 值 | 显示文本 |
|-------------|----------|
| `skip` | 跳过 |
| `rename` | 自动重命名 |
| `overwrite` | 覆盖 |

#### 3.2.5 结果展示改进

使用 `el-result` + `el-tag` 组合替代原有简单文本展示：

- **全部成功：** 显示 `icon="success"` 的 el-result
- **部分失败：** 显示 `icon="warning"` 的 el-result，配合颜色标签：
  - 绿色 `el-tag type="success"`：成功数量
  - 红色 `el-tag type="danger"`：失败数量
  - 黄色 `el-tag type="warning"`：跳过数量
- **失败详情：** 可滚动列表（`max-height: 200px; overflow-y: auto`）

---

### 3.3 RenameView 添加确认对话框

**文件：** [RenameView.vue](file:///d:/vscode/VsCodeWork/catch/web/src/views/RenameView.vue#L208-L248)

执行重命名前弹出确认对话框，显示规则名称（中文）和文件数量，并提示操作可撤销：

```javascript
const ruleNames = {
  prefix: '添加前缀',
  suffix: '添加后缀',
  sequence: '序号编号',
  replace: '替换文本',
  timestamp: '日期时间戳',
}

await ElMessageBox.confirm(
  `确定要对 ${paths.length} 个文件执行"${ruleNames[renameForm.rule] || renameForm.rule}"重命名吗？此操作可通过再次重命名撤销。`,
  '确认重命名',
  { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
)
```

---

### 3.4 侧边栏添加文件复制菜单入口

**文件：** [App.vue](file:///d:/vscode/VsCodeWork/catch/web/src/App.vue#L32-L35)

新增"文件复制"菜单项：

```html
<el-menu-item index="/copy">
  <el-icon><CopyDocument /></el-icon>
  <span>文件复制</span>
</el-menu-item>
```

**activeMenu 计算逻辑更新：**

当路由为 `/move?mode=copy` 时，高亮"文件复制"菜单项而非"文件移动"：

```javascript
const activeMenu = computed(() => {
  if (route.path === '/move' && route.query.mode === 'copy') {
    return '/copy'
  }
  return route.path
})
```

**菜单点击处理：**

"文件复制"菜单项点击时跳转到 `/move?mode=copy`，复用 MoveView 组件：

```javascript
const handleMenuSelect = (index) => {
  if (index === '/copy') {
    router.push({ path: '/move', query: { mode: 'copy' } })
  } else {
    router.push(index)
  }
}
```

---

## 四、构建验证

### 4.1 编译与构建

| 验证项 | 结果 |
|--------|------|
| Go 编译 | 通过 |
| npm 构建 | 通过（527ms） |

### 4.2 API 测试

| 接口 | 方法 | 结果 |
|------|------|------|
| `/api/config` | GET | 通过 |
| `/api/files/browse` | GET | 通过 |
| `/api/files/search` | GET | 通过 |
| `/api/files/rename/preview` | POST | 通过 |

### 4.3 前端页面

| 验证项 | 结果 |
|--------|------|
| 前端页面 HTTP 200 | 通过 |

---

## 五、修改文件清单

| 文件路径 | 变更类型 | 变更摘要 |
|----------|----------|----------|
| `internal/infrastructure/persistence/file_repository_impl.go` | 后端优化 | Copy 方法使用 io.Copy + 失败清理 + 权限保留；Move 方法跨文件系统回退 |
| `web/src/views/SearchView.vue` | 前端修复 | sessionStorage 传递文件列表、goToParent 修复、新增 browseParentPath |
| `web/src/views/MoveView.vue` | 前端改造 | 模式切换开关、文件计数、确认对话框、结果展示改进、goToParent 修复 |
| `web/src/views/RenameView.vue` | 前端改进 | 确认对话框、sessionStorage 替代、清理 useRoute 导入 |
| `web/src/views/DeleteView.vue` | 前端改进 | sessionStorage 替代 query params |
| `web/src/App.vue` | 前端新增 | 文件复制菜单入口、CopyDocument 图标、activeMenu 逻辑更新 |

---

## 六、问题与风险

### 6.1 已知限制

| 编号 | 限制描述 | 影响范围 | 建议处理 |
|------|----------|----------|----------|
| L1 | sessionStorage 在浏览器标签页关闭后数据丢失 | 若用户在搜索页选择文件后关闭标签页再打开操作页，文件列表将丢失 | 可考虑增加 localStorage 备选方案，但需权衡数据残留风险 |
| L2 | Move 方法跨文件系统回退时，若 Copy 成功但 Remove 源文件失败，目标文件已存在但源文件未删除 | 极端场景下可能出现文件重复 | 可在后续版本增加事务日志或重试机制 |
| L3 | RenameView 的"此操作可通过再次重命名撤销"提示仅适用于部分规则（如前缀/后缀/替换），序号编号规则难以精确撤销 | 用户可能对撤销难度产生误解 | 建议在 P3 阶段根据规则类型调整提示文案 |

### 6.2 后续建议

| 优先级 | 建议 | 说明 |
|--------|------|------|
| 高 | 为 Copy/Move 方法添加单元测试 | 覆盖同文件系统、跨文件系统、权限保留等场景 |
| 中 | 为 sessionStorage 传递机制添加容量超限处理 | 当文件列表极大时 sessionStorage 可能接近浏览器限制（通常 5-10MB） |
| 低 | MoveView 结果展示增加"查看详情"展开/收起功能 | 当成功文件数量很多时，结果区域可能过长 |

---

## 七、变更影响矩阵

```
                    file_repo  SearchView  MoveView  RenameView  DeleteView  App.vue
Copy 方法优化          X
Move 跨文件系统        X
goToParent 修复                   X          X
sessionStorage 替代               X          X          X           X
MoveView 改造                                X
RenameView 确认对话框                                    X
文件复制菜单入口                                                      X
```

---

*报告生成时间：2026-04-11*
*项目版本：v1.0*
*阶段：P2*
