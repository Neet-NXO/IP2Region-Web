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

- 🔍 **IP查询**: 快速查询IP地址的地理位置信息
- 📊 **数据管理**: 支持XDB文件的加载、卸载和内存管理
- 🛠️ **数据编辑**: 支持IP段数据的编辑和修改
- 🏗️ **数据库生成**: 从文本文件生成XDB数据库文件
- 📤 **数据导出**: 支持XDB文件导出为文本格式
- ⚡ **高性能**: 毫秒级查询响应，支持内存加载
- 📱 **现代界面**: 基于Vue 3 + Element Plus的响应式Web界面
- 🎯 **任务管理**: 支持异步任务处理和进度显示

## 🛠 技术栈

### 后端
- **语言**: Go 1.24
- **框架**: Gin Web框架
- **核心**: ip2region XDB数据库引擎
- **功能**: RESTful API、文件处理、异步任务

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

### 1. IP查询
- 在"IP查询"页面输入IP地址即可查询地理位置信息
- 支持IPv4地址查询
- 显示国家、省份、城市、运营商等信息

### 2. 数据库管理
- **加载XDB**: 将XDB文件加载到内存以提高查询速度
- **卸载XDB**: 释放内存中的XDB数据
- **状态查看**: 查看当前XDB文件的加载状态

### 3. 数据编辑
- **编辑IP段**: 修改特定IP段的地理位置信息
- **批量编辑**: 从文件导入IP段进行批量编辑
- **保存更改**: 保存编辑内容并可选择生成新的XDB文件

### 4. 数据库生成
- **从文本生成**: 将文本格式的IP数据生成为XDB数据库文件
- **进度监控**: 实时查看生成进度
- **任务管理**: 支持取消正在进行的生成任务

### 5. 数据导出
- **导出为文本**: 将XDB文件导出为文本格式
- **异步处理**: 大文件导出支持后台处理
- **进度跟踪**: 实时显示导出进度

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
- `POST /api/search` - IP地址查询

### 数据库管理
- `POST /api/load-xdb` - 加载XDB文件到内存
- `POST /api/unload-xdb` - 卸载内存中的XDB文件
- `GET /api/status` - 获取XDB加载状态

### 数据编辑
- `POST /api/edit/segment` - 编辑IP段
- `POST /api/edit/file` - 从文件编辑IP段
- `POST /api/list/segments` - 列出IP段
- `POST /api/edit/save` - 保存编辑
- `POST /api/edit/saveAndGenerate` - 保存并生成XDB

### 数据生成与导出
- `POST /api/generate` - 生成数据库
- `POST /api/generate-with-progress` - 异步生成数据库
- `POST /api/export-xdb` - 导出XDB文件
- `GET /api/task/:taskId` - 查询任务状态

## 📊 性能指标

- **查询响应时间**: < 10毫秒 (内存模式)
- **并发处理能力**: 高并发支持
- **内存占用**: 根据XDB文件大小调整
- **数据库大小**: 约94MB (标准ip2region.xdb)

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
