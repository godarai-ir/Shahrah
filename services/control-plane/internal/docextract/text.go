package docextract

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

func hasUTF16BOM(data []byte) bool {
	return bytes.HasPrefix(data, []byte{0xFF, 0xFE}) || bytes.HasPrefix(data, []byte{0xFE, 0xFF})
}

// decodeText reads UTF-8 and UTF-16 with a byte order mark. Other text is read both as UTF-8 with invalid bytes dropped
// and as Windows-1256, the legacy Persian code page, because a few stray invalid bytes must not turn UTF-8 Persian into
// unmatchable mojibake, and a Windows-1256 file must still be readable.
func decodeText(data []byte) (string, error) {
	if hasUTF16BOM(data) {
		decoded, err := unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder().Bytes(data)
		if err != nil {
			return "", fmt.Errorf("%w: invalid UTF-16 text: %w", ErrMalformed, err)
		}

		return string(decoded), nil
	}

	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	if utf8.Valid(data) {
		return string(data), nil
	}

	decoded, err := charmap.Windows1256.NewDecoder().Bytes(data)
	if err != nil {
		return "", fmt.Errorf("%w: undecodable text: %w", ErrMalformed, err)
	}

	return strings.ToValidUTF8(string(data), "") + "\n\n" + string(decoded), nil
}
