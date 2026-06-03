 package validator

import (
	"fmt"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)
 
func ExtractErrors(err error) map[string]string {
	result := make(map[string]string)

	var ve validator.ValidationErrors
	if ok := asValidationErrors(err, &ve); !ok {
		result["_"] = err.Error()
		return result
	}

	for _, fe := range ve {
		field := toSnakeCase(fe.Field())
		result[field] = buildMessage(fe)
	}
	return result
}

func buildMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", fe.Field(), fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fe.Field(), fe.Param())
	case "eqfield":
		return fmt.Sprintf("%s must match %s", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

 
func asValidationErrors(err error, target *validator.ValidationErrors) bool {
	ve, ok := err.(validator.ValidationErrors)
	if ok {
		*target = ve
	}
	return ok
}

// toSnakeCase sederhana: lowercase huruf pertama saja (Gin sudah pakai json tag).
func toSnakeCase(s string) string {
	if len(s) == 0 {
		return s
	}
	b := []byte(s)
	if b[0] >= 'A' && b[0] <= 'Z' {
		b[0] += 32
	}
	return string(b)
}

func Init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// contoh custom rule — bisa ditambah sesuai kebutuhan
		_ = v.RegisterValidation("nowhitespace", func(fl validator.FieldLevel) bool {
			for _, c := range fl.Field().String() {
				if c == ' ' {
					return false
				}
			}
			return true
		})
	}
}