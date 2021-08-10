package util

import (
	"github.com/gin-gonic/gin"
	"strings"
)

// SetCookie 设置 Cookies
func SetCookie(ctx *gin.Context, accessToken, refreshToken, domain string) {
	// 验证域名非 localhost 并且非 . 开头则增加 .
	if domain != "localhost" && !strings.HasPrefix(domain, ".") {
		domain = "." + domain
	}
	ctx.SetCookie("access_token", accessToken, 60*60*24, "/", domain, true, true)
	ctx.SetCookie("refresh_token", refreshToken, 60*60*24*30, "/", domain, true, true)
}

// SetRegisterCookie 设置注册 Cookies
func SetRegisterCookie(ctx *gin.Context, registerToken, domain string) {
	// 验证域名非 localhost 并且非 . 开头则增加 .
	if domain != "localhost" && !strings.HasPrefix(domain, ".") {
		domain = "." + domain
	}
	ctx.SetCookie("register_token", registerToken, 60*60, "/", domain, true, true)
}
