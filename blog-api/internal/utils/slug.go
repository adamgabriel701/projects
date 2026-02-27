package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// GenerateSlug cria slug URL-friendly a partir de texto
func GenerateSlug(text string) string {
	// Converter para minúsculas
	text = strings.ToLower(text)

	// Substituir espaços e underscores por hífen
	text = strings.ReplaceAll(text, " ", "-")
	text = strings.ReplaceAll(text, "_", "-")

	// Remover acentos (simplificado)
	text = removeAccents(text)

	// Manter apenas caracteres alfanuméricos e hífens
	var result strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' {
			result.WriteRune(r)
		}
	}

	// Remover hífens duplicos e extremos
	slug := result.String()
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	return slug
}

// ExtractKeywords extrai palavras-chave para busca
func ExtractKeywords(text string) []string {
	// Normalizar
	text = strings.ToLower(text)
	text = removeAccents(text)

	// Dividir em palavras
	words := strings.Fields(text)

	// Filtrar palavras muito curtas e comuns (stop words)
	stopWords := map[string]bool{
		"de": true, "a": true, "o": true, "que": true, "e": true,
		"do": true, "da": true, "em": true, "um": true, "para": true,
		"the": true, "an": true, "and": true, "or": true, "in": true,
	}

	var keywords []string
	seen := make(map[string]bool)

	for _, word := range words {
		word = strings.TrimFunc(word, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})

		if len(word) > 2 && !stopWords[word] && !seen[word] {
			keywords = append(keywords, word)
			seen[word] = true
		}
	}

	return keywords
}

func removeAccents(s string) string {
	// Tabela simples de substituição
	replacements := map[rune]rune{
		'á': 'a', 'à': 'a', 'ã': 'a', 'â': 'a', 'ä': 'a',
		'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
		'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i',
		'ó': 'o', 'ò': 'o', 'õ': 'o', 'ô': 'o', 'ö': 'o',
		'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u',
		'ç': 'c', 'ñ': 'n',
	}

	var result strings.Builder
	for _, r := range s {
		if replacement, ok := replacements[r]; ok {
			result.WriteRune(replacement)
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}
