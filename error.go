package logs

import (
	"errors"
	"fmt"
	"runtime"
)

// Error 异常
type Error struct {
	stack StackFrame
	err   error
}

// WrapError 包装异常
func WrapError(e error) *Error {
	_, file, line, _ := runtime.Caller(1)
	return &Error{
		stack: StackFrame{
			File: file,
			Line: uint(line),
		},
		err: e,
	}
}

// NewError 新建异常
func NewError(f string, a ...any) *Error {
	_, file, line, _ := runtime.Caller(1)
	return &Error{
		stack: StackFrame{
			File: file,
			Line: uint(line),
		},
		err: fmt.Errorf(f, a...),
	}
}

func (self *Error) Error() string {
	return self.err.Error()
}

// Stack 获取栈帧信息
func (self *Error) Stack() StackFrame {
	return self.stack
}

// Stacks 获取所有栈帧信息
func (self *Error) Stacks() (stacks []StackFrame) {
	var err *Error
	var cursor error = self
	for errors.As(cursor, &err) {
		stacks = append(stacks, err.Stack())
		cursor = err.err
	}
	return stacks
}

// SubError 获取子异常信息
func (self *Error) SubError() error {
	return self.err
}
