package provider

import (
	"io"
	"net/url"
	"testing"
)

func TestTranslationModelReportsWrappedEOF(t *testing.T) {
	for _, cause := range []error{io.EOF, io.ErrUnexpectedEOF} {
		t.Run(cause.Error(), func(t *testing.T) {
			requestErr := &url.Error{
				Op:  "Post",
				URL: "https://example.invalid/v1/chat/completions",
				Err: cause,
			}
			const want = "translation model API connection closed before a complete response (EOF)"
			if got := translationModelRequestError(requestErr).Error(); got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}
