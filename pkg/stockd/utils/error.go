package utils

import "fmt"

type errors struct {
	code    int
	message string
	data    []interface{}
}

type errorsInterface interface {
	Code() int
	Error() string
	Data() []interface{}
}

func GetCodeAndData(err error) (int, []interface{}) {
	if c, ok := err.(errorsInterface); ok {
		return c.Code(), c.Data()
	}
	return 0, nil
}

func New(code int, format string, messages ...interface{}) error {
	return Wrap(code, nil, format, messages...)
}

func Wrap(code int, err error, format string, messages ...interface{}) error {
	if err != nil {
		format = fmt.Sprintf("%s : %s", format, err.Error())
	}
	message := fmt.Sprintf(format, messages...)
	if format == "" {
		message = fmt.Sprintf(GetErrMsg(code, ""), messages...)
	}
	return &errors{code, message, messages}
}

func (e *errors) Code() int        { return e.code }
func (e *errors) Data() []interface{} { return e.data }

func (e *errors) Error() string {
	if e.message == "" {
		e.message = GetErrMsg(e.code, "")
	}
	return e.message
}
