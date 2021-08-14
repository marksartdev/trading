package log

import (
	"fmt"

	"go.uber.org/zap"
)

// Action for log.
type Action string

// Logger describes methods for logging actions.
type Logger interface {
	Info(action Action, msg string)
	Warn(action Action, msg string)
	Error(action Action, err error)
}

// Logger color wrapper on zap.Logger.
type cLogger struct {
	logger  *zap.SugaredLogger
	service string
	color   Color
}

// NewLogger creates new logger.
func NewLogger(logger *zap.SugaredLogger, service string, color Color) Logger {
	return &cLogger{logger: logger, service: service, color: color}
}

// Info logs info.
func (c *cLogger) Info(action Action, msg string) {
	c.logger.Info(c.prepareMsg(action, msg))
}

// Warn logs warnings.
func (c *cLogger) Warn(action Action, msg string) {
	c.logger.Warn(c.prepareMsg(action, msg))
}

// Error logs errors.
func (c *cLogger) Error(action Action, err error) {
	c.logger.Error(c.prepareMsg(action, err.Error()))
}

func (c *cLogger) prepareMsg(action Action, msg string) string {
	title := fmt.Sprintf("%s [%s]", c.service, action)
	title = fmt.Sprintf(string(c.color), title)

	return fmt.Sprintf("%-40s %s", title, msg)
}
