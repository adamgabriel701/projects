package validator

import (
	"errors"
	"regexp"
	"strings"
)

type Validator struct {
	errors map[string]string
}

func New() *Validator {
	return &Validator{
		errors: make(map[string]string),
	}
}

func (v *Validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.errors[field] = "campo obrigatório"
	}
}

func (v *Validator) Email(field, value string) {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, value)
	if !matched {
		v.errors[field] = "email inválido"
	}
}

func (v *Validator) MinLength(field, value string, min int) {
	if len(value) < min {
		v.errors[field] = "mínimo de " + string(rune('0'+min)) + " caracteres"
	}
}

func (v *Validator) MaxLength(field, value string, max int) {
	if len(value) > max {
		v.errors[field] = "máximo de " + string(rune('0'+max)) + " caracteres"
	}
}

func (v *Validator) GreaterThan(field string, value, min float64) {
	if value <= min {
		v.errors[field] = "deve ser maior que " + string(rune('0'+int(min)))
	}
}

func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

func (v *Validator) Errors() map[string]string {
	return v.errors
}

func (v *Validator) Error() error {
	if !v.HasErrors() {
		return nil
	}
	
	var msgs []string
	for field, msg := range v.errors {
		msgs = append(msgs, field+": "+msg)
	}
	return errors.New(strings.Join(msgs, "; "))
}
