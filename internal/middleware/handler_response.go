package middleware

import (
	"mime"
	"net/http"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

var streamContentTypes = []string{
	"text/event-stream",
	"application/octet-stream",
	"multipart/x-mixed-replace",
}

func HandlerResponse(r *ghttp.Request) {
	r.Middleware.Next()
	if r.Response.BufferLength() > 0 || r.Response.BytesWritten() > 0 {
		return
	}
	mediaType, _, _ := mime.ParseMediaType(r.Response.Header().Get("Content-Type"))
	for _, contentType := range streamContentTypes {
		if mediaType == contentType {
			return
		}
	}

	var (
		msg  string
		err  = r.GetError()
		res  = r.GetHandlerResponse()
		code = gerror.Code(err)
	)
	if err != nil {
		if code == gcode.CodeNil {
			code = gcode.CodeInternalError
		}
		msg = err.Error()
		r.Response.Status = statusFromCode(code)
	} else {
		if r.Response.Status > 0 && r.Response.Status != http.StatusOK {
			switch r.Response.Status {
			case http.StatusNotFound:
				code = gcode.CodeNotFound
			case http.StatusForbidden, http.StatusUnauthorized:
				code = gcode.CodeNotAuthorized
			case http.StatusBadRequest:
				code = gcode.CodeInvalidRequest
			default:
				code = gcode.CodeUnknown
			}
			err = gerror.NewCode(code, msg)
			r.SetError(err)
		} else {
			code = gcode.CodeOK
		}
		msg = code.Message()
	}
	r.Response.WriteJson(ghttp.DefaultHandlerResponse{Code: code.Code(), Message: msg, Data: res})
}

func statusFromCode(code gcode.Code) int {
	numericCode := code.Code()
	switch numericCode {
	case 409001:
		return http.StatusConflict
	case 429001:
		return http.StatusTooManyRequests
	}
	if numericCode >= http.StatusBadRequest && numericCode < http.StatusNetworkAuthenticationRequired {
		return numericCode
	}
	switch code {
	case gcode.CodeOK:
		return http.StatusOK
	case gcode.CodeValidationFailed, gcode.CodeInvalidParameter, gcode.CodeMissingParameter, gcode.CodeInvalidRequest, gcode.CodeBusinessValidationFailed:
		return http.StatusBadRequest
	case gcode.CodeNotAuthorized, gcode.CodeSecurityReason:
		return http.StatusForbidden
	case gcode.CodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
