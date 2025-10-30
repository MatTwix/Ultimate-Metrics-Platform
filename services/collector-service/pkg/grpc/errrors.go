package grpc

import "fmt"

type ParseTimeError struct {
	Field string
	Value string
	Err   error
}

func (e ParseTimeError) Error() string {
	return fmt.Sprintf("failed to parse %s '%s': %v", e.Field, e.Value, e.Err)
}

func (e ParseTimeError) Unwrap() error {
	return e.Err
}
