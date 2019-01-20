package sqi

import ()

var (
	badRequestErr = newBadRequestError("")
	conditionErr  = newConditionError("")
	evalErr       = newEvalError("")
	malformedErr  = newMalformedError("")
	mismatchErr   = newMismatchError("")
	parseErr      = newParseError("")
	unhandledErr  = newUnhandledError("")
)

// --------------------------------
// SQI-ERROR

func newBadRequestError(msg string) error {
	return &sqi_err_t{badRequestErrCode, msg, nil}
}

func newConditionError(msg string) error {
	return &sqi_err_t{conditionErrCode, msg, nil}
}

func newEvalError(msg string) error {
	return &sqi_err_t{evalErrCode, msg, nil}
}

func newMalformedError(msg string) error {
	return &sqi_err_t{malformedErrCode, msg, nil}
}

func newMismatchError(msg string) error {
	return &sqi_err_t{mismatchErrCode, msg, nil}
}

func newParseError(msg string) error {
	return &sqi_err_t{parseErrCode, msg, nil}
}

func newUnhandledError(msg string) error {
	return &sqi_err_t{unhandledErrCode, msg, nil}
}

type sqi_err_t struct {
	code int
	msg  string
	err  error
}

func (e *sqi_err_t) ErrorCode() int {
	return e.code
}

func (e *sqi_err_t) Error() string {
	var label string
	switch e.code {
	case badRequestErrCode:
		label = "sqi: bad request"
	case conditionErrCode:
		label = "sqi: condition"
	case evalErrCode:
		label = "sqi: eval"
	case malformedErrCode:
		label = "sqi: malformed"
	case mismatchErrCode:
		label = "sqi: mismatch"
	case parseErrCode:
		label = "sqi: parse"
	case unhandledErrCode:
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
	badRequestErrCode = 1000 + iota
	conditionErrCode
	evalErrCode
	malformedErrCode
	mismatchErrCode
	parseErrCode
	unhandledErrCode
)
