package controller

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LogOp prints an operation log.
// Format: [OPT] Time | IP | Role | Path | Params Extra
func LogOp(c *gin.Context, params any, extra ...any) {
	ip := c.ClientIP()
	path := c.Request.URL.Path
	now := time.Now().Format("2006/01/02 - 15:04:05")

	roleVal, exists := c.Get("role")
	role := "Unknown"
	if exists {
		if r, ok := roleVal.(string); ok {
			role = r
		}
	}

	var contentParts []string

	if params != nil {
		contentParts = append(contentParts, fmt.Sprintf("%+v", params))
	}

	for _, e := range extra {
		contentParts = append(contentParts, fmt.Sprintf("%+v", e))
	}

	content := strings.Join(contentParts, " ")

	fmt.Printf("[OPT] %s | %s | %s | %s | %s\n", now, ip, role, path, content)
}

// LogError prints an error log.
func LogError(c *gin.Context, args ...any) {
	msg := fmt.Sprint(args...)
	stack := string(debug.Stack())
	fmt.Printf("[ERR] %s\n%s\n", msg, stack)
}

// FailWithError logs the error and sends a string response.
func FailWithError(c *gin.Context, code int, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	LogError(c, msg)
	c.String(code, msg)
}
