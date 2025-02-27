// Copyright 2018 ETH Zurich
// Copyright 2020 ETH Zurich, Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Debug logs at debug level.
func Debug(msg string, ctx ...any) {
	if enabled(DebugLevel) {
		zap.L().Debug(msg, convertCtx(ctx)...)
	}
}

// Info logs at info level.
func Info(msg string, ctx ...any) {
	if enabled(InfoLevel) {
		zap.L().Info(msg, convertCtx(ctx)...)
	}
}

// Error logs at error level.
func Error(msg string, ctx ...any) {
	if enabled(ErrorLevel) {
		zap.L().Error(msg, convertCtx(ctx)...)
	}
}

func enabled(lvl Level) bool {
	return zap.L().Core().Enabled(zapcore.Level(lvl))
}

// WithOptions returns the logger with the options applied.
func WithOptions(opts ...Option) Logger {
	co := applyOptions(opts)
	return &logger{logger: zap.L().WithOptions(co.zapOptions()...)}
}

type Level zapcore.Level

const (
	DebugLevel Level = Level(zapcore.DebugLevel)
	InfoLevel  Level = Level(zapcore.InfoLevel)
	ErrorLevel Level = Level(zapcore.ErrorLevel)
)

// Logger describes the logger interface.
type Logger interface {
	New(ctx ...any) Logger
	Debug(msg string, ctx ...any)
	Info(msg string, ctx ...any)
	Error(msg string, ctx ...any)
	Enabled(lvl Level) bool
}

type logger struct {
	logger *zap.Logger
}

// New creates a logger with the given context.
func New(ctx ...any) Logger {
	return &logger{logger: zap.L().With(convertCtx(ctx)...)}
}

func (l *logger) New(ctx ...any) Logger {
	return &logger{logger: l.logger.With(convertCtx(ctx)...)}
}

func (l *logger) Debug(msg string, ctx ...any) {
	if l.Enabled(DebugLevel) {
		l.logger.Debug(msg, convertCtx(ctx)...)
	}
}

func (l *logger) Info(msg string, ctx ...any) {
	if l.Enabled(InfoLevel) {
		l.logger.Info(msg, convertCtx(ctx)...)
	}
}

func (l *logger) Error(msg string, ctx ...any) {
	if l.Enabled(ErrorLevel) {
		l.logger.Error(msg, convertCtx(ctx)...)
	}
}

func (l *logger) Enabled(lvl Level) bool {
	return l.logger.Core().Enabled(zapcore.Level(lvl))
}

// Root returns the root logger. It's a logger without any context.
func Root() Logger {
	return &logger{logger: zap.L()}
}

// Discard sets the logger up to discard all log entries. This is useful for
// testing.
func Discard() {
	Root().(*logger).logger = zap.NewNop()
}

// DiscardLogger implements the Logger interface and discards all messages.
// Subloggers created from this logger will also discard all messages and
// ignore the additional context.
//
// To see how to use this, see the example.
type DiscardLogger struct{}

func (d DiscardLogger) New(ctx ...any) Logger      { return d }
func (DiscardLogger) Debug(msg string, ctx ...any) {}
func (DiscardLogger) Info(msg string, ctx ...any)  {}
func (DiscardLogger) Error(msg string, ctx ...any) {}
func (DiscardLogger) Enabled(lvl Level) bool       { return false }

func convertCtx(ctx []any) []zap.Field {
	fields := make([]zap.Field, 0, len(ctx)/2)
	for i := 0; i+1 < len(ctx); i += 2 {
		fields = append(fields, zap.Any(ctx[i].(string), ctx[i+1]))
	}
	return fields
}
