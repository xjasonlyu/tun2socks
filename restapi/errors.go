package restapi

var (
	ErrBadRequest    = newError("Body invalid")
	ErrUnauthorized  = newError("Unauthorized")
	ErrUninitialized = newError("Uninitialized")
)

var _ error = (*HTTPError)(nil)

// HTTPError is custom HTTP error for API
type HTTPError struct {
	Message string `json:"message"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

func newError(msg string) *HTTPError {
	return &HTTPError{Message: msg}
}
