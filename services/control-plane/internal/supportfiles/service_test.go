package supportfiles

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type memoryStore struct {
	putErrorAt int
	puts       []string
	deletes    []string
	objects    []ObjectInfo
}

func (s *memoryStore) Put(_ context.Context, key string, _ io.Reader, _ int64, _ string) error {
	if s.putErrorAt > 0 && len(s.puts)+1 == s.putErrorAt {
		return errors.New("put failed")
	}

	s.puts = append(s.puts, key)

	return nil
}
func (*memoryStore) Get(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("content")), nil
}
func (s *memoryStore) Delete(_ context.Context, key string) error {
	s.deletes = append(s.deletes, key)
	return nil
}
func (s *memoryStore) List(context.Context, string) ([]ObjectInfo, error) {
	return s.objects, nil
}

type scannerResult struct {
	clean bool
	err   error
}

func (s scannerResult) Scan(context.Context, io.Reader) (bool, string, error) {
	if s.err != nil {
		return false, "", s.err
	}

	if !s.clean {
		return false, "Eicar-Test-Signature FOUND", nil
	}

	return true, "clean", nil
}

func TestUploadFilesAcceptsValidTextAndStoresGeneratedKey(t *testing.T) {
	store := &memoryStore{}
	service := New(store, scannerResult{clean: true})
	files := multipartFiles(t, map[string][]byte{"incident.log": []byte("request failed\n")})

	uploads, err := service.UploadFiles(context.Background(), "tenant-id", "ticket-id", files)
	if err != nil {
		t.Fatalf("UploadFiles returned error: %v", err)
	}

	if len(uploads) != 1 || uploads[0].ScanStatus != "clean" {
		t.Fatalf("unexpected uploads: %#v", uploads)
	}

	if strings.Contains(uploads[0].ObjectKey, "incident.log") || !strings.HasPrefix(uploads[0].ObjectKey, "support/tenant-id/ticket-id/") {
		t.Fatalf("unsafe or unexpected object key: %q", uploads[0].ObjectKey)
	}
}

func TestUploadFilesRejectsDisguisedExecutable(t *testing.T) {
	service := New(&memoryStore{}, scannerResult{clean: true})
	files := multipartFiles(t, map[string][]byte{"payload.png": []byte("MZ\x90\x00executable")})

	_, err := service.UploadFiles(context.Background(), "tenant", "ticket", files)
	if !errors.Is(err, ErrTypeNotAllowed) {
		t.Fatalf("got %v want ErrTypeNotAllowed", err)
	}
}

func TestUploadFilesRejectsMalformedJSON(t *testing.T) {
	service := New(&memoryStore{}, scannerResult{clean: true})
	files := multipartFiles(t, map[string][]byte{"payload.json": []byte(`{"unterminated":`)})

	_, err := service.UploadFiles(context.Background(), "tenant", "ticket", files)
	if !errors.Is(err, ErrTypeNotAllowed) {
		t.Fatalf("got %v want ErrTypeNotAllowed", err)
	}
}

func TestUploadFilesFailsClosedForInfectedOrUnavailableScanner(t *testing.T) {
	files := multipartFiles(t, map[string][]byte{"note.txt": []byte("hello")})
	for name, testCase := range map[string]struct {
		scanner  Scanner
		expected error
	}{
		"infected": {scanner: scannerResult{clean: false}, expected: ErrInfected},
		"failure":  {scanner: scannerResult{err: errors.New("scanner unavailable")}, expected: ErrScanFailed},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := New(&memoryStore{}, testCase.scanner).UploadFiles(context.Background(), "tenant", "ticket", files)
			if !errors.Is(err, testCase.expected) {
				t.Fatalf("got %v want %v", err, testCase.expected)
			}
		})
	}
}

func TestUploadFilesDeletesEarlierObjectsWhenLaterUploadFails(t *testing.T) {
	store := &memoryStore{putErrorAt: 2}
	service := New(store, scannerResult{clean: true})
	files := multipartFiles(t, map[string][]byte{"one.txt": []byte("one"), "two.txt": []byte("two")})

	if _, err := service.UploadFiles(context.Background(), "tenant", "ticket", files); err == nil {
		t.Fatal("expected storage failure")
	}

	if len(store.puts) != 1 || len(store.deletes) != 1 || store.puts[0] != store.deletes[0] {
		t.Fatalf("compensation did not delete the first object: puts=%v deletes=%v", store.puts, store.deletes)
	}
}

func TestUploadReceiptImageStoresPrivateBillingObject(t *testing.T) {
	store := &memoryStore{}
	service := New(store, scannerResult{clean: true})
	files := multipartFiles(t, map[string][]byte{
		"receipt.png": {0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00},
	})

	upload, err := service.UploadReceiptImage(context.Background(), "tenant-id", files[0])
	if err != nil {
		t.Fatalf("UploadReceiptImage returned error: %v", err)
	}

	if upload.ContentType != "image/png" || upload.SizeBytes != files[0].Size {
		t.Fatalf("unexpected receipt upload: %#v", upload)
	}

	if !strings.HasPrefix(upload.ObjectKey, "billing/receipts/tenant-id/") || len(store.puts) != 1 || store.puts[0] != upload.ObjectKey {
		t.Fatalf("receipt was not stored under the private billing prefix: upload=%#v puts=%v", upload, store.puts)
	}

	if contentType, ok := ValidateBankTransferReceiptFileName(upload.FileName); !ok || contentType != "image/png" {
		t.Fatalf("generated receipt filename is invalid: %q", upload.FileName)
	}
}

func TestUploadReceiptImageRejectsNonImageContent(t *testing.T) {
	service := New(&memoryStore{}, scannerResult{clean: true})
	files := multipartFiles(t, map[string][]byte{"receipt.png": []byte("not an image")})

	_, err := service.UploadReceiptImage(context.Background(), "tenant-id", files[0])
	if !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("got %v want ErrInvalidReceipt", err)
	}
}

func TestUploadReceiptImageFailsClosedForInfectedOrUnavailableScanner(t *testing.T) {
	files := multipartFiles(t, map[string][]byte{
		"receipt.png": {0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00},
	})

	for name, testCase := range map[string]struct {
		scanner  Scanner
		expected error
	}{
		"infected": {scanner: scannerResult{clean: false}, expected: ErrInfected},
		"failure":  {scanner: scannerResult{err: errors.New("scanner unavailable")}, expected: ErrScanFailed},
	} {
		t.Run(name, func(t *testing.T) {
			store := &memoryStore{}
			_, err := New(store, testCase.scanner).UploadReceiptImage(context.Background(), "tenant-id", files[0])
			if !errors.Is(err, testCase.expected) {
				t.Fatalf("got %v want %v", err, testCase.expected)
			}

			if len(store.puts) != 1 || len(store.deletes) != 1 || store.puts[0] != store.deletes[0] {
				t.Fatalf("failed receipt scan was not cleaned up: puts=%v deletes=%v", store.puts, store.deletes)
			}
		})
	}
}

func TestCleanupOrphansKeepsReferencedAndRecentObjects(t *testing.T) {
	now := time.Now().UTC()
	store := &memoryStore{objects: []ObjectInfo{
		{Key: "support/referenced", LastModified: now.Add(-48 * time.Hour)},
		{Key: "support/orphan", LastModified: now.Add(-48 * time.Hour)},
		{Key: "support/recent", LastModified: now.Add(-time.Hour)},
	}}
	service := New(store, scannerResult{clean: true})

	deleted, err := service.CleanupOrphans(context.Background(), map[string]struct{}{"support/referenced": {}}, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("CleanupOrphans returned error: %v", err)
	}

	if deleted != 1 || len(store.deletes) != 1 || store.deletes[0] != "support/orphan" {
		t.Fatalf("unexpected cleanup result: deleted=%d keys=%v", deleted, store.deletes)
	}
}

func multipartFiles(t *testing.T, values map[string][]byte) []*multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)
	for name, value := range values {
		part, err := writer.CreateFormFile("files", name)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := part.Write(value); err != nil {
			t.Fatal(err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	if err := request.ParseMultipartForm(MaxMessageBytes + 1024); err != nil {
		t.Fatal(err)
	}

	return request.MultipartForm.File["files"]
}
