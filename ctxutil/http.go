package ctxutil

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	clientIPKey    = "client_ip"
	userAgentKey   = "user_agent"
	sessionIDKey   = "session_id"
	httpRequestKey = "http_request"
)

// SetHTTPRequest sets HTTP request to context.Context
func SetHTTPRequest(ctx context.Context, req *http.Request) context.Context {
	return SetValue(ctx, httpRequestKey, req)
}

// GetHTTPRequest gets HTTP request from context.Context
func GetHTTPRequest(ctx context.Context) *http.Request {
	if req, ok := GetValue(ctx, httpRequestKey).(*http.Request); ok {
		return req
	}
	// Try to get from Gin context
	if ginCtx, ok := GetGinContext(ctx); ok && ginCtx.Request != nil {
		return ginCtx.Request
	}
	return nil
}

// SetClientIP sets client IP to context.Context
func SetClientIP(ctx context.Context, ip string) context.Context {
	return SetValue(ctx, clientIPKey, ip)
}

// GetClientIP gets client IP from context.Context
func GetClientIP(ctx context.Context) string {
	// First try to get from context value (if previously set)
	if ip, ok := GetValue(ctx, clientIPKey).(string); ok && ip != "" {
		return ip
	}

	// Try to get from Gin context
	if ginCtx, ok := GetGinContext(ctx); ok {
		return getClientIPFromGin(ginCtx)
	}

	// Try to get from HTTP request
	if req := GetHTTPRequest(ctx); req != nil {
		return getClientIPFromRequest(req)
	}

	return "unknown"
}

// SetUserAgent sets user agent to context.Context
func SetUserAgent(ctx context.Context, userAgent string) context.Context {
	return SetValue(ctx, userAgentKey, userAgent)
}

// GetUserAgent gets user agent from context.Context
func GetUserAgent(ctx context.Context) string {
	// First try to get from context value (if previously set)
	if ua, ok := GetValue(ctx, userAgentKey).(string); ok && ua != "" {
		return ua
	}

	// Try to get from Gin context
	if ginCtx, ok := GetGinContext(ctx); ok {
		if userAgent := ginCtx.GetHeader("User-Agent"); userAgent != "" {
			return userAgent
		}
	}

	// Try to get from HTTP request
	if req := GetHTTPRequest(ctx); req != nil {
		if userAgent := req.Header.Get("User-Agent"); userAgent != "" {
			return userAgent
		}
	}

	return "unknown"
}

// SetSessionID sets session ID to context.Context
func SetSessionID(ctx context.Context, sessionID string) context.Context {
	return SetValue(ctx, sessionIDKey, sessionID)
}

// GetSessionID gets session ID from context.Context
func GetSessionID(ctx context.Context) string {
	// First try to get from context value (if previously set)
	if sessionID, ok := GetValue(ctx, sessionIDKey).(string); ok && sessionID != "" {
		return sessionID
	}

	// Try to get from Gin context
	if ginCtx, ok := GetGinContext(ctx); ok {
		// From Cookie
		if sessionID, err := ginCtx.Cookie("session_id"); err == nil && sessionID != "" {
			return sessionID
		}
		// From Header
		if sessionID := ginCtx.GetHeader("X-Session-ID"); sessionID != "" {
			return sessionID
		}
	}

	// Try to get from HTTP request
	if req := GetHTTPRequest(ctx); req != nil {
		// From Cookie
		if cookie, err := req.Cookie("session_id"); err == nil && cookie.Value != "" {
			return cookie.Value
		}
		// From Header
		if sessionID := req.Header.Get("X-Session-ID"); sessionID != "" {
			return sessionID
		}
	}

	return ""
}

// SetClientInfo sets client information to context.Context
func SetClientInfo(ctx context.Context, ip, userAgent, sessionID string) context.Context {
	ctx = SetClientIP(ctx, ip)
	ctx = SetUserAgent(ctx, userAgent)
	if sessionID != "" {
		ctx = SetSessionID(ctx, sessionID)
	}
	return ctx
}

// GetClientInfo gets all client information from context.Context
func GetClientInfo(ctx context.Context) (ip, userAgent, sessionID string) {
	return GetClientIP(ctx), GetUserAgent(ctx), GetSessionID(ctx)
}

// getClientIPFromGin gets client IP from Gin context
func getClientIPFromGin(c *gin.Context) string {
	// 1. Check X-Forwarded-For header (for multi-level proxy)
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For may contain multiple IPs, take the first one
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" && !isPrivateIP(ip) {
				return ip
			}
		}
	}

	// 2. Check X-Real-IP header (for Nginx proxy)
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" && !isPrivateIP(xRealIP) {
		return xRealIP
	}

	// 3. Check CF-Connecting-IP header (for Cloudflare)
	cfConnectingIP := c.GetHeader("CF-Connecting-IP")
	if cfConnectingIP != "" && !isPrivateIP(cfConnectingIP) {
		return cfConnectingIP
	}

	// 4. Check other common proxy headers
	proxyHeaders := []string{
		"X-Forwarded",
		"Forwarded-For",
		"Forwarded",
		"X-Client-IP",
		"X-Cluster-Client-IP",
	}

	for _, header := range proxyHeaders {
		ip := c.GetHeader(header)
		if ip != "" && !isPrivateIP(ip) {
			return ip
		}
	}

	// 5. Use Gin's ClientIP method (automatically handles proxies)
	clientIP := c.ClientIP()
	if clientIP != "" {
		return clientIP
	}

	// 6. Finally use RemoteAddr
	if c.Request != nil {
		return getIPFromAddr(c.Request.RemoteAddr)
	}

	return "unknown"
}

// getClientIPFromRequest gets client IP from standard HTTP request
func getClientIPFromRequest(req *http.Request) string {
	// 1. Check X-Forwarded-For header
	xForwardedFor := req.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" && !isPrivateIP(ip) {
				return ip
			}
		}
	}

	// 2. Check X-Real-IP header
	xRealIP := req.Header.Get("X-Real-IP")
	if xRealIP != "" && !isPrivateIP(xRealIP) {
		return xRealIP
	}

	// 3. Check CF-Connecting-IP header
	cfConnectingIP := req.Header.Get("CF-Connecting-IP")
	if cfConnectingIP != "" && !isPrivateIP(cfConnectingIP) {
		return cfConnectingIP
	}

	// 4. Use RemoteAddr
	return getIPFromAddr(req.RemoteAddr)
}

// getIPFromAddr extracts IP from address string
func getIPFromAddr(addr string) string {
	if addr == "" {
		return "unknown"
	}

	// Remove port number
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}

	return addr
}

// isPrivateIP checks if IP is private
func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return true // Invalid IP is considered private
	}

	// Check if it's a private network IP
	privateRanges := []string{
		"10.0.0.0/8",     // Class A
		"172.16.0.0/12",  // Class B
		"192.168.0.0/16", // Class C
		"127.0.0.0/8",    // Loopback
		"169.254.0.0/16", // Link Local
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 unique local
		"fe80::/10",      // IPv6 link local
	}

	for _, rangeStr := range privateRanges {
		_, subnet, err := net.ParseCIDR(rangeStr)
		if err != nil {
			continue
		}
		if subnet.Contains(ip) {
			return true
		}
	}

	return false
}

// UserAgentInfo represents parsed user agent information
type UserAgentInfo struct {
	Browser  string `json:"browser"`
	Version  string `json:"version"`
	OS       string `json:"os"`
	Platform string `json:"platform"`
	Mobile   bool   `json:"mobile"`
}

// ParseUserAgent parses user agent string into structured information
func ParseUserAgent(userAgent string) *UserAgentInfo {
	if userAgent == "" {
		return &UserAgentInfo{}
	}

	info := &UserAgentInfo{}
	ua := strings.ToLower(userAgent)

	// Detect mobile device
	mobileKeywords := []string{"mobile", "android", "iphone", "ipad", "windows phone"}
	for _, keyword := range mobileKeywords {
		if strings.Contains(ua, keyword) {
			info.Mobile = true
			break
		}
	}

	// Detect operating system
	if strings.Contains(ua, "windows") {
		info.OS = "Windows"
	} else if strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os") {
		info.OS = "macOS"
	} else if strings.Contains(ua, "linux") {
		info.OS = "Linux"
	} else if strings.Contains(ua, "android") {
		info.OS = "Android"
	} else if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
		info.OS = "iOS"
	}

	// Detect browser
	if strings.Contains(ua, "edge") {
		info.Browser = "Microsoft Edge"
	} else if strings.Contains(ua, "chrome") {
		info.Browser = "Google Chrome"
	} else if strings.Contains(ua, "firefox") {
		info.Browser = "Mozilla Firefox"
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		info.Browser = "Safari"
	} else if strings.Contains(ua, "opera") {
		info.Browser = "Opera"
	}

	return info
}

// GetParsedUserAgent gets parsed user agent information from context
func GetParsedUserAgent(ctx context.Context) *UserAgentInfo {
	userAgent := GetUserAgent(ctx)
	if userAgent == "unknown" || userAgent == "" {
		return &UserAgentInfo{}
	}
	return ParseUserAgent(userAgent)
}

// GetReferer gets HTTP referer from context
func GetReferer(ctx context.Context) string {
	// Try to get from Gin context
	if ginCtx, ok := GetGinContext(ctx); ok {
		if referer := ginCtx.GetHeader("Referer"); referer != "" {
			return referer
		}
	}

	// Try to get from HTTP request
	if req := GetHTTPRequest(ctx); req != nil {
		if referer := req.Header.Get("Referer"); referer != "" {
			return referer
		}
	}

	return ""
}

// GetAcceptLanguage gets Accept-Language header from context
func GetAcceptLanguage(ctx context.Context) string {
	// Try to get from Gin context
	if ginCtx, ok := GetGinContext(ctx); ok {
		if lang := ginCtx.GetHeader("Accept-Language"); lang != "" {
			return lang
		}
	}

	// Try to get from HTTP request
	if req := GetHTTPRequest(ctx); req != nil {
		if lang := req.Header.Get("Accept-Language"); lang != "" {
			return lang
		}
	}

	return ""
}

// GetRequestMethod gets HTTP method from context
func GetRequestMethod(ctx context.Context) string {
	// Try to get from Gin context
	if ginCtx, ok := GetGinContext(ctx); ok && ginCtx.Request != nil {
		return ginCtx.Request.Method
	}

	// Try to get from HTTP request
	if req := GetHTTPRequest(ctx); req != nil {
		return req.Method
	}

	return ""
}

// GetRequestURI gets request URI from context
func GetRequestURI(ctx context.Context) string {
	// Try to get from Gin context
	if ginCtx, ok := GetGinContext(ctx); ok && ginCtx.Request != nil {
		return ginCtx.Request.RequestURI
	}

	// Try to get from HTTP request
	if req := GetHTTPRequest(ctx); req != nil {
		return req.RequestURI
	}

	return ""
}

// GetRequestHeaders gets all request headers from context
func GetRequestHeaders(ctx context.Context) map[string]string {
	headers := make(map[string]string)

	// Try to get from Gin context
	if ginCtx, ok := GetGinContext(ctx); ok && ginCtx.Request != nil {
		for key, values := range ginCtx.Request.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		return headers
	}

	// Try to get from HTTP request
	if req := GetHTTPRequest(ctx); req != nil {
		for key, values := range req.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
	}

	return headers
}
