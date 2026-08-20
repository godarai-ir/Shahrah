package parser

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTikaExtractPreservesXHTMLStructure(t *testing.T) {
	t.Parallel()

	xhtml := `<?xml version="1.1" encoding="UTF-8"?><html xmlns="http://www.w3.org/1999/xhtml" lang="fa-IR"><body>
<div class="page"><h1>شرایط بازگشت</h1><p>مهلت بازگشت ۷ روز است.</p>
<ul><li>کالا سالم باشد</li></ul><table><tr><th>استان</th><th>هزینه</th></tr><tr><td>تهران</td><td>۱۰۰٬۰۰۰ ریال</td></tr></table></div>
<div class="page"><h2>پشتیبانی</h2><p>با ما تماس بگیرید.</p></div></body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Accept") != "text/xml" {
			t.Errorf("unexpected Accept header: %q", request.Header.Get("Accept"))
		}

		if request.Header.Get("Content-Type") != "application/pdf" {
			t.Errorf("unexpected Content-Type header: %q", request.Header.Get("Content-Type"))
		}

		writer.Header().Set("Content-Type", "application/xhtml+xml")
		_, _ = writer.Write([]byte(xhtml))
	}))
	defer server.Close()

	client, err := NewTikaClient(server.URL, time.Second)
	if err != nil {
		t.Fatalf("create Tika client: %v", err)
	}

	document, err := client.Extract(context.Background(), "application/pdf", strings.NewReader("source"))
	if err != nil {
		t.Fatalf("extract document: %v", err)
	}

	if document.Language != "fa-IR" {
		t.Fatalf("unexpected document language: %q", document.Language)
	}

	if document.PageCount == nil || *document.PageCount != 2 {
		t.Fatalf("unexpected page count: %#v", document.PageCount)
	}

	if len(document.Blocks) != 7 {
		t.Fatalf("unexpected block count: %d", len(document.Blocks))
	}

	if document.Blocks[1].Heading != "شرایط بازگشت" || document.Blocks[1].Page == nil || *document.Blocks[1].Page != 1 {
		t.Fatalf("paragraph structure was not preserved: %#v", document.Blocks[1])
	}

	if document.Blocks[3].Text != "استان | هزینه" {
		t.Fatalf("table row was not preserved: %q", document.Blocks[3].Text)
	}

	if document.Blocks[4].Text != "تهران | ۱۰۰٬۰۰۰ ریال" {
		t.Fatalf("Persian table values changed: %q", document.Blocks[4].Text)
	}
}

func TestTikaExtractRequiresOCRForEmptyBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`<html><body></body></html>`))
	}))
	defer server.Close()

	client, err := NewTikaClient(server.URL, time.Second)
	if err != nil {
		t.Fatalf("create Tika client: %v", err)
	}

	_, err = client.Extract(context.Background(), "application/pdf", strings.NewReader("source"))
	if !errors.Is(err, ErrOCRRequired) {
		t.Fatalf("expected OCR-required error, got %v", err)
	}
}

func TestValidateExtractedDocumentRejectsMojibake(t *testing.T) {
	t.Parallel()

	document, err := parseXHTML(
		[]byte(`<html><body><p>Ã Ø Ù متن خراب Ã Ø Ù</p></body></html>`),
		"text/html",
	)
	if err != nil {
		t.Fatalf("parse XHTML: %v", err)
	}

	err = validateExtractedDocument(document)
	if !errors.Is(err, ErrLowExtractionQuality) {
		t.Fatalf("expected low-quality error, got %v", err)
	}
}

func TestParseXHTMLFallsBackToBodyText(t *testing.T) {
	t.Parallel()

	document, err := parseXHTML(
		[]byte(`<html lang="fa"><body>متن فارسی بدون تگ پاراگراف</body></html>`),
		"text/plain",
	)
	if err != nil {
		t.Fatalf("parse XHTML body fallback: %v", err)
	}

	if len(document.Blocks) != 1 || document.Blocks[0].Text != "متن فارسی بدون تگ پاراگراف" {
		t.Fatalf("unexpected fallback blocks: %#v", document.Blocks)
	}
}

func TestParseXHTMLSanitizesTikaXML11Controls(t *testing.T) {
	t.Parallel()

	document, err := parseXHTML(
		[]byte("<?xml version=\"1.1\" encoding=\"UTF-8\"?><html><body><p>بخش اول\x00بخش دوم&#0;بخش سوم</p></body></html>"),
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	)
	if err != nil {
		t.Fatalf("parse XML 1.1 response: %v", err)
	}

	if len(document.Blocks) != 1 || document.Blocks[0].Text != "بخش اول بخش دوم بخش سوم" {
		t.Fatalf("unexpected sanitized blocks: %#v", document.Blocks)
	}
}
