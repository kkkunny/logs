package logs

import (
	"errors"
	"fmt"
	"runtime"
)

// Error 异常
type Error interface {
	error
	Stack() StackFrame
}

type logError struct {
	stack StackFrame
	err   error
}

func newLogError(skip uint, err error) *logError {
	_, file, line, _ := runtime.Caller(int(skip) + 1)
	return &logError{
		stack: StackFrame{
			File: file,
			Line: uint(line),
		},
		err: err,
	}
}

// ErrorWrap 包装异常
func ErrorWrap(err error) Error {
	if err == nil {
		return nil
	}
	var logErr Error
	if errors.As(err, &logErr) {
		return logErr
	}
	return newLogError(1, err)
}

// ErrorWith 包装异常
func ErrorWith[T any](v T, err error) (T, Error) {
	if err == nil {
		return v, nil
	}
	var logErr Error
	if errors.As(err, &logErr) {
		return v, logErr
	}
	return v, newLogError(1, err)
}

// Errorf 新建异常
func Errorf(f string, a ...any) Error {
	return newLogError(1, fmt.Errorf(f, a...))
}

func (self *logError) Error() string {
	return self.err.Error()
}

// Stack 获取栈帧信息
func (self *logError) Stack() StackFrame {
	return self.stack
}
