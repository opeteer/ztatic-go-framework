package rapid

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"ztatic-go-framework/errors"
)

// StructValidator implements the echo.Validator interface, wrapping
// go-playground/validator/v10 for robust struct tag validation.
type StructValidator struct {
	validator *validator.Validate
}

// NewStructValidator initializes a new validator.
func NewStructValidator() *StructValidator {
	return &StructValidator{validator: validator.New()}
}

// Validate executes the validation tags on the given struct.
func (cv *StructValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		// Map struct validation errors into a standardized *errors.Error with field violations.
		return errors.Map(err)
	}
	return nil
}

// BindAndValidate is a DX (Developer Experience) helper that performs both 
// request payload binding (JSON/XML/Form/Query) and struct tag validation 
// in a single atomic call.
func BindAndValidate(c *echo.Context, i any) error {
	// 1. Bind payload to struct
	if err := c.Bind(i); err != nil {
		return errors.BadRequest(fmt.Sprintf("Binding error: %v", err)).WithInternal(err)
	}
	
	// 2. Validate struct tags using the Engine's registered Validator
	if err := c.Validate(i); err != nil {
		return err
	}
	
	return nil
}
