package ai

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tmc/langchaingo/schema"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func stringifyDocs(documents []schema.Document) string {
	var output string

	for i, doc := range documents {
		output += fmt.Sprintf("%d: %s", i, doc.PageContent)
	}

	return output
}

func deduplicate(documents []schema.Document) []schema.Document {

	// maybe pass in doc length

	seen := make(map[string]struct{})

	var newDocs []schema.Document

	for _, doc := range documents {
		if _, ok := seen[doc.PageContent]; !ok {
			newDocs = append(newDocs, doc)
		}
		seen[doc.PageContent] = struct{}{}
	}

	return newDocs
}

func cleanAIResponse(text string) string {

	if text == "" {
		return ""
	}

	// Remove line breaks and extra spaces
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")

	// Remove specific markup characters
	unwantedChars := []string{"*", "#", "-", "•"}
	for _, char := range unwantedChars {
		text = strings.ReplaceAll(text, char, "")
	}

	// Remove multiple spaces
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	return text
}

func titleCase(s string) string {

	caser := cases.Title(language.English)

	return caser.String(s)

}
