// Package logging 提供基于 slog + lumberjack 的日志基础设施。
//
// 功能：
//   - JSON 结构化输出（为后续接入 OpenObserve / ELK 做准备）
//   - 自动日志轮转（基于文件大小）、压缩归档、过期清理
//   - 实现 fiber v2 的 AllLogger 接口，业务代码无需任何改动
//   - 自定义字段顺序：time → level → service → caller → msg → 其他
package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	fiberlog "github.com/gofiber/fiber/v2/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config 日志配置
type Config struct {
	ServiceName string     // 服务名称，写入每条日志的 service 字段
	LogDir      string     // 日志目录，默认 "./log"
	FileName    string     // 日志文件名，默认 "server.log"
	MaxSizeMB   int        // 单文件最大 MB，超过触发轮转，默认 100
	MaxAgeDays  int        // 归档文件保留天数，默认 30
	MaxBackups  int        // 最多保留归档文件数，默认 10
	Compress    bool       // 是否 gzip 压缩归档，默认 true
	Level       slog.Level // 日志级别，默认 slog.LevelInfo
	Console     bool       // 是否同时输出到终端（foreground 模式用）
}

func (c *Config) applyDefaults() {
	if c.LogDir == "" {
		c.LogDir = "./log"
	}
	if c.FileName == "" {
		c.FileName = "server.log"
	}
	if c.MaxSizeMB <= 0 {
		c.MaxSizeMB = 100
	}
	if c.MaxAgeDays <= 0 {
		c.MaxAgeDays = 30
	}
	if c.MaxBackups <= 0 {
		c.MaxBackups = 10
	}
}

// orderedHandler 自定义 slog.Handler，控制 JSON 字段输出顺序：
// time → level → service → caller → msg → 其他 attrs
type orderedHandler struct {
	writer  io.Writer
	level   slog.Level
	service string
	mu      sync.Mutex
}

func (h *orderedHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *orderedHandler) Handle(_ context.Context, r slog.Record) error {
	buf := &bytes.Buffer{}
	buf.WriteByte('{')

	// 1. time
	buf.WriteString(`"time":`)
	writeJSONString(buf, r.Time.Format(time.RFC3339Nano))

	// 2. level
	buf.WriteString(`,"level":`)
	writeJSONString(buf, r.Level.String())

	// 3. service
	if h.service != "" {
		buf.WriteString(`,"service":`)
		writeJSONString(buf, h.service)
	}

	// 4. caller
	if r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		if f.File != "" {
			buf.WriteString(`,"caller":`)
			writeJSONString(buf, shortenPath(f.File)+":"+fmt.Sprint(f.Line))
		}
	}

	// 5. msg
	buf.WriteString(`,"msg":`)
	writeJSONString(buf, r.Message)

	// 6. 额外 attrs
	r.Attrs(func(a slog.Attr) bool {
		buf.WriteByte(',')
		writeJSONString(buf, a.Key)
		buf.WriteByte(':')
		writeAttrValue(buf, a.Value)
		return true
	})

	buf.WriteString("}\n")

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.writer.Write(buf.Bytes())
	return err
}

func (h *orderedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h // 预设 attrs 已通过 service 字段处理
}

func (h *orderedHandler) WithGroup(_ string) slog.Handler {
	return h
}

// writeJSONString 写入 JSON 转义字符串
func writeJSONString(buf *bytes.Buffer, s string) {
	b, _ := json.Marshal(s)
	buf.Write(b)
}

// writeAttrValue 写入 slog.Value 的 JSON 表示
func writeAttrValue(buf *bytes.Buffer, v slog.Value) {
	switch v.Kind() {
	case slog.KindString:
		writeJSONString(buf, v.String())
	case slog.KindInt64:
		fmt.Fprintf(buf, "%d", v.Int64())
	case slog.KindUint64:
		fmt.Fprintf(buf, "%d", v.Uint64())
	case slog.KindFloat64:
		fmt.Fprintf(buf, "%g", v.Float64())
	case slog.KindBool:
		fmt.Fprintf(buf, "%t", v.Bool())
	case slog.KindTime:
		writeJSONString(buf, v.Time().Format(time.RFC3339Nano))
	default:
		writeJSONString(buf, v.String())
	}
}

// shortenPath 将绝对路径截断为从 "oshit-go/" 之后的相对路径
func shortenPath(file string) string {
	const marker = "oshit-go" + string(filepath.Separator)
	if idx := strings.Index(file, marker); idx >= 0 {
		return file[idx+len(marker):]
	}
	return filepath.Base(file)
}

// Setup 初始化全局日志：slog 作为底层，lumberjack 做轮转，同时替换 fiber log 和标准库 log。
func Setup(cfg Config) {
	cfg.applyDefaults()

	lj := &lumberjack.Logger{
		Filename:   cfg.LogDir + "/" + cfg.FileName,
		MaxSize:    cfg.MaxSizeMB,
		MaxAge:     cfg.MaxAgeDays,
		MaxBackups: cfg.MaxBackups,
		Compress:   cfg.Compress,
	}

	var writer io.Writer = lj
	if cfg.Console {
		writer = io.MultiWriter(lj, os.Stdout)
	}

	handler := &orderedHandler{
		writer:  writer,
		level:   cfg.Level,
		service: cfg.ServiceName,
	}

	slogger := slog.New(handler)
	slog.SetDefault(slogger)

	// 替换 Go 标准库 log 的输出
	log.SetOutput(writer)
	log.SetFlags(0)

	// 替换 fiber log 的全局 logger
	adapter := newFiberAdapter(slogger, cfg.Level)
	fiberlog.SetLogger(adapter)
}
