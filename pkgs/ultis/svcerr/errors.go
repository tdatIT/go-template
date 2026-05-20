package svcerr

type Error struct {
	Status               int
	InternalErrorMessage string
	Code                 int    `json:"code"`
	Message              string `json:"message"`
}

func (e *Error) Error() string {
	if e.InternalErrorMessage != "" {
		return e.InternalErrorMessage
	}
	return e.Message
}

func (e *Error) WithError(err error) error {
	e.InternalErrorMessage = err.Error()
	return e
}
