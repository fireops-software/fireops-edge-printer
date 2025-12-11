package error

import "fmt"

type ErrIo string

// Error implements error.
func (e ErrIo) Error() string {
	return fmt.Sprintf("ErrIo: %s", string(e))
}

func NewErrIo(format string, args ...any) error {
	return ErrIo(fmt.Sprintf(format, args...))
}
