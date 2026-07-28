package billing

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/shopspring/decimal"
)

const (
	imageBillingModeOff     = "off"
	imageBillingModeShadow  = "shadow"
	imageBillingModeEnforce = "enforce"

	providerCostTicksPerUSD = int64(10_000_000_000)
)

func imageDimensionalBillingModeFromEnv() string {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("IMAGE_DIMENSIONAL_BILLING_MODE"))) {
	case imageBillingModeShadow:
		return imageBillingModeShadow
	case imageBillingModeEnforce:
		return imageBillingModeEnforce
	default:
		return imageBillingModeOff
	}
}

// EstimateImageUsage rates a pre-provider image usage estimate using either the
// legacy token-only path or the dimension-aware path selected by rollout mode.
func (s *Service) EstimateImageUsage(
	ctx context.Context,
	dep platform.Deployment,
	model string,
	usage platform.ImageUsage,
) (platform.UsageCostBreakdown, error) {
	usage = normalizeImageUsage(usage)
	legacy := func() (platform.UsageCostBreakdown, error) {
		return s.rateImageDeploymentUsage(ctx, dep, model, usage, false)
	}
	dimensional := func() (platform.UsageCostBreakdown, error) {
		return s.rateImageDeploymentUsage(ctx, dep, model, usage, true)
	}

	switch s.imageBillingMode {
	case imageBillingModeEnforce:
		rated, err := dimensional()
		observeImageBillingResult(dep.Provider, rated.RatingStatus, err)
		return rated, err
	case imageBillingModeShadow:
		legacyRated, legacyErr := legacy()
		dimensionalRated, dimensionalErr := dimensional()
		observeImageBillingShadow(
			dep.Provider,
			legacyRated,
			legacyErr,
			dimensionalRated,
			dimensionalErr,
		)
		observeImageBillingResult(dep.Provider, legacyRated.RatingStatus, legacyErr)
		return legacyRated, legacyErr
	default:
		rated, err := legacy()
		observeImageBillingResult(dep.Provider, rated.RatingStatus, err)
		return rated, err
	}
}

func (s *Service) rateImageDeploymentUsage(
	ctx context.Context,
	dep platform.Deployment,
	model string,
	usage platform.ImageUsage,
	dimensional bool,
) (platform.UsageCostBreakdown, error) {
	base := imageUsageBreakdown(model, dep.Provider, dep.RegionName, usage)
	var (
		measures []platform.UsageMeasure
		err      error
	)
	if dimensional {
		measures, err = buildImageUsageMeasures(
			dep.EffectivePricing,
			dep.RegionName,
			usage,
			time.Now().UTC(),
		)
	} else {
		measures = tokenUsageMeasures(usage.InputTokens, usage.OutputTokens)
	}
	if err != nil {
		if strings.EqualFold(strings.TrimSpace(dep.BillingMode), platform.BillingModeExternalBilled) &&
			(errors.Is(err, ErrPricingUnresolved) || errors.Is(err, ErrRateUnresolved)) {
			return s.unratedImageBreakdown(ctx, dep, base, measures), nil
		}
		return platform.UsageCostBreakdown{}, err
	}

	base.UsageMeasures = measures
	return s.RateDeploymentUsage(ctx, dep, base)
}

func (s *Service) unratedImageBreakdown(
	ctx context.Context,
	dep platform.Deployment,
	base platform.UsageCostBreakdown,
	measures []platform.UsageMeasure,
) platform.UsageCostBreakdown {
	rate := s.exchangeRate(ctx)
	exchangeRate := decimal.Zero
	if rate != nil {
		exchangeRate = decimal.NewFromFloat(rate.USDtoIRR)
		base.ExchangeRate = rate.USDtoIRR
	}
	base.ExchangeRateExact = exchangeRate.String()
	base.UsageMeasures = cloneUsageMeasures(measures)
	base.RatingStatus = "unrated"
	base.PricingSource = dep.PricingSource
	base.PricingSnapshot = cloneModelPrice(dep.EffectivePricing)
	return base
}

// FinalizeImageUsage reconciles image usage after the provider has returned.
// Provider-reported exact cost wins; otherwise actual dimensional usage is
// rated and the reserved estimate is the final fail-safe.
func (s *Service) FinalizeImageUsage(
	ctx context.Context,
	model string,
	usage platform.ImageUsage,
	estimated platform.UsageCostBreakdown,
) (platform.UsageCostBreakdown, error) {
	usage = applyEstimatedImageInputSplit(usage, estimated)
	usage = normalizeImageUsage(usage)
	provider := imageBillingProvider(estimated)

	if usage.ProviderCostUSDTicks != nil && *usage.ProviderCostUSDTicks >= 0 {
		base := s.bestEffortImageFinalization(model, usage, estimated)
		finalized := s.applyProviderReportedImageCost(ctx, base, usage)
		if s.imageBillingMode == imageBillingModeShadow {
			dimensional, dimensionalErr := s.rateFinalImageUsage(
				model,
				usage,
				estimated,
				true,
			)
			observeImageBillingShadow(
				provider,
				finalized,
				nil,
				dimensional,
				dimensionalErr,
			)
		}
		observeImageBillingResult(provider, finalized.RatingStatus, nil)
		return finalized, nil
	}

	legacy := func() (platform.UsageCostBreakdown, error) {
		return s.rateFinalImageUsage(model, usage, estimated, false)
	}
	dimensional := func() (platform.UsageCostBreakdown, error) {
		return s.rateFinalImageUsage(model, usage, estimated, true)
	}

	var (
		finalized platform.UsageCostBreakdown
		err       error
	)
	switch s.imageBillingMode {
	case imageBillingModeEnforce:
		finalized, err = dimensional()
	case imageBillingModeShadow:
		legacyRated, legacyErr := legacy()
		dimensionalRated, dimensionalErr := dimensional()
		observeImageBillingShadow(
			provider,
			legacyRated,
			legacyErr,
			dimensionalRated,
			dimensionalErr,
		)
		finalized, err = legacyRated, legacyErr
	default:
		finalized, err = legacy()
	}
	if err != nil {
		if canUseEstimatedImageFinal(estimated) {
			fallback := estimatedImageFinal(estimated, usage, err)
			observeImageBillingFallback(provider, "post_provider_rate_unresolved")
			observeImageBillingResult(provider, fallback.RatingStatus, nil)
			return fallback, nil
		}
		observeImageBillingResult(provider, "", err)
		return platform.UsageCostBreakdown{}, err
	}

	observeImageBillingResult(provider, finalized.RatingStatus, nil)
	return finalized, nil
}

func (s *Service) bestEffortImageFinalization(
	model string,
	usage platform.ImageUsage,
	estimated platform.UsageCostBreakdown,
) platform.UsageCostBreakdown {
	dimensional := s.imageBillingMode == imageBillingModeEnforce
	if finalized, err := s.rateFinalImageUsage(model, usage, estimated, dimensional); err == nil {
		return finalized
	}
	return estimatedImageFinal(estimated, usage, nil)
}

func (s *Service) rateFinalImageUsage(
	model string,
	usage platform.ImageUsage,
	estimated platform.UsageCostBreakdown,
	dimensional bool,
) (platform.UsageCostBreakdown, error) {
	if !imageUsageHasActualFacts(usage) {
		return platform.UsageCostBreakdown{}, ErrRateUnresolved
	}

	var (
		measures []platform.UsageMeasure
		err      error
	)
	if dimensional {
		measures, err = buildImageUsageMeasures(
			estimated.PricingSnapshot,
			imageBillingRegion(estimated),
			usage,
			time.Now().UTC(),
		)
	} else {
		measures = tokenUsageMeasures(usage.InputTokens, usage.OutputTokens)
		if len(measures) == 0 {
			return platform.UsageCostBreakdown{}, ErrRateUnresolved
		}
	}
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	rated, err := s.rateSnapshotUsage(model, measures, estimated)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}
	applyImageUsageTotals(&rated, usage)
	rated.Metadata = mergeImageUsageMetadata(rated.Metadata, usage)
	return rated, nil
}

func canUseEstimatedImageFinal(estimated platform.UsageCostBreakdown) bool {
	switch strings.ToLower(strings.TrimSpace(estimated.RatingStatus)) {
	case "rated", "free", "unrated", "estimated_final":
		return true
	default:
		return estimated.PricingSnapshot != nil
	}
}

func estimatedImageFinal(
	estimated platform.UsageCostBreakdown,
	usage platform.ImageUsage,
	cause error,
) platform.UsageCostBreakdown {
	out := estimated
	if strings.EqualFold(strings.TrimSpace(out.RatingStatus), "rated") {
		out.RatingStatus = "estimated_final"
	}
	applyImageUsageTotals(&out, usage)
	out.Metadata = mergeImageUsageMetadata(out.Metadata, usage)
	if cause != nil {
		out.Metadata["image_billing_fallback_reason"] = "actual_usage_rate_unresolved"
	}
	return out
}

func (s *Service) applyProviderReportedImageCost(
	ctx context.Context,
	base platform.UsageCostBreakdown,
	usage platform.ImageUsage,
) platform.UsageCostBreakdown {
	ticks := int64(0)
	if usage.ProviderCostUSDTicks != nil {
		ticks = *usage.ProviderCostUSDTicks
	}
	totalUSD := decimal.NewFromInt(ticks).Div(decimal.NewFromInt(providerCostTicksPerUSD))
	rate := s.exchangeRate(ctx)
	exchangeRate := decimal.Zero
	if rate != nil {
		exchangeRate = decimal.NewFromFloat(rate.USDtoIRR)
	}
	totalIRR := totalUSD.Mul(exchangeRate)

	base.TotalCostUSD, _ = totalUSD.Float64()
	base.TotalCostIRR, _ = totalIRR.Float64()
	base.ExchangeRate, _ = exchangeRate.Float64()
	base.TotalCostUSDExact = totalUSD.String()
	base.TotalCostIRRExact = totalIRR.String()
	base.ExchangeRateExact = exchangeRate.String()
	base.InputCostUSD = 0
	base.OutputCostUSD = base.TotalCostUSD
	base.RatingStatus = "provider_reported"
	base.PricingSource = "provider_reported"
	base.AppliedRates = []platform.AppliedRate{{
		ProfileKey:       "provider-reported",
		RateKey:          "provider-reported:request",
		Usage:            "output",
		Unit:             "request",
		RateAmount:       totalUSD.String(),
		RateQuantity:     1,
		UsageAmount:      "1",
		ComponentCostUSD: totalUSD.String(),
		Currency:         "USD",
		Resolution:       usage.Resolution,
		ServiceTier:      "standard",
		PricingSource:    "provider_reported",
		Qualifiers:       imageUsageQualifiers(usage),
	}}
	applyImageUsageTotals(&base, usage)
	base.Metadata = mergeImageUsageMetadata(base.Metadata, usage)
	return base
}

func buildImageUsageMeasures(
	price *platform.ModelPrice,
	region string,
	usage platform.ImageUsage,
	at time.Time,
) ([]platform.UsageMeasure, error) {
	usage = normalizeImageUsage(usage)
	if price == nil || strings.EqualFold(strings.TrimSpace(price.Status), "unknown") {
		return nil, ErrPricingUnresolved
	}
	if strings.EqualFold(strings.TrimSpace(price.Status), "free") {
		return canonicalFreeImageMeasures(usage), nil
	}

	profile, err := selectPricingProfile(price.Profiles, region, at)
	if err != nil {
		return nil, err
	}

	qualifiers := imageUsageQualifiers(usage)
	textTokens, imageTokens := resolvedImageInputTokens(usage)
	if usage.InputTokens > 0 &&
		usage.InputTextTokens == nil &&
		usage.InputImageTokens == nil {
		inputTokenAvailable := imageRateAvailable(
			profile.Rates,
			imageMeasure("input", "token", usage.InputTokens, "", qualifiers),
		)
		inputImageTokenAvailable := imageRateAvailable(
			profile.Rates,
			imageMeasure(
				"input",
				"image_token",
				usage.InputTokens,
				usage.Resolution,
				qualifiers,
			),
		)
		if !inputTokenAvailable && inputImageTokenAvailable {
			textTokens = 0
			imageTokens = usage.InputTokens
		}
	}
	inputTokenAmount := textTokens
	measures := make([]platform.UsageMeasure, 0, 3)

	if usage.InputImageCount > 0 {
		candidate := imageMeasure(
			"input",
			"image",
			usage.InputImageCount,
			usage.Resolution,
			qualifiers,
		)
		if imageRateAvailable(profile.Rates, candidate) {
			measures = append(measures, candidate)
			imageTokens = 0
		}
	}
	if imageTokens > 0 {
		candidate := imageMeasure(
			"input",
			"image_token",
			imageTokens,
			usage.Resolution,
			qualifiers,
		)
		if imageRateAvailable(profile.Rates, candidate) {
			measures = append(measures, candidate)
			imageTokens = 0
		}
	}
	inputTokenAmount += imageTokens
	if inputTokenAmount > 0 {
		candidate := imageMeasure("input", "token", inputTokenAmount, "", qualifiers)
		if imageRateAvailable(profile.Rates, candidate) {
			measures = append(measures, candidate)
		}
	}

	if usage.OutputImageCount > 0 {
		candidate := imageMeasure(
			"output",
			"image",
			usage.OutputImageCount,
			usage.Resolution,
			qualifiers,
		)
		if imageRateAvailable(profile.Rates, candidate) {
			return append(measures, candidate), nil
		}
	}
	if usage.OutputTokens > 0 {
		candidate := imageMeasure(
			"output",
			"image_token",
			usage.OutputTokens,
			usage.Resolution,
			qualifiers,
		)
		if imageRateAvailable(profile.Rates, candidate) {
			return append(measures, candidate), nil
		}
		candidate = imageMeasure("output", "token", usage.OutputTokens, "", qualifiers)
		if imageRateAvailable(profile.Rates, candidate) {
			return append(measures, candidate), nil
		}
	}

	return nil, ErrRateUnresolved
}

func canonicalFreeImageMeasures(usage platform.ImageUsage) []platform.UsageMeasure {
	textTokens, imageTokens := resolvedImageInputTokens(usage)
	measures := make([]platform.UsageMeasure, 0, 3)
	if textTokens > 0 {
		measures = append(measures, imageMeasure("input", "token", textTokens, "", nil))
	}
	if imageTokens > 0 {
		measures = append(measures, imageMeasure(
			"input",
			"image_token",
			imageTokens,
			usage.Resolution,
			imageUsageQualifiers(usage),
		))
	} else if usage.InputImageCount > 0 {
		measures = append(measures, imageMeasure(
			"input",
			"image",
			usage.InputImageCount,
			usage.Resolution,
			imageUsageQualifiers(usage),
		))
	}
	if usage.OutputTokens > 0 {
		measures = append(measures, imageMeasure(
			"output",
			"image_token",
			usage.OutputTokens,
			usage.Resolution,
			imageUsageQualifiers(usage),
		))
	} else if usage.OutputImageCount > 0 {
		measures = append(measures, imageMeasure(
			"output",
			"image",
			usage.OutputImageCount,
			usage.Resolution,
			imageUsageQualifiers(usage),
		))
	}
	return measures
}

func imageRateAvailable(rates []platform.PricingRate, measure platform.UsageMeasure) bool {
	amount, err := decimal.NewFromString(measure.Amount)
	if err != nil || amount.IsNegative() {
		return false
	}
	_, err = selectPricingRate(rates, measure, amount)
	return err == nil
}

func imageMeasure(
	usage string,
	unit string,
	amount int,
	resolution string,
	qualifiers map[string]interface{},
) platform.UsageMeasure {
	return platform.UsageMeasure{
		Usage:       usage,
		Unit:        unit,
		Amount:      strconv.Itoa(maxInt(amount, 0)),
		Resolution:  strings.ToLower(strings.TrimSpace(resolution)),
		ServiceTier: "standard",
		Qualifiers:  cloneInterfaceMap(qualifiers),
	}
}

func resolvedImageInputTokens(usage platform.ImageUsage) (int, int) {
	total := maxInt(usage.InputTokens, 0)
	textKnown := usage.InputTextTokens != nil
	imageKnown := usage.InputImageTokens != nil
	textTokens := 0
	imageTokens := 0
	if textKnown {
		textTokens = maxInt(*usage.InputTextTokens, 0)
	}
	if imageKnown {
		imageTokens = maxInt(*usage.InputImageTokens, 0)
	}
	switch {
	case textKnown && imageKnown:
		return textTokens, imageTokens
	case textKnown:
		return textTokens, maxInt(total-textTokens, 0)
	case imageKnown:
		return maxInt(total-imageTokens, 0), imageTokens
	default:
		return total, 0
	}
}

func normalizeImageUsage(usage platform.ImageUsage) platform.ImageUsage {
	usage.InputTokens = maxInt(usage.InputTokens, 0)
	usage.OutputTokens = maxInt(usage.OutputTokens, 0)
	usage.TotalTokens = maxInt(usage.TotalTokens, 0)
	usage.InputImageCount = maxInt(usage.InputImageCount, 0)
	usage.OutputImageCount = maxInt(usage.OutputImageCount, 0)
	if usage.InputTextTokens != nil {
		value := maxInt(*usage.InputTextTokens, 0)
		usage.InputTextTokens = &value
	}
	if usage.InputImageTokens != nil {
		value := maxInt(*usage.InputImageTokens, 0)
		usage.InputImageTokens = &value
	}

	detailTotal := 0
	if usage.InputTextTokens != nil {
		detailTotal += *usage.InputTextTokens
	}
	if usage.InputImageTokens != nil {
		detailTotal += *usage.InputImageTokens
	}
	if detailTotal > usage.InputTokens {
		usage.InputTokens = detailTotal
	}
	if usage.InputTokens == 0 && usage.OutputTokens == 0 && usage.TotalTokens > 0 {
		usage.OutputTokens = usage.TotalTokens
	} else if usage.OutputTokens == 0 && usage.TotalTokens > usage.InputTokens {
		usage.OutputTokens = usage.TotalTokens - usage.InputTokens
	}
	if usage.TotalTokens < usage.InputTokens+usage.OutputTokens {
		usage.TotalTokens = usage.InputTokens + usage.OutputTokens
	}
	usage.Resolution = strings.ToLower(strings.TrimSpace(usage.Resolution))
	usage.Quality = strings.ToLower(strings.TrimSpace(usage.Quality))
	return usage
}

func applyEstimatedImageInputSplit(
	usage platform.ImageUsage,
	estimated platform.UsageCostBreakdown,
) platform.ImageUsage {
	if usage.InputTokens <= 0 ||
		usage.InputTextTokens != nil ||
		usage.InputImageTokens != nil ||
		estimated.Metadata == nil {
		return usage
	}

	estimatedText, textOK := metadataImageInt(
		estimated.Metadata,
		"image_estimated_input_text_tokens",
	)
	estimatedImage, imageOK := metadataImageInt(
		estimated.Metadata,
		"image_estimated_input_image_tokens",
	)
	if !textOK && !imageOK {
		return usage
	}
	estimatedTotal := maxInt(estimatedText, 0) + maxInt(estimatedImage, 0)
	if estimatedTotal <= 0 {
		text := usage.InputTokens
		image := 0
		usage.InputTextTokens = &text
		usage.InputImageTokens = &image
		return usage
	}

	image := usage.InputTokens * maxInt(estimatedImage, 0) / estimatedTotal
	text := usage.InputTokens - image
	usage.InputTextTokens = &text
	usage.InputImageTokens = &image
	return usage
}

func metadataImageInt(metadata map[string]interface{}, key string) (int, bool) {
	value, ok := metadata[key]
	if !ok {
		return 0, false
	}
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), int64(int(typed)) == typed
	case float64:
		return int(typed), float64(int(typed)) == typed
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		return parsed, err == nil
	default:
		return 0, false
	}
}

func imageUsageBreakdown(
	model string,
	provider string,
	region string,
	usage platform.ImageUsage,
) platform.UsageCostBreakdown {
	return platform.UsageCostBreakdown{
		Model:        model,
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		TotalTokens:  usage.TotalTokens,
		RatingStatus: "pending",
		Metadata: mergeImageUsageMetadata(map[string]interface{}{
			"image_billing_provider": provider,
			"image_billing_region":   region,
		}, usage),
	}
}

func applyImageUsageTotals(out *platform.UsageCostBreakdown, usage platform.ImageUsage) {
	if out == nil {
		return
	}
	if usage.InputTokens > 0 || usage.OutputTokens > 0 || usage.TotalTokens > 0 {
		out.InputTokens = usage.InputTokens
		out.OutputTokens = usage.OutputTokens
		out.TotalTokens = usage.TotalTokens
	}
}

func imageUsageHasActualFacts(usage platform.ImageUsage) bool {
	if strings.EqualFold(
		strings.TrimSpace(usage.Source),
		platform.ImageUsageSourceEstimated,
	) {
		return false
	}
	return usage.InputTokens > 0 ||
		usage.OutputTokens > 0 ||
		usage.TotalTokens > 0 ||
		usage.InputImageCount > 0 ||
		usage.OutputImageCount > 0
}

func imageUsageQualifiers(usage platform.ImageUsage) map[string]interface{} {
	if strings.TrimSpace(usage.Quality) == "" {
		return nil
	}
	return map[string]interface{}{"quality": strings.ToLower(strings.TrimSpace(usage.Quality))}
}

func mergeImageUsageMetadata(
	metadata map[string]interface{},
	usage platform.ImageUsage,
) map[string]interface{} {
	out := cloneInterfaceMap(metadata)
	if out == nil {
		out = map[string]interface{}{}
	}
	if usage.Source != "" {
		out["image_usage_source"] = usage.Source
	}
	if usage.InputImageCount > 0 {
		out["input_image_count"] = usage.InputImageCount
	}
	if usage.InputTextTokens != nil {
		out["image_estimated_input_text_tokens"] = *usage.InputTextTokens
	}
	if usage.InputImageTokens != nil {
		out["image_estimated_input_image_tokens"] = *usage.InputImageTokens
	}
	if usage.OutputImageCount > 0 {
		out["output_count"] = usage.OutputImageCount
	}
	if usage.Resolution != "" {
		out["resolution"] = usage.Resolution
	}
	if usage.Quality != "" {
		out["quality"] = usage.Quality
	}
	if usage.ProviderCostUSDTicks != nil {
		out["provider_cost_in_usd_ticks"] = *usage.ProviderCostUSDTicks
	}
	return out
}

func imageBillingProvider(estimated platform.UsageCostBreakdown) string {
	if estimated.Metadata == nil {
		return "unknown"
	}
	if provider, ok := estimated.Metadata["image_billing_provider"].(string); ok &&
		strings.TrimSpace(provider) != "" {
		return strings.ToLower(strings.TrimSpace(provider))
	}
	return "unknown"
}

func imageBillingRegion(estimated platform.UsageCostBreakdown) string {
	if len(estimated.AppliedRates) > 0 &&
		strings.TrimSpace(estimated.AppliedRates[0].Region) != "" {
		return estimated.AppliedRates[0].Region
	}
	if estimated.Metadata != nil {
		if region, ok := estimated.Metadata["image_billing_region"].(string); ok {
			return strings.TrimSpace(region)
		}
	}
	return ""
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
