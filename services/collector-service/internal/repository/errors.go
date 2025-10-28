package repository

import (
	"fmt"
	"strings"
)

type BatchInsertError struct {
	SuccessfullCount int
	FailedCount      int
	Errors           []error
}

func (e *BatchInsertError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"batch insertion completed with %d successes and %d failures. Failures:",
		e.SuccessfullCount, e.FailedCount,
	))

	for _, err := range e.Errors {
		sb.WriteString(fmt.Sprintf("\n- %v", err))
	}

	return sb.String()
}
