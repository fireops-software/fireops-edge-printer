package error

import "fmt"

type ErrPdfGeneration string

// Error implements error.
func (e ErrPdfGeneration) Error() string {
	return fmt.Sprintf("ErrPdfGeneration: %s", string(e))
}

func NewErrPdfGeneration(format string, args ...any) error {
	return ErrPdfGeneration(fmt.Sprintf(format, args...))
}
