package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM), // GitHub Flavored Markdown
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
	),
)

// ToHTML converte markdown para HTML
func ToHTML(source string) string {
	var buf bytes.Buffer
	if err := md.Convert([]byte(source), &buf); err != nil {
		return source // retorna original em caso de erro
	}
	return buf.String()
}

// ToPlainText extrai texto plano do markdown (simplificado)
func ToPlainText(source string) string {
	// Remover sintaxe markdown básica
	text := source
	
	// Remover headers
	for i := 6; i >= 1; i-- {
		prefix := ""
		for j := 0; j < i; j++ {
			prefix += "#"
		}
		text = replaceAll(text, prefix+" ", "")
	}
	
	// Remover negrito e itálico
	text = replaceAll(text, "**", "")
	text = replaceAll(text, "__", "")
	text = replaceAll(text, "*", "")
	text = replaceAll(text, "_", "")
	
	// Remover links [text](url) -> text
	// Simplificado - em produção usar regex
	
	return text
}

func replaceAll(s, old, new string) string {
	for {
		newS := bytes.ReplaceAll([]byte(s), []byte(old), []byte(new))
		if string(newS) == s {
			return s
		}
		s = string(newS)
	}
}
