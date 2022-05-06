package logger

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	xRequestIDHeaderKey = "x-request-id"
)

// NewLogger returns a commons.Logger from a zap.Logger
func NewLogger(z *zap.Logger) *Logger {
	return &Logger{z: z}
}

// Logger is a wrapper to zap.Logger and will handle some common requirements
type Logger struct {
	z     *zap.Logger
	reqID string
}

// GorillaMiddleware returns a middleware function for a gorilla router
func (log *Logger) GorillaMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.reqID = r.Header.Get(xRequestIDHeaderKey)
			log.reqID = strings.ReplaceAll(log.reqID, "-", "")
			next.ServeHTTP(w, r)
		})
	}
}

func (log *Logger) zapLogger() *zap.Logger {
	if log.reqID == "" {
		return log.z
	}
	return log.z.With(zap.String("reqID", log.reqID))
}

func (log *Logger) Sugar() *zap.SugaredLogger {
	return log.zapLogger().Sugar()
}

func (log *Logger) Named(s string) *zap.Logger {
	return log.zapLogger().Named(s)
}

func (log *Logger) WithOptions(opts ...zap.Option) *zap.Logger {
	return log.zapLogger().WithOptions(opts...)
}

func (log *Logger) With(fields ...zap.Field) *zap.Logger {
	return log.zapLogger().With(fields...)
}

func (log *Logger) Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	return log.zapLogger().Check(lvl, msg)
}

func (log *Logger) Debug(msg string, fields ...zap.Field) {
	log.zapLogger().Debug(msg, fields...)
}

func (log *Logger) Info(msg string, fields ...zap.Field) {
	log.zapLogger().Info(msg, fields...)
}

func (log *Logger) Warn(msg string, fields ...zap.Field) {
	log.zapLogger().Warn(msg, fields...)
}

func (log *Logger) Error(msg string, fields ...zap.Field) {
	log.zapLogger().Error(msg, fields...)
}

func (log *Logger) DPanic(msg string, fields ...zap.Field) {
	log.zapLogger().DPanic(msg, fields...)
}

func (log *Logger) Panic(msg string, fields ...zap.Field) {
	log.zapLogger().Panic(msg, fields...)
}

func (log *Logger) Fatal(msg string, fields ...zap.Field) {
	log.zapLogger().Fatal(msg, fields...)
}

func (log *Logger) Sync() error {
	return log.zapLogger().Sync()
}

func (log *Logger) Core() zapcore.Core {
	return log.zapLogger().Core()
}
