package retry

import (
	"errors"
	"testing"
	"time"
)

func TestWithRetrySuccess(t *testing.T) {
	callCount := 0
	operation := func() error {
		callCount++
		if callCount <= 2 {
			return errors.New("temporary error")
		}
		return nil
	}

	config := NewRetryConfig(3, time.Millisecond*10)
	err := WithRetry(operation, config)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}

func TestWithRetryFailure(t *testing.T) {
	callCount := 0
	expectedError := errors.New("persistent error")
	operation := func() error {
		callCount++
		return expectedError
	}

	config := NewRetryConfig(2, time.Millisecond*10)
	err := WithRetry(operation, config)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if callCount != 3 { // maxRetries + 1
		t.Errorf("Expected 3 calls, got %d", callCount)
	}

	if !errors.Is(err, expectedError) {
		t.Errorf("Expected wrapped error to contain original error")
	}
}

func TestWithRetryNoRetries(t *testing.T) {
	callCount := 0
	operation := func() error {
		callCount++
		return errors.New("error")
	}

	config := NewRetryConfig(0, time.Millisecond*10)
	err := WithRetry(operation, config)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestWithRetryImmediateSuccess(t *testing.T) {
	callCount := 0
	operation := func() error {
		callCount++
		return nil
	}

	config := NewRetryConfig(3, time.Millisecond*10)
	err := WithRetry(operation, config)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestNewRetryConfig(t *testing.T) {
	maxRetries := 5
	delay := time.Second * 2

	config := NewRetryConfig(maxRetries, delay)

	if config.MaxRetries != maxRetries {
		t.Errorf("Expected MaxRetries %d, got %d", maxRetries, config.MaxRetries)
	}

	if config.Delay != delay {
		t.Errorf("Expected Delay %v, got %v", delay, config.Delay)
	}
}
