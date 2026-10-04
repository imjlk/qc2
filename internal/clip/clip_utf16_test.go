package clip

import (
	"slices"
	"strings"
	"testing"
)

func TestEncodeClipboardUTF16(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		text    string
		want    []uint16
		wantErr string
	}{
		{name: "empty", text: "", want: []uint16{0}},
		{name: "ascii", text: "qc2", want: []uint16{'q', 'c', '2', 0}},
		{name: "space", text: "a b", want: []uint16{'a', ' ', 'b', 0}},
		{name: "korean", text: "한글", want: []uint16{0xD55C, 0xAE00, 0}},
		{name: "emoji", text: "🙂", want: []uint16{0xD83D, 0xDE42, 0}},
		{name: "nul", text: "a\x00b", wantErr: "NUL"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := encodeClipboardUTF16(test.text)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("encodeClipboardUTF16() error = %v, want %q", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("encodeClipboardUTF16() error = %v", err)
			}
			if !slices.Equal(got, test.want) {
				t.Fatalf("encodeClipboardUTF16() = %#x, want %#x", got, test.want)
			}
		})
	}
}
