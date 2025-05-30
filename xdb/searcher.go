// Copyright 2022 The Ip2Region Authors. All rights reserved.
// Use of this source code is governed by a Apache2.0-style
// license that can be found in the LICENSE file.

// ---
// ip2region database v2.0 searcher.
// this is part of the maker for testing and validate.
// please use the searcher in binding/golang for production use.
// And this is a Not thread safe implementation.

package xdb

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type Searcher struct {
	handle *os.File

	// header info
	header []byte

	// use it only when this feature enabled.
	// Preload the vector index will reduce the number of IO operations
	// thus speedup the search process
	vectorIndex []byte

	// 内存模式标志
	memoryMode bool

	// 内容缓冲区大小
	contentBufferSize int64

	// 完全内存模式：整个XDB文件内容缓冲区
	contentBuffer []byte
}

func NewSearcher(dbFile string) (*Searcher, error) {
	return NewWithFileOnly(dbFile)
}

// NewSearcherWithVectorIndex 创建一个带有向量索引的搜索器
func NewSearcherWithVectorIndex(dbFile string) (*Searcher, error) {
	s, err := NewSearcher(dbFile)
	if err != nil {
		return nil, err
	}

	// 加载向量索引
	err = s.LoadVectorIndex()
	if err != nil {
		s.Close()
		return nil, err
	}

	return s, nil
}

// LoadContentFromFile 从文件加载整个XDB内容到内存缓冲区
func LoadContentFromFile(dbFile string) ([]byte, error) {
	file, err := os.Open(dbFile)
	if err != nil {
		return nil, fmt.Errorf("打开XDB文件失败: %w", err)
	}
	defer file.Close()

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 读取整个文件内容到内存
	buffer := make([]byte, fileInfo.Size())
	_, err = io.ReadFull(file, buffer)
	if err != nil {
		return nil, fmt.Errorf("读取XDB文件内容失败: %w", err)
	}

	return buffer, nil
}

// NewWithBuffer 使用内存缓冲区创建完全基于内存的搜索器
func NewWithBuffer(contentBuffer []byte) (*Searcher, error) {
	if len(contentBuffer) < HeaderInfoLength {
		return nil, fmt.Errorf("XDB内容缓冲区太小，至少需要 %d 字节", HeaderInfoLength)
	}

	s := &Searcher{
		handle:            nil, // 内存模式不需要文件句柄
		header:            nil,
		vectorIndex:       nil,
		memoryMode:        true,
		contentBufferSize: int64(len(contentBuffer)),
		contentBuffer:     contentBuffer,
	}

	// 从内存缓冲区加载向量索引
	err := s.loadVectorIndexFromBuffer()
	if err != nil {
		return nil, fmt.Errorf("从内存缓冲区加载向量索引失败: %w", err)
	}

	return s, nil
}

// NewSearcherWithMemoryMode 创建一个内存模式的搜索器（兼容旧接口，但推荐使用NewWithBuffer）
func NewSearcherWithMemoryMode(dbFile string) (*Searcher, error) {
	// 加载整个文件内容到内存
	contentBuffer, err := LoadContentFromFile(dbFile)
	if err != nil {
		return nil, err
	}

	// 使用内存缓冲区创建搜索器
	return NewWithBuffer(contentBuffer)
}

// 从内存缓冲区加载向量索引
func (s *Searcher) loadVectorIndexFromBuffer() error {
	if len(s.contentBuffer) < HeaderInfoLength+VectorIndexLength {
		return fmt.Errorf("内容缓冲区太小，无法包含向量索引")
	}

	// 从内存缓冲区中提取向量索引
	s.vectorIndex = make([]byte, VectorIndexLength)
	copy(s.vectorIndex, s.contentBuffer[HeaderInfoLength:HeaderInfoLength+VectorIndexLength])

	return nil
}

// IsMemoryMode 检查是否为内存模式
func (s *Searcher) IsMemoryMode() bool {
	return s.memoryMode
}

// GetContentBufferSize 获取内容缓冲区大小
func (s *Searcher) GetContentBufferSize() int64 {
	if s.memoryMode && s.contentBuffer != nil {
		return int64(len(s.contentBuffer))
	}

	if s.handle == nil {
		return 0
	}

	if s.contentBufferSize > 0 {
		return s.contentBufferSize
	}

	// 如果未设置，获取文件大小
	fileInfo, err := s.handle.Stat()
	if err != nil {
		return 0
	}
	s.contentBufferSize = fileInfo.Size()
	return s.contentBufferSize
}

// IsVectorIndexLoaded 检查向量索引是否已加载
func (s *Searcher) IsVectorIndexLoaded() bool {
	return s.vectorIndex != nil
}

// GetVectorIndexSize 获取向量索引大小
func (s *Searcher) GetVectorIndexSize() int {
	if s.vectorIndex == nil {
		return 0
	}
	return len(s.vectorIndex)
}

func (s *Searcher) Close() {
	if s.handle != nil {
		err := s.handle.Close()
		if err != nil {
			return
		}
	}
	// 内存模式下清理内存
	if s.memoryMode {
		s.contentBuffer = nil
	}
}

// LoadVectorIndex load and cache the vector index for search speedup.
// this will take up VectorIndexRows x VectorIndexCols x VectorIndexSize bytes memory.
func (s *Searcher) LoadVectorIndex() error {
	// loaded already
	if s.vectorIndex != nil {
		return nil
	}

	if s.memoryMode {
		// 内存模式下从缓冲区加载
		return s.loadVectorIndexFromBuffer()
	}

	// 文件模式下从文件加载
	// load all the vector index block
	_, err := s.handle.Seek(HeaderInfoLength, 0)
	if err != nil {
		return fmt.Errorf("seek to vector index: %w", err)
	}

	var buff = make([]byte, VectorIndexLength)
	rLen, err := s.handle.Read(buff)
	if err != nil {
		return err
	}

	if rLen != len(buff) {
		return fmt.Errorf("incomplete read: readed bytes should be %d", len(buff))
	}

	s.vectorIndex = buff
	return nil
}

// ClearVectorIndex clear preloaded vector index cache
func (s *Searcher) ClearVectorIndex() {
	s.vectorIndex = nil
}

// readFromBuffer 从内存缓冲区读取数据
func (s *Searcher) readFromBuffer(offset int64, length int) ([]byte, error) {
	if s.contentBuffer == nil {
		return nil, fmt.Errorf("内容缓冲区为空")
	}

	if offset < 0 || offset >= int64(len(s.contentBuffer)) {
		return nil, fmt.Errorf("偏移量超出缓冲区范围: %d", offset)
	}

	if int64(length) > int64(len(s.contentBuffer))-offset {
		return nil, fmt.Errorf("读取长度超出缓冲区范围")
	}

	data := make([]byte, length)
	copy(data, s.contentBuffer[offset:offset+int64(length)])
	return data, nil
}

// Search find the region for the specified ip address
func (s *Searcher) Search(ip uint32) (string, int, error) {
	// locate the segment index block based on the vector index
	var ioCount = 0
	var il0 = (ip >> 24) & 0xFF
	var il1 = (ip >> 16) & 0xFF
	var idx = il0*VectorIndexCols*VectorIndexSize + il1*VectorIndexSize
	var sPtr, ePtr = uint32(0), uint32(0)

	if s.vectorIndex != nil {
		sPtr = binary.LittleEndian.Uint32(s.vectorIndex[idx:])
		ePtr = binary.LittleEndian.Uint32(s.vectorIndex[idx+4:])
	} else {
		// 如果向量索引未加载，需要从存储中读取
		var buffVec []byte
		var err error

		if s.memoryMode {
			// 从内存缓冲区读取
			buffVec, err = s.readFromBuffer(int64(HeaderInfoLength+idx), VectorIndexSize)
			if err != nil {
				return "", ioCount, fmt.Errorf("read vector index from buffer at %d: %w", HeaderInfoLength+idx, err)
			}
		} else {
			// 从文件读取
			pos, err := s.handle.Seek(int64(HeaderInfoLength+idx), 0)
			if err != nil {
				return "", ioCount, fmt.Errorf("seek to vector index %d: %w", HeaderInfoLength+idx, err)
			}
			ioCount++
			buffVec = make([]byte, VectorIndexSize)
			rLenVec, err := s.handle.Read(buffVec)
			if err != nil {
				return "", ioCount, fmt.Errorf("read vector index at %d: %w", pos, err)
			}
			if rLenVec != len(buffVec) {
				return "", ioCount, fmt.Errorf("incomplete read for vector index: readed bytes should be %d", len(buffVec))
			}
		}

		sPtr = binary.LittleEndian.Uint32(buffVec)
		ePtr = binary.LittleEndian.Uint32(buffVec[4:])
	}

	// binary search the segment index to get the region
	var dataLen, dataPtr = 0, uint32(0)
	var buff = make([]byte, SegmentIndexSize)
	var l, h = 0, int((ePtr - sPtr) / SegmentIndexSize)

	if sPtr == 0 || ePtr == 0 || sPtr >= ePtr { // sPtr can be 0 if a /16 prefix has no IPs
		// No need to search if the range is invalid or empty
		// return "", ioCount, nil // This would indicate not found
	}

	for l <= h {
		m := (l + h) >> 1
		p := sPtr + uint32(m*SegmentIndexSize)

		var err error
		if s.memoryMode {
			// 从内存缓冲区读取
			buff, err = s.readFromBuffer(int64(p), SegmentIndexSize)
			if err != nil {
				return "", ioCount, fmt.Errorf("read segment index from buffer at %d: %w", p, err)
			}
		} else {
			// 从文件读取
			_, err := s.handle.Seek(int64(p), 0)
			if err != nil {
				return "", ioCount, fmt.Errorf("seek to segment block at %d: %w", p, err)
			}

			ioCount++
			rLen, err := s.handle.Read(buff)
			if err != nil {
				return "", ioCount, fmt.Errorf("read segment index at %d: %w", p, err)
			}

			if rLen != len(buff) {
				return "", ioCount, fmt.Errorf("incomplete read: readed bytes should be %d", len(buff))
			}
		}

		sip := binary.LittleEndian.Uint32(buff)
		eipRead := binary.LittleEndian.Uint32(buff[4:]) // Renamed to avoid conflict with outer ip
		// decode the data step by step to reduce the unnecessary calculations
		// sip := binary.LittleEndian.Uint32(buff)
		if ip < sip {
			h = m - 1
		} else {
			// eip := binary.LittleEndian.Uint32(buff[4:])
			if ip > eipRead {
				l = m + 1
			} else {
				dataLen = int(binary.LittleEndian.Uint16(buff[8:]))
				dataPtr = binary.LittleEndian.Uint32(buff[10:])
				break
			}
		}
	}

	if dataLen == 0 {
		return "", ioCount, nil
	}

	// load and return the region data
	var regionBuff []byte
	var err error

	if s.memoryMode {
		// 从内存缓冲区读取地区数据
		regionBuff, err = s.readFromBuffer(int64(dataPtr), dataLen)
		if err != nil {
			return "", ioCount, fmt.Errorf("read region data from buffer at %d: %w", dataPtr, err)
		}
	} else {
		// 从文件读取地区数据
		_, err := s.handle.Seek(int64(dataPtr), 0)
		if err != nil {
			return "", ioCount, fmt.Errorf("seek to data block at %d: %w", dataPtr, err)
		}

		ioCount++
		regionBuff = make([]byte, dataLen)
		rLen, err := s.handle.Read(regionBuff)
		if err != nil {
			return "", ioCount, fmt.Errorf("read region data at %d: %w", dataPtr, err)
		}

		if rLen != dataLen {
			return "", ioCount, fmt.Errorf("incomplete read: readed bytes should be %d", dataLen)
		}
	}

	return string(regionBuff), ioCount, nil
}

// NewWithFileOnly 创建一个完全基于文件的搜索器（每次查询都进行IO操作）
func NewWithFileOnly(dbFile string) (*Searcher, error) {
	handle, err := os.OpenFile(dbFile, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}

	return &Searcher{
		handle:            handle,
		header:            nil,
		vectorIndex:       nil, // 不预加载向量索引
		memoryMode:        false,
		contentBufferSize: 0,
		contentBuffer:     nil,
	}, nil
}
