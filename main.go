// Copyright 2022 The Ip2Region Authors. All rights reserved.
// Use of this source code is governed by a Apache2.0-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ip2region-web/api"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	port       = flag.Int("port", 8080, "Web服务监听端口")
	staticPath = flag.String("static", "./frontend/dist", "前端静态文件目录")
)

// 设置路由
func setupRouter() *gin.Engine {
	r := gin.Default()

	// 跨域中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 静态文件服务
	if _, err := os.Stat(*staticPath); !os.IsNotExist(err) {
		// 先注册API路由组
		apiGroup := r.Group("/api")
		{
			// IP搜索
			apiGroup.POST("/search", api.SearchIP)

			// 加载XDB文件到内存 - 支持两种路径格式
			apiGroup.POST("/load-xdb", api.LoadXdbToMemory)

			// 获取XDB文件加载状态
			apiGroup.GET("/xdb-status", api.GetXdbStatus)

			// 卸载内存中的XDB文件
			apiGroup.POST("/unload-xdb", api.UnloadXdb)

			// 导出XDB文件到文本文件
			apiGroup.POST("/export-xdb", api.ExportXdb)

			// 获取导出任务状态
			apiGroup.GET("/export-task/:taskId", api.GetExportTaskStatusHandler)

			// 取消导出任务
			apiGroup.POST("/export-task/:taskId/cancel", api.CancelExportTask)

			// 异步生成数据库（带进度显示）
			apiGroup.POST("/generate-with-progress", api.GenerateDbWithProgress)

			// 获取生成任务状态
			apiGroup.GET("/generate-task/:taskId", api.GetGenerateTaskStatusHandler)

			// 取消生成任务
			apiGroup.POST("/generate-task/:taskId/cancel", api.CancelGenerateTask)

			// 数据库生成
			apiGroup.POST("/generate", api.GenerateDb)

			// 查询任务状态（新增）
			apiGroup.GET("/task/:taskId", api.GetTaskStatus)

			// 编辑IP段
			apiGroup.POST("/edit/segment", api.EditSegment)

			// PUT方法编辑IP段
			apiGroup.PUT("/edit/segment", api.EditSegment)

			// 从文件编辑IP段
			apiGroup.POST("/edit/file", api.EditFromFile)

			// 列出IP段
			apiGroup.POST("/list/segments", api.ListSegments)

			// 保存编辑
			apiGroup.POST("/edit/save", api.SaveEdit)

			// 保存编辑并生成xdb文件
			apiGroup.POST("/edit/saveAndGenerate", api.SaveAndGenerateDb)

			// 获取当前编辑的源文件信息
			apiGroup.GET("/edit/current-file", api.GetCurrentEditFile)

			// 卸载当前编辑的源文件
			apiGroup.POST("/edit/unload-file", api.UnloadEditFile)

			// 新增调试接口
			apiGroup.GET("/debug/status", api.GetDebugStatus)
			apiGroup.POST("/force-load-memory", api.ForceLoadToMemory)
		}

		// 然后再设置静态文件服务和NoRoute处理
		// 使用前缀路由而非根路由
		r.Static("/static", *staticPath)

		// 根路径重定向到静态文件目录
		r.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/static/")
		})

		// 配置SPA应用，让所有未匹配的路由都返回index.html
		r.NoRoute(func(c *gin.Context) {
			// 如果请求的是API路径，不处理
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				return
			}

			// 如果是根路径访问，重定向到/static/
			if c.Request.URL.Path == "/" {
				c.Redirect(http.StatusMovedPermanently, "/static/")
				return
			}

			// 其他路径尝试返回index.html
			indexPath := filepath.Join(*staticPath, "index.html")
			c.File(indexPath)
		})
	} else {
		// API路由组 - 当静态文件不存在时仍需要注册API路由
		apiGroup := r.Group("/api")
		{
			// IP搜索
			apiGroup.POST("/search", api.SearchIP)

			// 加载XDB文件到内存 - 支持两种路径格式
			apiGroup.POST("/load-xdb", api.LoadXdbToMemory)

			// 获取XDB文件加载状态
			apiGroup.GET("/xdb-status", api.GetXdbStatus)

			// 卸载内存中的XDB文件
			apiGroup.POST("/unload-xdb", api.UnloadXdb)

			// 导出XDB文件到文本文件
			apiGroup.POST("/export-xdb", api.ExportXdb)

			// 获取导出任务状态
			apiGroup.GET("/export-task/:taskId", api.GetExportTaskStatusHandler)

			// 取消导出任务
			apiGroup.POST("/export-task/:taskId/cancel", api.CancelExportTask)

			// 异步生成数据库
			apiGroup.POST("/generate-with-progress", api.GenerateDbWithProgress)

			// 获取生成任务状态
			apiGroup.GET("/generate-task/:taskId", api.GetGenerateTaskStatusHandler)

			// 取消生成任务
			apiGroup.POST("/generate-task/:taskId/cancel", api.CancelGenerateTask)

			// 数据库生成
			apiGroup.POST("/generate", api.GenerateDb)

			// 查询任务状态
			apiGroup.GET("/task/:taskId", api.GetTaskStatus)

			// 编辑IP段
			apiGroup.POST("/edit/segment", api.EditSegment)

			// PUT方法编辑IP段
			apiGroup.PUT("/edit/segment", api.EditSegment)

			// 从文件编辑IP段
			apiGroup.POST("/edit/file", api.EditFromFile)

			// 列出IP段
			apiGroup.POST("/list/segments", api.ListSegments)

			// 保存编辑
			apiGroup.POST("/edit/save", api.SaveEdit)

			// 保存编辑并生成xdb文件
			apiGroup.POST("/edit/saveAndGenerate", api.SaveAndGenerateDb)

			// 获取当前编辑的源文件信息
			apiGroup.GET("/edit/current-file", api.GetCurrentEditFile)

			// 卸载当前编辑的源文件
			apiGroup.POST("/edit/unload-file", api.UnloadEditFile)

			// 新增调试接口
			apiGroup.GET("/debug/status", api.GetDebugStatus)
			apiGroup.POST("/force-load-memory", api.ForceLoadToMemory)
		}
	}

	return r
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 设置日志格式
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// 设置Gin为release模式，关闭debug输出
	gin.SetMode(gin.ReleaseMode)

	// 创建router
	r := setupRouter()

	// 启动Web服务器
	log.Printf("Starting web server on port %d...\n", *port)
	log.Printf("Static files directory: %s\n", *staticPath)

	err := r.Run(fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("启动Web服务器失败: %v", err)
	}
}
