package video

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type staticVideoOutputResolver struct {
	addresses []netip.Addr
	err       error
}

func (r staticVideoOutputResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	return r.addresses, r.err
}

type recordingVideoOutputDialer struct {
	addresses []string
	err       error
}

func (d *recordingVideoOutputDialer) DialContext(_ context.Context, _, address string) (net.Conn, error) {
	d.addresses = append(d.addresses, address)
	return nil, d.err
}

func TestCopyFromURLRejectsLoopbackBeforeSendingRequest(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests.Add(1)
	}))
	defer server.Close()

	store := &S3ObjectStore{
		client:     &s3.Client{},
		httpClient: newVideoOutputHTTPClient(time.Second, true),
		allowHTTP:  true,
		maxBytes:   1024,
	}
	_, err := store.CopyFromURL(context.Background(), "tenant", "job", platform.VideoArtifact{
		ProviderURL: server.URL + "/output.mp4",
	})
	if !errors.Is(err, ErrUnsafeVideoOutputURL) {
		t.Fatalf("CopyFromURL() error = %v, want ErrUnsafeVideoOutputURL", err)
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("loopback server received %d requests, want 0", got)
	}
}

func TestValidateVideoOutputHTTPURL(t *testing.T) {
	tests := []struct {
		name      string
		rawURL    string
		allowHTTP bool
		wantErr   bool
	}{
		{name: "public HTTPS", rawURL: "https://cdn.provider.example/output.mp4"},
		{name: "public IP", rawURL: "https://93.184.216.34/output.mp4"},
		{name: "HTTP disabled", rawURL: "http://cdn.provider.example/output.mp4", wantErr: true},
		{name: "HTTP enabled", rawURL: "http://cdn.provider.example/output.mp4", allowHTTP: true},
		{name: "loopback IPv4", rawURL: "https://127.0.0.1/output.mp4", wantErr: true},
		{name: "cloud metadata", rawURL: "https://169.254.169.254/latest/meta-data", wantErr: true},
		{name: "private IPv6", rawURL: "https://[fd00::1]/output.mp4", wantErr: true},
		{name: "localhost name", rawURL: "https://worker.localhost/output.mp4", wantErr: true},
		{name: "credentials", rawURL: "https://user:password@cdn.provider.example/output.mp4", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source, err := url.Parse(test.rawURL)
			if err != nil {
				t.Fatalf("url.Parse() error = %v", err)
			}
			err = validateVideoOutputHTTPURL(source, test.allowHTTP)
			if test.wantErr && !errors.Is(err, ErrUnsafeVideoOutputURL) {
				t.Fatalf("validateVideoOutputHTTPURL() error = %v, want ErrUnsafeVideoOutputURL", err)
			}
			if !test.wantErr && err != nil {
				t.Fatalf("validateVideoOutputHTTPURL() unexpected error = %v", err)
			}
		})
	}
}

func TestResolvePublicVideoOutputIPsRejectsAnyNonPublicAnswer(t *testing.T) {
	tests := []struct {
		name      string
		addresses []netip.Addr
		wantErr   bool
	}{
		{
			name:      "public",
			addresses: []netip.Addr{netip.MustParseAddr("93.184.216.34")},
		},
		{
			name:      "private",
			addresses: []netip.Addr{netip.MustParseAddr("10.0.0.1")},
			wantErr:   true,
		},
		{
			name: "mixed public and private",
			addresses: []netip.Addr{
				netip.MustParseAddr("93.184.216.34"),
				netip.MustParseAddr("127.0.0.1"),
			},
			wantErr: true,
		},
		{
			name:      "IPv4 mapped private",
			addresses: []netip.Addr{netip.MustParseAddr("::ffff:192.168.1.1")},
			wantErr:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := resolvePublicVideoOutputIPs(
				context.Background(),
				staticVideoOutputResolver{addresses: test.addresses},
				"tcp",
				"cdn.provider.example",
			)
			if test.wantErr && !errors.Is(err, ErrUnsafeVideoOutputURL) {
				t.Fatalf("resolvePublicVideoOutputIPs() error = %v, want ErrUnsafeVideoOutputURL", err)
			}
			if !test.wantErr {
				if err != nil {
					t.Fatalf("resolvePublicVideoOutputIPs() unexpected error = %v", err)
				}
				if len(got) != len(test.addresses) {
					t.Fatalf("resolvePublicVideoOutputIPs() returned %d addresses, want %d", len(got), len(test.addresses))
				}
			}
		})
	}
}

func TestSafeVideoOutputDialPinsResolvedPublicIP(t *testing.T) {
	dialErr := errors.New("dial stopped by test")
	dialer := &recordingVideoOutputDialer{err: dialErr}
	dial := safeVideoOutputDialContext(
		staticVideoOutputResolver{addresses: []netip.Addr{netip.MustParseAddr("93.184.216.34")}},
		dialer,
	)

	_, err := dial(context.Background(), "tcp", "cdn.provider.example:443")
	if !errors.Is(err, dialErr) {
		t.Fatalf("safe dial error = %v, want wrapped test dial error", err)
	}
	if len(dialer.addresses) != 1 || dialer.addresses[0] != "93.184.216.34:443" {
		t.Fatalf("dialed addresses = %v, want resolved IP only", dialer.addresses)
	}
}

func TestVideoOutputRedirectPolicyRejectsPrivateTarget(t *testing.T) {
	client := newVideoOutputHTTPClient(time.Second, true)
	target, err := url.Parse("http://169.254.169.254/latest/meta-data")
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	err = client.CheckRedirect(
		&http.Request{URL: target},
		[]*http.Request{{URL: &url.URL{Scheme: "https", Host: "cdn.provider.example"}}},
	)
	if !errors.Is(err, ErrUnsafeVideoOutputURL) {
		t.Fatalf("CheckRedirect() error = %v, want ErrUnsafeVideoOutputURL", err)
	}
}
