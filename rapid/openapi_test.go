package rapid

import (
	"reflect"
	"testing"
)

type DummyUser struct {
	ID        int      `json:"id" validate:"required"`
	Email     string   `json:"email_address" validate:"required,email"`
	Age       int      `json:"age,omitempty" validate:"min=18,max=120"`
	Password  string   `json:"-"`
	Roles     []string `json:"roles"`
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
