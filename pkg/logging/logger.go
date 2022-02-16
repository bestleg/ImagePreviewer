package logging

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger() (*zap.SugaredLogger, error) {
	var logger *zap.Logger
	var sugarLogger *zap.SugaredLogger

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     "\n",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}

	config := zap.Config{
		// todo : в конфиг можно выкинуть
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
	}

	// Can construct a logger
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	// The then Sugarlogger
	sugarLogger = logger.Sugar()
	return sugarLogger, nil
}

// MiddleWareLogger middleware setup logger and logs all requests.
func MiddleWareLogger(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Infow("http request handled",
				"request", fmt.Sprintf("%s %s %s", r.Method, r.RequestURI, r.Proto),
				"request_method", r.Method,
				"request_uri", r.RequestURI,
				"request_proto", r.Proto,
				"request_duration_ms", int(time.Since(start).Seconds()*1000),
				"real_ip", r.Header.Get("X-Real-IP"),
				"proxy_add_x_forwarded_for", r.Header.Get("X-Forwarded-For"),
				"remote_addr", r.RemoteAddr,
				"http_referrer", r.Referer(),
				"http_user_agent", r.UserAgent(),
			)
		})
	}
}
