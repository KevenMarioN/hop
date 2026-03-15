package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestKeepTryingWithImmediateSuccess tests KeepTrying with immediate success
func TestKeepTryingWithImmediateSuccess(t *testing.T) {
	ctx := context.Background()
	called := false
	fn := func() error {
		called = true
		return nil
	}

	err := KeepTrying(ctx, fn)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !called {
		t.Error("Function was not called")
	}
}

// TestKeepTryingWithTemporaryFailure tests KeepTrying with temporary failure
func TestKeepTryingWithTemporaryFailure(t *testing.T) {
	ctx := context.Background()
	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("falha temporária")
		}

		return nil
	}

	err := KeepTrying(ctx, fn)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls, but got %d", callCount)
	}
}

// TestKeepTryingWithCancelledContext tests KeepTrying with cancelled context
func TestKeepTryingWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancela imediatamente

	fn := func() error {
		return errors.New("falha")
	}

	err := KeepTrying(ctx, fn)
	if err == nil {
		t.Error("Expected cancelled context error, but got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestKeepTryingWithCustomConfig tests KeepTrying with custom configuration
func TestKeepTryingWithCustomConfig(t *testing.T) {
	ctx := context.Background()
	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("falha")
		}

		return nil
	}

	cfg := BackoffConfig{
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	err := KeepTrying(ctx, fn,
		WithInitialDelay(cfg.InitialDelay), WithMaxDelay(cfg.MaxDelay), WithMultiplier(cfg.Multiplier))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls, but got %d", callCount)
	}
}

// TestKeepTryingWithTimeout tests KeepTrying with timeout
func TestKeepTryingWithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	fn := func() error {
		time.Sleep(100 * time.Millisecond) // Simula trabalho demorado
		return nil
	}

	err := KeepTrying(ctx, fn)
	if err == nil {
		t.Error("Expected timeout error, but got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestRetryOnceWithSuccess tests RetryOnce with success
func TestRetryOnceWithSuccess(t *testing.T) {
	called := false
	fn := func() error {
		called = true
		return nil
	}

	err := RetryOnce(fn)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !called {
		t.Error("Function was not called")
	}
}

// TestRetryOnceWithFailure tests RetryOnce with failure
func TestRetryOnceWithFailure(t *testing.T) {
	expectedErr := errors.New("falha")
	fn := func() error {
		return expectedErr
	}

	err := RetryOnce(fn)
	if err == nil {
		t.Error("Expected error, but got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestWithInitialDelay testa opção WithInitialDelay
func TestWithInitialDelay(t *testing.T) {
	cfg := defaultConfig()
	initialDelay := 200 * time.Millisecond

	opt := WithInitialDelay(initialDelay)
	opt(&cfg)

	if cfg.InitialDelay != initialDelay {
		t.Errorf("InitialDelay not configured correctly. Expected: %v, Got: %v", initialDelay, cfg.InitialDelay)
	}
}

// TestWithMaxDelay testa opção WithMaxDelay
func TestWithMaxDelay(t *testing.T) {
	cfg := defaultConfig()
	maxDelay := 60 * time.Second

	opt := WithMaxDelay(maxDelay)
	opt(&cfg)

	if cfg.MaxDelay != maxDelay {
		t.Errorf("MaxDelay not configured correctly. Expected: %v, Got: %v", maxDelay, cfg.MaxDelay)
	}
}

// TestDefaultConfig testa configuração padrão
func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if cfg.InitialDelay != 100*time.Millisecond {
		t.Errorf("InitialDelay default incorrect. Expected: 100ms, Got: %v", cfg.InitialDelay)
	}

	if cfg.MaxDelay != 30*time.Second {
		t.Errorf("MaxDelay default incorrect. Expected: 30s, Got: %v", cfg.MaxDelay)
	}

	if cfg.Multiplier != 2.0 {
		t.Errorf("Multiplier default incorrect. Expected: 2.0, Got: %v", cfg.Multiplier)
	}
}
