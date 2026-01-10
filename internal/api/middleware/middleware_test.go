package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"v/internal/logger"
)

// Property 7: Error Logging with Request Context
// For any error that occurs during request processing, the error log SHALL contain
// the request ID, HTTP method, path, and error details.
// **Validates: Requirements 9.5**

func init() {
	gin.SetMode(gin.TestMode)
}

type testLogger struct {
	output *bytes.Buffer
	level  logger.Level
}

func newTestLogger() *testLogger {
	return &testLogger{output: &bytes.Buffer{}, level: logger.DebugLevel}
}

func (l *testLogger) Debug(msg string, fields ...logger.Field) { l.log("debug", msg, fields...) }
func (l *testLogger) Info(msg string, fields ...logger.Field)  { l.log("info", msg, fields...) }
func (l *testLogger) Warn(msg string, fields ...logger.Field)  { l.log("warn", msg, fields...) }
func (l *testLogger) Error(msg string, fields ...logger.Field) { l.log("error", msg, fields...) }
func (l *testLogger) Fatal(msg string, fields ...logger.Field) { l.log("fatal", msg, fields...) }
func (l *testLogger) With(fields ...logger.Field) logger.Logger { return l }
func (l *testLogger) SetLevel(level logger.Level)               { l.level = level }
func (l *testLogger) GetLevel() logger.Level                    { return l.level }

func (l *testLogger) log(level, msg string, fields ...logger.Field) {
	entry := map[string]any{"level": level, "message": msg}
	if len(fields) > 0 {
		fieldsMap := make(map[string]any)
		for _, f := range fields {
			fieldsMap[f.Key] = f.Value
		}
		entry["fields"] = fieldsMap
	}
	data, _ := json.Marshal(entry)
	l.output.Write(data)
	l.output.WriteString("\n")
}

func (l *testLogger) getOutput() string { return l.output.String() }


func TestErrorLoggingWithRequestContext(t *testing.T) {
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("error logs contain request context fields", prop.ForAll(
		func(method, path string) bool {
			if path == "" || !strings.HasPrefix(path, "/") {
				return true
			}

			testLog := newTestLogger()
			router := gin.New()
			router.Use(RequestID())
			router.Use(Logger(testLog))

			router.Handle(method, path, func(c *gin.Context) {
				c.Error(&gin.Error{Err: http.ErrBodyNotAllowed, Type: gin.ErrorTypePrivate})
				c.JSON(http.StatusInternalServerError, gin.H{"error": "test error"})
			})

			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			output := testLog.getOutput()
			var logEntry map[string]any
			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) == 0 {
				return false
			}

			if err := json.Unmarshal([]byte(lines[len(lines)-1]), &logEntry); err != nil {
				return false
			}

			fields, ok := logEntry["fields"].(map[string]any)
			if !ok {
				return false
			}

			if _, ok := fields["request_id"]; !ok {
				return false
			}
			if m, ok := fields["method"]; !ok || m != method {
				return false
			}
			if p, ok := fields["path"]; !ok || p != path {
				return false
			}

			return true
		},
		gen.OneConstOf("GET", "POST", "PUT", "DELETE"),
		gen.OneConstOf("/api/test", "/api/users", "/api/proxies", "/health"),
	))

	properties.TestingRun(t)
}

func TestErrorLoggingWithRequestContext_ErrorsIncluded(t *testing.T) {
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("error details are included in logs", prop.ForAll(
		func(errorMsg string) bool {
			if errorMsg == "" {
				return true
			}

			testLog := newTestLogger()
			router := gin.New()
			router.Use(RequestID())
			router.Use(Logger(testLog))

			router.GET("/test", func(c *gin.Context) {
				c.Error(&gin.Error{Err: http.ErrBodyNotAllowed, Meta: errorMsg, Type: gin.ErrorTypePrivate})
				c.JSON(http.StatusInternalServerError, gin.H{"error": errorMsg})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			output := testLog.getOutput()
			var logEntry map[string]any
			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) == 0 {
				return false
			}

			if err := json.Unmarshal([]byte(lines[len(lines)-1]), &logEntry); err != nil {
				return false
			}

			fields, ok := logEntry["fields"].(map[string]any)
			if !ok {
				return false
			}

			if _, ok := fields["errors"]; !ok {
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	properties.TestingRun(t)
}

func TestErrorLoggingWithRequestContext_RequestIDPropagated(t *testing.T) {
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("request ID is propagated and logged", prop.ForAll(
		func(customRequestID string) bool {
			if customRequestID == "" {
				return true
			}

			testLog := newTestLogger()
			router := gin.New()
			router.Use(RequestID())
			router.Use(Logger(testLog))

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Request-ID", customRequestID)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Header().Get("X-Request-ID") != customRequestID {
				return false
			}

			output := testLog.getOutput()
			var logEntry map[string]any
			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) == 0 {
				return false
			}

			if err := json.Unmarshal([]byte(lines[len(lines)-1]), &logEntry); err != nil {
				return false
			}

			fields, ok := logEntry["fields"].(map[string]any)
			if !ok {
				return false
			}

			if rid, ok := fields["request_id"]; !ok || rid != customRequestID {
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
	))

	properties.TestingRun(t)
}

func TestErrorLoggingWithRequestContext_LogLevelBasedOnStatus(t *testing.T) {
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("log level matches HTTP status severity", prop.ForAll(
		func(status int) bool {
			if status < 200 || status > 599 {
				return true
			}

			testLog := newTestLogger()
			router := gin.New()
			router.Use(RequestID())
			router.Use(Logger(testLog))

			router.GET("/test", func(c *gin.Context) {
				c.Status(status)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			output := testLog.getOutput()
			var logEntry map[string]any
			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) == 0 {
				return false
			}

			if err := json.Unmarshal([]byte(lines[len(lines)-1]), &logEntry); err != nil {
				return false
			}

			level, ok := logEntry["level"].(string)
			if !ok {
				return false
			}

			if status >= 500 && level != "error" {
				return false
			}
			if status >= 400 && status < 500 && level != "warn" {
				return false
			}
			if status < 400 && level != "info" {
				return false
			}

			return true
		},
		gen.IntRange(200, 599),
	))

	properties.TestingRun(t)
}
