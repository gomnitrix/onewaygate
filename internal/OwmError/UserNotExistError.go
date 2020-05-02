package OwmError

import "fmt"

type UserNotExistError struct {
	name string
}

func (err UserNotExistError) Error() string {
	return fmt.Sprintf("No Such User: %s", err.name)
}

func GetUserNotExistError(name string) UserNotExistError {
	return UserNotExistError{name: name}
}
