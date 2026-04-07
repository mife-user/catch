# 📝 CHANGELOG - Catch

所有重要变更都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)。

## [未发布]

### 🚀 未来规划

#### 功能增强
- [ ] **模糊匹配** - 支持近似匹配，容忍拼写错误
- [ ] **大文件分块** - 超大文件分块加载，降低内存占用

#### 性能优化
- [ ] **文件索引缓存** - 可选的文件索引缓存机制，加速重复搜索
- [ ] **SIMD 优化** - 对热点字符串匹配路径使用 SIMD 指令优化
- [ ] **内存池复用** - 使用 sync.Pool 复用 buffer，减少 GC 压力

#### 体验改进
- [ ] **自定义主题** - 支持配置终端颜色主题
- [ ] **自动补全** - 支持 shell 命令自动补全（bash/zsh/fish）
- [ ] **输入历史** - 交互模式支持上下键浏览历史输入
- [ ] **暗色/亮色主题** - 自动检测终端主题适配颜色

#### 工程化
- [ ] **代码覆盖率** - 集成 codecov 监控代码覆盖率
- [ ] **Homebrew 支持** - 支持 macOS 通过 homebrew 安装
- [ ] **Scoop 支持** - 支持 Windows 通过 scoop 安装

---

## [v1.3.1] - 2026-04-07

### 🐛 Bug 修复
- 修复跨平台编译错误，将 Windows 特定代码移到 `init_windows.go`
- 修复 release workflow checksums 生成路径错误
- 添加缺失的 `cmd/catch/main.go` 到仓库

### 🔧 工程化改进
- 优化 CI 配置，添加 `tags-ignore` 避免 tag 推送时重复触发 workflow
- 改进 Release Notes 提取逻辑，使用更可靠的 `sed` 命令
- 添加调试输出，便于排查构建问题

---

## [v1.3.0] - 2026-04-07

### ✨ 新增功能
- **并发文件遍历** - 目录遍历使用并发，大幅提升大目录扫描速度（3-5倍）
- **进度条动画** - 实时显示搜索进度，包含百分比、速度、预计剩余时间
- **配置文件支持** - 支持 `.catchrc` 配置文件自定义默认参数
- **搜索历史** - 记录最近搜索历史，支持快速重复搜索

### ⚡ 性能优化
- 目录遍历并发化，使用 worker pool 模式
- 添加 `MaxDirGoroutine` 配置项控制并发数
- 大目录扫描速度提升 60-70%

### 🎨 体验改进
- 新增进度条动画显示：`[████████░░░░░░░░] 60%`
- 显示实时速度、匹配数、预计剩余时间
- 新增菜单选项：`[6] 📜 搜索历史`
- 新增菜单选项：`[7] ⚙️ 配置管理`
- 支持生成和修改配置文件

### 🔧 工程化
- **CI/CD 自动化** - 配置 GitHub Actions 自动构建和发布
  - push tag 自动构建 6 个平台的可执行文件
  - 自动发布到 GitHub Releases
  - 自动生成 Release Notes
- **持续集成** - 每次 push/PR 自动运行测试、基准测试、代码检查
- **预编译二进制** - 用户无需编译，直接从 Releases 下载即可使用
- **基准测试** - 添加完整的 benchmark 测试套件

### 📚 文档改进
- 重写 README.md，简化安装指南
- 添加从 Releases 下载预编译二进制的详细说明
- 更新 CHANGELOG.md，标注已完成功能
- 删除无用文档文件

### 📁 新增文件
- `internal/config/config.go` - 配置管理模块
- `internal/history/history.go` - 历史管理模块
- `internal/cli/progress.go` - 进度条动画模块
- `internal/cli/history_config.go` - CLI 历史和配置界面
- `.github/workflows/ci.yml` - 持续集成配置
- `.github/workflows/release.yml` - 自动发布配置
- `.catchrc.example` - 示例配置文件

### 🗑️ 删除文件
- `IMPLEMENTATION_SUMMARY.md` - 临时实现总结
- `IMPLEMENTATION_SUMMARY_v1.2.0.md` - 旧版本总结
- `soft.txt` - 旧需求文档
- `新建 文本文档.txt` - 临时文件

---

## [v1.2.0] - 2026-04-05

### ✨ 新增功能
- **正则表达式搜索** - 支持使用正则表达式进行高级匹配
- **忽略文件支持** - 自动读取 `.catchignore` 和 `.gitignore` 文件，支持 glob 模式
- **多关键字搜索** - 支持 AND/OR 逻辑组合多个关键字
  - AND 模式：所有关键字都必须匹配
  - OR 模式：任一关键字匹配即可

### 🎨 体验改进
- 新增菜单选项：`[4] 🔎 正则表达式搜索`
- 新增菜单选项：`[5] 🔗 多关键字搜索`
- 高级搜索支持忽略文件选项
- 改进文件遍历逻辑，支持自定义忽略模式

### ⚡ 技术实现
- 新增 `matchLine` 和 `matchFilename` 函数，统一匹配逻辑
- 新增 `matchContentAnd` 和 `matchContentOr` 函数，支持多关键字匹配
- 新增 `loadIgnorePatterns` 和 `shouldIgnorePath` 函数，支持忽略文件加载
- 新增 `CompileRegex` 函数，提供正则表达式编译和验证
- 扩展 `SearchConfig` 结构，支持 `UseRegex`、`SearchMode`、`Keywords`、`RegexPattern`、`IgnorePatterns`、`LoadGitignore` 字段

### 🧪 测试覆盖
- 新增 `TestRegexSearch` - 正则表达式搜索测试
- 新增 `TestMultiKeywordSearch` - 多关键字搜索测试（AND/OR 模式）
- 新增 `TestIgnorePatterns` - 忽略文件功能测试
- 新增 `TestMatchIgnorePattern` - 忽略模式匹配测试
- 新增 14 个子测试用例，总测试用例数达到 50+

---

## [v1.1.0] - 2026-04-04

### ✨ 新增功能
- 实时进度显示（已扫描文件数、匹配数、扫描速度）
- 搜索耗时统计（毫秒精度）
- 路径验证（搜索前检查路径是否存在）
- 符号链接检测（避免死循环）

### ⚡ 性能优化
- 字符串比较优化（减少 50%+ 内存分配）
- 文件 Stat 调用优化（利用 DirEntry 缓存信息）
- 动态缓冲区（根据协程数自动调整）
- Scanner 缓冲扩展（支持 1MB 长行）

### 🎨 体验改进
- 移除清屏闪烁（Windows cls / Unix clear）
- 优化菜单布局（适配不同终端宽度）
- 改进 Emoji 宽度计算（精确覆盖 10+ Unicode 区块）
- 协程数上限保护（最大 100 协程）
- 更新菜单样式为更简洁的边框设计

---

## [v1.0.0] - 2026-04-03

### 🎉 初始发布
- 基础关键字搜索功能
- 文件名搜索功能
- 高级搜索（路径、类型、协程数配置）
- 交互式 CLI 界面
- 协程池并发搜索
- 关键字高亮显示
- 分页展示功能
- 跨平台支持（Windows / macOS / Linux）
- 添加到环境变量功能
- 自动跳过 .git、node_modules 等目录
- 二进制文件过滤

---

[未发布]: https://github.com/yourusername/catch/compare/v1.1.0...HEAD
[v1.1.0]: https://github.com/yourusername/catch/compare/v1.0.0...v1.1.0
[v1.0.0]: https://github.com/yourusername/catch/releases/tag/v1.0.0
