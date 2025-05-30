# IP2Region Web

<div align="center">

![IP2Region Web](https://img.shields.io/badge/IP2Region-Web-blue)
![Go Version](https://img.shields.io/badge/Go-1.24+-green)
![Vue Version](https://img.shields.io/badge/Vue-3.0+-brightgreen)
![License](https://img.shields.io/badge/License-Apache%202.0-blue)
![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen)

一个现代化的 IP 地理位置查询 Web 应用，基于 ip2region 数据库构建

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

## ✨ 特性

- 🚀 **快速查询**: 毫秒级 IP 地理位置查询
- 🌐 **现代界面**: 基于 Vue 3 + TypeScript 的响应式前端
- 🎯 **高性能**: Go 语言构建的高性能后端

## 🛠 技术栈

### 后端
- **语言**: Go 1.24+
- **框架**: Gin (HTTP 路由)
### 前端
- **框架**: Vue 3 + TypeScript
- **构建工具**: Vite
- **UI 库**: Element Plus
- **路由**: Vue Router 4
- **HTTP 客户端**: Axios

## 📦 快速开始

### 环境要求

- Go 1.24+
- Node.js 18+
- MySQL 8.0+ (可选，也可使用 SQLite)

### 安装与运行

#### 1. 克隆项目
```bash
git clone https://github.com/Neet-NXO/IP2Region-Web.git
cd IP2Region-Web
```

#### 2. 后端启动
```bash
# 下载依赖
go mod download

# 复制配置文件
cp config.example.yaml config.yaml

# 编辑配置文件
vim config.yaml

# 运行项目
go run main.go
```

#### 3. 前端启动
```bash
# 进入前端目录
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
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


## 🔧 开发指南

### 项目结构
```
IP2Region-Web/
├── api/                    # API 接口层
│   └── handler.go          # HTTP 处理器
├── xdb/                    # IP2Region XDB 核心包
│   ├── searcher.go         # IP 查询器
│   ├── maker.go            # XDB 数据生成器
│   ├── editor.go           # 数据编辑器
│   ├── segment.go          # 数据段处理
│   ├── util.go             # 工具函数
│   └── index.go            # 索引处理
├── frontend/               # 前端代码
│   ├── src/
│   │   ├── api/            # API 接口
│   │   ├── assets/         # 静态资源
│   │   ├── router/         # 路由配置
│   │   ├── views/          # 页面组件
│   │   ├── App.vue         # 根组件
│   │   └── main.js         # 入口文件
│   ├── dist/               # 构建输出
│   ├── index.html          # HTML 模板
│   ├── package.json        # 依赖配置
│   └── vite.config.js      # Vite 配置
├── data/                   # 数据文件 (*.xdb)
├── go.mod                  # Go 模块配置
├── go.sum                  # Go 依赖校验
├── main.go                 # 主程序入口
├── Makefile                # Linux/macOS 构建脚本
├── make.bat                # Windows 构建脚本
├── LICENSE                 # 许可证文件
├── ReadMe.md               # 项目说明
└── .gitignore              # Git 忽略配置
```

### 贡献指南

我们欢迎所有形式的贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详细信息。

#### 开发流程
1. Fork 项目
2. 创建特性分支
3. 提交代码
4. 创建 Pull Request

#### 代码规范
- 后端遵循 Go 标准代码风格
- 前端使用 ESLint + Prettier
- 提交信息遵循 Conventional Commits

## 📊 性能指标

- **查询响应时间**: < 10微秒 (本地查询)
- **并发处理能力**: 100,000+ QPS
- **内存占用**: < 100MB (基础运行)
- **数据库大小**: ~30MB (ip2region.xdb)

## 🤝 贡献者

感谢所有为项目做出贡献的开发者！

<a href="https://github.com/Neet-NXO/IP2Region-Web/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=Neet-NXO/IP2Region-Web" />
</a>

## 📄 许可证

本项目基于 [Apache License 2.0](LICENSE) 许可证开源。

## 🔗 相关链接

- **[ip2region 原始项目](https://github.com/lionsoul2014/ip2region)** - 离线IP地址定位库和数据管理框架
- [Go 官方文档](https://golang.org/doc/)
- [Vue 3 文档](https://v3.vuejs.org/)
- [Element Plus](https://element-plus.org/)

## 📮 联系我们

- 提交 Issue: [GitHub Issues](https://github.com/Neet-NXO/IP2Region-Web/issues)
- GitHub: [@Neet-NXO](https://github.com/Neet-NXO)

---

<div align="center">
  Made with ❤️ by IP2Region Web Contributors
</div>
