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
		if e.Wrapped {
			panic(errors.WithMessagef(e.Prev, e.Message))
		} else {
			panic(errors.Wrapf(e.Prev, e.Message))
		}
	}
	if p != nil {
		panic(p)
	}
}

//func CheckWithWrap(err error, format string, args ...interface{}) {
//	if err != nil {
//		panic(errors.Wrapf(err, format, args...))
//	}
//}
//func CheckWithMessage(err error, format string, args ...interface{}) {
//	if err != nil {
//		panic(errors.WithMessagef(err, format, args...))
//	}
//}
//
//func MyRecover() {
//	err := recover()
//	if err == nil {
//		return
//	}
//	e, ok := err.(Error)
//	if ok {
//		for {
//			//w.Header().Add("ErrorCode", e.Code)
//			//w.Header().Add("ErrorMessage", e.Message)
//			if e.Prev == nil {
//				break
//			}
//			prev, ok := e.Prev.(Error)
//			if !ok {
//				break
//			}
//			e = prev
//		}
//	} else {
//		panic(err)
//	}
//}
