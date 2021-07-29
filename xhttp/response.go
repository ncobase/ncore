package xhttp

import (
	"net/http"
	"stone/common/ecode"
	"stone/common/types"

	"github.com/gin-gonic/gin"
)

// ResponseException Response ResponseException Struct
type ResponseException struct {
	Status  int         `json:"status,omitempty"`  // http 状态
	Code    int         `json:"code,omitempty"`    // 业务代码
	Message string      `json:"message,omitempty"` // 消息
	Data    interface{} `json:"data,omitempty"`    // 返回数据
}

func response(code int, message string, data interface{}) *ResponseException {
	return &ResponseException{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func fail(r *ResponseException) (int, interface{}) {
	// 默认返回请求失败状态
	status := http.StatusBadRequest
	// 默认错误代码
	code := ecode.RequestErr
	message := ecode.Text(code)

	// 自定义状态码 ref: http.StatusCode
	if r.Status != 0 {
		status = r.Status
	}

	// 自定义业务代码
	if r.Code != 0 {
		code = r.Code
	}

	// 自定义消息
	if r.Message != "" {
		message = r.Message
	}

	return status, response(code, message, nil)
}

func success(r *ResponseException) (int, interface{}) {
	status := http.StatusOK
	// message := http.StatusText(status)

	if r == nil {
		return status, types.JSON{"message": "ok"}
	}
	// 如传入 status 且 不等于 0 则使用
	if r.Status != 0 {
		status = r.Status
	}

	// 非 200 - 400 范围内的状态码走 fail 方法
	if !(status >= 200 && status < 400) {
		return fail(r)
	}

	return status, r.Data
}

func write(context *gin.Context, contextType string, code int, res interface{}) {
	switch contextType {
	case "IndentedJSON":
		context.IndentedJSON(code, res)
	case "SecureJSON":
		context.SecureJSON(code, res)
	case "JSON":
		context.JSON(code, res)
	case "AsciiJSON":
		context.AsciiJSON(code, res)
	case "PureJSON":
		context.PureJSON(code, res)
	case "XML":
		context.XML(code, res)
	case "YAML":
		context.YAML(code, res)
	case "ProtoBuf":
		context.ProtoBuf(code, res)
	}
	context.Abort()
}

// Fail 执行失败后调用
func Fail(context *gin.Context, r *ResponseException) {
	// 默认输出 JSON 数据
	contextType := "JSON"
	// 构建输出数据
	statusCode, result := fail(r)
	write(context, contextType, statusCode, result)
}

// Success 执行成功后调用
func Success(context *gin.Context, r *ResponseException) {
	// 默认输出 JSON 数据
	contextType := "JSON"
	// 构建输出数据
	statusCode, result := success(r)
	write(context, contextType, statusCode, result)
}
