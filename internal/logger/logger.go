package logger

import "go.uber.org/zap"

type Field = zap.Field

var (
	String = zap.String
	Int    = zap.Int
	Int64  = zap.Int64
	Bool   = zap.Bool
	Error  = zap.Error
)

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
}

type zapLogger struct {
	l *zap.Logger
}

func New(env string) Logger {
	var (
		l   *zap.Logger
		err error
	)
	if env == "local" || env == "dev" {
		l, err = zap.NewDevelopment()
	} else {
		l, err = zap.NewProduction()
	}
	if err != nil {
		panic(err)
	}
	return &zapLogger{l: l}
}

func (z *zapLogger) Debug(msg string, fields ...Field) {
	z.l.Debug(msg, fields...)
}

func (z *zapLogger) Info(msg string, fields ...Field) {
	z.l.Info(msg, fields...)
}

func (z *zapLogger) Warn(msg string, fields ...Field) {
	z.l.Warn(msg, fields...)
}

func (z *zapLogger) Error(msg string, fields ...Field) {
	z.l.Error(msg, fields...)
}

func (z *zapLogger) Fatal(msg string, fields ...Field) {
	z.l.Fatal(msg, fields...)
}
