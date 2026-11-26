package provider

import (
	"testing"

	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/request"
)

func TestLegacyAIProvidersRequireConfiguration(t *testing.T) {
	website := &Website{System: &config.SystemConfig{Language: "en"}, AiGenerateConfig: &config.AiGenerateConfig{}}
	t.Run("translation", func(t *testing.T) {
		_, err := website.AnqiTranslateString(&AnqiTranslateTextRequest{
			Text: []string{"Hello world"}, ToLanguage: "zh",
		})
		if err == nil || err.Error() != website.Tr("TranslationEngineRequired") {
			t.Fatalf("expected a local configuration error, got %v", err)
		}
	})
	t.Run("streaming writing", func(t *testing.T) {
		_, err := website.AnqiAiGenerateStream(&request.KeywordRequest{Title: "Hello world"})
		if err == nil || err.Error() != website.Tr("NoAiGenerationSourceSelected") {
			t.Fatalf("expected a local configuration error, got %v", err)
		}
	})
}
