package error

import "fmt"

type ErrPrinter string

// Error implements error.
func (e ErrPrinter) Error() string {
	return fmt.Sprintf("ErrPrinter: %s", string(e))
}

func NewErrPrinter(format string, arg ...any) error {
	return ErrPrinter(fmt.Sprintf(format, arg...))
}
