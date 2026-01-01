package generator

import (
	"errors"
	"testing"
	"time"
)

func TestRunStep_Success(t *testing.T) {
	duration, err := runStep("Test step", func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if duration < 10*time.Millisecond {
		t.Errorf("expected duration >= 10ms, got %v", duration)
	}
}

func TestRunStep_Error(t *testing.T) {
	expectedErr := errors.New("test error")

	duration, err := runStep("Test step with error", func() error {
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if duration == 0 {
		t.Error("expected non-zero duration")
	}
}

func TestRunStep_TracksDuration(t *testing.T) {
	sleepTime := 50 * time.Millisecond

	duration, err := runStep("Duration test", func() error {
		time.Sleep(sleepTime)
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Allow some tolerance for timing
	if duration < sleepTime {
		t.Errorf("expected duration >= %v, got %v", sleepTime, duration)
	}

	if duration > sleepTime+100*time.Millisecond {
		t.Errorf("expected duration < %v, got %v", sleepTime+100*time.Millisecond, duration)
	}
}
