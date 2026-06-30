package config

import "testing"

func TestValidateRequiresServiceTokens(t *testing.T) {
	err := Config{}.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateAcceptsRequiredServiceTokens(t *testing.T) {
	cfg := Config{
		ServiceToken:    "gateway-secret",
		APIServiceToken: "api-secret",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
