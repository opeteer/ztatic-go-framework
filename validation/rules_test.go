package validation

import (
	"context"
	"testing"
)

type rulesSample struct {
	Password string `validate:"strong_password"`
	SafeText string `validate:"no_html"`
	SQLSafe  string `validate:"no_sql_injection"`
	Path     string `validate:"safe_path"`
	Slug     string `validate:"slug"`
	Phone    string `validate:"phone"`
	Version  string `validate:"semver"`
	Cron     string `validate:"cron"`
	RawJSON  string `validate:"json"`
	Card     string `validate:"credit_card"`
}

type mockDBResolver struct {
	existingRecords map[string]bool
}

func (m *mockDBResolver) Exists(ctx context.Context, table, column string, value any) (bool, error) {
	key := table + ":" + column + ":" + value.(string)
	return m.existingRecords[key], nil
}

type dbSample struct {
	Email    string `validate:"unique=users:email"`
	TenantID string `validate:"exists=tenants:id"`
}

func TestRules_ZeroTrustAndDomain(t *testing.T) {
	eng := New()

	valid := rulesSample{
		Password: "P@ssw0rdSecure!",
		SafeText: "Just a clean normal text without markup",
		SQLSafe:  "SELECT normal column from somewhere",
		Path:     "uploads/avatar.png",
		Slug:     "my-awesome-post-2026",
		Phone:    "+14155552671",
		Version:  "v1.2.3-beta.1",
		Cron:     "0 12 * * ?",
		RawJSON:  `{"status":"active"}`,
		Card:     "4532015112830366", // valid Luhn VISA
	}

	if err := eng.Validate(&valid); err != nil {
		t.Fatalf("expected valid struct to pass, got: %v", err)
	}

	// Test failures
	invalidPassword := valid
	invalidPassword.Password = "weak"
	if err := eng.Validate(&invalidPassword); err == nil {
		t.Errorf("expected weak password to fail")
	}

	invalidHTML := valid
	invalidHTML.SafeText = "<script>alert(1)</script>"
	if err := eng.Validate(&invalidHTML); err == nil {
		t.Errorf("expected script tag to fail no_html")
	}

	invalidSQL := valid
	invalidSQL.SQLSafe = "admin' OR '1'='1' --"
	if err := eng.Validate(&invalidSQL); err == nil {
		t.Errorf("expected SQL injection to fail no_sql_injection")
	}

	invalidPath := valid
	invalidPath.Path = "../../../etc/shadow"
	if err := eng.Validate(&invalidPath); err == nil {
		t.Errorf("expected traversal path to fail safe_path")
	}

	invalidSlug := valid
	invalidSlug.Slug = "invalid slug with spaces!"
	if err := eng.Validate(&invalidSlug); err == nil {
		t.Errorf("expected invalid slug to fail")
	}

	invalidPhone := valid
	invalidPhone.Phone = "not-a-phone-000"
	if err := eng.Validate(&invalidPhone); err == nil {
		t.Errorf("expected invalid phone to fail")
	}

	invalidSemver := valid
	invalidSemver.Version = "v1..2"
	if err := eng.Validate(&invalidSemver); err == nil {
		t.Errorf("expected invalid semver to fail")
	}

	invalidJSON := valid
	invalidJSON.RawJSON = "{invalid-json"
	if err := eng.Validate(&invalidJSON); err == nil {
		t.Errorf("expected invalid JSON to fail")
	}

	invalidCard := valid
	invalidCard.Card = "4532015112830367" // invalid Luhn checksum
	if err := eng.Validate(&invalidCard); err == nil {
		t.Errorf("expected invalid credit card to fail")
	}
}

func TestRules_DatabaseUniquenessAndExistence(t *testing.T) {
	mockDB := &mockDBResolver{
		existingRecords: map[string]bool{
			"users:email:taken@example.com": true,
			"tenants:id:tenant-123":        true,
		},
	}

	eng := New()
	eng.SetDatabaseResolver(mockDB)

	// Test 1: Email is unique, Tenant exists -> OK
	valid := dbSample{
		Email:    "new@example.com",
		TenantID: "tenant-123",
	}
	if err := eng.Validate(&valid); err != nil {
		t.Fatalf("expected valid db checks to pass, got: %v", err)
	}

	// Test 2: Email is taken -> FAIL
	takenEmail := dbSample{
		Email:    "taken@example.com",
		TenantID: "tenant-123",
	}
	if err := eng.Validate(&takenEmail); err == nil {
		t.Fatalf("expected taken email to fail unique validation")
	}

	// Test 3: Tenant does not exist -> FAIL
	missingTenant := dbSample{
		Email:    "new@example.com",
		TenantID: "tenant-999",
	}
	if err := eng.Validate(&missingTenant); err == nil {
		t.Fatalf("expected missing tenant to fail exists validation")
	}
}
