// Любой тип с методом Error() string считается ошибкой, проблема простых ошибок: нет инвы о хосте,
// нельзя проверить тип программно, приходится парсить строку. Кастомные ошибки: есть инфа о хосте,
// можно проверить через errors.ls() и извлечь детали через errors.As()
// *PingError - передаем ошибки как указатель, не копируем структуру и указатели сравниваем по адресу
package errors

import (
	"fmt"
)

// Типы ошибок
// PingError - кастомная ошибка пинга хоста
type PingError struct {
	Host    string //IP или домен где произошла ошибка
	Message string //Краткое описание ошибки
	Err     error  // Оригинальная ошибка(для обертки)
}

// Error() - реализация интерфейса error
func (e *PingError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("ping %s: %s: %v", e.Host, e.Message, e.Err)
	}
	return fmt.Sprintf("ping %s: %s", e.Host, e.Message)
}

// Unwrap() - для поддержки errors.ls() и errors.As()
func (e *PingError) Unwrap() error {
	return e.Err
}

// Предопределенные ошибки

var (
	ErrTimeout         = &PingError{Message: "timeout"}
	ErrHostUnreachable = &PingError{Message: "host unreachable"}
	ErrDNSFailed       = &PingError{Message: "DNS resolution failed"}
	ErrCanceled        = &PingError{Message: "operation canceled"}
)

// Фабрики ошибок

// NewTimeoutError - создает ошибку таймаута
func NewTimeoutError(host string, timeout interface{}) *PingError {
	return &PingError{
		Host:    host,
		Message: fmt.Sprintf("timeout after %v", timeout),
		Err:     ErrTimeout,
	}
}

// NewHostUnreachableError - создает ошибку недоступности
func NewHostUnreachableError(host string, err error) *PingError {
	return &PingError{
		Host:    host,
		Message: "host unreachable",
		Err:     err,
	}
}

// NewDNSError - создает ошибку DNS
func NewDNSError(host string, err error) *PingError {
	return &PingError{
		Host:    host,
		Message: "DNS resolution failed",
		Err:     err,
	}
}

// NewCanceledError - создает ошибку отмены
func NewCanceledError(host string) *PingError {
	return &PingError{
		Host:    host,
		Message: "operation canceled by user",
		Err:     ErrCanceled,
	}
}
