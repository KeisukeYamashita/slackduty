package log

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// New creates a new zap logger with the given log level.
func New(level string) (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	err := config.Level.UnmarshalText([]byte(level))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse log level")
	}
	config.DisableStacktrace = true
	config.Sampling = nil
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	return config.Build()
}

// NewDiscard creates logger which output to ioutil.Discard.
// This can be used for testing.
func NewDiscard() *zap.Logger {
	return zap.NewNop()
}
