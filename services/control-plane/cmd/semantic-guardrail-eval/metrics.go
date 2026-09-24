package main

import (
	"math"
	"sort"
)

const (
	groupAll     = "all"
	groupPersian = "persian"
	groupHard    = "hard"

	defaultReviewThreshold = 0.35
	defaultBlockThreshold  = 0.70
)

// prediction is one rule probability for one sample under one language variant.
type prediction struct {
	SampleID string  `json:"sample_id"`
	Variant  string  `json:"variant"`
	Rule     string  `json:"rule"`
	Register string  `json:"register"`
	Kind     string  `json:"kind"`
	Hard     bool    `json:"hard"`
	Label    bool    `json:"label"`
	Score    float64 `json:"score"`
}

// countsForRule excludes samples that are positive only for another rule. Attacks often match several rules,
// and any rule firing on an attack is a correct block, so those samples say nothing about false positives.
func (p prediction) countsForRule() bool {
	return p.Label || p.Kind != kindPositive
}

// operatingPoint is the confusion matrix at one threshold. A score at or above the threshold is a positive prediction.
type operatingPoint struct {
	Threshold float64 `json:"threshold"`
	Recall    float64 `json:"recall"`
	FPR       float64 `json:"false_positive_rate"`
	Precision float64 `json:"precision"`
	TP        int     `json:"tp"`
	FP        int     `json:"fp"`
	FN        int     `json:"fn"`
	TN        int     `json:"tn"`
}

// groupMetrics summarizes one rule for one register group and one language variant.
type groupMetrics struct {
	Variant   string          `json:"variant"`
	Rule      string          `json:"rule"`
	Group     string          `json:"group"`
	Positives int             `json:"positives"`
	Negatives int             `json:"negatives"`
	AUC       *float64        `json:"auc,omitempty"`
	Best      *operatingPoint `json:"best_at_max_fpr,omitempty"`
	AtReview  operatingPoint  `json:"at_review_threshold"`
	AtBlock   operatingPoint  `json:"at_block_threshold"`
}

func registerGroups() []string {
	groups := make([]string, 0, 3+len(validRegisters))
	groups = append(groups, groupAll, groupPersian, groupHard)

	return append(groups, validRegisters...)
}

func inGroup(p prediction, group string) bool {
	switch group {
	case groupAll:
		return true
	case groupPersian:
		return p.Register == "fa_formal" || p.Register == "fa_colloquial"
	case groupHard:
		return p.Hard
	default:
		return p.Register == group
	}
}

func computeGroupMetrics(predictions []prediction, variants, rules []string, maxFPR float64) []groupMetrics {
	var out []groupMetrics

	for _, variant := range variants {
		for _, ruleKey := range rules {
			for _, group := range registerGroups() {
				selected := filterPredictions(predictions, variant, ruleKey, group)
				if len(selected) == 0 {
					continue
				}

				out = append(out, summarize(selected, variant, ruleKey, group, maxFPR))
			}
		}
	}

	return out
}

func filterPredictions(predictions []prediction, variant, ruleKey, group string) []prediction {
	var out []prediction

	for _, p := range predictions {
		if p.Variant == variant && p.Rule == ruleKey && p.countsForRule() && inGroup(p, group) {
			out = append(out, p)
		}
	}

	return out
}

func summarize(predictions []prediction, variant, ruleKey, group string, maxFPR float64) groupMetrics {
	metrics := groupMetrics{
		Variant:  variant,
		Rule:     ruleKey,
		Group:    group,
		AtReview: confusionAt(predictions, defaultReviewThreshold),
		AtBlock:  confusionAt(predictions, defaultBlockThreshold),
	}

	for _, p := range predictions {
		if p.Label {
			metrics.Positives++
		} else {
			metrics.Negatives++
		}
	}

	if auc, ok := rocAUC(predictions); ok {
		metrics.AUC = &auc
	}

	if best, ok := bestAtMaxFPR(predictions, maxFPR); ok {
		metrics.Best = &best
	}

	return metrics
}

func confusionAt(predictions []prediction, threshold float64) operatingPoint {
	point := operatingPoint{Threshold: threshold}

	for _, p := range predictions {
		predicted := p.Score >= threshold

		switch {
		case p.Label && predicted:
			point.TP++
		case p.Label:
			point.FN++
		case predicted:
			point.FP++
		default:
			point.TN++
		}
	}

	point.Recall = ratio(point.TP, point.TP+point.FN)
	point.FPR = ratio(point.FP, point.FP+point.TN)
	point.Precision = ratio(point.TP, point.TP+point.FP)

	return point
}

// bestAtMaxFPR returns the threshold with the highest recall whose false positive rate stays within maxFPR.
func bestAtMaxFPR(predictions []prediction, maxFPR float64) (operatingPoint, bool) {
	thresholds := make([]float64, 0, len(predictions)+1)
	for _, p := range predictions {
		thresholds = append(thresholds, p.Score)
	}

	thresholds = append(thresholds, math.Nextafter(1, 2))
	sort.Float64s(thresholds)

	var (
		best  operatingPoint
		found bool
	)

	for _, threshold := range thresholds {
		point := confusionAt(predictions, threshold)
		if point.TP+point.FN == 0 || point.FPR > maxFPR {
			continue
		}

		// Thresholds ascend, so on equal recall the later point has the same or a lower false positive rate.
		if !found || point.Recall >= best.Recall {
			best = point
			found = true
		}
	}

	return best, found
}

// rocAUC is the Mann-Whitney estimate of the probability that a positive scores above a negative. It is undefined without both classes.
func rocAUC(predictions []prediction) (float64, bool) {
	var positives, negatives []float64

	for _, p := range predictions {
		if p.Label {
			positives = append(positives, p.Score)
		} else {
			negatives = append(negatives, p.Score)
		}
	}

	if len(positives) == 0 || len(negatives) == 0 {
		return 0, false
	}

	var wins float64

	for _, positive := range positives {
		for _, negative := range negatives {
			switch {
			case positive > negative:
				wins++
			case positive == negative:
				wins += 0.5
			}
		}
	}

	return wins / float64(len(positives)*len(negatives)), true
}

// percentile uses the nearest-rank method on a copy of values.
func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)

	rank := int(math.Ceil(p / 100 * float64(len(sorted))))
	if rank < 1 {
		rank = 1
	}

	return sorted[rank-1]
}

func ratio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}

	return float64(numerator) / float64(denominator)
}
