package video

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var ErrUnsafeVideoOutputURL = errors.New("unsafe video provider output URL")

type videoOutputIPResolver interface {
	LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error)
}

type videoOutputDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

var blockedVideoOutputPrefixes = mustVideoOutputPrefixes(
	"0.0.0.0/8",
	"10.0.0.0/8",
	"100.64.0.0/10",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"192.88.99.0/24",
	"192.168.0.0/16",
	"198.18.0.0/15",
	"198.51.100.0/24",
	"203.0.113.0/24",
	"224.0.0.0/4",
	"240.0.0.0/4",
	"::/96",
	"64:ff9b::/96",
	"64:ff9b:1::/48",
	"100::/64",
	"2001:2::/48",
	"2001:10::/28",
	"2001:20::/28",
	"2001:db8::/32",
	"2002::/16",
	"fc00::/7",
	"fe80::/10",
	"fec0::/10",
	"ff00::/8",
)

func newVideoOutputHTTPClient(timeout time.Duration, allowHTTP bool) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// An environment proxy would become the real network destination and bypass
	// the address validation performed by the pinned dialer below.
	transport.Proxy = nil
	transport.DialContext = safeVideoOutputDialContext(
		net.DefaultResolver,
		&net.Dialer{Timeout: min(timeout, 30*time.Second), KeepAlive: 30 * time.Second},
	)

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("%w: too many redirects", ErrUnsafeVideoOutputURL)
			}
			return validateVideoOutputHTTPURL(request.URL, allowHTTP)
		},
	}
}

func validateVideoOutputHTTPURL(source *url.URL, allowHTTP bool) error {
	if source == nil || !source.IsAbs() || source.Host == "" || source.Opaque != "" {
		return fmt.Errorf("%w: URL must be absolute", ErrUnsafeVideoOutputURL)
	}
	if source.User != nil {
		return fmt.Errorf("%w: URL credentials are not allowed", ErrUnsafeVideoOutputURL)
	}

	scheme := strings.ToLower(source.Scheme)
	if scheme != "https" && !(allowHTTP && scheme == "http") {
		return fmt.Errorf("%w: URL scheme is not allowed", ErrUnsafeVideoOutputURL)
	}

	host := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(source.Hostname())), ".")
	if host == "" {
		return fmt.Errorf("%w: URL host is required", ErrUnsafeVideoOutputURL)
	}
	if host == "localhost" ||
		strings.HasSuffix(host, ".localhost") ||
		strings.HasSuffix(host, ".local") ||
		strings.HasSuffix(host, ".internal") {
		return fmt.Errorf("%w: local host is not allowed", ErrUnsafeVideoOutputURL)
	}
	if port := source.Port(); port != "" {
		portNumber, err := strconv.Atoi(port)
		if err != nil || portNumber < 1 || portNumber > 65535 {
			return fmt.Errorf("%w: invalid port", ErrUnsafeVideoOutputURL)
		}
	}
	if address, err := netip.ParseAddr(host); err == nil && !isPublicVideoOutputIP(address) {
		return fmt.Errorf("%w: non-public address is not allowed", ErrUnsafeVideoOutputURL)
	}
	return nil
}

func safeVideoOutputDialContext(resolver videoOutputIPResolver, dialer videoOutputDialer) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid network address", ErrUnsafeVideoOutputURL)
		}

		addresses, err := resolvePublicVideoOutputIPs(ctx, resolver, network, host)
		if err != nil {
			return nil, err
		}

		var dialErrors []error
		for _, resolved := range addresses {
			connection, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(resolved.String(), port))
			if dialErr == nil {
				return connection, nil
			}
			dialErrors = append(dialErrors, dialErr)
		}
		return nil, fmt.Errorf("provider output connection failed: %w", errors.Join(dialErrors...))
	}
}

func resolvePublicVideoOutputIPs(
	ctx context.Context,
	resolver videoOutputIPResolver,
	network string,
	host string,
) ([]netip.Addr, error) {
	lookupNetwork := "ip"
	switch network {
	case "tcp4":
		lookupNetwork = "ip4"
	case "tcp6":
		lookupNetwork = "ip6"
	}
	addresses, err := resolver.LookupNetIP(ctx, lookupNetwork, host)
	if err != nil {
		return nil, fmt.Errorf("resolve provider output host: %w", err)
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("resolve provider output host: no addresses")
	}

	result := make([]netip.Addr, 0, len(addresses))
	for _, address := range addresses {
		address = address.Unmap()
		if !isPublicVideoOutputIP(address) {
			return nil, fmt.Errorf("%w: host resolves to a non-public address", ErrUnsafeVideoOutputURL)
		}
		result = append(result, address)
	}
	return result, nil
}

func isPublicVideoOutputIP(address netip.Addr) bool {
	address = address.Unmap()
	if !address.IsValid() || address.Zone() != "" || !address.IsGlobalUnicast() {
		return false
	}
	for _, prefix := range blockedVideoOutputPrefixes {
		if prefix.Contains(address) {
			return false
		}
	}
	return true
}

func mustVideoOutputPrefixes(values ...string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		prefixes = append(prefixes, netip.MustParsePrefix(value))
	}
	return prefixes
}
