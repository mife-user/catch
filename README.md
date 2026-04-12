# Catch 文件整理工具

## 项目概述

Catch 是一款面向非技术用户的文件整理工具，通过"命令行启动 + 网页界面操作"的混合架构，帮助用户高效管理本地文件。用户只需记住一个命令 `catch`，即可通过直观的网页界面完成文件查找、删除、重命名、移动/复制等操作。

### 核心价值

- **零学习成本**：无需记忆复杂命令行参数，所有操作通过网页界面完成
- **安全可靠**：多重安全机制保障用户数据安全
- **高效便捷**：利用 Go 并发特性实现快速文件处理

### 目标用户

- 普通电脑用户（非技术人员）
- 需要定期整理文件的用户
- 对命令行有抵触情绪的用户

## 技术架构

| 层级 | 技术 | 说明 |
|------|------|------|
| 前端 | Vue 3 + Element Plus | 响应式UI框架 |
| 后端 | Go + Gin | RESTful API |
| 架构 | DDD | 领域驱动设计 |
| 存储 | JSON文件 | 配置和数据存储 |

## 安装说明

### 系统要求

- Windows/macOS/Linux
- Go 1.18+（仅开发环境需要）
- Node.js 16+（仅开发环境需要）

### 安装方法

#### 方法一：直接下载可执行文件

1. 从 [Releases](https://github.com/yourusername/catch/releases) 页面下载对应操作系统的可执行文件
2. 将可执行文件添加到系统 PATH 环境变量中
3. 在终端中输入 `catch` 命令启动

#### 方法二：从源码构建

1. 克隆仓库：
   ```bash
   git clone https://github.com/yourusername/catch.git
   cd catch
   ```

2. 构建前端：
   ```bash
   cd web
   npm install
   npm run build
   cd ..
   ```

3. 构建后端：
   ```bash
   go build -o catch ./cmd/catch
   ```

4. 将构建产物添加到系统 PATH 环境变量中

## 使用指南

### 启动应用

在任意终端窗口中输入以下命令：

```bash
catch
```

系统会自动：
1. 检测配置文件是否存在，不存在则创建默认配置
2. 检测可用端口（范围：3000-3100）
3. 启动 Gin API 服务
4. 自动打开默认浏览器访问 `http://localhost:{端口}`

### 核心功能

#### 1. 文件查找

1. 在左侧导航点击"文件查找"
2. 右侧显示查找条件表单
3. 填写查找条件（搜索路径、文件名、文件类型、文件大小、修改日期）
4. 点击"开始查找"
5. 查看查找结果，可对结果进行批量操作

#### 2. 文件删除

1. 在文件查找结果中选择要删除的文件
2. 点击"删除选中"按钮
3. 选择删除方式：
   - **移至回收站**：调用系统回收站，可恢复
   - **直接删除（过期清理）**：移至 `.catch-trash/`，到期自动清理
   - **永久删除**：立即彻底删除，需密码验证

#### 3. 文件重命名

1. 在文件查找结果中选择要重命名的文件
2. 点击"重命名选中"按钮
3. 选择重命名规则：
   - 添加前缀
   - 添加后缀
   - 序号编号
   - 替换文本
   - 日期时间戳
4. 填写规则参数
5. 预览重命名结果
6. 确认执行

#### 4. 文件移动/复制

1. 在文件查找结果中选择要移动/复制的文件
2. 点击"移动选中"或"复制选中"按钮
3. 选择目标路径：
   - 手动输入
   - 浏览选择
   - 从收藏目录中选择
4. 处理冲突（如目标文件已存在）
5. 确认执行

#### 5. 设置管理

1. 在左侧导航点击"设置"
2. 选择设置类型：
   - **基础设置**：默认过期时间、默认搜索路径、收藏目录列表、服务端口
   - **安全设置**：安全密码设置（用于永久删除验证）
   - **SMTP设置**：邮件服务器配置（用于发送反馈）
   - **关于/反馈**：版本信息、开发者信息、许可证信息、反馈功能

## 故障排除

### 常见问题

1. **浏览器无法自动打开**
   - 手动访问终端中显示的地址（如 `http://localhost:3000`）

2. **端口全部占用**
   - 关闭占用端口的程序
   - 或修改配置文件中的端口设置

3. **配置文件损坏**
   - 系统会自动备份并重建配置

4. **权限不足**
   - 无权限访问的文件会自动跳过
   - 查找完成后会显示"跳过的文件"列表

5. **永久删除功能不可用**
   - 需在首次启动时设置安全密码
   - 或在设置页面的安全设置中配置

## 贡献指南

### 开发环境设置

1. 克隆仓库：
   ```bash
   git clone https://github.com/yourusername/catch.git
   cd catch
   ```

2. 安装前端依赖：
   ```bash
   cd web
   npm install
   cd ..
   ```

3. 安装后端依赖：
   ```bash
   go mod download
   ```

### 开发流程

1. 创建分支：
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. 开发代码

3. 运行测试：
   ```bash
   # 运行后端测试
   go test ./...
   
   # 运行前端测试
   cd web
   npm test
   cd ..
   ```

4. 提交代码：
   ```bash
   git add .
   git commit -m "Add your commit message"
   git push origin feature/your-feature-name
   ```

5. 创建 Pull Request

### 代码规范

- 后端：遵循 Go 代码规范
- 前端：遵循 Vue 代码规范
- 提交信息：使用清晰、简洁的描述

## 许可证

[MIT License](LICENSE)

## 联系方式

- 反馈邮箱：15723556393@163.com
- GitHub：[https://github.com/yourusername/catch](https://github.com/yourusername/catch)

---

# Catch File Organizer Tool

## Project Overview

Catch is a file organization tool designed for non-technical users, using a hybrid architecture of "command-line启动 + web interface operation" to help users efficiently manage local files. Users only need to remember one command `catch` to complete file search, deletion, renaming, moving/copying operations through an intuitive web interface.

### Core Values

- **Zero learning cost**: No need to memorize complex command-line parameters, all operations are completed through the web interface
- **Safe and reliable**: Multiple security mechanisms to ensure user data security
- **Efficient and convenient**: Using Go's concurrency features for fast file processing

### Target Users

- General computer users (non-technical personnel)
- Users who need to organize files regularly
- Users who have resistance to command-line interfaces

## Technical Architecture

| Layer | Technology | Description |
|-------|------------|-------------|
| Frontend | Vue 3 + Element Plus | Responsive UI framework |
| Backend | Go + Gin | RESTful API |
| Architecture | DDD | Domain-driven design |
| Storage | JSON files | Configuration and data storage |

## Installation Instructions

### System Requirements

- Windows/macOS/Linux
- Go 1.18+ (only needed for development environment)
- Node.js 16+ (only needed for development environment)

### Installation Methods

#### Method 1: Direct Download of Executable

1. Download the executable file for your operating system from the [Releases](https://github.com/yourusername/catch/releases) page
2. Add the executable to your system PATH environment variable
3. Start by entering the `catch` command in the terminal

#### Method 2: Build from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/catch.git
   cd catch
   ```

2. Build the frontend:
   ```bash
   cd web
   npm install
   npm run build
   cd ..
   ```

3. Build the backend:
   ```bash
   go build -o catch ./cmd/catch
   ```

4. Add the built product to your system PATH environment variable

## Usage Guide

### Starting the Application

Enter the following command in any terminal window:

```bash
catch
```

The system will automatically:
1. Check if the configuration file exists, create a default configuration if it doesn't
2. Detect available ports (range: 3000-3100)
3. Start the Gin API service
4. Automatically open the default browser to access `http://localhost:{port}`

### Core Features

#### 1. File Search

1. Click "File Search" in the left navigation
2. The search criteria form is displayed on the right
3. Fill in the search criteria (search path, file name, file type, file size, modification date)
4. Click "Start Search"
5. View search results and perform batch operations on results

#### 2. File Deletion

1. Select the files to delete in the file search results
2. Click the "Delete Selected" button
3. Choose deletion method:
   - **Move to Recycle Bin**: Use system recycle bin, recoverable
   - **Direct Delete (Expired Cleanup)**: Move to `.catch-trash/`, automatically clean up when expired
   - **Permanent Delete**: Immediately and completely delete, requires password verification

#### 3. File Renaming

1. Select the files to rename in the file search results
2. Click the "Rename Selected" button
3. Choose renaming rules:
   - Add prefix
   - Add suffix
   - Sequence numbering
   - Replace text
   - Date timestamp
4. Fill in rule parameters
5. Preview renaming results
6. Confirm execution

#### 4. File Moving/Copying

1. Select the files to move/copy in the file search results
2. Click the "Move Selected" or "Copy Selected" button
3. Select target path:
   - Manual input
   - Browse selection
   - Select from favorite directories
4. Handle conflicts (if target file already exists)
5. Confirm execution

#### 5. Settings Management

1. Click "Settings" in the left navigation
2. Select settings type:
   - **Basic Settings**: Default expiration time, default search path, favorite directory list, service port
   - **Security Settings**: Security password setting (for permanent deletion verification)
   - **SMTP Settings**: Mail server configuration (for sending feedback)
   - **About/Feedback**: Version information, developer information, license information, feedback function

## Troubleshooting

### Common Issues

1. **Browser cannot open automatically**
   - Manually access the address displayed in the terminal (e.g., `http://localhost:3000`)

2. **All ports are occupied**
   - Close programs occupying ports
   - Or modify port settings in the configuration file

3. **Configuration file is corrupted**
   - The system will automatically back up and rebuild the configuration

4. **Insufficient permissions**
   - Files without access permissions will be automatically skipped
   - A "Skipped Files" list will be displayed after the search is completed

5. **Permanent deletion function is unavailable**
   - Security password needs to be set during first startup
   - Or configured in Security Settings in the settings page

## Contribution Guidelines

### Development Environment Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/catch.git
   cd catch
   ```

2. Install frontend dependencies:
   ```bash
   cd web
   npm install
   cd ..
   ```

3. Install backend dependencies:
   ```bash
   go mod download
   ```

### Development Process

1. Create a branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Develop code

3. Run tests:
   ```bash
   # Run backend tests
   go test ./...
   
   # Run frontend tests
   cd web
   npm test
   cd ..
   ```

4. Commit code:
   ```bash
   git add .
   git commit -m "Add your commit message"
   git push origin feature/your-feature-name
   ```

5. Create a Pull Request

### Code Standards

- Backend: Follow Go code standards
- Frontend: Follow Vue code standards
- Commit messages: Use clear, concise descriptions

## License

[MIT License](LICENSE)

## Contact Information

- Feedback email: 15723556393@163.com
- GitHub: [https://github.com/yourusername/catch](https://github.com/yourusername/catch)