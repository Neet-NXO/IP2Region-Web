@echo off
:: ip2region-web Windows构建脚本

setlocal enabledelayedexpansion

set APP_NAME=ip2region-web
set FRONTEND_DIR=frontend
set DIST_DIR=%FRONTEND_DIR%\dist

if "%1"=="" goto build-all

if "%1"=="clean" (
    call :clean
) else if "%1"=="build" (
    call :build-backend
) else if "%1"=="build-all" (
    call :build-all
) else if "%1"=="build-backend" (
    call :build-backend
) else if "%1"=="build-frontend" (
    call :build-frontend
) else if "%1"=="deps" (
    call :deps
) else if "%1"=="deps-frontend" (
    call :deps-frontend
) else if "%1"=="dev" (
    call :dev
) else if "%1"=="help" (
    call :help
) else (
    echo 未知命令: %1
    call :help
)
exit /b 0

:check-env
echo 检查构建环境...
go version >nul 2>&1 || (echo 错误: 需要安装Go语言 && exit /b 1)
node --version >nul 2>&1 || (echo 错误: 需要安装Node.js && exit /b 1)
npm --version >nul 2>&1 || (echo 错误: 需要安装npm && exit /b 1)
echo 环境检查通过
exit /b 0

:deps
echo 下载Go依赖...
go mod tidy
go mod download
exit /b 0

:deps-frontend
echo 安装前端依赖...
cd %FRONTEND_DIR%
npm install
cd ..
exit /b 0

:build-backend
echo 构建后端...
go build -ldflags="-s -w" -o %APP_NAME%.exe
echo 后端构建完成: %APP_NAME%.exe
exit /b 0

:build-frontend
echo 构建前端...
cd %FRONTEND_DIR%
npm run build
cd ..
echo 前端构建完成: %DIST_DIR%
exit /b 0

:build-all
call :check-env
call :deps
call :deps-frontend
call :build-frontend
call :build-backend
echo.
echo ====================================
echo 构建完成: %APP_NAME%.exe
echo 前端文件: %DIST_DIR%
echo 使用方法: %APP_NAME%.exe -port=8080 -static=.\%DIST_DIR%
echo ====================================
exit /b 0

:dev
call :deps
echo 启动开发模式...
go run main.go
exit /b 0

:clean
echo 清理构建产物...
if exist %APP_NAME%.exe del /f /q %APP_NAME%.exe
if exist %DIST_DIR% rmdir /s /q %DIST_DIR%
echo 清理完成
exit /b 0

:help
echo.
echo 可用的构建命令:
echo   build-all        - 完整构建（默认）
echo   build            - 仅构建后端
echo   build-backend    - 仅构建后端
echo   build-frontend   - 仅构建前端
echo   deps             - 下载Go依赖
echo   deps-frontend    - 安装前端依赖
echo   dev              - 开发模式
echo   clean            - 清理构建产物
echo   help             - 显示此帮助信息
echo.
echo 示例:
echo   make.bat build-all
echo   make.bat clean
echo   make.bat dev
exit /b 0
