package util

import (
	"github.com/gin-gonic/gin"
	"strings"
)

// SetCookie - Setting cookies
func SetCookie(ctx *gin.Context, accessToken, refreshToken, domain string) {
	// verify domain is not localhost and is not dot prefix add dot
	if domain != "localhost" && !strings.HasPrefix(domain, ".") {
		domain = "." + domain
	}
	ctx.SetCookie("access_token", accessToken, 60*60*24, "/", domain, true, true)
	ctx.SetCookie("refresh_token", refreshToken, 60*60*24*30, "/", domain, true, true)
}

// SetRegisterCookie - Setting register cookie
func SetRegisterCookie(ctx *gin.Context, registerToken, domain string) {
	// verify domain is not localhost and is not dot prefix add dot
	if domain != "localhost" && !strings.HasPrefix(domain, ".") {
		domain = "." + domain
	}
	ctx.SetCookie("register_token", registerToken, 60*60, "/", domain, true, true)
}
