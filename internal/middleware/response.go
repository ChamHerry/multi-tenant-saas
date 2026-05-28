package middleware

import (
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"
)

type ErrorEnvelope struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func writeError(r *ghttp.Request, status int, code string, err error) {
	message := http.StatusText(status)
	if err != nil && err.Error() != "" {
		message = err.Error()
	}
	r.Response.Status = status
	r.Response.WriteJsonExit(ErrorEnvelope{Code: code, Message: message, Data: nil})
}
