package OwmError

type AccessDeniedError struct {
	Info string
}

func (e AccessDeniedError) Error() string {
	return e.Info
}

func GetAccessDeniedError(message string) AccessDeniedError {
	return AccessDeniedError{
		Info: message,
	}
}
