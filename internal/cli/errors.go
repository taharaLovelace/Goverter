package cli

import "fmt"

type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string {
	return e.err.Error()
}

func (e *exitError) Unwrap() error {
	return e.err
}

func usageError(format string, args ...any) error {
	return &exitError{code: 2, err: fmt.Errorf(format, args...)}
}

func runtimeError(err error) error {
	return &exitError{code: 1, err: err}
}
