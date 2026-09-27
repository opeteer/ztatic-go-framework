package validation

import (
	"context"
	"encoding/json"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var (
	slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	semverRegex = regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	xssCheckRegex = regexp.MustCompile(`(?is)<\s*script\b[^>]*>|<\s*/\s*script\s*>|javascript\s*:|\bon(load|error|click|dblclick|mouse\w+|key\w+|focus\w*|blur|change|submit|input)\s*=|<\s*(iframe|object|embed|applet)\b[^>]*>`)
	sqliCheckRegexes = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bunion\s*(all\s+|distinct\s+)?select\b`),
		regexp.MustCompile(`(?i)(;\s*|\b(union|select)\b.*?)\b(drop\s+(table|database)|truncate\s+table|alter\s+table)\b`),
		regexp.MustCompile(`(?i)\b(exec|execute)\s*(xp_\w+|sp_\w+|immediate|master\.|\(@?|@\w+)`),
		regexp.MustCompile(`(?i)('\s*(or|and)\s*'?\d+'?\s*=\s*'?\d+'?|'\s*(or|and)\s*'\w+'\s*=\s*'\w+')`),
		regexp.MustCompile(`(?i)('\s*(or|and)\s*.*?--|'\s*--|;\s*--)`),
	}

	customRulesMu  sync.RWMutex
	customRules    = make(map[string]validator.Func)
	customCtxRules = make(map[string]validator.FuncCtx)
)

// RegisterRule registers a custom validation rule tag globally.
func RegisterRule(tag string, fn validator.Func) {
	customRulesMu.Lock()
	defer customRulesMu.Unlock()
	customRules[tag] = fn
}

// RegisterRuleWithContext registers a custom context-aware validation rule tag globally.
func RegisterRuleWithContext(tag string, fn validator.FuncCtx) {
	customRulesMu.Lock()
	defer customRulesMu.Unlock()
	customCtxRules[tag] = fn
}

func registerBuiltInRules(v *validator.Validate, getResolver func() DatabaseResolver) {
	// Zero-Trust Security Rules
	_ = v.RegisterValidation("strong_password", validateStrongPassword)
	_ = v.RegisterValidation("no_html", validateNoHTML)
	_ = v.RegisterValidation("xss_safe", validateNoHTML)
	_ = v.RegisterValidation("no_sql_injection", validateNoSQLInjection)
	_ = v.RegisterValidation("safe_path", validateSafePath)

	// Domain & Format Rules
	_ = v.RegisterValidation("slug", validateSlug)
	_ = v.RegisterValidation("phone", validatePhone)
	_ = v.RegisterValidation("semver", validateSemVer)
	_ = v.RegisterValidation("cron", validateCron)
	_ = v.RegisterValidation("json", validateJSON)
	_ = v.RegisterValidation("credit_card", validateCreditCard)

	// Context-Aware Database Rules
	_ = v.RegisterValidationCtx("unique", makeValidateUnique(getResolver))
	_ = v.RegisterValidationCtx("exists", makeValidateExists(getResolver))

	// Register user-defined custom rules
	customRulesMu.RLock()
	defer customRulesMu.RUnlock()
	for tag, fn := range customRules {
		_ = v.RegisterValidation(tag, fn)
	}
	for tag, fn := range customCtxRules {
		_ = v.RegisterValidationCtx(tag, fn)
	}
}

// Rule implementations

func validateStrongPassword(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	minLen := 8
	if param := fl.Param(); param != "" {
		if p, err := strconv.Atoi(param); err == nil && p > 0 {
			minLen = p
		}
	}

	if len(val) < minLen {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range val {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}

func validateNoHTML(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}
	if xssCheckRegex.MatchString(val) {
		return false
	}
	// Check for raw html tags like <div ...> or <b>
	return !htmlTagRegex.MatchString(val)
}

func validateNoSQLInjection(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}
	for _, re := range sqliCheckRegexes {
		if re.MatchString(val) {
			return false
		}
	}
	return true
}

func validateSafePath(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}
	if strings.Contains(val, "\x00") {
		return false
	}
	// Check directory traversal
	cleaned := filepath.Clean(val)
	if strings.HasPrefix(cleaned, "..") || strings.Contains(val, "../") || strings.Contains(val, `..\`) {
		return false
	}
	return true
}

func validateSlug(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}
	return slugRegex.MatchString(val)
}

func validatePhone(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}
	return phoneRegex.MatchString(val)
}

func validateSemVer(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}
	return semverRegex.MatchString(val)
}

func validateCron(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}
	fields := strings.Fields(val)
	if len(fields) != 5 {
		return false
	}
	cronFieldPattern := regexp.MustCompile(`^[\d\*\-\,\/\?a-zA-Z]+$`)
	for _, field := range fields {
		if !cronFieldPattern.MatchString(field) {
			return false
		}
	}
	return true
}

func validateJSON(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}
	return json.Valid([]byte(val))
}

func validateCreditCard(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	val = strings.ReplaceAll(val, " ", "")
	val = strings.ReplaceAll(val, "-", "")
	if len(val) < 13 || len(val) > 19 {
		return false
	}
	// Luhn Algorithm
	var sum int
	alternate := false
	for i := len(val) - 1; i >= 0; i-- {
		r := rune(val[i])
		if !unicode.IsDigit(r) {
			return false
		}
		digit := int(r - '0')
		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alternate = !alternate
	}
	return sum%10 == 0
}

func parseTableAndCol(param string) (string, string, bool) {
	parts := strings.FieldsFunc(param, func(r rune) bool {
		return r == ':' || r == '/' || r == '.' || r == ','
	})
	if len(parts) < 2 {
		return "", "", false
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), true
}

func makeValidateUnique(getResolver func() DatabaseResolver) validator.FuncCtx {
	return func(ctx context.Context, fl validator.FieldLevel) bool {
		table, column, ok := parseTableAndCol(fl.Param())
		if !ok {
			return true // misconfigured rule tag, do not block
		}

		resolver := ResolverFromContext(ctx)
		if resolver == nil && getResolver != nil {
			resolver = getResolver()
		}
		if resolver == nil {
			// No database resolver configured; treat as passing
			return true
		}

		exists, err := resolver.Exists(ctx, table, column, fl.Field().Interface())
		if err != nil {
			return false
		}
		return !exists
	}
}

func makeValidateExists(getResolver func() DatabaseResolver) validator.FuncCtx {
	return func(ctx context.Context, fl validator.FieldLevel) bool {
		table, column, ok := parseTableAndCol(fl.Param())
		if !ok {
			return true
		}

		resolver := ResolverFromContext(ctx)
		if resolver == nil && getResolver != nil {
			resolver = getResolver()
		}
		if resolver == nil {
			return true
		}

		exists, err := resolver.Exists(ctx, table, column, fl.Field().Interface())
		if err != nil {
			return false
		}
		return exists
	}
}
