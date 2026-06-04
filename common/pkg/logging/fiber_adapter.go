package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"

	fiberlog "github.com/gofiber/fiber/v2/log"
)

// callerSkip 是从 emit() 到业务代码的栈帧数：
// 0: runtime.Callers → 1: emit → 2: adapter 方法(Infof等) → 3: fiber 全局函数(log.Infof) → 4: 业务代码
const callerSkip = 4

// fiberAdapter 实现 fiber v2 的 AllLogger 接口，将所有日志调用委托给 slog。
// 通过 runtime.Callers 跳过 adapter 层，在 JSON 中输出真正的业务代码文件名和行号。
type fiberAdapter struct {
	handler slog.Handler
	level   slog.Level
}

func newFiberAdapter(logger *slog.Logger, level slog.Level) *fiberAdapter {
	return &fiberAdapter{handler: logger.Handler(), level: level}
}

// emit 构造 slog.Record 并手动设置正确的调用方信息
func (a *fiberAdapter) emit(level slog.Level, msg string) {
	if !a.handler.Enabled(context.Background(), level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(callerSkip, pcs[:])

	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	_ = a.handler.Handle(context.Background(), r)
}

// callerAttr 返回简短的 file:line 属性，用于 WithLogger 系列方法
func callerAttr() slog.Attr {
	_, file, line, ok := runtime.Caller(callerSkip - 1) // 比 emit 少一层
	if !ok {
		return slog.String("caller", "???")
	}
	// 使用相对路径：从 oshit-go/ 开始
	if idx := findModuleRoot(file); idx >= 0 {
		file = file[idx:]
	}
	return slog.String("caller", fmt.Sprintf("%s:%d", file, line))
}

// findModuleRoot 在路径中找 "oshit-go/" 的起始位置，返回其后的偏移
func findModuleRoot(path string) int {
	const marker = "oshit-go" + string(filepath.Separator)
	for i := 0; i <= len(path)-len(marker); i++ {
		if path[i:i+len(marker)] == marker {
			return i
		}
	}
	return -1
}

// ---- ControlLogger ----

func (a *fiberAdapter) SetLevel(lv fiberlog.Level) {
	switch lv {
	case fiberlog.LevelTrace:
		a.level = slog.LevelDebug - 4
	case fiberlog.LevelDebug:
		a.level = slog.LevelDebug
	case fiberlog.LevelInfo:
		a.level = slog.LevelInfo
	case fiberlog.LevelWarn:
		a.level = slog.LevelWarn
	case fiberlog.LevelError, fiberlog.LevelFatal, fiberlog.LevelPanic:
		a.level = slog.LevelError
	}
}

func (a *fiberAdapter) SetOutput(_ io.Writer) {}

// ---- Logger (variadic) ----

func (a *fiberAdapter) Trace(v ...any) { a.emit(slog.LevelDebug-4, fmt.Sprint(v...)) }
func (a *fiberAdapter) Debug(v ...any) { a.emit(slog.LevelDebug, fmt.Sprint(v...)) }
func (a *fiberAdapter) Info(v ...any)  { a.emit(slog.LevelInfo, fmt.Sprint(v...)) }
func (a *fiberAdapter) Warn(v ...any)  { a.emit(slog.LevelWarn, fmt.Sprint(v...)) }
func (a *fiberAdapter) Error(v ...any) { a.emit(slog.LevelError, fmt.Sprint(v...)) }
func (a *fiberAdapter) Fatal(v ...any) { a.emit(slog.LevelError, fmt.Sprint(v...)); os.Exit(1) }
func (a *fiberAdapter) Panic(v ...any) {
	msg := fmt.Sprint(v...)
	a.emit(slog.LevelError, msg)
	panic(msg)
}

// ---- FormatLogger (printf-style) ----

func (a *fiberAdapter) Tracef(format string, v ...any) {
	a.emit(slog.LevelDebug-4, fmt.Sprintf(format, v...))
}
func (a *fiberAdapter) Debugf(format string, v ...any) {
	a.emit(slog.LevelDebug, fmt.Sprintf(format, v...))
}
func (a *fiberAdapter) Infof(format string, v ...any) {
	a.emit(slog.LevelInfo, fmt.Sprintf(format, v...))
}
func (a *fiberAdapter) Warnf(format string, v ...any) {
	a.emit(slog.LevelWarn, fmt.Sprintf(format, v...))
}
func (a *fiberAdapter) Errorf(format string, v ...any) {
	a.emit(slog.LevelError, fmt.Sprintf(format, v...))
}
func (a *fiberAdapter) Fatalf(format string, v ...any) {
	a.emit(slog.LevelError, fmt.Sprintf(format, v...))
	os.Exit(1)
}
func (a *fiberAdapter) Panicf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	a.emit(slog.LevelError, msg)
	panic(msg)
}

// ---- WithLogger (structured key-value) ----

func (a *fiberAdapter) Tracew(msg string, keysAndValues ...any) {
	a.emit(slog.LevelDebug-4, msg)
}
func (a *fiberAdapter) Debugw(msg string, keysAndValues ...any) {
	a.emit(slog.LevelDebug, msg)
}
func (a *fiberAdapter) Infow(msg string, keysAndValues ...any) {
	a.emit(slog.LevelInfo, msg)
}
func (a *fiberAdapter) Warnw(msg string, keysAndValues ...any) {
	a.emit(slog.LevelWarn, msg)
}
func (a *fiberAdapter) Errorw(msg string, keysAndValues ...any) {
	a.emit(slog.LevelError, msg)
}
func (a *fiberAdapter) Fatalw(msg string, keysAndValues ...any) {
	a.emit(slog.LevelError, msg)
	os.Exit(1)
}
func (a *fiberAdapter) Panicw(msg string, keysAndValues ...any) {
	a.emit(slog.LevelError, msg)
	panic(msg)
}

// ---- WithContext ----

func (a *fiberAdapter) WithContext(_ context.Context) fiberlog.CommonLogger {
	return a
}
