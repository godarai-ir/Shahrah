package docextract

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

const docxMainPart = "word/document.xml"

// docxPartOrder puts metadata first, because classification labels usually live in document properties, then the
// body, then headers, footers, notes, and comments, where confidentiality stamps are often placed.
var docxPartOrder = []struct {
	name   string
	prefix bool
}{
	{name: "docProps/core.xml"},
	{name: "docProps/custom.xml"},
	{name: docxMainPart},
	{name: "word/header", prefix: true},
	{name: "word/footer", prefix: true},
	{name: "word/footnotes.xml"},
	{name: "word/endnotes.xml"},
	{name: "word/comments.xml"},
}

func isDOCX(data []byte) bool {
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return false
	}

	for _, file := range archive.File {
		if file.Name == docxMainPart {
			return true
		}
	}

	return false
}

func extractDOCX(data []byte, limits Limits) (string, error) {
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("%w: invalid Word archive: %w", ErrMalformed, err)
	}

	if len(archive.File) > limits.MaxZipEntries {
		return "", fmt.Errorf("%w: Word archive has %d parts, limit is %d", ErrTooLarge, len(archive.File), limits.MaxZipEntries)
	}

	budget := limits.MaxUnzippedBytes

	var out strings.Builder

	for _, file := range docxTextParts(archive.File) {
		raw, err := readZipPart(file, budget)
		if err != nil {
			return "", err
		}

		budget -= int64(len(raw))

		text, err := wordXMLText(raw)
		if err != nil {
			return "", fmt.Errorf("%w: %s: %w", ErrMalformed, file.Name, err)
		}

		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		if out.Len() > 0 {
			out.WriteString("\n\n")
		}

		out.WriteString(text)

		if out.Len() > limits.MaxTextBytes {
			return "", fmt.Errorf("%w: extracted text exceeds %d bytes", ErrTooLarge, limits.MaxTextBytes)
		}
	}

	return out.String(), nil
}

// docxTextParts returns the XML parts that carry user-visible text or document properties, in docxPartOrder.
func docxTextParts(files []*zip.File) []*zip.File {
	rank := func(name string) int {
		for i, part := range docxPartOrder {
			if name == part.name || (part.prefix && strings.HasPrefix(name, part.name) && strings.HasSuffix(name, ".xml")) {
				return i
			}
		}

		return -1
	}

	parts := make([]*zip.File, 0, len(docxPartOrder))

	for _, file := range files {
		if rank(file.Name) >= 0 {
			parts = append(parts, file)
		}
	}

	sort.SliceStable(parts, func(i, j int) bool {
		ri, rj := rank(parts[i].Name), rank(parts[j].Name)
		if ri != rj {
			return ri < rj
		}

		return parts[i].Name < parts[j].Name
	})

	return parts
}

// readZipPart reads at most budget bytes and never trusts the declared uncompressed size.
func readZipPart(file *zip.File, budget int64) (raw []byte, err error) {
	if budget <= 0 {
		return nil, fmt.Errorf("%w: Word archive expands beyond the inspection limit", ErrTooLarge)
	}

	reader, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrMalformed, file.Name, err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil && err == nil {
			raw, err = nil, fmt.Errorf("%w: %s: %w", ErrMalformed, file.Name, closeErr)
		}
	}()

	raw, err = io.ReadAll(io.LimitReader(reader, budget+1))
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrMalformed, file.Name, err)
	}

	if int64(len(raw)) > budget {
		return nil, fmt.Errorf("%w: Word archive expands beyond the inspection limit", ErrTooLarge)
	}

	return raw, nil
}

// wordXMLText collects run text from WordprocessingML parts and every value from property parts. Deleted tracked
// changes are kept, because they are still inside the file the model receives.
func wordXMLText(raw []byte) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(raw))

	var (
		out       strings.Builder
		inRunText bool
		wordPart  = bytes.Contains(raw, []byte("wordprocessingml"))
	)

	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			return out.String(), nil
		}

		if err != nil {
			return "", err
		}

		switch element := token.(type) {
		case xml.StartElement:
			switch element.Name.Local {
			case "t", "delText":
				inRunText = true
			case "tab":
				out.WriteByte('\t')
			case "br", "cr":
				out.WriteByte('\n')
			}
		case xml.EndElement:
			switch element.Name.Local {
			case "t", "delText":
				inRunText = false
			case "p":
				out.WriteByte('\n')
			default:
				if !wordPart {
					out.WriteByte('\n')
				}
			}
		case xml.CharData:
			if inRunText || (!wordPart && len(bytes.TrimSpace(element)) > 0) {
				out.Write(element)
			}
		}
	}
}
