// Copyright 2022 The Ip2Region Authors. All rights reserved.
// Use of this source code is governed by a Apache2.0-style
// license that can be found in the LICENSE file.

package api

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	_ "unsafe" // 用于go:linkname

	"ip2region-web/xdb"

	"github.com/gin-gonic/gin"
)

// 标准响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// IP查询请求
type SearchRequest struct {
	IP         string `json:"ip" binding:"required"`
	DbPath     string `json:"dbPath,omitempty"`     // 可选的数据库文件路径
	SearchMode string `json:"searchMode,omitempty"` // 查询模式：file, vector, memory
}

// 加载XDB文件到内存请求
type LoadXdbRequest struct {
	DbPath     string `json:"dbPath" binding:"required"`
	SearchMode string `json:"searchMode" binding:"required"` // 查询模式：vector, memory
}

// 加载XDB文件结果
type LoadXdbResult struct {
	DbPath        string `json:"dbPath"`
	SearchMode    string `json:"searchMode"` // 当前加载的模式
	InMemoryMode  bool   `json:"inMemoryMode"`
	BufferSizeKB  int64  `json:"bufferSizeKB"`
	VectorLoaded  bool   `json:"vectorLoaded"`
	VectorSizeKB  int    `json:"vectorSizeKB"`
	LoadTimeTaken string `json:"loadTimeTaken"`
}

// IP查询结果
type SearchResult struct {
	Region          string `json:"region"`
	IoCount         int    `json:"ioCount"`
	TookNanoseconds int64  `json:"tookNanoseconds"` // 纳秒级精度的查询耗时
	SearchMode      string `json:"searchMode"`      // 使用的查询模式
	QueryTime       string `json:"queryTime"`       // 新增：查询完成时的服务器时间
}

// 数据库生成请求
type GenDbRequest struct {
	SrcFile string `json:"srcFile" binding:"required"`
	DstFile string `json:"dstFile" binding:"required"`
}

// 编辑IP段请求
type EditSegmentRequest struct {
	Segment string `json:"segment" binding:"required"`
	SrcFile string `json:"srcFile" binding:"required"`
}

// 编辑文件请求
type EditFileRequest struct {
	File    string `json:"file" binding:"required"`
	SrcFile string `json:"srcFile" binding:"required"`
}

// 查看IP段请求
type ListSegmentsRequest struct {
	Offset  int    `json:"offset"`
	Size    int    `json:"size"`
	SrcFile string `json:"srcFile" binding:"required"`
}

// 保存编辑请求
type SaveEditRequest struct {
	SrcFile string `json:"srcFile" binding:"required"`
}

// 保存编辑并生成数据库请求
type SaveAndGenerateRequest struct {
	SrcFile string `json:"srcFile" binding:"required"`
	DstFile string `json:"dstFile" binding:"required"`
}

// 单个IP查询修改请求
type SingleIPRequest struct {
	IP     string `json:"ip" binding:"required"`
	DbPath string `json:"dbPath" binding:"required"`
}

// IP查询修改结果
type IPModifyResult struct {
	IP          string   `json:"ip"`
	StartIP     string   `json:"startIP"`
	EndIP       string   `json:"endIP"`
	Region      string   `json:"region"`
	RegionParts []string `json:"regionParts"`
	OrigSegment string   `json:"origSegment"`
}

// 修改IP数据请求
type ModifyIPRequest struct {
	OrigSegment string `json:"origSegment" binding:"required"`
	NewSegment  string `json:"newSegment" binding:"required"`
	SrcFile     string `json:"srcFile" binding:"required"`
	DbPath      string `json:"dbPath" binding:"required"`
}

// 使用atomic操作优化的全局变量
var (
	searcher     *xdb.Searcher
	searcherPath string
	searcherMode string       // 当前搜索器模式：file, vector, memory
	inMemoryMode int32        // 使用atomic操作，0表示false，1表示true
	searcherLock sync.RWMutex // 保护searcher和searcherPath的读写锁
)

// 全局编辑文件路径（使用atomic.Value保护）
var (
	currentEditFilePath atomic.Value // 存储string类型
)

// IP段结构体定义（全局）
type IPSegment struct {
	StartIP uint32
	EndIP   uint32
	Region  string
}

// 性能统计结构体（使用atomic计数器）
type SearchStats struct {
	totalSearches     int64 // 总搜索次数
	totalErrors       int64 // 总错误次数
	totalIoOperations int64 // 总IO操作次数
}

var globalStats SearchStats

// GetSearchStats 获取搜索统计信息
func GetSearchStats() (searches, errors, ioOps int64) {
	return atomic.LoadInt64(&globalStats.totalSearches),
		atomic.LoadInt64(&globalStats.totalErrors),
		atomic.LoadInt64(&globalStats.totalIoOperations)
}

// 获取或创建指定模式的搜索器
func getSearcherByMode(dbPath string, mode string) (*xdb.Searcher, error) {
	// 文件模式不使用全局缓存，应该由调用方自己管理生命周期
	if mode == "file" {
		return xdb.NewWithFileOnly(dbPath)
	}

	// 先使用读锁检查（仅限向量和内存模式）
	searcherLock.RLock()
	if searcherPath == dbPath && searcher != nil && searcherMode == mode {
		searcherLock.RUnlock()
		return searcher, nil
	}
	searcherLock.RUnlock()

	// 需要创建新的搜索器，使用写锁
	searcherLock.Lock()
	defer searcherLock.Unlock()

	// 双重检查锁定模式
	if searcherPath == dbPath && searcher != nil && searcherMode == mode {
		return searcher, nil
	}

	// 关闭现有的搜索器
	if searcher != nil {
		searcher.Close()
		searcher = nil
		searcherPath = ""
		searcherMode = ""
		atomic.StoreInt32(&inMemoryMode, 0)
	}

	// 根据模式创建新的搜索器（排除文件模式）
	var err error
	switch mode {
	case "vector":
		searcher, err = xdb.NewSearcherWithVectorIndex(dbPath)
	case "memory":
		searcher, err = xdb.NewSearcherWithMemoryMode(dbPath)
	default:
		return nil, fmt.Errorf("不支持的搜索模式: %s", mode)
	}

	if err != nil {
		return nil, err
	}

	// 设置全局变量
	searcherPath = dbPath
	searcherMode = mode
	if searcher.IsMemoryMode() {
		atomic.StoreInt32(&inMemoryMode, 1)
	} else {
		atomic.StoreInt32(&inMemoryMode, 0)
	}

	return searcher, nil
}

// 加载XDB文件到内存
func LoadXdbToMemory(c *gin.Context) {
	var req LoadXdbRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证搜索模式
	if req.SearchMode != "vector" && req.SearchMode != "memory" {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "不支持的搜索模式，只支持: vector, memory",
		})
		return
	}

	// 开始计时
	tStart := time.Now()

	// 根据模式加载搜索器
	s, err := getSearcherByMode(req.DbPath, req.SearchMode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "加载XDB文件失败: " + err.Error(),
		})
		return
	}

	// 获取加载结果信息
	result := LoadXdbResult{
		DbPath:        req.DbPath,
		SearchMode:    req.SearchMode,
		InMemoryMode:  s.IsMemoryMode(),
		BufferSizeKB:  s.GetContentBufferSize() / 1024,
		VectorLoaded:  s.IsVectorIndexLoaded(),
		VectorSizeKB:  s.GetVectorIndexSize() / 1024,
		LoadTimeTaken: time.Since(tStart).String(),
	}

	var modeDesc string
	switch req.SearchMode {
	case "vector":
		modeDesc = "向量索引模式"
	case "memory":
		modeDesc = "完全内存模式"
	}

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  fmt.Sprintf("XDB文件已加载到%s", modeDesc),
		Data: result,
	})
}

// 卸载内存中的XDB文件
func UnloadXdb(c *gin.Context) {
	searcherLock.Lock()
	defer searcherLock.Unlock()

	if searcher != nil {
		searcher.Close()
	}

	searcher = nil
	searcherPath = ""
	atomic.StoreInt32(&inMemoryMode, 0)

	// 强制垃圾回收，确保释放文件句柄
	runtime.GC()

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "XDB文件已卸载",
	})
}

// 获取XDB文件加载状态
func GetXdbStatus(c *gin.Context) {
	searcherLock.RLock()
	defer searcherLock.RUnlock()

	status := map[string]interface{}{
		"loaded":      false,
		"dbPath":      "",
		"searchMode":  "",
		"inMemory":    false,
		"vectorIndex": false,
		"bufferSize":  int64(0),
		"vectorSize":  0,
	}

	// 只有向量模式和内存模式才显示为已加载状态
	// 文件模式不保持加载状态，因为它是用完即关的
	if searcher != nil && (searcherMode == "vector" || searcherMode == "memory") {
		status["loaded"] = true
		status["dbPath"] = searcherPath
		status["searchMode"] = searcherMode
		status["inMemory"] = atomic.LoadInt32(&inMemoryMode) == 1
		status["vectorIndex"] = searcher.IsVectorIndexLoaded()
		status["bufferSize"] = searcher.GetContentBufferSize()
		status["vectorSize"] = searcher.GetVectorIndexSize()
	}

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "获取状态成功",
		Data: status,
	})
}

// SearchIP 搜索IP地址信息
func SearchIP(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		atomic.AddInt64(&globalStats.totalErrors, 1)
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	// 增加搜索计数
	atomic.AddInt64(&globalStats.totalSearches, 1)

	result, err := SearchIPFunc(req.IP, req.DbPath, req.SearchMode)
	if err != nil {
		atomic.AddInt64(&globalStats.totalErrors, 1)
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "搜索失败: " + err.Error(),
		})
		return
	}

	// 增加IO操作计数
	atomic.AddInt64(&globalStats.totalIoOperations, int64(result.IoCount))

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "搜索成功",
		Data: result,
	})
}

// SearchIPFunc 内部IP搜索函数
func SearchIPFunc(ip string, dbPath string, searchMode string) (*SearchResult, error) {
	var s *xdb.Searcher
	var err error
	var usedMode string
	var shouldCloseSearcher bool = false // 标记是否需要在函数结束时关闭searcher

	// 如果是文件模式，每次都创建新的searcher，用完即关
	if searchMode == "file" {
		if dbPath == "" {
			return nil, fmt.Errorf("文件模式需要指定数据库文件路径")
		}

		s, err = xdb.NewWithFileOnly(dbPath)
		if err != nil {
			return nil, fmt.Errorf("加载数据库失败: %s", err.Error())
		}
		usedMode = "file"
		shouldCloseSearcher = true // 文件模式需要关闭
	} else {
		// 对于向量模式和内存模式，先检查是否有已加载的数据库可以使用
		searcherLock.RLock()
		hasLoadedSearcher := searcher != nil
		loadedPath := searcherPath
		loadedMode := searcherMode
		searcherLock.RUnlock()

		// 优先使用已加载的数据库（仅限于向量和内存模式）
		if hasLoadedSearcher && (dbPath == "" || dbPath == loadedPath) && (loadedMode == "vector" || loadedMode == "memory") {
			// 如果未指定数据库路径，或指定的路径与已加载的相同，且已加载的是向量或内存模式
			searcherLock.RLock()
			if searcher != nil {
				s = searcher
				usedMode = loadedMode
				searcherLock.RUnlock()
			} else {
				searcherLock.RUnlock()
				return nil, fmt.Errorf("数据库连接已断开，请重新加载")
			}
		} else if dbPath == "" {
			// 如果未指定数据库路径且没有已加载的数据库
			return nil, fmt.Errorf("未指定数据库文件，且没有加载数据库")
		} else {
			// 需要加载指定路径的数据库
			if searchMode == "" {
				searchMode = "file" // 默认使用文件模式
			}

			// 验证搜索模式
			if searchMode != "file" && searchMode != "vector" && searchMode != "memory" {
				return nil, fmt.Errorf("不支持的搜索模式: %s，支持的模式: file, vector, memory", searchMode)
			}

			// 如果是文件模式，创建临时searcher
			if searchMode == "file" {
				s, err = xdb.NewWithFileOnly(dbPath)
				if err != nil {
					return nil, fmt.Errorf("加载数据库失败: %s", err.Error())
				}
				usedMode = "file"
				shouldCloseSearcher = true
			} else {
				// 向量模式和内存模式使用全局缓存
				s, err = getSearcherByMode(dbPath, searchMode)
				if err != nil {
					return nil, fmt.Errorf("加载数据库失败: %s", err.Error())
				}
				usedMode = searchMode
			}
		}
	}

	// 确保文件模式的searcher在函数结束时被关闭
	if shouldCloseSearcher {
		defer func() {
			if s != nil {
				s.Close()
			}
		}()
	}

	// 检查和转换IP
	ipUint32, err := xdb.IP2Long(ip)
	if err != nil {
		return nil, fmt.Errorf("无效的IP地址: %s", err.Error())
	}

	startTime := time.Now().UnixNano()
	region, ioCount, err := s.Search(ipUint32)
	endTime := time.Now().UnixNano()
	elapsed := endTime - startTime

	if err != nil {
		return nil, fmt.Errorf("搜索失败: %s", err.Error())
	}

	return &SearchResult{
		Region:          region,
		IoCount:         ioCount,
		TookNanoseconds: elapsed,
		SearchMode:      usedMode,
		QueryTime:       time.Now().Format("2006/01/02 15:04:05"),
	}, nil
}

// 生成数据库
func GenerateDb(c *gin.Context) {
	var req GenDbRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查源文件是否存在
	if _, err := os.Stat(req.SrcFile); os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "源文件不存在: " + req.SrcFile,
		})
		return
	}

	// 确保目标目录存在
	dstDir := filepath.Dir(req.DstFile)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "创建目标目录失败: " + err.Error(),
		})
		return
	}

	// 创建数据库生成器
	tStart := time.Now()
	maker, err := xdb.NewMaker(xdb.VectorIndexPolicy, req.SrcFile, req.DstFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "创建生成器失败: " + err.Error(),
		})
		return
	}
	defer maker.Close()

	// 初始化
	if err := maker.Init(); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "初始化失败: " + err.Error(),
			Data: nil,
		})
		return
	}

	// 开始处理
	if err := maker.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "处理失败: " + err.Error(),
			Data: nil,
		})
		return
	}

	// 结束处理
	if err := maker.End(); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "结束处理失败: " + err.Error(),
			Data: nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "生成成功",
		Data: gin.H{
			"elapsed": time.Since(tStart).String(),
			"srcFile": req.SrcFile,
			"dstFile": req.DstFile,
		},
	})
}

// 编辑器实例缓存
var (
	editors     = make(map[string]*xdb.Editor)
	editorsLock sync.RWMutex
)

// 获取编辑器实例
func getEditor(srcFile string) (*xdb.Editor, error) {
	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在: %s", srcFile)
	}

	// 先使用读锁检查
	editorsLock.RLock()
	if editor, ok := editors[srcFile]; ok {
		if editor.IsHandleValid() {
			editorsLock.RUnlock()
			setCurrentEditFilePath(srcFile)
			return editor, nil
		}
	}
	editorsLock.RUnlock()

	// 需要创建新的编辑器，使用写锁
	editorsLock.Lock()
	defer editorsLock.Unlock()

	// 双重检查锁定模式
	if editor, ok := editors[srcFile]; ok {
		if editor.IsHandleValid() {
			setCurrentEditFilePath(srcFile)
			return editor, nil
		} else {
			// 先关闭旧的编辑器
			editor.Close()
			delete(editors, srcFile)
		}
	}

	// 创建新的编辑器
	editor, err := xdb.NewEditor(srcFile)
	if err != nil {
		return nil, err
	}

	// 保存到全局编辑器集合
	editors[srcFile] = editor
	setCurrentEditFilePath(srcFile)
	return editor, nil
}

// 编辑单个IP段
func EditSegment(c *gin.Context) {
	var req EditSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "参数错误: " + err.Error(),
			Data: nil,
		})
		return
	}

	// 获取编辑器
	editor, err := getEditor(req.SrcFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "创建编辑器失败: " + err.Error(),
			Data: nil,
		})
		return
	}

	// 编辑IP段
	oldCount, newCount, err := editor.Put(req.Segment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "编辑IP段失败: " + err.Error(),
			Data: nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "编辑成功",
		Data: gin.H{
			"oldCount": oldCount,
			"newCount": newCount,
			"segment":  req.Segment,
		},
	})
}

// 从文件批量编辑IP段
func EditFromFile(c *gin.Context) {
	var req EditFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "参数错误: " + err.Error(),
			Data: nil,
		})
		return
	}

	// 验证文件存在
	if _, err := os.Stat(req.File); os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "文件不存在: " + req.File,
			Data: nil,
		})
		return
	}

	// 获取编辑器
	editor, err := getEditor(req.SrcFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "创建编辑器失败: " + err.Error(),
			Data: nil,
		})
		return
	}

	// 从文件编辑
	oldCount, newCount, err := editor.PutFile(req.File)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "从文件编辑失败: " + err.Error(),
			Data: nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "编辑成功",
		Data: gin.H{
			"oldCount": oldCount,
			"newCount": newCount,
			"file":     req.File,
		},
	})
}

// 列出IP段
func ListSegments(c *gin.Context) {
	var req ListSegmentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "参数错误: " + err.Error(),
			Data: nil,
		})
		return
	}

	// 设置默认值
	if req.Size <= 0 {
		req.Size = 10
	}

	// 获取编辑器
	editor, err := getEditor(req.SrcFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "创建编辑器失败: " + err.Error(),
			Data: nil,
		})
		return
	}

	// 获取IP段列表
	segments := editor.Slice(req.Offset, req.Size)

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "获取成功",
		Data: gin.H{
			"offset":   req.Offset,
			"size":     req.Size,
			"total":    editor.SegLen(),
			"segments": segments,
		},
	})
}

// 保存编辑
func SaveEdit(c *gin.Context) {
	var req SaveEditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "参数错误: " + err.Error(),
			Data: nil,
		})
		return
	}

	// 获取编辑器
	editor, ok := editors[req.SrcFile]
	if !ok {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "编辑器不存在，请先进行编辑操作",
			Data: nil,
		})
		return
	}

	// 保存编辑
	if err := editor.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "保存失败: " + err.Error(),
			Data: nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "保存成功",
		Data: gin.H{
			"srcFile": req.SrcFile,
		},
	})
}

// 保存编辑并生成数据库文件
func SaveAndGenerateDb(c *gin.Context) {
	var req SaveAndGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取编辑器
	editor, err := getEditor(req.SrcFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "获取编辑器失败: " + err.Error(),
		})
		return
	}

	// 如果编辑器需要保存，先保存更改
	if editor.NeedSave() {
		if err := editor.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Code: 500,
				Msg:  "保存文件失败: " + err.Error(),
			})
			return
		}
	}

	// 使用编辑器中的内存数据直接生成XDB文件
	tStart := time.Now()
	if err := editor.SaveToXdbFile(req.DstFile); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "生成XDB文件失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "源文件已保存并生成XDB文件",
		Data: map[string]interface{}{
			"srcFile":   req.SrcFile,
			"dstFile":   req.DstFile,
			"segLen":    editor.SegLen(),
			"timeTaken": time.Since(tStart).String(),
		},
	})
}

// 清理资源函数
func Cleanup() {
	// 关闭搜索器
	if searcher != nil {
		searcher.Close()
		searcher = nil
	}
	// 编辑器不需要关闭
	editors = make(map[string]*xdb.Editor)
}

// 导出任务状态结构（优化版本，使用atomic计数器）
type ExportTaskStatus struct {
	TaskID            string  `json:"taskId"`
	XdbPath           string  `json:"xdbPath"`
	ExportPath        string  `json:"exportPath"`
	Status            string  `json:"status"`            // "pending", "processing", "completed", "failed"
	Progress          float64 `json:"progress"`          // 进度百分比 0-100
	CurrentAClass     uint32  `json:"currentAClass"`     // 当前处理的A类网段
	ProcessedAClasses int     `json:"processedAClasses"` // 已处理的A类网段数量
	TotalAClasses     int     `json:"totalAClasses"`     // 总A类网段数量

	// 新增：明确的导出字段用于JSON序列化
	RecordCount  int64 `json:"recordCount"`  // 直接使用这个名字，确保前端也用它
	SegmentCount int64 `json:"segmentCount"` // 直接使用这个名字

	recordCount  int64 // 内部原子计数器，保持小写非导出
	segmentCount int64 // 内部原子计数器，保持小写非导出

	ErrorMessage    string    `json:"errorMessage"`
	StartTime       time.Time `json:"startTime"`
	EndTime         time.Time `json:"endTime"`
	DurationSeconds float64   `json:"durationSeconds,omitempty"` // 可选字段，改为秒数
	lastUpdateTime  int64     `json:"-"`                         // 使用atomic存储unix时间戳
	DetailedStatus  string    `json:"detailedStatus"`            // 详细状态描述
}

// GetRecordCountInternal 原子获取记录数 (内部使用)
func (e *ExportTaskStatus) GetRecordCountInternal() int64 {
	return atomic.LoadInt64(&e.recordCount)
}

// AddRecordCountInternal 原子增加记录数 (内部使用) - 注意：原AddRecordCount功能不变，仅改名以匹配内部使用约定
func (e *ExportTaskStatus) AddRecordCountInternal(delta int64) int64 {
	return atomic.AddInt64(&e.recordCount, delta)
}

// SetRecordCountInternal 原子设置记录数 (内部使用)
func (e *ExportTaskStatus) SetRecordCountInternal(count int64) {
	atomic.StoreInt64(&e.recordCount, count)
}

// GetSegmentCountInternal 原子获取段数 (内部使用)
func (e *ExportTaskStatus) GetSegmentCountInternal() int64 {
	return atomic.LoadInt64(&e.segmentCount)
}

// AddSegmentCountInternal 原子增加段数 (内部使用) - 注意：原AddSegmentCount功能不变，仅改名以匹配内部使用约定
func (e *ExportTaskStatus) AddSegmentCountInternal(delta int64) int64 {
	return atomic.AddInt64(&e.segmentCount, delta)
}

// SetSegmentCountInternal 原子设置段数 (内部使用)
func (e *ExportTaskStatus) SetSegmentCountInternal(count int64) {
	atomic.StoreInt64(&e.segmentCount, count)
}

// GetLastUpdateTime 获取最后更新时间
func (e *ExportTaskStatus) GetLastUpdateTime() time.Time {
	timestamp := atomic.LoadInt64(&e.lastUpdateTime)
	if timestamp == 0 {
		return time.Time{}
	}
	return time.Unix(timestamp, 0)
}

// UpdateLastUpdateTime 更新最后更新时间为当前时间
func (e *ExportTaskStatus) UpdateLastUpdateTime() {
	atomic.StoreInt64(&e.lastUpdateTime, time.Now().Unix())
}

// LastUpdateTime 为了兼容JSON序列化，提供getter方法 - 这个可以保留，因为它处理的是非原子字段的逻辑转换
func (e *ExportTaskStatus) LastUpdateTime() time.Time {
	return e.GetLastUpdateTime()
}

// 导出任务管理器
var (
	exportTasks     = make(map[string]*ExportTaskStatus)
	exportTasksLock = sync.RWMutex{}
	// 添加取消信号通道的映射
	cancelChans = make(map[string]chan bool)
)

// 获取任务状态
func GetExportTaskStatus(taskID string) *ExportTaskStatus {
	exportTasksLock.RLock()
	defer exportTasksLock.RUnlock()

	if task, exists := exportTasks[taskID]; exists {
		// 创建一个任务状态的副本用于返回，以确保我们填充的导出字段是最新的
		taskCopy := *task // Dereference to get a copy

		// 计算已运行时间 (这部分逻辑保持不变)
		var duration time.Duration
		if taskCopy.Status == "completed" || taskCopy.Status == "failed" {
			if !taskCopy.EndTime.IsZero() && !taskCopy.StartTime.IsZero() {
				duration = taskCopy.EndTime.Sub(taskCopy.StartTime)
			} else {
				duration = time.Since(taskCopy.StartTime)
			}
		} else {
			duration = time.Since(taskCopy.StartTime)
		}

		durationSeconds := math.Round(duration.Seconds())
		if durationSeconds < 0 || math.IsNaN(durationSeconds) {
			durationSeconds = 0
		}
		taskCopy.DurationSeconds = durationSeconds

		taskCopy.RecordCount = task.GetRecordCountInternal()
		taskCopy.SegmentCount = task.GetSegmentCountInternal()

		return &taskCopy
	}
	return nil
}

// 更新任务状态
func updateExportTaskStatus(taskID string, updater func(*ExportTaskStatus)) {
	exportTasksLock.Lock()
	defer exportTasksLock.Unlock()

	if task, exists := exportTasks[taskID]; exists {
		updater(task)
	} else {
		log.Printf("任务 %s: updateExportTaskStatus - 任务不存在，无法更新", taskID)
	}
}

// ExportXdb 导出XDB文件中的数据到文本文件
func ExportXdb(c *gin.Context) {
	var req struct {
		XdbPath    string `json:"xdbPath" binding:"required"`
		ExportPath string `json:"exportPath" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "请求参数错误: " + err.Error(),
			Data: nil,
		})
		return
	}

	// 创建导出任务ID
	taskID := fmt.Sprintf("export_%s", time.Now().Format("20060102150405"))

	// 创建任务取消通道
	exportTasksLock.Lock()
	cancelChan := make(chan bool, 1)
	cancelChans[taskID] = cancelChan

	// 初始化任务状态
	exportTasks[taskID] = &ExportTaskStatus{
		TaskID:         taskID,
		XdbPath:        req.XdbPath,
		ExportPath:     req.ExportPath,
		Status:         "pending",
		StartTime:      time.Now(),
		lastUpdateTime: time.Now().Unix(),
	}
	exportTasksLock.Unlock()

	// 异步执行导出
	go executeExportTask(taskID, req.XdbPath, req.ExportPath)

	// 返回任务ID
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "导出任务已创建",
		Data: map[string]interface{}{
			"taskId": taskID,
		},
	})
}

func executeExportTask(taskID string, xdbPath string, exportPath string) {
	log.Printf("开始执行导出任务: %s, XDB: %s, 导出至: %s", taskID, xdbPath, exportPath)

	// 获取取消通道
	var cancelChan chan bool
	exportTasksLock.RLock()
	if ch, exists := cancelChans[taskID]; exists {
		cancelChan = ch
	} else {
		cancelChan = make(chan bool, 1)
		log.Printf("警告: 任务 %s 的取消通道未找到，已重新创建。", taskID)
	}
	exportTasksLock.RUnlock()

	// 清理函数
	defer func() {
		exportTasksLock.Lock()
		delete(cancelChans, taskID)
		exportTasksLock.Unlock()
		log.Printf("导出任务清理完成: %s", taskID)
	}()

	// 更新任务状态为处理中
	updateExportTaskStatus(taskID, func(task *ExportTaskStatus) {
		task.Status = "processing"
		task.DetailedStatus = "正在加载XDB文件..."
		task.StartTime = time.Now()
		task.Progress = 0
		task.SetRecordCountInternal(0)
		task.SetSegmentCountInternal(0)
		task.UpdateLastUpdateTime()
	})

	var searcherInstance *xdb.Searcher
	var err error
	var localSearcherCreated bool = false

	// 尝试使用全局已加载的 vector 或 memory 模式的 searcher
	searcherLock.RLock()
	if searcher != nil && searcherPath == xdbPath && (searcherMode == "vector" || searcherMode == "memory") {
		searcherInstance = searcher
		log.Printf("任务 %s: 使用已加载的 %s 模式搜索器: %s", taskID, searcherMode, searcherPath)
	}
	searcherLock.RUnlock()

	if searcherInstance == nil {
		// 如果没有匹配的全局搜索器，或者全局搜索器是文件模式（不应在此处使用），则为本次任务创建临时的文件模式搜索器
		log.Printf("任务 %s: 未匹配到已加载的向量/内存模式搜索器，将创建临时文件模式搜索器用于导出: %s", taskID, xdbPath)
		searcherInstance, err = xdb.NewWithFileOnly(xdbPath) // 直接使用 NewWithFileOnly
		if err != nil {
			errMsg := fmt.Sprintf("创建临时文件模式搜索器失败: %v", err)
			log.Printf("任务 %s: %s", taskID, errMsg)
			updateExportTaskStatus(taskID, func(task *ExportTaskStatus) {
				task.Status = "failed"
				task.ErrorMessage = errMsg
				task.EndTime = time.Now()
			})
			return
		}
		localSearcherCreated = true // 标记此搜索器是本地创建的，需要关闭
		log.Printf("任务 %s: 临时文件模式XDB文件加载成功: %s", taskID, xdbPath)
	}

	// 如果是本地创建的临时搜索器，确保在使用完毕后关闭
	if localSearcherCreated && searcherInstance != nil {
		defer searcherInstance.Close()
	}

	updateExportTaskStatus(taskID, func(task *ExportTaskStatus) {
		task.DetailedStatus = "正在导出IP段..."
		task.UpdateLastUpdateTime()
	})

	// 用于跟踪已处理的段数量
	var processedSegments int64 = 0

	allSegments, err := dumpAllIPsFromXDB(searcherInstance, taskID, cancelChan, func(processedIP uint32, totalIPs uint32, segmentCount int) {
		var progress float64
		if totalIPs > 0 {
			progress = float64(processedIP) / float64(totalIPs) * 100
		}

		// 更新已处理的段数量
		processedSegments = int64(segmentCount)

		// 准备详细状态字符串，不包括百分比
		detailedStatus := fmt.Sprintf("正在扫描 IP: %s - 已发现 %d 个IP段",
			xdb.Long2IP(processedIP), segmentCount)

		updateExportTaskStatus(taskID, func(task *ExportTaskStatus) {
			// RecordCount 表示当前处理到的IP地址
			// SegmentCount 表示已发现的IP段数量
			task.SetRecordCountInternal(int64(processedIP)) // 当前处理的IP地址
			task.SetSegmentCountInternal(processedSegments) // 已发现的段数量
			task.Progress = progress
			task.CurrentAClass = 0
			task.ProcessedAClasses = 0
			task.TotalAClasses = 0
			task.DetailedStatus = detailedStatus
			task.UpdateLastUpdateTime()
		})
		log.Printf("任务 %s: 扫描进度 - %s", taskID, detailedStatus)
	})

	if err != nil {
		errMsg := fmt.Sprintf("导出IP段失败: %v", err)
		log.Printf("任务 %s: %s", taskID, errMsg)
		// 检查错误是否由于取消操作导致
		if errors.Is(err, context.Canceled) || errors.Is(err, errTaskCancelled) || strings.Contains(err.Error(), "任务已取消") {
			updateExportTaskStatus(taskID, func(task *ExportTaskStatus) {
				task.Status = "failed"
				task.ErrorMessage = "导出任务已取消"
				task.EndTime = time.Now()
			})
		} else {
			updateExportTaskStatus(taskID, func(task *ExportTaskStatus) {
				task.Status = "failed"
				task.ErrorMessage = errMsg
				task.EndTime = time.Now()
			})
		}
		return
	}

	select {
	case <-cancelChan:
		log.Printf("任务 %s 在数据收集后、写入文件前被取消", taskID)
		updateExportTaskStatus(taskID, func(task *ExportTaskStatus) {
			task.Status = "failed"
			task.ErrorMessage = "导出任务已取消"
			task.EndTime = time.Now()
		})
		return
	default:
	}

	updateExportTaskStatus(taskID, func(task *ExportTaskStatus) {
		task.DetailedStatus = fmt.Sprintf("准备写入 %d 个IP段到文件...", len(allSegments))
		task.SetSegmentCountInternal(int64(len(allSegments)))
		task.Progress = 99 // 扫描完成，准备写入
		task.UpdateLastUpdateTime()
	})

	expectedFields := 5 // 默认值
	if len(allSegments) > 0 {
		firstRegionStr := allSegments[0].Region
		if firstRegionStr != "" {
			parts := strings.Split(firstRegionStr, "|")
			allZeros := true
			nonZeroPartsCount := 0
			for _, part := range parts {
				trimmedPart := strings.TrimSpace(part)
				if trimmedPart != "0" && trimmedPart != "" {
					allZeros = false
				}
				if trimmedPart != "" {
					nonZeroPartsCount++
				}
			}
			if !allZeros && nonZeroPartsCount > 0 {
				expectedFields = nonZeroPartsCount
			} else if allZeros && len(parts) > 0 {
				expectedFields = len(parts)
			}
		}
		// 确保字段数在合理范围内
		if expectedFields < 1 {
			expectedFields = 1
		} else if expectedFields > 15 {
			expectedFields = 15
		}
		log.Printf("任务 %s: 根据首个有效段推断的区域字段数量: %d (首段Region: '%s')", taskID, expectedFields, firstRegionStr)
	} else {
		log.Printf("任务 %s: 未发现任何IP段，使用默认区域字段数量: %d", taskID, expectedFields)
	}

	err = writeResultsToFile(allSegments, exportPath, expectedFields, taskID, cancelChan, func(writtenCount, totalCount int) {
		if writtenCount == 1 {
			// 开始写入
			updateExportTaskStatus(taskID, func(task *ExportTaskStatus) {
				task.Progress = 99
				task.DetailedStatus = fmt.Sprintf("正在写入 %d 个IP段到文件...", totalCount)
				task.UpdateLastUpdateTime()
			})
			log.Printf("任务 %s: 开始写入 %d 个IP段到文件", taskID, totalCount)
		}
	})

	if err != nil {
		errMsg := fmt.Sprintf("写入导出文件失败: %v", err)
		log.Printf("任务 %s: %s", taskID, errMsg)
		if errors.Is(err, context.Canceled) || errors.Is(err, errTaskCancelled) || strings.Contains(err.Error(), "任务已取消") {
			updateExportTaskStatus(taskID, func(task *ExportTaskStatus) {
				task.Status = "failed"
				task.ErrorMessage = "导出任务已取消 (写入阶段)"
				task.EndTime = time.Now()
			})
		} else {
			updateExportTaskStatus(taskID, func(task *ExportTaskStatus) {
				task.Status = "failed"
				task.ErrorMessage = errMsg
				task.EndTime = time.Now()
			})
		}
		return
	}

	log.Printf("任务 %s: 导出成功完成", taskID)
	updateExportTaskStatus(taskID, func(task *ExportTaskStatus) {
		task.Status = "completed"
		task.Progress = 100
		task.EndTime = time.Now()
		task.DetailedStatus = "导出完成"
		task.UpdateLastUpdateTime()
	})
}

var errTaskCancelled = errors.New("任务已取消")

// dumpAllIPsFromXDB 从 xdb.Searcher 实例中逐个IP地址导出数据。
func dumpAllIPsFromXDB(s *xdb.Searcher, taskID string, cancelChan chan bool, progressCallback func(processedIP, totalIPs uint32, segmentCount int)) ([]*IPSegment, error) {
	log.Printf("任务 %s: 开始从XDB逐IP转储所有数据", taskID)
	segments := make([]*IPSegment, 0, 14000000) // 预分配1400万容量

	var currentIP uint32 = 0x01000000 // 1.0.0.0
	const lastIP uint32 = 0xFFFFFFFF
	const stepSize uint32 = 256 // 每256个IP为一个步长，可以调整这个值

	if currentIP > lastIP {
		log.Printf("任务 %s: 起始扫描IP (1.0.0.0) 大于 IPv4 最大IP，不执行扫描。", taskID)
		return segments, nil
	}

	log.Printf("任务 %s: 逐IP扫描将从 IP %s 开始，步长为 %d", taskID, xdb.Long2IP(currentIP), stepSize)

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	go func() {
		select {
		case <-cancelChan:
			cancelCtx()
		case <-ctx.Done():
		}
	}()

	var segmentCount int = 0
	var lastRegion string = ""
	var segmentStartIP uint32 = currentIP

	for currentIP <= lastIP {
		if ctx.Err() != nil {
			log.Printf("任务 %s: XDB转储导出被取消 (当前IP: %s)", taskID, xdb.Long2IP(currentIP))
			return nil, errTaskCancelled
		}

		// 查询当前IP的区域信息
		currentRegion, _, err := s.Search(currentIP)
		if err != nil {
			log.Printf("警告: 任务 %s: 查询 IP %s 失败: %v", taskID, xdb.Long2IP(currentIP), err)
			// 检查是否会发生溢出
			if currentIP > lastIP-stepSize {
				// 如果加上stepSize会溢出，直接跳出循环
				log.Printf("任务 %s: IP %s 接近最大值，停止扫描以避免溢出", taskID, xdb.Long2IP(currentIP))
				break
			}
			currentIP += stepSize
			continue
		}

		// 如果区域为空，使用默认值
		if currentRegion == "" {
			currentRegion = "0|0|0|0|0|0|0|0"
		}

		// 如果区域发生变化，保存上一个段
		if lastRegion != "" && currentRegion != lastRegion {
			segments = append(segments, &IPSegment{
				StartIP: segmentStartIP,
				EndIP:   currentIP - 1,
				Region:  lastRegion,
			})
			segmentCount++
			segmentStartIP = currentIP
		}

		lastRegion = currentRegion

		// 每处理一定数量的IP后更新进度
		if currentIP%256 == 0 || currentIP == lastIP {
			progressCallback(currentIP, lastIP, segmentCount)
		}

		// 检查是否会发生溢出
		if currentIP > lastIP-stepSize {
			// 如果加上stepSize会溢出，直接跳出循环
			log.Printf("任务 %s: IP %s 接近最大值，完成扫描", taskID, xdb.Long2IP(currentIP))
			break
		}
		currentIP += stepSize
	}

	// 添加最后一个段
	if lastRegion != "" {
		segments = append(segments, &IPSegment{
			StartIP: segmentStartIP,
			EndIP:   lastIP,
			Region:  lastRegion,
		})
		segmentCount++
	}

	progressCallback(lastIP, lastIP, segmentCount)
	log.Printf("任务 %s: XDB转储完成，共发现 %d 个段 (从 %s 开始扫描)", taskID, segmentCount, xdb.Long2IP(0x01000000))
	return segments, nil
}

// writeResultsToFile 将IP段写入文件。
// 添加了 taskID 和 cancelChan 用于检查取消信号，以及一个简单的进度回调。
func writeResultsToFile(results []*IPSegment, filePath string, expectedFields int, taskID string, cancelChan chan bool, progressCallback func(writtenCount, totalCount int)) error {
	log.Printf("任务 %s: 开始将 %d 个IP段写入文件 %s", taskID, len(results), filePath)

	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建导出文件 %s 失败: %w", filePath, err)
	}
	defer outFile.Close()

	bufWriter := bufio.NewWriterSize(outFile, 4*1024*1024) // 4MB缓冲区
	var finalErr error                                     // 用于捕获 flush时的错误

	defer func() {
		if errFlush := bufWriter.Flush(); errFlush != nil {
			log.Printf("任务 %s: 刷新缓冲区到文件 %s 失败: %v", taskID, filePath, errFlush)
			if finalErr == nil { // 只在没有其他错误时设置
				finalErr = fmt.Errorf("刷新缓冲区失败: %w", errFlush)
			}
		}
	}()

	if len(results) == 0 {
		log.Printf("任务 %s: 没有结果可写入文件 %s", taskID, filePath)
		return nil // finalErr 仍然可能由 Flush 产生
	}

	totalSegments := len(results)
	for i, segment := range results {
		select {
		case <-cancelChan:
			log.Printf("任务 %s: 写入文件时检测到取消信号 (段 %d/%d)", taskID, i+1, totalSegments)
			finalErr = errTaskCancelled // 使用预定义的取消错误
			return finalErr
		default:
		}

		region := segment.Region
		if region == "" {
			log.Printf("任务 %s: 警告 - 段 (%s - %s) Region为空，将使用 %d 字段的默认全零值", taskID, xdb.Long2IP(segment.StartIP), xdb.Long2IP(segment.EndIP), expectedFields)
			if expectedFields <= 1 {
				region = "0"
			} else {
				region = strings.Repeat("0|", expectedFields-1) + "0"
			}
		}

		line := fmt.Sprintf("%s|%s|%s",
			xdb.Long2IP(segment.StartIP),
			xdb.Long2IP(segment.EndIP),
			region)

		if _, errw := bufWriter.WriteString(line); errw != nil {
			finalErr = fmt.Errorf("写入文件失败 (段 %d, IP: %s): %w", i, xdb.Long2IP(segment.StartIP), errw)
			return finalErr
		}
		// 每行都写入换行符，包括最后一行
		if _, errw := bufWriter.WriteString("\n"); errw != nil {
			finalErr = fmt.Errorf("写入换行符失败 (段 %d): %w", i, errw)
			return finalErr
		}

		if (i+1)%1000 == 0 || i == totalSegments-1 { // 每1000条或最后一条时回调进度
			progressCallback(i+1, totalSegments)
		}
	}

	log.Printf("任务 %s: 所有 %d 段已写入缓冲区，准备刷新到文件 %s", taskID, totalSegments, filePath)
	return finalErr // 可能被 defer中的Flush错误覆盖
}

// GetExportTaskStatusHandler 获取导出任务状态
func GetExportTaskStatusHandler(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "任务ID不能为空",
			Data: nil,
		})
		return
	}

	task := GetExportTaskStatus(taskID)
	if task == nil {
		c.JSON(http.StatusNotFound, Response{
			Code: 404,
			Msg:  "找不到指定的导出任务",
			Data: nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "获取任务状态成功",
		Data: task,
	})
}

// CancelExportTask 取消导出任务
func CancelExportTask(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "任务ID不能为空",
			Data: nil,
		})
		return
	}

	task := GetExportTaskStatus(taskID)
	if task == nil {
		c.JSON(http.StatusNotFound, Response{
			Code: 404,
			Msg:  "找不到指定的导出任务",
			Data: nil,
		})
		return
	}

	// 只能取消 pending 或 processing 状态的任务
	if task.Status != "pending" && task.Status != "processing" {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "任务已完成或已失败，无法取消",
			Data: nil,
		})
		return
	}

	// 发送取消信号
	exportTasksLock.RLock()
	cancelChan, exists := cancelChans[taskID]
	exportTasksLock.RUnlock()

	if exists {
		// 关闭通道通知导出协程终止
		close(cancelChan)
	}

	// 更新任务状态
	updateExportTaskStatus(taskID, func(task *ExportTaskStatus) {
		task.Status = "failed"
		task.ErrorMessage = "用户取消任务"
		task.EndTime = time.Now()
	})

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "任务已取消",
		Data: nil,
	})
}

// GenerateTaskStatus任务状态结构体
type GenerateTaskStatus struct {
	TaskID          string    `json:"taskId"`
	SrcFile         string    `json:"srcFile"`
	DstFile         string    `json:"dstFile"`
	Status          string    `json:"status"`             // "pending", "processing", "completed", "failed"
	Progress        float64   `json:"progress,omitempty"` // 不再使用，保留字段以兼容旧版本
	SegmentCount    int64     `json:"segmentCount"`
	ErrorMessage    string    `json:"errorMessage"`
	StartTime       time.Time `json:"startTime"`
	EndTime         time.Time `json:"endTime"`
	DurationSeconds float64   `json:"durationSeconds,omitempty"` // 秒数
	LastUpdateTime  time.Time `json:"lastUpdateTime,omitempty"`  // 最后更新时间
}

// 生成任务管理器
var (
	generateTasks       = make(map[string]*GenerateTaskStatus)
	generateTasksLock   = sync.RWMutex{}
	generateCancelChans = make(map[string]chan bool)
)

// 获取生成任务状态
func GetGenerateTaskStatus(taskID string) *GenerateTaskStatus {
	generateTasksLock.RLock()
	defer generateTasksLock.RUnlock()

	if task, exists := generateTasks[taskID]; exists {
		// 计算已运行时间
		var duration time.Duration
		if task.Status == "completed" || task.Status == "failed" {
			if !task.EndTime.IsZero() && !task.StartTime.IsZero() {
				duration = task.EndTime.Sub(task.StartTime)
			} else {
				// 如果开始或结束时间未设置，使用当前时间
				duration = time.Since(task.StartTime)
			}
		} else {
			duration = time.Since(task.StartTime)
		}

		// 更新运行时间信息 - 秒数
		durationSeconds := duration.Seconds()
		if durationSeconds < 0 || math.IsNaN(durationSeconds) {
			durationSeconds = 0
		}
		// 四舍五入到整数
		durationSeconds = math.Round(durationSeconds)
		task.DurationSeconds = durationSeconds

		return task
	}
	return nil
}

// 更新生成任务状态
func updateGenerateTaskStatus(taskID string, updater func(*GenerateTaskStatus)) {
	generateTasksLock.Lock()
	defer generateTasksLock.Unlock()

	if task, exists := generateTasks[taskID]; exists {
		updater(task)
	}
}

// GenerateDbWithProgress 生成XDB文件并返回任务ID，以便前端轮询进度
func GenerateDbWithProgress(c *gin.Context) {
	var req SaveAndGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "参数错误: " + err.Error(),
			Data: nil,
		})
		return
	}

	// 创建生成任务ID
	taskID := fmt.Sprintf("generate_%s", time.Now().Format("20060102150405"))

	// 创建任务取消通道
	generateTasksLock.Lock()
	cancelChan := make(chan bool, 1)
	generateCancelChans[taskID] = cancelChan

	// 初始化任务状态
	generateTasks[taskID] = &GenerateTaskStatus{
		TaskID:         taskID,
		SrcFile:        req.SrcFile,
		DstFile:        req.DstFile,
		Status:         "pending",
		StartTime:      time.Now(),
		LastUpdateTime: time.Now(),
	}
	generateTasksLock.Unlock()

	// 异步执行生成
	go executeGenerateDbTask(taskID, req.SrcFile, req.DstFile)

	// 返回任务ID
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "生成任务已创建",
		Data: map[string]interface{}{
			"taskId": taskID,
		},
	})
}

// 执行生成任务
func executeGenerateDbTask(taskID, srcFile, dstFile string) {
	// 获取取消通道
	var cancelChan chan bool

	generateTasksLock.RLock()
	if ch, exists := generateCancelChans[taskID]; exists {
		cancelChan = ch
	} else {
		cancelChan = make(chan bool, 1) // 防御性编程：如果没找到，创建一个新的
	}
	generateTasksLock.RUnlock()

	// 设置清理函数，在任务结束时删除任务取消通道
	defer func() {
		generateTasksLock.Lock()
		delete(generateCancelChans, taskID)
		generateTasksLock.Unlock()
	}()

	// 更新任务状态为处理中
	updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
		task.Status = "processing"
		task.StartTime = time.Now()      // 确保开始时间被设置
		task.LastUpdateTime = time.Now() // 初始化最后更新时间
	})

	// 设置超时控制
	timeoutTimer := time.NewTimer(10 * time.Minute)
	defer timeoutTimer.Stop()

	// 创建通道用于传递结果
	doneChan := make(chan bool, 1)

	// 启动生成协程
	go func() {
		// 检查是否有对该文件的编辑，如果有，先保存
		if editor, ok := editors[srcFile]; ok && editor.NeedSave() {
			// 有未保存的编辑内容，先保存
			if err := editor.Save(); err != nil {
				updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
					task.Status = "failed"
					task.ErrorMessage = "保存编辑内容失败: " + err.Error()
					task.EndTime = time.Now()
				})
				doneChan <- true
				return
			}
		}

		// 检查文件是否存在
		if _, err := os.Stat(srcFile); os.IsNotExist(err) {
			updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
				task.Status = "failed"
				task.ErrorMessage = "源文件不存在: " + srcFile
				task.EndTime = time.Now()
			})
			doneChan <- true
			return
		}

		// 确保目标目录存在
		dstDir := filepath.Dir(dstFile)
		if _, err := os.Stat(dstDir); os.IsNotExist(err) {
			if err := os.MkdirAll(dstDir, 0755); err != nil {
				updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
					task.Status = "failed"
					task.ErrorMessage = "创建目标目录失败: " + err.Error()
					task.EndTime = time.Now()
				})
				doneChan <- true
				return
			}
		}

		// 创建maker
		maker, err := xdb.NewMaker(xdb.VectorIndexPolicy, srcFile, dstFile)
		if err != nil {
			updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
				task.Status = "failed"
				task.ErrorMessage = "创建Maker失败: " + err.Error()
				task.EndTime = time.Now()
			})
			doneChan <- true
			return
		}
		defer maker.Close()

		// 更新任务状态
		updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
			// 使用新添加的GetSegmentsCount方法获取段数量
			task.SegmentCount = int64(maker.GetSegmentsCount())
			task.LastUpdateTime = time.Now()
		})

		// 检查是否已取消
		select {
		case <-cancelChan:
			updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
				task.Status = "failed"
				task.ErrorMessage = "用户取消任务"
				task.EndTime = time.Now()
			})
			doneChan <- true
			return
		default:
			// 继续执行
		}

		// 初始化
		if err := maker.Init(); err != nil {
			updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
				task.Status = "failed"
				task.ErrorMessage = "初始化失败: " + err.Error()
				task.EndTime = time.Now()
			})
			doneChan <- true
			return
		}

		// 更新任务状态
		updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
			// 使用新添加的GetSegmentsCount方法获取段数量
			task.SegmentCount = int64(maker.GetSegmentsCount())
			task.LastUpdateTime = time.Now()
		})

		// 检查是否已取消
		select {
		case <-cancelChan:
			updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
				task.Status = "failed"
				task.ErrorMessage = "用户取消任务"
				task.EndTime = time.Now()
			})
			doneChan <- true
			return
		default:
			// 继续执行
		}

		// 创建一个停止进度更新的通道
		progressStopChan := make(chan bool, 1)
		defer close(progressStopChan)

		// 启动一个goroutine来定期更新进度
		go func() {
			// 定期更新进度的计时器
			ticker := time.NewTicker(500 * time.Millisecond) // 每500毫秒更新一次
			defer ticker.Stop()

			// 模拟处理进度的计数器
			var processedCounter int64 = 0
			// 获取总段数
			totalSegments := int64(maker.GetSegmentsCount())
			if totalSegments <= 0 {
				totalSegments = 1 // 防止除零错误
			}

			// 估算每次需要增加的段数（基于总段数的百分比）
			incrementPerTick := totalSegments / 100 // 每次增加1%的进度
			if incrementPerTick < 1 {
				incrementPerTick = 1 // 确保至少增加1
			}

			for {
				select {
				case <-progressStopChan:
					// 收到停止信号，退出goroutine
					return
				case <-ticker.C:
					// 模拟处理进度
					if processedCounter < totalSegments {
						// 增加进度计数器
						processedCounter += incrementPerTick
						// 确保不超过总数
						if processedCounter > totalSegments {
							processedCounter = totalSegments
						}
						// 更新进度
						updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
							task.SegmentCount = processedCounter
							task.LastUpdateTime = time.Now()
						})
					}
				case <-cancelChan:
					// 任务被取消，退出goroutine
					return
				}
			}
		}()

		// 开始处理
		if err := maker.Start(); err != nil {
			// 发送信号停止进度更新
			progressStopChan <- true

			updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
				task.Status = "failed"
				task.ErrorMessage = "处理失败: " + err.Error()
				task.EndTime = time.Now()
			})
			doneChan <- true
			return
		}

		// 发送信号停止进度更新
		progressStopChan <- true

		// 更新任务状态
		updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
			// 确保最终段数是正确的
			task.SegmentCount = int64(maker.GetSegmentsCount())
			task.LastUpdateTime = time.Now()
		})

		// 检查是否已取消
		select {
		case <-cancelChan:
			updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
				task.Status = "failed"
				task.ErrorMessage = "用户取消任务"
				task.EndTime = time.Now()
			})
			doneChan <- true
			return
		default:
			// 继续执行
		}

		// 结束处理
		if err := maker.End(); err != nil {
			updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
				task.Status = "failed"
				task.ErrorMessage = "结束处理失败: " + err.Error()
				task.EndTime = time.Now()
			})
			doneChan <- true
			return
		}

		// 更新任务完成状态
		updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
			task.Status = "completed"
			task.EndTime = time.Now()
		})

		doneChan <- true
	}()

	// 等待生成完成或超时
	select {
	case <-doneChan:
		// 生成正常完成，状态已在任务中更新
	case <-timeoutTimer.C:
		// 超时
		close(cancelChan) // 发送取消信号
		updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
			if task.Status != "completed" {
				task.Status = "failed"
				task.ErrorMessage = "生成任务超时（10分钟）"
				task.EndTime = time.Now()
			}
		})
	}
}

// GetGenerateTaskStatusHandler 获取生成任务状态
func GetGenerateTaskStatusHandler(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "任务ID不能为空",
			Data: nil,
		})
		return
	}

	task := GetGenerateTaskStatus(taskID)
	if task == nil {
		c.JSON(http.StatusNotFound, Response{
			Code: 404,
			Msg:  "找不到指定的生成任务",
			Data: nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "获取任务状态成功",
		Data: task,
	})
}

// CancelGenerateTask 取消生成任务
func CancelGenerateTask(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "任务ID不能为空",
			Data: nil,
		})
		return
	}

	task := GetGenerateTaskStatus(taskID)
	if task == nil {
		c.JSON(http.StatusNotFound, Response{
			Code: 404,
			Msg:  "找不到指定的生成任务",
			Data: nil,
		})
		return
	}

	// 只能取消 pending 或 processing 状态的任务
	if task.Status != "pending" && task.Status != "processing" {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "任务已完成或已失败，无法取消",
			Data: nil,
		})
		return
	}

	// 发送取消信号
	generateTasksLock.RLock()
	cancelChan, exists := generateCancelChans[taskID]
	generateTasksLock.RUnlock()

	if exists {
		// 关闭通道通知生成协程终止
		close(cancelChan)
	}

	// 更新任务状态
	updateGenerateTaskStatus(taskID, func(task *GenerateTaskStatus) {
		task.Status = "failed"
		task.ErrorMessage = "用户取消任务"
		task.EndTime = time.Now()
	})

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "任务已取消",
		Data: nil,
	})
}

// LoadXdbToMemoryFunc 提供给外部调用的加载XDB文件到内存的函数
func LoadXdbToMemoryFunc(dbPath string) (*LoadXdbResult, error) {
	// 开始计时
	tStart := time.Now()

	// 加载XDB文件到内存
	s, err := getSearcherByMode(dbPath, "memory") // 直接使用 getSearcherByMode
	if err != nil {
		return nil, fmt.Errorf("加载XDB文件失败: %v", err)
	}

	// 获取加载结果信息
	result := &LoadXdbResult{
		DbPath:        dbPath,
		SearchMode:    "memory", // 明确指定是内存模式
		InMemoryMode:  s.IsMemoryMode(),
		BufferSizeKB:  s.GetContentBufferSize() / 1024,
		VectorLoaded:  s.IsVectorIndexLoaded(),
		VectorSizeKB:  s.GetVectorIndexSize() / 1024,
		LoadTimeTaken: time.Since(tStart).String(),
	}

	return result, nil
}

// 辅助函数：安全获取当前编辑文件路径
func getCurrentEditFilePath() string {
	if val := currentEditFilePath.Load(); val != nil {
		return val.(string)
	}
	return ""
}

// 辅助函数：安全设置当前编辑文件路径
func setCurrentEditFilePath(path string) {
	currentEditFilePath.Store(path)
}

// 辅助函数：清空当前编辑文件路径
func clearCurrentEditFilePath() {
	currentEditFilePath.Store("")
}

// GetCurrentEditFile 获取当前正在编辑的源文件信息
func GetCurrentEditFile(c *gin.Context) {
	currentPath := getCurrentEditFilePath()

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "获取当前编辑文件信息成功",
		Data: map[string]interface{}{
			"currentEditFile": currentPath,
			"fileLoaded":      currentPath != "",
		},
	})
}

// UnloadEditFile 卸载当前编辑的源文件
func UnloadEditFile(c *gin.Context) {
	currentPath := getCurrentEditFilePath()

	if currentPath == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "当前没有加载的编辑文件",
		})
		return
	}

	// 记录旧路径用于返回信息
	oldPath := currentPath

	// 清理编辑器实例
	editorsLock.Lock()
	if editor, ok := editors[currentPath]; ok {
		editor.Close()
		delete(editors, currentPath)
	}
	editorsLock.Unlock()

	// 清空当前文件路径
	clearCurrentEditFilePath()

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "编辑文件已卸载",
		Data: map[string]string{
			"unloadedFile": oldPath,
		},
	})
}

// 直接通过IP检查向量索引
func CheckVectorIndexByIP(ipStr string, dbPath string) (map[string]interface{}, error) {
	// 检查数据库文件是否存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("数据库文件不存在: %s", dbPath)
	}

	// 解析IP地址
	ip, err := xdb.IP2Long(ipStr)
	if err != nil {
		return nil, fmt.Errorf("无效的IP地址: %v", err)
	}

	// 加载数据库
	s, err := xdb.NewSearcherWithVectorIndex(dbPath)
	if err != nil {
		return nil, fmt.Errorf("加载数据库失败: %v", err)
	}
	defer s.Close()

	// 使用常规方法查询
	region, ioCount, err := s.Search(ip)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %v", err)
	}

	return map[string]interface{}{
		"ip":            ipStr,
		"ip_int":        ip,
		"region":        region,
		"io_count":      ioCount,
		"vector_loaded": s.IsVectorIndexLoaded(),
	}, nil
}

// 转储向量索引信息
func DumpVectorIndexInfo(c *gin.Context) {
	var req SingleIPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	result, err := CheckVectorIndexByIP(req.IP, req.DbPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "向量索引信息",
		Data: result,
	})
}

// 创建数据库生成器
func CreateDbMaker(srcFile, dstFile string) (*xdb.Maker, error) {
	// 验证源文件存在
	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("源文件不存在: %s", srcFile)
	}

	// 确保目标目录存在
	dstDir := filepath.Dir(dstFile)
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return nil, fmt.Errorf("创建目标目录失败: %s", err.Error())
		}
	}

	// 创建maker
	maker, err := xdb.NewMaker(xdb.VectorIndexPolicy, srcFile, dstFile)
	if err != nil {
		return nil, fmt.Errorf("创建Maker失败: %s", err.Error())
	}

	return maker, nil
}

// 构建数据库
func MakeDb(m *xdb.Maker) error {
	// 初始化
	if err := m.Init(); err != nil {
		return fmt.Errorf("初始化失败: %s", err.Error())
	}

	// 开始处理
	if err := m.Start(); err != nil {
		return fmt.Errorf("处理失败: %s", err.Error())
	}

	// 结束处理
	if err := m.End(); err != nil {
		return fmt.Errorf("结束处理失败: %s", err.Error())
	}

	return nil
}

// 异步数据库生成请求
type AsyncGenDbRequest struct {
	SrcFile string `json:"srcFile" binding:"required"`
	DstFile string `json:"dstFile" binding:"required"`
}

// 异步生成结果
type AsyncGenDbResult struct {
	TaskId string `json:"taskId"`
}

func GetTaskStatus(c *gin.Context) {
	taskId := c.Param("taskId")
	if taskId == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "缺少taskId参数",
		})
		return
	}

	// 查询任务状态
	status, err := xdb.QueryTaskStatus(taskId)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code: 404,
			Msg:  "未找到该任务: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "查询任务状态成功",
		Data: status,
	})
}

// GetDebugStatus 获取调试状态信息，帮助诊断内存模式问题
func GetDebugStatus(c *gin.Context) {
	searcherLock.RLock()
	defer searcherLock.RUnlock()

	// 获取详细状态信息
	debugInfo := map[string]interface{}{
		"searcher_loaded":    searcher != nil,
		"searcher_path":      searcherPath,
		"in_memory_mode":     atomic.LoadInt32(&inMemoryMode) == 1,
		"in_memory_mode_raw": atomic.LoadInt32(&inMemoryMode),
		"vector_index":       false,
		"buffer_size":        int64(0),
		"vector_size":        0,
		"is_memory_mode":     false,
	}

	if searcher != nil {
		debugInfo["vector_index"] = searcher.IsVectorIndexLoaded()
		debugInfo["buffer_size"] = searcher.GetContentBufferSize()
		debugInfo["vector_size"] = searcher.GetVectorIndexSize()
		debugInfo["is_memory_mode"] = searcher.IsMemoryMode()
	}

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "调试状态信息",
		Data: debugInfo,
	})
}

// ForceLoadToMemory 强制重新加载XDB文件到内存模式
func ForceLoadToMemory(c *gin.Context) {
	var req LoadXdbRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "请求参数错误: " + err.Error(),
		})
		return
	}

	// 强制卸载现有的searcher
	searcherLock.Lock()
	if searcher != nil {
		searcher.Close()
		searcher = nil
		searcherPath = ""
		searcherMode = "" // 清除模式
		atomic.StoreInt32(&inMemoryMode, 0)
	}
	searcherLock.Unlock()

	// 强制垃圾回收
	runtime.GC()

	// 重新加载到内存模式
	tStart := time.Now()
	s, err := getSearcherByMode(req.DbPath, "memory") // 直接使用 getSearcherByMode
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "强制加载XDB文件失败: " + err.Error(),
		})
		return
	}

	// 验证加载结果
	result := LoadXdbResult{
		DbPath:        req.DbPath,
		SearchMode:    "memory", // 明确是内存模式
		InMemoryMode:  s.IsMemoryMode(),
		BufferSizeKB:  s.GetContentBufferSize() / 1024,
		VectorLoaded:  s.IsVectorIndexLoaded(),
		VectorSizeKB:  s.GetVectorIndexSize() / 1024,
		LoadTimeTaken: time.Since(tStart).String(),
	}

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "XDB文件已强制重新加载到内存",
		Data: result,
	})
}
