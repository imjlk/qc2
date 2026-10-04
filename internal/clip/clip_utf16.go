package clip

import (
	"errors"
	"strings"
	"unicode/utf16"
)

// encodeClipboardUTF16 returns NUL-terminated UTF-16 for the Windows clipboard.
// CF_UNICODETEXT is a C string, so an embedded NUL is rejected.
func encodeClipboardUTF16(text string) ([]uint16, error) {
	if strings.Contains(text, "\x00") {
		return nil, errors.New("text contains a NUL character")
	}

	encoded := utf16.Encode([]rune(text))
	return append(encoded, 0), nil
}
