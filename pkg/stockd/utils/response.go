package utils

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type HTTPResponse struct {
	RequestID string      `json:"requestID"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
}

func httpResponse(requestID string, code int, msg string, data interface{}) *HTTPResponse {
	return &HTTPResponse{RequestID: requestID, Code: code, Message: msg, Data: data}
}

func getCaller() string {
	pc, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s:%d:%s", file, line, runtime.FuncForPC(pc).Name())
}

func HTTPRequestFailed(c *gin.Context, err error) {
	HTTPRequestFailedV4(c, err, ERROR)
}

func HTTPRequestFailedV4(c *gin.Context, err error, code int, data ...interface{}) {
	if err != nil {
		logrus.WithField(string(ContextKeyReqID), GetReqID(c.Request.Context())).
			Errorf("%s -> Error: %s", getCaller(), err)
	}
	var (
		errCode int
		errData []interface{}
		msg     string
	)
	errCode, errData = GetCodeAndData(err)
	if errCode == 0 {
		errCode = code
		errData = data
	}
	if errCode == 0 {
		errCode = ERROR
	}
	if errData == nil {
		errData = data
	}
	if errCode == 600 {
		errCode = ERROR
		msg = err.Error()
	} else {
		msg = GetMsg(errCode, GetLang(c))
	}
	if len(errData) > 0 && (errCode != ERROR && errCode != 400) && strings.Contains(msg, "%") {
		msg = fmt.Sprintf(msg, errData...)
		msg = strings.TrimSpace(msg)
	}
	if err != nil && !strings.Contains(err.Error(), "%!(EXTRA") && err.Error() != msg {
		msg += " " + err.Error()
	}
	c.JSON(http.StatusOK, httpResponse(GetReqID(c.Request.Context()), errCode, msg, nil))
}

func HTTPRequestFailedV5(c *gin.Context, err error) {
	c.JSON(http.StatusOK, httpResponse(GetReqID(c.Request.Context()), ERROR, err.Error(), ""))
}

func HTTPRequestSuccess(c *gin.Context, code int, data interface{}) {
	c.JSON(http.StatusOK, httpResponse(
		GetReqID(c.Request.Context()), code,
		GetSuccessMsg(code, GetLang(c)), data,
	))
}
