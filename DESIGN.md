# 设计文档

## 架构决策记录 (ADR)

### ADR-001: cmd/internal 目录结构

**状态**: 已采纳

**上下文**: 
原始代码将 main.go 和 pool.go 放在根目录，随着功能增加，代码组织变得混乱。

**决策**: 
采用 Go 标准项目布局：
- `cmd/catch/` - 程序入口
- `internal/searcher/` - 核心业务逻辑

**结果**: 
- 主程序与业务逻辑解耦
- internal 包不可被外部导入
- 便于后续添加其他入口 (如 GUI)

---

### ADR-002: 组件化设计

**状态**: 已采纳

**上下文**: 
原始 pool.go 包含所有逻辑，难以维护和测试。

**决策**: 
按职责分离为独立文件：
- `types.go` - 公共类型
- `scanner.go` - 目录遍历
- `worker.go` - 并发处理
- `collector.go` - 结果收集
- `searcher.go` - 统一入口

**结果**: 
- 每个文件 < 200 行
- 职责清晰，易于理解
- 便于单元测试

---

### ADR-003: 权限不足跳过策略

**状态**: 已采纳

**上下文**: 
用户要求"没有权限打开则跳过"，原始代码遇到错误直接返回。

**决策**: 
- 使用 `os.IsPermission(err)` 检测权限错误
- 跳过无权限文件/目录，记录到 `FilesSkipped`
- 继续遍历其他文件，不中断整体流程

**结果**: 
- 工具更健壮，不会因权限问题中断
- 用户可看到跳过了多少文件
- 适合扫描系统目录等场景

---

### ADR-004: 原子操作统计

**状态**: 已采纳

**上下文**: 
多协程并发更新统计信息需要加锁，影响性能。

**决策**: 
使用 `sync/atomic` 原子操作更新 Stats:
```go
atomic.AddInt64(&w.stats.FilesScanned, 1)
```

**结果**: 
- 无锁统计，性能更好
- 代码更简洁
- 避免死锁风险

---

## 接口设计

### SearchEngine (统一入口)

```go
type SearchEngine struct {
    config    Config
    scanner   *DirScanner
    pool      *WorkerPool
    collector *Collector
}

func NewSearchEngine(rootPath, keyword string, searchType bool, cfg Config) *SearchEngine
func (se *SearchEngine) Run() Stats
func (se *SearchEngine) RunSilent() ([]SearchResult, Stats)
```

### Config (配置)

```go
type Config struct {
    WorkerNum      int // 工作协程数量
    ChanBufferSize int // 通道缓冲区大小
}

func DefaultConfig() Config // 返回默认配置
```

### Stats (统计)

```go
type Stats struct {
    FilesScanned   int64 // 扫描文件数
    LinesScanned   int64 // 扫描行数
    MatchesFound   int64 // 匹配结果数
    ErrorsOccurred int64 // 发生错误数
    FilesSkipped   int64 // 跳过文件数
}
```

---

## 并发流程

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ DirScanner  │ ──→ │ WorkerPool  │ ──→ │  Collector  │
│ (遍历目录)   │     │ (4 个 worker) │     │ (输出结果)   │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │
       │              ┌────┴────┐
       │              │ Worker  │
       │              │ Worker  │
       │              │ Worker  │
       │              │ Worker  │
       │              └─────────┘
       │
       └──→ filepath.WalkDir (递归遍历)
```

---

## 错误处理策略

| 错误类型 | 处理方式 | 记录位置 |
|----------|----------|----------|
| 目录不存在 | 退出程序 | main.go |
| 权限不足 (目录) | 跳过该目录 | `FilesSkipped++` |
| 权限不足 (文件) | 跳过该文件 | `FilesSkipped++` |
| 文件打开失败 | 跳过该文件 | `ErrorsOccurred++` |
| 扫描错误 | 记录错误 | `ErrorsOccurred++` |

---

## 性能优化点

1. **缓冲区大小**: 默认 100，平衡内存和性能
2. **Worker 数量**: 默认 4，适合大多数场景
3. **原子操作**: 无锁统计，避免竞争
4. **延迟关闭**: `defer file.Close()` 确保资源释放
5. **Context 取消**: 支持优雅退出

---

## 测试建议

```bash
# 1. 基本功能测试
./catch . func

# 2. 文件名搜索测试
./catch -v . go

# 3. 多协程测试
./catch -w 8 . func

# 4. 大缓冲区测试
./catch -b 500 . func

# 5. 权限测试 (需要管理员权限的目录)
./catch C:/Windows/System32 dll
```
