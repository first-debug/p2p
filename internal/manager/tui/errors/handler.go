// Package errors предоставляет обработчики ошибок для TUI менеджера.
package errors

import (
	"fmt"
	"log"
	"os"
)

// logger - логгер для вывода отладочной информации об ошибках.
var Logger = log.New(os.Stderr, "[TUI Errors] ", log.LstdFlags)

// ErrorType определяет тип ошибки для классификации.
type ErrorType int

const (
	// ErrorTypeUnknown - неизвестная ошибка
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeNetwork - сетевая ошибка
	ErrorTypeNetwork
	// ErrorTypeStorage - ошибка хранилища
	ErrorTypeStorage
	// ErrorTypeSession - ошибка сессии
	ErrorTypeSession
	// ErrorTypeUI - ошибка UI
	ErrorTypeUI
	// ErrorTypeValidation - ошибка валидации
	ErrorTypeValidation
)

// String возвращает строковое представление типа ошибки.
func (e ErrorType) String() string {
	switch e {
	case ErrorTypeNetwork:
		return "Network"
	case ErrorTypeStorage:
		return "Storage"
	case ErrorTypeSession:
		return "Session"
	case ErrorTypeUI:
		return "UI"
	case ErrorTypeValidation:
		return "Validation"
	default:
		return "Unknown"
	}
}

// TUIError представляет ошибку TUI с дополнительной информацией.
type TUIError struct {
	Type    ErrorType
	Message string
	Cause   error
	Context map[string]interface{}
}

// Error возвращает строковое представление ошибки.
func (e *TUIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap возвращает причину ошибки.
func (e *TUIError) Unwrap() error {
	return e.Cause
}

// NewError создаёт новую ошибку TUI.
func NewError(errType ErrorType, message string, cause error) *TUIError {
	return &TUIError{
		Type:    errType,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// NewErrorf создаёт новую ошибку TUI с форматированием.
func NewErrorf(errType ErrorType, format string, args ...interface{}) *TUIError {
	return &TUIError{
		Type:    errType,
		Message: fmt.Sprintf(format, args...),
		Context: make(map[string]interface{}),
	}
}

// WithContext добавляет контекст к ошибке.
func (e *TUIError) WithContext(key string, value interface{}) *TUIError {
	e.Context[key] = value
	return e
}

// Log логирует ошибку.
func (e *TUIError) Log() {
	Logger.Printf("Error: %v", e)
	if len(e.Context) > 0 {
		Logger.Printf("Context: %v", e.Context)
	}
}

// HandleError обрабатывает ошибку и возвращает пользовательское сообщение.
func HandleError(err error, defaultMsg string) string {
	if err == nil {
		return ""
	}

	// Логируем ошибку
	if tuiErr, ok := err.(*TUIError); ok {
		tuiErr.Log()
		return tuiErr.Message
	}

	Logger.Printf("Error: %v", err)
	return defaultMsg
}

// HandleErrorWithFallback обрабатывает ошибку и возвращает fallback значение.
func HandleErrorWithFallback[T any](err error, fallback T, msg string) T {
	if err != nil {
		HandleError(err, msg)
		return fallback
	}
	return fallback
}

// SuppressError подавляет ошибку, логируя её.
func SuppressError(err error, context string) {
	if err != nil {
		Logger.Printf("[%s] Error suppressed: %v", context, err)
	}
}

// Must паникует, если ошибка не nil.
func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// WrapError оборачивает существующую ошибку в TUIError.
func WrapError(err error, errType ErrorType, message string) *TUIError {
	if err == nil {
		return nil
	}
	return &TUIError{
		Type:    errType,
		Message: message,
		Cause:   err,
		Context: make(map[string]interface{}),
	}
}

// IsErrorType проверяет, является ли ошибка указанным типом.
func IsErrorType(err error, errType ErrorType) bool {
	if tuiErr, ok := err.(*TUIError); ok {
		return tuiErr.Type == errType
	}
	return false
}
