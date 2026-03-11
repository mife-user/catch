# 🎯 Catch - 本地文件夹内容搜索工具

一个用 Go 语言开发的轻量级本地文件夹内容搜索命令行工具，支持交互式界面和协程池并发搜索。

## ✨ 功能特性

- 🔍 **关键字搜索** - 支持按关键字搜索文件内容和文件名
- 🔄 **递归/非递归搜索** - 可选择是否搜索子目录
- ⚡ **协程池并发** - 使用固定数量协程处理文件遍历，避免资源占用过高
- 🎨 **交互式界面** - 类似 Claude CLI 风格，支持数字/方向键选择功能
- 🌈 **高亮展示** - 搜索结果关键字高亮显示
- 📄 **分页展示** - 大量结果时分页查看
- 🌍 **跨平台支持** - Windows / macOS / Linux
- 🚀 **一键配置** - 支持添加到系统环境变量

## 📥 下载

### 方式 1: 克隆仓库

```bash
git clone https://github.com/yourusername/catch.git
cd catch
```

### 方式 2: 下载源码

从 [Releases](https://github.com/yourusername/catch/releases) 页面下载最新源码

## 🔨 编译

### 环境要求

- Go 1.18 或更高版本

### 编译步骤

```bash
# 进入项目目录
cd catch

# 下载依赖
go mod tidy

# 编译
go build -o catch ./cmd/catch
```

### 跨平台编译

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o catch.exe ./cmd/catch

# macOS
GOOS=darwin GOARCH=amd64 go build -o catch ./cmd/catch

# Linux
GOOS=linux GOARCH=amd64 go build -o catch ./cmd/catch
```

## 🚀 运行

### 直接运行

```bash
# 进入交互模式
./catch

# 命令行搜索
./catch search <关键字> [选项]
```

### 示例

```bash
# 搜索关键字 "hello"
./catch search hello

# 递归搜索
./catch search hello -r

# 指定文件类型
./catch search func -t .go

# 指定搜索路径
./catch search config -p ./src -r
```

## ➕ 添加到环境变量

### 自动添加

运行以下命令：

```bash
# Linux/macOS
./catch add-path

# Windows (需要管理员权限)
catch.exe add-path
```

### 手动添加

#### Windows

1. 右键点击"此电脑" -> "属性"
2. 点击"高级系统设置"
3. 点击"环境变量"
4. 在"系统变量"中找到"Path"
5. 点击"编辑" -> "新建"
6. 添加 catch 可执行文件所在路径

#### macOS/Linux

在 `~/.bashrc` 或 `~/.zshrc` 中添加：

```bash
export PATH="$PATH:/path/to/catch"
```

然后执行：

```bash
source ~/.bashrc  # 或 source ~/.zshrc
```

## 📖 使用说明

### 交互模式

运行 `catch` 进入交互模式：

```
╔════════════════════════════════════════╗
║     🎯 Catch - 文件搜索工具            ║
╠════════════════════════════════════════╣
║  [1] 🔍 搜索文件内容                   ║
║  [2] 📁 搜索文件名                     ║
║  [3] ⚙️  高级搜索                      ║
║  [4] ➕ 添加到环境变量                 ║
║  [q] ❌ 退出                           ║
╚════════════════════════════════════════╝

请选择功能 (1-4 或 q):
```

### 命令行模式

```bash
# 基本用法
catch search <关键字>

# 选项
catch search <关键字> [选项]

# 选项说明
-r, --recursive    递归搜索子目录
-t, --type         文件类型过滤 (如：.go,.txt)
-p, --path         搜索路径 (默认为当前目录)
```

### 搜索结果显示

- 📄 表示文件内容匹配
- 📁 表示文件名匹配
- 关键字会以黄色高亮显示

## 📁 项目结构

```
catch/
├── cmd/
│   └── catch/
│       └── main.go          # 程序入口
├── internal/
│   ├── cli/
│   │   └── cli.go           # 交互式 CLI 界面
│   └── search/
│       └── search.go        # 搜索核心逻辑
├── go.mod                   # Go 模块定义
├── README.md                # 项目说明
└── soft.txt                 # 需求文档
```

## ⚙️ 技术实现

- **协程池**: 固定数量协程处理文件遍历与内容搜索
- **并发控制**: 使用 channel 和 sync.WaitGroup 管理并发
- **跨平台**: 纯 Go 实现，无第三方依赖（除 golang.org/x/term）
- **文件过滤**: 自动跳过 .git、node_modules 等目录

## 📝 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📧 联系方式

如有问题或建议，请通过 GitHub Issues 联系。
