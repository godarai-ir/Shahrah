package docextract

import (
	"bytes"
	"fmt"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

func hasUTF16BOM(data []byte) bool {
	return bytes.HasPrefix(data, []byte{0xFF, 0xFE}) || bytes.HasPrefix(data, []byte{0xFE, 0xFF})
}

// decodeText reads UTF-8, UTF-16 with a byte order mark, and falls back to Windows-1256, the legacy Persian code page.
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

	return string(decoded), nil
}
