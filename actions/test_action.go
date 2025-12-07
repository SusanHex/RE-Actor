package actions

import (
	"log/slog"
	"time"
)

type TestAction struct {
	Delay uint
}

func (ta TestAction) Act(message string) error {
	time.Sleep(time.Duration(ta.Delay) * time.Millisecond)
	slog.Info("Test Action:", "message", message)
	return nil
}
