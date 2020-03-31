package log

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNew(t *testing.T) {
	if l, err := New("INFO"); err != nil {
		t.Fatal(err)
	} else if l.Check(zapcore.InfoLevel, "should log") == nil {
		t.Fatal("info not logged")
	} else if l.Check(zapcore.DebugLevel, "should not log") != nil {
		t.Fatal("debug logged")
	}

	if l, err := New("error"); err != nil {
		t.Fatal(err)
	} else if l.Check(zapcore.ErrorLevel, "should log") == nil {
		t.Fatal("error not logged")
	} else if l.Check(zapcore.WarnLevel, "should not log") != nil {
		t.Fatal("warning logged")
	}

	if _, err := New("INFOOO"); err == nil {
		t.Fatal("expect to be failed")
	}
}

func TestNewDiscard(t *testing.T) {
	logger := NewDiscard()
	logger.Info("test discard", zap.String("output", "discard"))
}
