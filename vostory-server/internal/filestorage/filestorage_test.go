package filestorage

import "testing"

func TestNormalizeKey(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"audio/12/345_v2.wav", "audio/12/345_v2.wav"},
		{"storage/audio/12/345_v2.wav", "audio/12/345_v2.wav"},
		{`storage\audio\12\345_v2.wav`, "audio/12/345_v2.wav"},
		{"./storage/reference-audio/a.wav", "reference-audio/a.wav"},
		{"storage/storage/audio/1.wav", "audio/1.wav"},
	}
	for _, c := range cases {
		if got := NormalizeKey(c.in); got != c.want {
			t.Fatalf("NormalizeKey(%q)=%q, want %q", c.in, got, c.want)
		}
	}
}
