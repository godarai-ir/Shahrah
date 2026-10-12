package credits

import (
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// ProductMetadataKey lets request setup (for example on-behalf-of service keys)
// pin the product label for a request.
const ProductMetadataKey = "godarai_product"

// TrustedProduct is a product label set by the control plane itself. Product
// only honors values of this type, so a string a client manages to place in
// request metadata can never relabel its usage.
type TrustedProduct string

// MarshalText keeps the label readable when request metadata is persisted.
func (p TrustedProduct) MarshalText() ([]byte, error) {
	return []byte(p), nil
}

// Product labels the GodarAI surface that produced usage from request metadata
// and the request type. The first matching rule wins.
func Product(metadata map[string]interface{}, requestType string) string {
	if product, ok := metadata[ProductMetadataKey].(TrustedProduct); ok && strings.TrimSpace(string(product)) != "" {
		return strings.TrimSpace(string(product))
	}

	if metadataText(metadata, "playground_source") != "" {
		return platform.CreditProductPlayground
	}

	switch strings.ToLower(strings.TrimSpace(requestType)) {
	case "embeddings":
		return platform.CreditProductEmbeddings
	case "ocr":
		return platform.CreditProductOCR
	}

	return platform.CreditProductGateway
}

func metadataText(metadata map[string]interface{}, key string) string {
	value, ok := metadata[key].(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(value)
}
