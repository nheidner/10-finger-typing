package errors

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
)

// operation
type Op string

// user facing messages
type Messages map[string]string

type Error struct {
	err      error
	op       Op
	status   int
	username string
	stack    string
	messages Messages
}

func New(op Op, details ...any) error {
	e := &Error{
		op:       op,
		stack:    string(debug.Stack()),
		status:   http.StatusInternalServerError,
		messages: Messages{},
	}

	for _, detail := range details {
		switch t := detail.(type) {
		case *Error:
			for k, v := range t.Message() {
				e.messages[k] = v
			}

			e.err = t

			if e.status == http.StatusInternalServerError {
				e.status = t.Status()
			}
		case error:
			e.err = t
		case string:
			e.username = t
		case int:
			e.status = t
		case Messages:
			for k, v := range t {
				e.messages[k] = v
			}
		}
	}

	return e
}

// for debugging
func (e *Error) Error() string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s: ", string(e.op))

	if e.err != nil {
		b.WriteString(e.err.Error())
	}

	if e.username != "" {
		fmt.Fprintf(&b, "\n%s", e.username)
	}

	fmt.Fprintf(&b, "\n%s", e.stack)

	return b.String()
}

func (e *Error) Message() Messages {
	if len(e.messages) == 0 {
		switch e.status {
		case http.StatusBadRequest:
			return Messages{"message": "The request was invalid"}
		case http.StatusUnauthorized:
			return Messages{"message": "Unauthorized"}
		case http.StatusForbidden:
			return Messages{"message": "You do not have permission to perform this action"}
		case http.StatusNotFound:
			return Messages{"message": "The requested resource was not found"}
		case http.StatusUnsupportedMediaType:
			return Messages{"message": "Unsupported content-type"}
		default:
			return Messages{"message": "Something went wrong"}
		}
	}

	return e.messages
}

func (e *Error) Status() int {
	if e.status != 0 {
		return e.status
	}

	return http.StatusInternalServerError
}

func (e *Error) Unwrap() error {
	return e.err
}
