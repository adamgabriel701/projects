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

// ValidateStruct valida struct com tags (simplificado)
func (v *Validator) ValidateStruct(s interface{}) error {
	// Implementação básica - em produção usar go-playground/validator
	return nil
}

// Required verifica campo obrigatório
func (v *Validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.errors[field] = "campo obrigatório"
	}
}

// Email valida formato de email
func (v *Validator) Email(field, value string) {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, value)
	if !matched {
		v.errors[field] = "email inválido"
	}
}

// MinLength verifica tamanho mínimo
func (v *Validator) MinLength(field, value string, min int) {
	if len(value) < min {
		v.errors[field] = "mínimo de " + string(rune(min)) + " caracteres"
	}
}

// GreaterThan verifica se valor é maior
func (v *Validator) GreaterThan(field string, value, min float64) {
	if value <= min {
		v.errors[field] = "deve ser maior que zero"
	}
}

// HasErrors retorna true se há erros
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors retorna mapa de erros
func (v *Validator) Errors() map[string]string {
	return v.errors
}

// Error retorna erro formatado
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
