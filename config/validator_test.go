package config

import (
	"errors"
	"strings"
	"testing"
)

type ValidatedConfig struct {
	Env       string       `validate:"config_profile"`
	Port      int          `validate:"required,min=1024,max=65535"`
	SecretKey SecretString `validate:"secure_secret"`
	APIURL    string       `validate:"https_url"`
	DSN       string       `validate:"db_dsn"`
}

type SelfValidatingConfig struct {
	Mode string
}

func (s SelfValidatingConfig) Validate() error {
	if s.Mode == "invalid" {
		return errors.New("mode cannot be invalid")
	}
	return nil
}

func TestConfigValidator_Development(t *testing.T) {
	v := NewConfigValidator(ProfileDevelopment)

	cfg := ValidatedConfig{
		Env:       "development",
		Port:      8080,
		SecretKey: NewSecretString("short-dev-secret"), // allowed in dev
		APIURL:    "https://api.example.com",
		DSN:       "postgres://user:pass@localhost:5432/db",
	}

	if err := v.Validate(cfg); err != nil {
		t.Errorf("expected valid config to pass, got: %v", err)
	}
}

func TestConfigValidator_Production_SecureSecretEnforced(t *testing.T) {
	v := NewConfigValidator(ProfileProduction)

	cfgShort := ValidatedConfig{
		Env:       "production",
		Port:      8080,
		SecretKey: NewSecretString("short-secret"), // < 32 chars, should fail in prod
		APIURL:    "https://api.example.com",
		DSN:       "postgres://user:pass@localhost:5432/db",
	}

	if err := v.Validate(cfgShort); err == nil {
		t.Error("expected short secret to fail in production, got nil error")
	}

	cfgValid := ValidatedConfig{
		Env:       "production",
		Port:      8080,
		SecretKey: NewSecretString("12345678901234567890123456789012"), // 32 chars
		APIURL:    "https://api.example.com",
		DSN:       "postgres://user:pass@localhost:5432/db",
	}

	if err := v.Validate(cfgValid); err != nil {
		t.Errorf("expected 32-byte secret to pass in production, got: %v", err)
	}
}

func TestConfigValidator_HTTPSURL(t *testing.T) {
	v := NewConfigValidator(ProfileProduction)

	cfg := ValidatedConfig{
		Env:       "production",
		Port:      8080,
		SecretKey: NewSecretString("12345678901234567890123456789012"),
		APIURL:    "http://insecure.example.com", // Not HTTPS
		DSN:       "postgres://user:pass@localhost:5432/db",
	}

	err := v.Validate(cfg)
	if err == nil {
		t.Error("expected http URL to fail https_url validation")
	}
}

func TestConfigValidator_SelfValidator(t *testing.T) {
	v := NewConfigValidator(ProfileDevelopment)

	valid := SelfValidatingConfig{Mode: "safe"}
	if err := v.Validate(valid); err != nil {
		t.Errorf("expected valid mode to pass, got: %v", err)
	}

	invalid := SelfValidatingConfig{Mode: "invalid"}
	err := v.Validate(invalid)
	if err == nil {
		t.Error("expected invalid mode to fail SelfValidator")
	} else if !strings.Contains(err.Error(), "mode cannot be invalid") {
		t.Errorf("unexpected error message: %v", err)
	}
}
