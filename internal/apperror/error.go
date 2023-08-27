package apperror

import (
	"encoding/json"
	"fmt"
)

var (
	ErrNotFound = NewAppError(nil,"Not Found","")
)

type AppError struct {
	Err              error  `json:"-"`
	Message          string `json:"message"`
	DeveloperMessage string `json:"developerMessage"`
}

func (e *AppError) Error() string {
	
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func (e *AppError) Marshal() []byte {
	marshal, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	return marshal
}
func NewAppError(err error, message, developerMessage string) *AppError {
	return &AppError{
		Err: fmt.Errorf(message),
		Message: message,
		DeveloperMessage: developerMessage,
	}
}

func systemError(err error) *AppError {
	return NewAppError(err,"Internal system error",err.Error())
}