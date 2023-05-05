package logs

import (
	"errors"
	"fmt"
	"runtime"
)

// Error 异常
type Error interface {
	error
	Stack() runtime.Frame
	Stacks() []runtime.Frame
	Unwrap() error
}

type logError struct {
	stacks []runtime.Frame
	err   error
}

func newLogError(skip uint, err error) *logError {
	var reverseStacks []runtime.Frame
	pcs := make([]uintptr, 20)

	n := runtime.Callers(int(skip)+2, pcs)
	frames := runtime.CallersFrames(pcs[:n-1])
	for frame, exist := frames.Next(); exist; frame, exist = frames.Next() {
		if !exist {
			break
		}
		reverseStacks = append(reverseStacks, frame)
	}

	stacks := make([]runtime.Frame, len(reverseStacks))
	for i, s := range reverseStacks{
		stacks[len(reverseStacks)-i-1] = s
	}

	return &logError{
		stacks: stacks,
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

// Stacks 获取栈帧信息
func (self *logError) Stacks() []runtime.Frame {
	return self.stacks
}

// Stack 获取栈帧信息
func (self *logError) Stack() runtime.Frame {
	return self.stacks[len(self.stacks)-1]
}

func (self *logError) Unwrap() error {
	return self.err
}
