package OwmError

import "fmt"

type UserExistError struct {
	name string
}

func (err *UserExistError) Error() string {
	return fmt.Sprintf("User %s already exists", err.name)
}
