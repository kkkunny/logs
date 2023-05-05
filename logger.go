package logs

import (
	"errors"
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
	LogLevelTrace                 // trace
	LogLevelInfo                  // info
	LogLevelWarn                  // warn
	LogLevelError                 // error
)

var logLevelStringMap = [...]string{
	LogLevelDebug: " DEBUG ",
	LogLevelTrace: " TRACE ",
	LogLevelInfo:  " INFO  ",
	LogLevelWarn:  " WARN  ",
	LogLevelError: " ERROR ",
}

var logLevelColorMap = [...]color.Color{
	LogLevelDebug: color.Blue,
	LogLevelTrace: color.Cyan,
	LogLevelInfo:  color.Green,
	LogLevelWarn:  color.Yellow,
	LogLevelError: color.Red,
}

var logLevelStyleMap = [...]color.Style{
	LogLevelDebug: color.New(color.OpBold, color.White, color.BgBlue),
	LogLevelTrace: color.New(color.OpBold, color.White, color.BgCyan),
	LogLevelInfo:  color.New(color.OpBold, color.White, color.BgGreen),
	LogLevelWarn:  color.New(color.OpBold, color.White, color.BgYellow),
	LogLevelError: color.New(color.OpBold, color.White, color.BgRed),
}

// Logger 日志管理器
type Logger struct {
	level  LogLevel
	values *linkedhashmap.LinkedHashMap[string, string]
	writer *log.Logger
}

// DefaultLogger 默认日志管理器
func DefaultLogger(debug bool, values ...any) *Logger {
	if debug {
		return NewLogger(LogLevelDebug, os.Stdout, values...)
	}
	return NewLogger(LogLevelInfo, os.Stdout, values...)
}

// NewLogger 新建日志管理器
func NewLogger(level LogLevel, writer io.Writer, values ...any) *Logger {
	if len(values)%2 != 0 {
		panic("The length of the values must be an even number")
	}
	valueMap := linkedhashmap.NewLinkedHashMap[string, string]()
	for i, value := range values {
		if i%2 != 0 {
			valueMap.Set(values[i-1].(string), fmt.Sprintf("%v", value))
		}
	}
	return &Logger{
		level:  level,
		values: valueMap,
		writer: log.New(writer, "", 0),
	}
}

func (self *Logger) NewGroup(values ...any) *Logger {
	if len(values)%2 != 0 {
		panic("The length of the values must be an even number")
	}
	valueMap := linkedhashmap.NewLinkedHashMap[string, string]()
	for iter := self.values.Begin(); iter != nil; iter.Next() {
		valueMap.Set(iter.Key(), iter.Value())
		if !iter.HasNext() {
			break
		}
	}
	for i, value := range values {
		if i%2 != 0 {
			valueMap.Set(values[i-1].(string), fmt.Sprintf("%v", value))
		}
	}
	return &Logger{
		level:  self.level,
		values: valueMap,
		writer: self.writer,
	}
}

// 输出
func (self *Logger) output(level LogLevel, pos string, values *linkedhashmap.LinkedHashMap[string, string]) error {
	var globalValueBuf strings.Builder
	var i int
	for iter := self.values.Begin(); iter != nil; iter.Next() {
		globalValueBuf.WriteByte('[')
		globalValueBuf.WriteString(iter.Key())
		globalValueBuf.WriteByte(']')
		globalValueBuf.WriteString(iter.Value())
		if !iter.HasNext() {
			break
		}
		globalValueBuf.WriteString(" | ")
		i++
	}

	var valueBuf strings.Builder
	i = 0
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
			"| %s | %s | %s | %s",
			timeStr,
			pos,
			globalValueBuf.String(),
			valueBuf.String(),
		)
		suffix = logLevelColorMap[level].Text(suffix)
		s = logLevelStyleMap[level].Sprintf(logLevelStringMap[level]) + suffix
	} else {
		s = fmt.Sprintf(
			"%s| %s | %s | %s | %s",
			logLevelStringMap[level],
			timeStr,
			pos,
			globalValueBuf.String(),
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

// 打印
func (self *Logger) print(level LogLevel, skip uint, a ...any) error {
	items := self.checkItems(a...)
	if self.level > level {
		return nil
	}
	return self.outputByStack(level, skip+1, items)
}

// 格式化打印
func (self *Logger) printf(level LogLevel, skip uint, f string, a ...any) error {
	return self.print(level, skip+1, "msg", fmt.Sprintf(f, a...))
}

// 打印异常
func (self *Logger) printError(level LogLevel, skip uint, err error) error {
	var logerr Error
	if errors.As(err, &logerr) {
		return self.printLogError(level, logerr)
	} else {
		return self.print(level, skip+1, "error", err.Error())
	}
}

// 打印带栈异常
func (self *Logger) printLogError(level LogLevel, err Error) error {
	if self.level > level {
		return nil
	}

	stacks := err.Stacks()

	var stackBuffer strings.Builder
	stackBuffer.WriteByte('\n')
	for i, s := range stacks {
		stackBuffer.WriteString(fmt.Sprintf("\t%s:%d", s.File, s.Line))
		if i < len(stacks)-1 {
			stackBuffer.WriteByte('\n')
		}
	}

	values := linkedhashmap.NewLinkedHashMap[string, string]()
	values.Set("error", err.Error())
	values.Set("stack", stackBuffer.String())
	stack := stacks[len(stacks)-1]
	return self.output(level, fmt.Sprintf("%s:%d", stack.File, stack.Line), values)
}

// Debug 输出Debug信息
func (self *Logger) Debug(skip uint, a ...any) error {
	return self.print(LogLevelDebug, skip+1, a...)
}

// Debugf 输出Debugf格式化信息
func (self *Logger) Debugf(skip uint, f string, a ...any) error {
	return self.printf(LogLevelDebug, skip+1, f, a...)
}

// DebugError 输出Debug异常信息
func (self *Logger) DebugError(skip uint, err error) error {
	return self.printError(LogLevelDebug, skip+1, err)
}

// Trace 输出Trace信息
func (self *Logger) Trace(skip uint, a ...any) error {
	return self.print(LogLevelTrace, skip+1, a...)
}

// Tracef 输出Trace格式化信息
func (self *Logger) Tracef(skip uint, f string, a ...any) error {
	return self.printf(LogLevelTrace, skip+1, f, a...)
}

// TraceError 输出Trace异常信息
func (self *Logger) TraceError(skip uint, err error) error {
	return self.printError(LogLevelTrace, skip+1, err)
}

// Info 输出Info信息
func (self *Logger) Info(skip uint, a ...any) error {
	return self.print(LogLevelInfo, skip+1, a...)
}

// Infof 输出Info格式化信息
func (self *Logger) Infof(skip uint, f string, a ...any) error {
	return self.printf(LogLevelInfo, skip+1, f, a...)
}

// InfoError 输出Info异常信息
func (self *Logger) InfoError(skip uint, err error) error {
	return self.printError(LogLevelInfo, skip+1, err)
}

// Warn 输出Warn信息
func (self *Logger) Warn(skip uint, a ...any) error {
	return self.print(LogLevelWarn, skip+1, a...)
}

// Warnf 输出Warn格式化信息
func (self *Logger) Warnf(skip uint, f string, a ...any) error {
	return self.printf(LogLevelWarn, skip+1, f, a...)
}

// WarnError 输出Warn异常信息
func (self *Logger) WarnError(skip uint, err error) error {
	return self.printError(LogLevelWarn, skip+1, err)
}

// Error 输出Error信息
func (self *Logger) Error(skip uint, a ...any) error {
	return self.print(LogLevelError, skip+1, a...)
}

// Errorf 输出Error格式化信息
func (self *Logger) Errorf(skip uint, f string, a ...any) error {
	return self.printf(LogLevelError, skip+1, f, a...)
}

// ErrorError 输出Error异常信息
func (self *Logger) ErrorError(skip uint, err error) error {
	return self.printError(LogLevelError, skip+1, err)
}
