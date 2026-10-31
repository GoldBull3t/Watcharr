package tvdb

import "testing"

func TestISO3(t *testing.T) {
	cases := map[string]string{
		"pt-BR": "por",
		"en-US": "eng",
		"de-DE": "deu",
		"xx-XX": "eng",
		"":      "eng",
	}
	for in, want := range cases {
		if got := ISO3(in); got != want {
			t.Errorf("ISO3(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLoginWithoutKey(t *testing.T) {
	if err := New("").Validate(); err == nil {
		t.Error("expected error when api key is empty")
	}
}

