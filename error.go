package uof

import "fmt"

// Error Inspiration:
// https://peter.bourgon.org/blog/2019/09/11/programming-with-errors.html
// https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html
// https://middlemost.com/failure-is-your-domain/
type Error struct {
	Severity ErrorSeverity
	Op       string // logical operation
	Inner    error  // nested error
}

func (e Error) Error() string {
	s := fmt.Sprintf("uof error op: %s", e.Op)
	if e.Severity == NoticeSeverity {
		s = fmt.Sprintf("NOTICE %s", s)
	}
	if e.Inner != nil {
		s = fmt.Sprintf("%s, inner: %v", s, e.Inner)
	}
	return s
}

// Unwrap inner error
func (e Error) Unwrap() error {
	return e.Inner
}

// APIError from API
type APIError struct {
	URL        string
	StatusCode int
	Response   string
	Inner      error
}

// Error status code
func (e APIError) Error() string {
	s := fmt.Sprintf("uof api error url: %s", e.URL)
	if e.StatusCode != 0 {
		s = fmt.Sprintf("%s, status code: %d", s, e.StatusCode)
	}
	if e.Response != "" {
		s = fmt.Sprintf("%s, response: %s", s, e.Response)
	}
	if e.Inner != nil {
		s = fmt.Sprintf("%s, inner: %v", s, e.Inner)
	}
	return s

}

// Unwrap API inner error
func (e APIError) Unwrap() error {
	return e.Inner
}

// E returns severity and inner error
func E(op string, inner error) Error {
	return Error{
		Severity: LogSeverity,
		Op:       op,
		Inner:    inner,
	}
}

// Notice error
func Notice(op string, inner error) Error {
	return Error{
		Severity: NoticeSeverity,
		Op:       op,
		Inner:    inner,
	}
}

// ErrorSeverity log level
type ErrorSeverity int8

// Log for later or notifiy
const (
	LogSeverity    ErrorSeverity = iota // log for later
	NoticeSeverity                      // oprator should be notified about this error
)
