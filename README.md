# Searching - Go 并发文件搜索工具

一个使用 Go 语言编写的高性能并发文件搜索工具，支持搜索文件内容和文件名。

## 功能特性

### 核心功能
- **文件内容搜索**: 递归遍历目录，搜索文件内容中包含关键词的行
- **文件名搜索**: 搜索文件名中包含关键词的文件 (`-v` 标志)
- **并发处理**: 使用 worker pool 模式，多协程并行搜索
- **权限处理**: 自动跳过无权限访问的文件/目录，记录跳过数量
- **统计信息**: 实时显示扫描文件数、行数、匹配数、错误数

### 命令行参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-v` | 搜索文件名（而非文件内容） | false |
| `-w` | 工作协程数量 | 4 |
| `-b` | 通道缓冲区大小 | 100 |

### 使用示例

```bash
# 搜索文件内容
./catch ./src func

# 搜索文件名
./catch -v ./src .go

# 自定义协程数和缓冲区
./catch -w 8 -b 200 ./src main
```

### 输出示例

```
开始搜索 - 目录：./src, 关键词：func, 模式：文件内容
工作协程数：4, 缓冲区大小：100

[1] 路径：src/main.go
    行数：10
    内容：func main() {

========== 搜索完成 ==========
扫描文件数：100
扫描行数：5000
匹配结果：25
发生错误：2
跳过文件 (权限不足): 1
==============================
```

---

## 设计理念

### 1. 架构分层 (cmd/internal 分离)

```
cmd/catch/          # 程序入口：CLI 解析、参数验证
internal/searcher/  # 业务逻辑：搜索核心实现
```

**优势**:
- 主程序与业务逻辑解耦
- 便于后续扩展 (如添加 GUI、Web 界面)
- internal 包不可被外部导入，保证封装性

### 2. 组件化设计

| 组件 | 职责 |
|------|------|
| `DirScanner` | 目录遍历，生成搜索任务，权限检查 |
| `Worker` | 执行单个搜索任务 |
| `WorkerPool` | 管理 worker 协程生命周期 |
| `Collector` | 收集和输出搜索结果 |
| `SearchEngine` | 统一入口，协调各组件 |

**优势**:
- 单职责原则，每个组件职责清晰
- 易于单元测试
- 便于独立优化和替换

### 3. 并发模型

```
DirScanner → TasksChan → WorkerPool → ResultsChan → Collector
     ↓                                              ↓
  (遍历目录)                                    (输出结果)
```

- 使用 Go channel 进行协程间通信
- 背压机制：缓冲区控制生产/消费速度
- Context 支持：可优雅取消搜索

### 4. 错误处理

- **权限不足**: 检测 `os.IsPermission(err)`，跳过并记录到 `FilesSkipped`
- **文件打开失败**: 记录到 `ErrorsOccurred`，继续处理其他文件
- **目录遍历错误**: 返回错误信息，不中断整体流程

### 5. 统计信息

使用 `sync/atomic` 原子操作，无锁统计：
- `FilesScanned`: 扫描文件数
- `LinesScanned`: 扫描行数
- `MatchesFound`: 匹配结果数
- `ErrorsOccurred`: 发生错误数
- `FilesSkipped`: 跳过文件数 (权限不足)

---

## 项目结构

```
Searching/
├── cmd/
│   └── catch/
│       └── main.go          # 程序入口
├── internal/
│   └── searcher/
│       ├── types.go         # 公共类型定义
│       ├── scanner.go       # 目录扫描器
│       ├── worker.go        # 工作池
│       ├── collector.go     # 结果收集器
│       └── searcher.go      # 搜索引擎入口
├── go.mod
├── README.md
└── catch.exe                # 编译产物
```

---

## 对话上下文压缩

### 迭代历史

| 版本 | 变更内容 |
|------|----------|
| v1 (原始) | 单文件 pool.go + main.go，基础 worker pool 实现 |
| v2 | 分离 DirScanner/Worker/Collector/SearchEngine，添加 Config 和 Stats |
| v3 (当前) | cmd/internal 目录结构，权限检查跳过逻辑，完善统计 |

### 关键决策

1. **放弃手动通道关闭** → 使用结构化生命周期管理
2. **放弃无意义 worker 结构体** → Worker 绑定 stats，无状态处理
3. **添加权限检查** → `os.Stat` + `os.IsPermission` 检测
4. **双收集器模式** → `Collector`(实时输出) + `SilentCollector`(静默收集)

### 技术选型

- **Go 1.25.4**: 使用最新稳定版
- **标准库**: 仅依赖标准库，无第三方依赖
- **原子操作**: `sync/atomic` 无锁统计
- **Context**: 支持优雅取消

---

## 构建与运行

```bash
# 编译
go build -o catch.exe ./cmd/catch

# 运行测试
./catch -h
./catch -v . go
./catch . func
```

---

## 后续扩展方向

1. **输出格式**: 支持 JSON/XML 格式输出
2. **正则搜索**: 添加 `-regex` 标志支持正则表达式
3. **文件过滤**: 添加 `--include`/`--exclude` 过滤文件扩展名
4. **进度显示**: 添加进度条显示搜索进度
5. **缓存机制**: 对大目录建立索引缓存
