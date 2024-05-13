package logger

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	log := New()

	assert.NotNil(t, log, "Expected logger instance, got nil")

	_, isJSONFormatter := log.Formatter.(*logrus.JSONFormatter)
	assert.True(t, isJSONFormatter, "Expected formatter to be *logrus.JSONFormatter")
}
