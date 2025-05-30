# IP2Region Web

<div align="center">

![IP2Region Web](https://img.shields.io/badge/IP2Region-Web-blue)
![Go Version](https://img.shields.io/badge/Go-1.24+-green)
![Vue Version](https://img.shields.io/badge/Vue-3.0+-brightgreen)
![License](https://img.shields.io/badge/License-Apache%202.0-blue)

一个现代化的 IP 地理位置查询和管理 Web 应用，基于 ip2region 数据库构建

</div>

## 🙏 致谢

> **本项目基于 [ip2region](https://github.com/lionsoul2014/ip2region) 构建**
> 
> 感谢 [@lionsoul2014](https://github.com/lionsoul2014) 及其团队开发的优秀 ip2region 离线IP地址定位库。
> 
> ip2region 是一个离线IP地址定位库和IP定位数据管理框架，支持亿级别数据段，十微秒级查询性能。
> 
> - 原始项目: https://github.com/lionsoul2014/ip2region
> - 许可证: Apache License 2.0

## ✨ 功能特性

- 🔍 **IP查询**: 
    - 支持多种搜索模式：文件直查 (`file`)、向量索引缓存 (`vector`)、完全内存加载 (`memory`)，用户可按需选择。
    - 快速准确地查询IP地址的地理位置信息。
- 📊 **XDB数据库管理**:
    - 支持XDB文件的动态加载、卸载，并能查看当前加载状态。
    - 智能管理内存和文件句柄，提供强制内存加载选项。
- 🛠️ **高级数据编辑**:
    - 支持对IP段数据进行在线编辑和修改。
    - 支持从文件批量导入IP段数据进行编辑。
    - 编辑后可直接保存并生成新的XDB数据库文件。
- 🏗️ **异步数据库生成**: 
    - 支持从文本格式的IP数据源异步生成XDB数据库文件。
    - 提供任务ID，支持查询生成进度和状态，可取消生成任务。
- 📤 **异步数据导出**: 
    - 支持将XDB文件内容异步导出为文本格式。
    - 提供任务ID，支持查询导出进度（已处理IP数、IP段数）、状态，可取消导出任务。
- ⚡ **高性能与监控**:
    - 毫秒级甚至微秒级的查询响应（取决于加载模式）。
    - 提供全局搜索统计（总搜索次数、错误次数、IO操作数）。
    - API响应包含纳秒级查询耗时。
    - 提供调试接口 (`/api/debug/status`)，帮助诊断和监控服务状态。
- 📱 **现代化Web界面**: 
    - 基于 Vue 3 + Element Plus 构建，提供响应式和用户友好的操作体验。
    - 实时显示任务（导出/生成）进度。
- 🛡️ **并发安全与稳定性**:
    - 后端采用线程安全设计，使用`sync.RWMutex`和`atomic`操作保护共享资源。
    - 细致的错误处理和日志记录。

## 🛠 技术栈

### 后端
- **语言**: Go 1.24
- **框架**: Gin Web框架
- **核心**: ip2region XDB数据库引擎
- **并发处理**: 使用`sync.RWMutex`和`atomic`包保证并发安全
- **内存管理**: 智能管理XDB文件加载（文件、向量索引、完全内存），支持垃圾回收优化
- **功能**: RESTful API、文件处理、异步任务队列、性能监控

### 前端
- **框架**: Vue 3 (Composition API)
- **UI库**: Element Plus
- **构建工具**: Vite
- **路由**: Vue Router 4
- **HTTP客户端**: Axios
- **状态管理**: Pinia

## 📦 快速开始

### 环境要求

- Go 1.24+
- Node.js 18+

### 安装与运行

#### 1. 克隆项目
```bash
git clone <repository-url>
cd ip2region-web
```

#### 2. 后端启动
```bash
# 安装Go依赖
go mod download

# 启动后端服务 (默认端口8080)
go run main.go

# 自定义端口启动
go run main.go -port=9000

# 自定义静态文件目录
go run main.go -static=./frontend/dist
```

#### 3. 前端开发
```bash
# 进入前端目录
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建生产版本
npm run build
```

#### 4. 使用构建脚本

**Linux/macOS:**
```bash
# 检查环境
make check

# 完整构建
make build

# 运行项目
make run
```

**Windows:**
```batch
# 检查环境
make.bat check

# 完整构建
make.bat build

# 运行项目
make.bat run
```

## 🚀 使用说明

本Web应用提供了直观的界面来操作IP2Region的各项功能。

### 1. XDB数据库加载与管理 (首页)
- **选择XDB文件**: 在首页输入您的 `ip2region.xdb` 文件路径 (例如: `./ip2region.xdb` 或绝对路径)。
- **选择加载模式**:
    - **向量模式 (推荐)**: 将XDB文件的向量索引加载到内存。这是性能和内存占用的良好平衡点。适用于大多数生产环境。
    - **内存模式**: 将整个XDB文件加载到内存。提供最佳查询性能，但会消耗更多内存。适用于对查询速度有极致要求的场景。
    - **文件模式 (通过API)**: API `/api/search` 在请求时可以指定 `searchMode: "file"` 和 `dbPath`。这种模式不将数据常驻内存，每次查询都会读文件，适合内存极其有限或不常查询的场景。
- **加载/卸载**: 点击 "加载数据库" 将选定的XDB文件按选定模式加载。加载成功后，按钮会变为 "卸载数据库"。
- **状态查看**: 加载成功后，会显示当前加载模式、内存占用、向量索引等信息。也可以通过 `/api/xdb-status` 接口获取详细状态和统计信息。
- **强制加载到内存**: 若遇到加载问题或需要确保最佳性能，可以通过 `POST /api/force-load-memory` 接口（请求体包含 `dbPath`）强制将指定XDB文件以完全内存模式加载。

### 2. IP查询 (IP查询页面 / API)
- **界面查询**: 在 "IP查询" 页面输入IP地址，系统会使用当前已加载的XDB数据库进行查询。
- **API查询**: 
    - 若已有加载的XDB (向量/内存模式)，直接调用 `POST /api/search` 并提供 `ip` 参数。
    - 若要使用特定的XDB文件或文件模式查询，调用 `POST /api/search` 时需额外提供 `dbPath` 和 `searchMode: "file"` 参数。
- **结果**: 显示国家、省份、城市、运营商等信息，以及查询耗时 (纳秒级)。

### 3. 数据库生成 (生成数据库页面 / API)
- **界面操作**: 
    1. 访问 "生成数据库" 页面。
    2. 输入源文本文件路径 (包含IP段和区域信息，每行格式通常为 `IP段|区域信息` 或 `起始IP|结束IP|区域信息`)。
    3. 输入目标XDB文件路径 (例如: `./new_ip2region.xdb`)。
    4. 点击 "开始生成"。生成过程为异步，会显示任务ID和进度条。
- **API操作**: 使用 `POST /api/generate-with-progress` 接口，请求体包含 `srcFile` 和 `dstFile`。
- **进度与取消**: 通过 `GET /api/generate-task/:taskId` 查看进度，通过 `POST /api/generate-task/:taskId/cancel` 取消任务。

### 4. 数据编辑 (编辑数据页面 / API)
- **加载源文件**: 在 "编辑数据" 页面，首先需要通过 `POST /api/edit/file` (请求体包含 `file` 指向源文本文件路径，`srcFile` 可用于临时文件名) 或在前端界面选择并上传源文本文件 (通常是用于生成XDB的原始IP段数据文件)。成功后，服务器会缓存此文件用于后续编辑。
- **编辑操作**:
    - **列出IP段**: 使用 `POST /api/list/segments` (请求体包含 `srcFile` 和分页参数 `offset`, `size`) 查看和搜索源文件中的IP段。
    - **修改IP段**: 使用 `POST /api/edit/segment` (请求体包含 `segment` 如 `1.2.3.4|中国|广东|深圳|电信`, 和 `srcFile`) 或 `PUT /api/edit/segment` 来修改单个IP段。前端界面通常会简化此操作。
- **保存更改**:
    - `POST /api/edit/save` (请求体包含 `srcFile`): 仅保存对当前编辑的源文本文件的修改到服务器缓存的路径。
    - `POST /api/edit/saveAndGenerate` (请求体包含 `srcFile` 和 `dstFile`): 保存修改到源文件，并立即使用修改后的源文件生成新的XDB数据库到 `dstFile`。
- **状态管理**: 
    - `GET /api/edit/current-file`: 查看当前服务器正在编辑的源文件信息。
    - `POST /api/edit/unload-file` (请求体包含 `srcFile`): 清除服务器当前编辑的源文件状态，放弃未保存的更改。

### 5. 数据导出 (首页 / API)
- **界面操作** (首页加载XDB后出现导出按钮):
    1. 确保已加载一个XDB文件。
    2. 点击 "导出XDB" 按钮。
    3. 在弹窗中指定导出的文本文件路径 (例如: `ip2region_export.txt`)。
    4. 点击 "导出"。导出过程为异步，会显示任务ID和进度条。
- **API操作**: 使用 `POST /api/export-xdb` 接口，请求体包含 `xdbPath` (要导出的XDB文件) 和 `exportPath` (目标文本文件)。
- **进度与取消**: 通过 `GET /api/export-task/:taskId` 查看进度，通过 `POST /api/export-task/:taskId/cancel` 取消任务。

### 6. 监控与调试
- **常规状态**: `GET /api/xdb-status` 提供基础的加载状态和搜索统计信息。
- **详细调试**: `GET /api/debug/status` 提供更深度的内部状态信息，包括加载器详情、内存模式状态、向量索引详情等，用于问题排查和性能分析。

## 🏗 项目结构

```
ip2region-web/
├── main.go                 # 程序入口
├── go.mod                  # Go模块配置
├── go.sum                  # Go依赖锁定
├── LICENSE                 # 许可证文件
├── ReadMe.md              # 项目说明
├── CHANGELOG.md           # 更新日志
├── .gitignore             # Git忽略文件
├── Makefile               # Linux/macOS构建脚本
├── make.bat               # Windows构建脚本
├── api/                   # API接口层
│   └── handler.go         # HTTP请求处理器
├── xdb/                   # IP2Region核心包
│   ├── searcher.go        # IP查询引擎
│   ├── maker.go           # XDB文件生成器
│   ├── editor.go          # 数据编辑器
│   ├── segment.go         # IP段处理
│   ├── util.go            # 工具函数
│   └── index.go           # 索引处理
├── frontend/              # 前端代码
│   ├── src/
│   │   ├── views/         # 页面组件
│   │   │   ├── Home.vue   # 首页
│   │   │   ├── Search.vue # IP查询页
│   │   │   ├── Edit.vue   # 数据编辑页
│   │   │   └── Generate.vue # 数据库生成页
│   │   ├── router/        # 路由配置
│   │   ├── api/           # API接口封装
│   │   ├── assets/        # 静态资源
│   │   ├── App.vue        # 根组件
│   │   └── main.js        # 入口文件
│   ├── dist/              # 构建输出目录
│   ├── index.html         # HTML模板
│   ├── package.json       # 前端依赖配置
│   └── vite.config.js     # Vite配置
├── *.xdb                  # XDB数据库文件
└── *.txt                  # 文本格式IP数据
```

## 🔧 API接口

### IP查询
- `POST /api/search` - IP地址查询 (支持指定 `dbPath` 和 `searchMode`)

### XDB数据库管理
- `POST /api/load-xdb` - 加载XDB文件到指定模式 (vector/memory)
- `POST /api/unload-xdb` - 卸载当前加载的XDB文件
- `GET /api/xdb-status` - 获取当前XDB加载状态和统计信息
- `POST /api/force-load-memory` - 强制重新加载XDB文件到完全内存模式

### 数据编辑
- `POST /api/edit/segment` - 编辑指定源文件的IP段
- `POST /api/edit/file` - 从上传的文件内容编辑IP段 (指定源文件)
- `POST /api/list/segments` - 列出指定源文件的IP段 (支持分页)
- `POST /api/edit/save` - 保存对指定源文件的编辑
- `POST /api/edit/saveAndGenerate` - 保存编辑并生成新的XDB文件
- `GET /api/edit/current-file` - 获取当前正在编辑的源文件信息
- `POST /api/edit/unload-file` - 卸载当前编辑的源文件，放弃未保存的更改

### 异步任务：数据生成与导出
- `POST /api/generate-with-progress` - 异步生成XDB数据库文件
- `GET /api/generate-task/:taskId` - 获取数据库生成任务的状态和进度
- `POST /api/generate-task/:taskId/cancel` - 取消正在进行的数据库生成任务
- `POST /api/export-xdb` - 异步导出XDB文件为文本格式
- `GET /api/export-task/:taskId` - 获取数据导出任务的状态和进度
- `POST /api/export-task/:taskId/cancel` - 取消正在进行的数据导出任务
- `GET /api/task/:taskId` - (通用)查询任务状态 (可用于检查xdb.Maker内部任务状态)

### 调试与监控
- `GET /api/debug/status` - 获取详细的调试状态信息 (内存、加载器、向量索引等)

## 📊 性能指标

- **查询响应时间**:
    - **内存模式 (`memory`)**: 通常在10微秒以内，接近内存直接访问速度。
    - **向量模式 (`vector`)**: 通常在几十微秒到几百微秒，具体取决于CPU和内存速度。
    - **文件模式 (`file`)**: 毫秒级，受磁盘I/O限制。
    - API提供纳秒级 (`tookNanoseconds`) 耗时统计。
- **并发处理能力**: 设计支持高并发，具体取决于服务器硬件配置。
- **内存占用**:
    - **文件模式**: 极低，仅缓存少量元数据。
    - **向量模式**: 数MB到数十MB（取决于XDB文件中的向量索引大小）。
    - **内存模式**: 等同于XDB文件大小 (例如，标准ip2region.xdb约94MB)。
- **数据库大小**: 标准 `ip2region.xdb` 文件约 94MB。
- **任务处理**: 异步任务（生成/导出）性能取决于磁盘I/O和CPU处理能力。

## 🛠 开发指南

### 命令行参数
```bash
go run main.go -port=8080 -static=./frontend/dist
```

- `-port`: Web服务监听端口 (默认: 8080)
- `-static`: 前端静态文件目录 (默认: ./frontend/dist)

### 构建部署
```bash
# 构建前端
cd frontend && npm run build

# 构建后端
go build -o ip2region-web main.go

# 运行
./ip2region-web -port=8080
```

## 📄 许可证

本项目基于 [Apache License 2.0](LICENSE) 许可证开源。

## 🔗 相关链接

- **[ip2region 原始项目](https://github.com/lionsoul2014/ip2region)** - 离线IP地址定位库
- [Go 官方文档](https://golang.org/doc/)
- [Vue 3 文档](https://v3.vuejs.org/)
- [Element Plus](https://element-plus.org/)

---

<div align="center">
  Made with ❤️ by IP2Region Web Contributors
</div>
