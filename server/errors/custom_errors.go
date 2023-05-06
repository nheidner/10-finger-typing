package errors

import "fmt"

type HTTPError struct {
	Status  int
	Message string
	Details string
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Status, e.Message)
}
