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
		ServiceToken:        "gateway-secret",
		APIServiceToken:     "api-secret",
		RecordBaseURL:       "http://port-record:8080",
		RecordInternalToken: "record-secret",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestLoadReadsRecordConfig(t *testing.T) {
	t.Setenv("PORT_RECORD_BASE_URL", "http://port-record:8080/")
	t.Setenv("PORT_RECORD_START_PATH", "/internal/recordings/start")
	t.Setenv("PORT_RECORD_INTERNAL_API_TOKEN", "record-secret")

	cfg := Load()

	if cfg.RecordBaseURL != "http://port-record:8080/" {
		t.Fatalf("RecordBaseURL = %q", cfg.RecordBaseURL)
	}
	if cfg.RecordStartPath != "/internal/recordings/start" {
		t.Fatalf("RecordStartPath = %q", cfg.RecordStartPath)
	}
	if cfg.RecordInternalToken != "record-secret" {
		t.Fatalf("RecordInternalToken = %q", cfg.RecordInternalToken)
	}
}
