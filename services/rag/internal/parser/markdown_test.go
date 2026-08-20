package parser

import "testing"

func TestParseOCRMarkdownPreservesPersianPagesAndTables(t *testing.T) {
	t.Parallel()

	document, err := ParseOCRMarkdown([]OCRPage{
		{Index: 1, Markdown: "## ادامه\nمتن صفحه دوم"},
		{Index: 0, Markdown: "# فاکتور\nمبلغ **۱۲۰٬۰۰۰ ریال** است.\n\n| شرح | مبلغ |\n| --- | ---: |\n| کالا | ۱۲۰٬۰۰۰ |"},
	})
	if err != nil {
		t.Fatalf("parse OCR Markdown: %v", err)
	}

	if document.PageCount == nil || *document.PageCount != 2 {
		t.Fatalf("unexpected page count: %#v", document.PageCount)
	}

	if document.Language != "fa" {
		t.Fatalf("unexpected language: %q", document.Language)
	}

	if len(document.Blocks) != 6 {
		t.Fatalf("unexpected block count: %d", len(document.Blocks))
	}

	if document.Blocks[1].Text != "مبلغ ۱۲۰٬۰۰۰ ریال است." {
		t.Fatalf("Persian OCR text was not cleaned safely: %q", document.Blocks[1].Text)
	}

	if document.Blocks[2].Text != "شرح | مبلغ" || document.Blocks[3].Text != "کالا | ۱۲۰٬۰۰۰" {
		t.Fatalf("OCR table structure was not preserved: %#v", document.Blocks[2:4])
	}

	if document.Blocks[4].Page == nil || *document.Blocks[4].Page != 2 {
		t.Fatalf("OCR page mapping was not preserved: %#v", document.Blocks[4])
	}
}
