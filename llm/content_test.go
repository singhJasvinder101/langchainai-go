package llm

import "testing"

func TestJoinTextParts(t *testing.T) {
	text, err := JoinTextParts([]ContentPart{TextPart("a"), TextPart("b")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "a\nb" {
		t.Fatalf("unexpected text: %q", text)
	}
}

func TestJoinTextPartsRejectsImage(t *testing.T) {
	_, err := JoinTextParts([]ContentPart{ImageURLPart("https://example.com/x.png")})
	if err == nil {
		t.Fatal("expected error for non-text part")
	}
}
