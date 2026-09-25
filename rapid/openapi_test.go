package rapid

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

type DummyUser struct {
	ID       int      `json:"id" validate:"required"`
	Email    string   `json:"email_address" validate:"required,email"`
	Age      int      `json:"age,omitempty" validate:"min=18,max=120"`
	Password string   `json:"-"`
	Roles    []string `json:"roles"`
}

type BaseEntity struct {
	ID string `json:"id"`
}

type ExtendedUser struct {
	BaseEntity
	Name string `json:"name"`
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/users/:id", "/users/{id}"},
		{"/api/v1/posts/:postId/comments/:commentId", "/api/v1/posts/{postId}/comments/{commentId}"},
		{"/", "/"},
		{"/about", "/about"},
	}

	for _, tc := range tests {
		got := normalizePath(tc.input)
		if got != tc.want {
			t.Errorf("normalizePath(%q) = %q; want %q", tc.input, got, tc.want)
		}
	}
}

func TestBuildSchemaFromType(t *testing.T) {
	visited := make(map[reflect.Type]string)
	schema, defs := buildSchemaFromType(reflect.TypeOf(DummyUser{}), visited)

	if schema["$ref"] != "#/components/schemas/DummyUser" {
		t.Errorf("Expected root ref to be #/components/schemas/DummyUser, got %v", schema["$ref"])
	}

	dummyDef := defs["DummyUser"].(map[string]any)
	if dummyDef["type"] != "object" {
		t.Errorf("Expected type object")
	}

	props := dummyDef["properties"].(map[string]any)

	idProp := props["id"].(map[string]any)
	if idProp["type"] != "integer" {
		t.Errorf("Expected id to be integer")
	}

	emailProp := props["email_address"].(map[string]any)
	if emailProp["type"] != "string" || emailProp["format"] != "email" {
		t.Errorf("Expected email_address to be string with email format")
	}

	ageProp := props["age"].(map[string]any)
	if ageProp["minimum"] != 18 || ageProp["maximum"] != 120 {
		t.Errorf("Expected age constraints to be min 18, max 120")
	}

	if _, ok := props["Password"]; ok {
		t.Errorf("Expected Password to be ignored due to json:'-'")
	}

	req := dummyDef["required"].([]string)
	if len(req) != 2 || (req[0] != "id" && req[1] != "id") || (req[0] != "email_address" && req[1] != "email_address") {
		t.Errorf("Expected required fields to be id, email_address, got %v", req)
	}
}

func TestServeDocs(t *testing.T) {
	gen := NewOpenAPIGenerator()
	e := echo.New()
	gen.ServeDocs(e, "/docs")

	// 1. Check /docs/openapi.json
	reqJSON := httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil)
	recJSON := httptest.NewRecorder()
	e.ServeHTTP(recJSON, reqJSON)

	if recJSON.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for openapi.json, got %d", recJSON.Code)
	}

	var schema map[string]any
	if err := json.Unmarshal(recJSON.Body.Bytes(), &schema); err != nil {
		t.Fatalf("openapi.json returned invalid JSON: %v", err)
	}
	if schema["openapi"] != "3.0.3" {
		t.Errorf("expected openapi version 3.0.3, got %v", schema["openapi"])
	}

	// 2. Check /docs UI HTML
	reqUI := httptest.NewRequest(http.MethodGet, "/docs", nil)
	recUI := httptest.NewRecorder()
	e.ServeHTTP(recUI, reqUI)

	if recUI.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for /docs UI, got %d", recUI.Code)
	}
	if !strings.Contains(recUI.Body.String(), "<title>") || !strings.Contains(recUI.Body.String(), "@scalar/api-reference") {
		t.Errorf("expected /docs HTML to contain Scalar API Reference scripts")
	}
}

// Dummy Resource Implementation for RegisterResource Test
type mockResource struct{}

func (m *mockResource) FindAll(c *echo.Context) ([]DummyUser, error) {
	return []DummyUser{{ID: 1, Email: "user@example.com"}}, nil
}
func (m *mockResource) FindByID(c *echo.Context, id string) (DummyUser, error) {
	return DummyUser{ID: 1, Email: "user@example.com"}, nil
}
func (m *mockResource) Create(c *echo.Context, item *DummyUser) (DummyUser, error) {
	return *item, nil
}
func (m *mockResource) Update(c *echo.Context, id string, item *DummyUser) (DummyUser, error) {
	return *item, nil
}
func (m *mockResource) Delete(c *echo.Context, id string) error {
	return nil
}

func TestRegisterResource(t *testing.T) {
	e := echo.New()
	e.Validator = NewStructValidator()
	g := e.Group("/api")

	RegisterResource[DummyUser](g, "/users", &mockResource{})

	// Test GET /api/users
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 for FindAll resource, got %d", rec.Code)
	}

	// Test DELETE /api/users/1
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/users/1", nil)
	recDel := httptest.NewRecorder()
	e.ServeHTTP(recDel, reqDel)

	if recDel.Code != http.StatusNoContent {
		t.Errorf("expected HTTP 204 for Delete resource, got %d", recDel.Code)
	}
}

// TestBuildSchemaFromType_AnonymousStructWeakness demonstrates that embedded anonymous struct properties
// are not currently merged into the parent schema in buildSchemaFromType.
func TestBuildSchemaFromType_AnonymousStructWeakness(t *testing.T) {
	visited := make(map[reflect.Type]string)
	_, defs := buildSchemaFromType(reflect.TypeOf(ExtendedUser{}), visited)

	userDef, ok := defs["ExtendedUser"].(map[string]any)
	if !ok {
		t.Fatalf("ExtendedUser definition missing")
	}

	props, ok := userDef["properties"].(map[string]any)
	if !ok {
		t.Fatalf("ExtendedUser properties missing")
	}

	// Patched: BaseEntity field `id` should now be present in ExtendedUser properties
	if _, hasID := props["id"]; !hasID {
		t.Errorf("expected embedded struct property 'id' to be merged into ExtendedUser schema properties")
	}
}
