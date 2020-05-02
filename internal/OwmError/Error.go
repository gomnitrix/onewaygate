package OwmError

import (
	"fmt"

	"github.com/pkg/errors"
)

type Error struct {
	Wrapped bool
	Message string
	Prev    error
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) Unwrap() error {
	return e.Prev
}

func Check(err error, wrapped bool, format string, args ...interface{}) {
	if err != nil {
		panic(Error{
			Wrapped: wrapped,
			Message: fmt.Sprintf(format, args),
			Prev:    err,
		})
	}
}

func Pack() {
	p := recover()
	if e, ok := p.(Error); ok {
		if !e.Wrapped {
			panic(Error{
				Wrapped: true,
				Message: "",
				Prev:    errors.Wrap(e.Prev, e.Message),
			})
			//panic(errors.WithMessagef(e.Prev, e.Message)) 如果以后想在每次recover的时候加上一条消息就改这里
		}
	}
	if p != nil {
		panic(p)
	}
}
