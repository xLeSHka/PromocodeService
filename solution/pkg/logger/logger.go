package logger

import (
	"context"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	Fatal(ctx context.Context, msg string, fields ...zap.Field)
}
type logger struct {
	logger *zap.Logger
}

const (
	LoggerKey = "loggerKey"
	RequestID = "requestID"
)

func (l logger) Info(ctx context.Context, msg string, fields ...zap.Field) {

	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}
	l.logger.Info(msg, fields...)

}

func (l logger) Error(ctx context.Context, msg string, fields ...zap.Field) {

	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}
	l.logger.Error(msg, fields...)
}
func (l logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {

	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}
	l.logger.Warn(msg, fields...)
}
func (l logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {

	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}
	l.logger.Fatal(msg, fields...)
}
func New() Logger {
	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",
			LevelKey:   "level",
			TimeKey:    "ts",
			EncodeTime: zapcore.ISO8601TimeEncoder,
		},
	}
	zapLogger, err := cfg.Build()
	if err != nil {
		log.Fatal("failed build logger", err)
	}
	defer zapLogger.Sync()
	return &logger{logger: zapLogger}
}
func GetLoggerFromCtx(ctx context.Context) Logger {
	return ctx.Value(LoggerKey).(Logger)
}
