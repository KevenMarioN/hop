package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestKeepTryingComSucessoImediato testa KeepTrying com sucesso imediato
func TestKeepTryingComSucessoImediato(t *testing.T) {
	ctx := context.Background()
	called := false
	fn := func() error {
		called = true
		return nil
	}

	err := KeepTrying(ctx, fn)
	if err != nil {
		t.Errorf("Não esperado erro, mas obteve: %v", err)
	}
	if !called {
		t.Error("Função não foi chamada")
	}
}

// TestKeepTryingComFalhaTemporária testa KeepTrying com falha temporária
func TestKeepTryingComFalhaTemporária(t *testing.T) {
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
		t.Errorf("Não esperado erro, mas obteve: %v", err)
	}
	if callCount != 3 {
		t.Errorf("Esperado 3 chamadas, mas obteve %d", callCount)
	}
}

// TestKeepTryingComContextCancelado testa KeepTrying com contexto cancelado
func TestKeepTryingComContextCancelado(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancela imediatamente

	fn := func() error {
		return errors.New("falha")
	}

	err := KeepTrying(ctx, fn)
	if err == nil {
		t.Error("Esperado erro de contexto cancelado, mas obteve nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Erro inesperado: %v", err)
	}
}

// TestKeepTryingComConfiguraçãoPersonalizada testa KeepTrying com configuração personalizada
func TestKeepTryingComConfiguraçãoPersonalizada(t *testing.T) {
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

	err := KeepTrying(ctx, fn, WithInitialDelay(cfg.InitialDelay), WithMaxDelay(cfg.MaxDelay))
	if err != nil {
		t.Errorf("Não esperado erro, mas obteve: %v", err)
	}
	if callCount != 3 {
		t.Errorf("Esperado 3 chamadas, mas obteve %d", callCount)
	}
}

// TestKeepTryingComTimeout testa KeepTrying com timeout
func TestKeepTryingComTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	fn := func() error {
		time.Sleep(100 * time.Millisecond) // Simula trabalho demorado
		return nil
	}

	err := KeepTrying(ctx, fn)
	if err == nil {
		t.Error("Esperado erro de timeout, mas obteve nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Erro inesperado: %v", err)
	}
}

// TestRetryOnceComSucesso testa RetryOnce com sucesso
func TestRetryOnceComSucesso(t *testing.T) {
	called := false
	fn := func() error {
		called = true
		return nil
	}

	err := RetryOnce(fn)
	if err != nil {
		t.Errorf("Não esperado erro, mas obteve: %v", err)
	}
	if !called {
		t.Error("Função não foi chamada")
	}
}

// TestRetryOnceComFalha testa RetryOnce com falha
func TestRetryOnceComFalha(t *testing.T) {
	expectedErr := errors.New("falha")
	fn := func() error {
		return expectedErr
	}

	err := RetryOnce(fn)
	if err == nil {
		t.Error("Esperado erro, mas obteve nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("Erro inesperado: %v", err)
	}
}

// TestWithInitialDelay testa opção WithInitialDelay
func TestWithInitialDelay(t *testing.T) {
	cfg := defaultConfig()
	initialDelay := 200 * time.Millisecond

	opt := WithInitialDelay(initialDelay)
	opt(&cfg)

	if cfg.InitialDelay != initialDelay {
		t.Errorf("InitialDelay não configurado corretamente. Esperado: %v, Obtido: %v", initialDelay, cfg.InitialDelay)
	}
}

// TestWithMaxDelay testa opção WithMaxDelay
func TestWithMaxDelay(t *testing.T) {
	cfg := defaultConfig()
	maxDelay := 60 * time.Second

	opt := WithMaxDelay(maxDelay)
	opt(&cfg)

	if cfg.MaxDelay != maxDelay {
		t.Errorf("MaxDelay não configurado corretamente. Esperado: %v, Obtido: %v", maxDelay, cfg.MaxDelay)
	}
}

// TestDefaultConfig testa configuração padrão
func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if cfg.InitialDelay != 100*time.Millisecond {
		t.Errorf("InitialDelay padrão incorreto. Esperado: 100ms, Obtido: %v", cfg.InitialDelay)
	}
	if cfg.MaxDelay != 30*time.Second {
		t.Errorf("MaxDelay padrão incorreto. Esperado: 30s, Obtido: %v", cfg.MaxDelay)
	}
	if cfg.Multiplier != 2.0 {
		t.Errorf("Multiplier padrão incorreto. Esperado: 2.0, Obtido: %v", cfg.Multiplier)
	}
}
