package sqi

import ()

var (
	BadRequestErr = NewBadRequestError("")
	MismatchErr   = NewMismatchError("")
	UnhandledErr  = NewUnhandledError("")
)

// --------------------------------
// PHLY-ERROR

func NewBadRequestError(msg string) error {
	return &SqiError{BadRequestErrCode, msg, nil}
}

func NewMismatchError(msg string) error {
	return &SqiError{MismatchErrCode, msg, nil}
}

func NewUnhandledError(msg string) error {
	return &SqiError{UnhandledErrCode, msg, nil}
}

type SqiError struct {
	code int
	msg  string
	err  error
}

func (e *SqiError) ErrorCode() int {
	return e.code
}

func (e *SqiError) Error() string {
	var label string
	switch e.code {
	case BadRequestErrCode:
		label = "sqi: bad request"
	case MismatchErrCode:
		label = "sqi: mismatch"
	case UnhandledErrCode:
		label = "sqi: unhandled"
	default:
		label = "sqi: error"
	}
	if e.msg != "" {
		label += " (" + e.msg + ")"
	}
	if e.err != nil {
		label += " (" + e.err.Error() + ")"
	}
	return label
}

// --------------------------------
// MISC

// mergeErrors() answers the first non-nil error in the list.
func mergeErrors(err ...error) error {
	for _, a := range err {
		if a != nil {
			return a
		}
	}
	return nil
}

// --------------------------------
// CONST and VAR

const (
	BadRequestErrCode = 1000 + iota
	MismatchErrCode
	UnhandledErrCode
)
