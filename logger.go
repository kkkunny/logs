package logs

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/kkkunny/containers/linkedhashmap"
)

// LogLevel 日志等级
type LogLevel uint8

const (
	LogLevelDebug LogLevel = iota // debug
	LogLevelInfo                  // info
	LogLevelWarn                  // warn
	LogLevelError                 // error
)

var logLevelStringMap = [...]string{
	LogLevelDebug: " DEBUG ",
	LogLevelInfo:  " INFO  ",
	LogLevelWarn:  " WARN  ",
	LogLevelError: " ERROR ",
}

var logLevelColorMap = [...]color.Color{
	LogLevelDebug: color.Blue,
	LogLevelInfo:  color.Green,
	LogLevelWarn:  color.Yellow,
	LogLevelError: color.Red,
}

var logLevelStyleMap = [...]color.Style{
	LogLevelDebug: color.New(color.White, color.BgBlue),
	LogLevelInfo:  color.New(color.White, color.BgGreen),
	LogLevelWarn:  color.New(color.White, color.BgYellow),
	LogLevelError: color.New(color.White, color.BgRed),
}

// Logger 日志管理器
type Logger struct {
	level    LogLevel
	writer   *log.Logger
}

// DefaultLogger 默认日志管理器
func DefaultLogger(debug bool) *Logger {
	if debug {
		return NewLogger(LogLevelDebug, os.Stdout)
	}
	return NewLogger(LogLevelInfo, os.Stdout)
}

// NewLogger 新建日志管理器
func NewLogger(level LogLevel, writer io.Writer) *Logger {
	return &Logger{
		level:    level,
		writer:   log.New(writer, "", 0),
	}
}

// 输出
func (self *Logger) output(level LogLevel, pos string, values *linkedhashmap.LinkedHashMap[string, string]) error {
	var valueBuf strings.Builder
	var i int
	for iter := values.Begin(); iter != nil; iter.Next() {
		valueBuf.WriteString(iter.Key())
		valueBuf.WriteByte('=')
		valueBuf.WriteString(iter.Value())
		if !iter.HasNext() {
			break
		}
		valueBuf.WriteByte(' ')
		i++
	}

	timeStr := time.Now().Format("2006-01-02 15:04:05")
	var s string
	writer := self.writer.Writer()
	if writer == os.Stdout || writer == os.Stderr {
		suffix := fmt.Sprintf(
			"| %s | %s | %s",
			timeStr,
			pos,
			valueBuf.String(),
		)
		suffix = logLevelColorMap[level].Text(suffix)
		s = logLevelStyleMap[level].Sprintf(logLevelStringMap[level]) + suffix
	} else {
		s = fmt.Sprintf(
			"%s| %s | %s | %s",
			logLevelStringMap[level],
			timeStr,
			pos,
			valueBuf.String(),
		)
	}
	return self.writer.Output(0, s)
}

func (self *Logger) outputByStack(
	level LogLevel, skip uint, values *linkedhashmap.LinkedHashMap[string, string],
) error {
	_, file, line, _ := runtime.Caller(int(skip + 1))
	return self.output(level, fmt.Sprintf("%s:%d", file, line), values)
}

// 检查item
func (self *Logger) checkItems(a ...any) *linkedhashmap.LinkedHashMap[string, string] {
	if len(a)%2 != 0 {
		panic("The number of items needs to be an even number")
	}

	items := linkedhashmap.NewLinkedHashMap[string, string]()
	for i, aa := range a {
		if i%2 != 0 {
			items.Set(fmt.Sprintf("%v", a[i-1]), fmt.Sprintf("%v", aa))
		}
	}
	return items
}

// Debug 输出Debug信息
func (self *Logger) Debug(skip uint, a ...any) error {
	items := self.checkItems(a...)
	if self.level > LogLevelDebug {
		return nil
	}
	return self.outputByStack(LogLevelDebug, skip+1, items)
}

// Debugf 输出Debugf格式化信息
func (self *Logger) Debugf(skip uint, f string, a ...any) error {
	return self.Debug(skip+1, "msg", fmt.Sprintf(f, a...))
}

// DebugError 输出Debug异常信息
func (self *Logger) DebugError(err Error) error {
	if self.level > LogLevelDebug {
		return nil
	}
	values := linkedhashmap.NewLinkedHashMap[string, string]()
	values.Set("msg", err.Error())
	stack := err.Stack()
	return self.output(LogLevelDebug, fmt.Sprintf("%s:%d", stack.File, stack.Line), values)
}

// Info 输出Info信息
func (self *Logger) Info(skip uint, a ...any) error {
	items := self.checkItems(a...)
	if self.level > LogLevelInfo {
		return nil
	}
	return self.outputByStack(LogLevelInfo, skip+1, items)
}

// Infof 输出Infof格式化信息
func (self *Logger) Infof(skip uint, f string, a ...any) error {
	return self.Info(skip+1, "msg", fmt.Sprintf(f, a...))
}

// InfoError 输出Info异常信息
func (self *Logger) InfoError(err Error) error {
	if self.level > LogLevelInfo {
		return nil
	}
	values := linkedhashmap.NewLinkedHashMap[string, string]()
	values.Set("msg", err.Error())
	stack := err.Stack()
	return self.output(LogLevelInfo, fmt.Sprintf("%s:%d", stack.File, stack.Line), values)
}

// Warn 输出Warn信息
func (self *Logger) Warn(skip uint, a ...any) error {
	items := self.checkItems(a...)
	if self.level > LogLevelWarn {
		return nil
	}
	return self.outputByStack(LogLevelWarn, skip+1, items)
}

// Warnf 输出Warnf格式化信息
func (self *Logger) Warnf(skip uint, f string, a ...any) error {
	return self.Warn(skip+1, "msg", fmt.Sprintf(f, a...))
}

// WarnError 输出Warn异常信息
func (self *Logger) WarnError(err Error) error {
	if self.level > LogLevelWarn {
		return nil
	}
	values := linkedhashmap.NewLinkedHashMap[string, string]()
	values.Set("msg", err.Error())
	stack := err.Stack()
	return self.output(LogLevelWarn, fmt.Sprintf("%s:%d", stack.File, stack.Line), values)
}

// Error 输出Error信息
func (self *Logger) Error(skip uint, a ...any) error {
	items := self.checkItems(a...)
	if self.level > LogLevelError {
		return nil
	}
	return self.outputByStack(LogLevelError, skip+1, items)
}

// Errorf 输出Errorf格式化信息
func (self *Logger) Errorf(skip uint, f string, a ...any) error {
	return self.Error(skip+1, "msg", fmt.Sprintf(f, a...))
}

// ErrorError 输出Error异常信息
func (self *Logger) ErrorError(err Error) error {
	if self.level > LogLevelError {
		return nil
	}
	values := linkedhashmap.NewLinkedHashMap[string, string]()
	values.Set("msg", err.Error())
	stack := err.Stack()
	return self.output(LogLevelError, fmt.Sprintf("%s:%d", stack.File, stack.Line), values)
}
