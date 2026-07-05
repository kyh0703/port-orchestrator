package config

import "testing"

func TestValidateDoesNotRequireInternalServiceTokens(t *testing.T) {
	if err := (Config{}).Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
