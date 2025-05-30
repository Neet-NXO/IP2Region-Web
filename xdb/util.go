// Copyright 2022 The Ip2Region Authors. All rights reserved.
// Use of this source code is governed by a Apache2.0-style
// license that can be found in the LICENSE file.

package xdb

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"unsafe"
)

// Long2IP 将长整数转换为IP地址
func Long2IP(ip uint32) string {
	// 预分配最大可能的字节数组
	buf := make([]byte, 0, 15) // "255.255.255.255" = 15字符

	// 提取4个字节
	a := (ip >> 24) & 0xFF
	b := (ip >> 16) & 0xFF
	c := (ip >> 8) & 0xFF
	d := ip & 0xFF

	// 手动转换数字为字符串，避免strconv的开销
	buf = appendUint8(buf, uint8(a))
	buf = append(buf, '.')
	buf = appendUint8(buf, uint8(b))
	buf = append(buf, '.')
	buf = appendUint8(buf, uint8(c))
	buf = append(buf, '.')
	buf = appendUint8(buf, uint8(d))

	return string(buf)
}

// appendUint8 手动将uint8转换为字符串并追加到字节数组
func appendUint8(buf []byte, val uint8) []byte {
	if val >= 100 {
		buf = append(buf, '0'+val/100)
		val %= 100
		buf = append(buf, '0'+val/10)
		buf = append(buf, '0'+val%10)
	} else if val >= 10 {
		buf = append(buf, '0'+val/10)
		buf = append(buf, '0'+val%10)
	} else {
		buf = append(buf, '0'+val)
	}
	return buf
}

// Long2IPPool 池化版本：重用缓冲区减少内存分配
var ipBufPool = make(chan []byte, 100) // 缓冲区池

func Long2IPPool(ip uint32) string {
	// 从池中获取缓冲区
	var buf []byte
	select {
	case buf = <-ipBufPool:
		buf = buf[:0] // 重置长度但保持容量
	default:
		buf = make([]byte, 0, 15)
	}

	// 提取4个字节
	a := (ip >> 24) & 0xFF
	b := (ip >> 16) & 0xFF
	c := (ip >> 8) & 0xFF
	d := ip & 0xFF

	// 手动转换
	buf = appendUint8(buf, uint8(a))
	buf = append(buf, '.')
	buf = appendUint8(buf, uint8(b))
	buf = append(buf, '.')
	buf = appendUint8(buf, uint8(c))
	buf = append(buf, '.')
	buf = appendUint8(buf, uint8(d))

	result := string(buf)

	// 归还缓冲区到池
	select {
	case ipBufPool <- buf:
	default:
		// 池满了，丢弃缓冲区
	}

	return result
}

func MidIP(sip uint32, eip uint32) uint32 {
	return uint32((uint64(sip) + uint64(eip)) >> 1)
}

func IterateSegments(handle *os.File, before func(l string), cb func(seg *Segment) error) error {
	var last *Segment = nil
	var scanner = bufio.NewScanner(handle)
	scanner.Split(bufio.ScanLines)

	// 添加行号跟踪和前后文信息
	var lineNumber int = 0
	var previousLines []string = make([]string, 0, 3) // 保存前3行
	var currentLine string
	var nextLines []string = make([]string, 0, 3) // 预读后3行

	// 预读所有行以便提供上下文
	var allLines []string
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	// 重新设置文件指针到开头
	handle.Seek(0, 0)
	scanner = bufio.NewScanner(handle)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		lineNumber++
		currentLine = strings.TrimSpace(strings.TrimSuffix(scanner.Text(), "\n"))

		// 更新前后文信息
		if lineNumber > 1 {
			// 获取前面的行
			previousLines = nil
			start := lineNumber - 4 // 前3行
			if start < 1 {
				start = 1
			}
			for i := start; i < lineNumber; i++ {
				if i-1 < len(allLines) {
					previousLines = append(previousLines, fmt.Sprintf("第%d行: %s", i, allLines[i-1]))
				}
			}
		}

		// 获取后面的行
		nextLines = nil
		for i := lineNumber; i < lineNumber+3 && i < len(allLines); i++ {
			nextLines = append(nextLines, fmt.Sprintf("第%d行: %s", i+1, allLines[i]))
		}

		if len(currentLine) < 1 { // ignore empty line
			continue
		}

		if currentLine[0] == '#' { // ignore the comment line
			continue
		}

		if before != nil {
			before(currentLine)
		}

		var ps = strings.SplitN(currentLine, "|", 3)
		if len(ps) != 3 {
			// 构建详细的错误信息
			var errorMsg strings.Builder
			errorMsg.WriteString(fmt.Sprintf("第%d行格式错误: `%s`\n", lineNumber, currentLine))

			if len(previousLines) > 0 {
				errorMsg.WriteString("\n前面的行:\n")
				for _, line := range previousLines {
					errorMsg.WriteString(fmt.Sprintf("  %s\n", line))
				}
			}

			errorMsg.WriteString(fmt.Sprintf("\n>>> 错误行: 第%d行: %s <<<\n", lineNumber, currentLine))

			if len(nextLines) > 0 {
				errorMsg.WriteString("\n后面的行:\n")
				for _, line := range nextLines {
					errorMsg.WriteString(fmt.Sprintf("  %s\n", line))
				}
			}

			return fmt.Errorf("%s", errorMsg.String())
		}

		sip, err := IP2Long(ps[0])
		if err != nil {
			var errorMsg strings.Builder
			errorMsg.WriteString(fmt.Sprintf("第%d行起始IP格式错误: `%s`\n", lineNumber, ps[0]))
			errorMsg.WriteString(fmt.Sprintf("错误原因: %s\n", err))
			errorMsg.WriteString(fmt.Sprintf("完整行内容: %s\n", currentLine))

			if len(previousLines) > 0 {
				errorMsg.WriteString("\n前面的行:\n")
				for _, line := range previousLines {
					errorMsg.WriteString(fmt.Sprintf("  %s\n", line))
				}
			}

			if len(nextLines) > 0 {
				errorMsg.WriteString("\n后面的行:\n")
				for _, line := range nextLines {
					errorMsg.WriteString(fmt.Sprintf("  %s\n", line))
				}
			}

			return fmt.Errorf("%s", errorMsg.String())
		}

		eip, err := IP2Long(ps[1])
		if err != nil {
			var errorMsg strings.Builder
			errorMsg.WriteString(fmt.Sprintf("第%d行结束IP格式错误: `%s`\n", lineNumber, ps[1]))
			errorMsg.WriteString(fmt.Sprintf("错误原因: %s\n", err))
			errorMsg.WriteString(fmt.Sprintf("完整行内容: %s\n", currentLine))

			if len(previousLines) > 0 {
				errorMsg.WriteString("\n前面的行:\n")
				for _, line := range previousLines {
					errorMsg.WriteString(fmt.Sprintf("  %s\n", line))
				}
			}

			if len(nextLines) > 0 {
				errorMsg.WriteString("\n后面的行:\n")
				for _, line := range nextLines {
					errorMsg.WriteString(fmt.Sprintf("  %s\n", line))
				}
			}

			return fmt.Errorf("%s", errorMsg.String())
		}

		if sip > eip {
			var errorMsg strings.Builder
			errorMsg.WriteString(fmt.Sprintf("第%d行IP范围错误: 起始IP(%s)不能大于结束IP(%s)\n", lineNumber, ps[0], ps[1]))
			errorMsg.WriteString(fmt.Sprintf("完整行内容: %s\n", currentLine))

			if len(previousLines) > 0 {
				errorMsg.WriteString("\n前面的行:\n")
				for _, line := range previousLines {
					errorMsg.WriteString(fmt.Sprintf("  %s\n", line))
				}
			}

			if len(nextLines) > 0 {
				errorMsg.WriteString("\n后面的行:\n")
				for _, line := range nextLines {
					errorMsg.WriteString(fmt.Sprintf("  %s\n", line))
				}
			}

			return fmt.Errorf("%s", errorMsg.String())
		}

		if len(ps[2]) < 1 {
			var errorMsg strings.Builder
			errorMsg.WriteString(fmt.Sprintf("第%d行区域信息为空\n", lineNumber))
			errorMsg.WriteString(fmt.Sprintf("完整行内容: %s\n", currentLine))

			if len(previousLines) > 0 {
				errorMsg.WriteString("\n前面的行:\n")
				for _, line := range previousLines {
					errorMsg.WriteString(fmt.Sprintf("  %s\n", line))
				}
			}

			if len(nextLines) > 0 {
				errorMsg.WriteString("\n后面的行:\n")
				for _, line := range nextLines {
					errorMsg.WriteString(fmt.Sprintf("  %s\n", line))
				}
			}

			return fmt.Errorf("%s", errorMsg.String())
		}

		var seg = &Segment{
			StartIP: sip,
			EndIP:   eip,
			Region:  ps[2],
		}

		// check and automatic merging the Consecutive Segments which means:
		// 1, region info is the same
		// 2, last.eip+1 = cur.sip
		if last == nil {
			last = seg
			continue
		} else if last.Region == seg.Region {
			if err = seg.AfterCheck(last); err == nil {
				last.EndIP = seg.EndIP
				continue
			}
		}

		if err = cb(last); err != nil {
			return fmt.Errorf("第%d行处理段时出错: %s\n段内容: %s", lineNumber, err, last.String())
		}

		// reset the last
		last = seg
	}

	// process the last segment
	if last != nil {
		if err := cb(last); err != nil {
			return fmt.Errorf("处理最后一个段时出错: %s\n段内容: %s", err, last.String())
		}
	}

	return nil
}

func CheckSegments(segList []*Segment) error {
	var last *Segment
	for _, seg := range segList {
		// sip must <= eip
		if seg.StartIP > seg.EndIP {
			return fmt.Errorf("segment `%s`: start ip should not be greater than end ip", seg.String())
		}

		// check the continuity of the data segment
		if last != nil {
			if last.EndIP+1 != seg.StartIP {
				return fmt.Errorf("discontinuous segment `%s`: last.eip+1 != cur.sip", seg.String())
			}
		}

		last = seg
	}

	return nil
}

// IP2Long 将IP地址转换为长整数
func IP2Long(ipStr string) (uint32, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0, fmt.Errorf("无效的IP地址: %s", ipStr)
	}
	ip = ip.To4()
	if ip == nil {
		return 0, fmt.Errorf("不支持IPv6地址: %s", ipStr)
	}
	val := *(*uint32)(unsafe.Pointer(&ip[0]))
	return (val&0xFF)<<24 | ((val>>8)&0xFF)<<16 | ((val>>16)&0xFF)<<8 | ((val >> 24) & 0xFF), nil
}
