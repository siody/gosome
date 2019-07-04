package http

import (
	"bytes"
	"encoding/json"
	"github.com/kataras/iris"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"io"
	"minerva.devops.letv.com/scloud/stargazer-base-lib/logzap"
	"net/http"
)

//CodeMsg code msg
var CodeMsg map[int]string

//Response template response of stargazer
type Response struct {
	ErrNo  int         `json:"errno"`
	ErrMsg string      `json:"errmsg"`
	Data   interface{} `json:"data"`
}

//LoadErrorCode load from viper
func LoadErrorCode() {
	viper.UnmarshalKey("errorcode", &CodeMsg)
}

//IrisError iris framework return error
func IrisError(c iris.Context, code int, e error) {
	var msg string
	if m, ok := CodeMsg[code]; ok {
		msg = m + " " + e.Error()
	} else {
		msg = "error: " + e.Error()
	}
	encodelog(c, &Response{ErrNo: code, ErrMsg: msg, Data: nil}, "error response")
}

//IrisSuccess iris framework return success
func IrisSuccess(c iris.Context, data interface{}) {
	encodelog(c, &Response{ErrNo: 10000, ErrMsg: "", Data: data}, "success response")
}

func encodelog(c iris.Context, response *Response, logprefix string) {
	var logstring bytes.Buffer
	responslog := io.MultiWriter(c, &logstring)
	enc := json.NewEncoder(responslog)
	enc.SetEscapeHTML(true)
	err := enc.Encode(response)
	if err != nil {
		c.StatusCode(http.StatusInternalServerError)
		logzap.Logger.Error(logprefix+" marshal failed", zap.Error(err))
		return
	}
	logzap.Logger.Info(logprefix+" OK", zap.String("data", logstring.String()))
}
