package error

import "fmt"

const (
	RequestParamsInvalid string = "request params invalid. error: %s"
)

type RequestError string

func (me RequestError) Error() string {
	return string(me)
}

func NewRequestError(err error) RequestError {
	return RequestError(fmt.Sprintf(RequestParamsInvalid, err.Error()))
}
