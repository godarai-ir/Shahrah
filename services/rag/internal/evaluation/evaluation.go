// Package evaluation measures the Persian managed RAG release gates.
package evaluation

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/rag/internal/domain"
)

const (
	minimumGroundedAccuracy = 0.85
	minimumRefusalAccuracy  = 0.90
	minimumRetrievalRecall  = 0.85
)

// Case is one reproducible Persian or bilingual evaluation query.
type Case struct {
	ID                     string   `json:"id"`
	Mode                   string   `json:"mode"`
	Query                  string   `json:"query"`
	KnowledgeBaseKeys      []string `json:"knowledge_base_keys"`
	PublicBotKey           string   `json:"public_bot_key"`
	Answerable             bool     `json:"answerable"`
	ExpectedDocumentTitles []string `json:"expected_document_titles,omitempty"`
	ExpectedLanguage       string   `json:"expected_language"`
	RequiredTerms          []string `json:"required_terms,omitempty"`
	ForbiddenTerms         []string `json:"forbidden_terms,omitempty"`
}

// Client invokes a provisioned evaluation tenant through public product contracts.
type Client interface {
	Retrieve(context.Context, Case) (platform.RetrievalResponse, error)
	Answer(context.Context, Case) (platform.ChatbotResponse, error)
}

// Report contains aggregate release-gate metrics without source or query content.
type Report struct {
	Cases            int      `json:"cases"`
	Failures         []string `json:"failures"`
	RetrievalTotal   int      `json:"retrieval_total"`
	RetrievalPassed  int      `json:"retrieval_passed"`
	RetrievalRecall  float64  `json:"retrieval_recall"`
	GroundedTotal    int      `json:"grounded_total"`
	GroundedPassed   int      `json:"grounded_passed"`
	GroundedAccuracy float64  `json:"grounded_accuracy"`
	RefusalTotal     int      `json:"refusal_total"`
	RefusalPassed    int      `json:"refusal_passed"`
	RefusalAccuracy  float64  `json:"refusal_accuracy"`
	CitationTotal    int      `json:"citation_total"`
	CitationPassed   int      `json:"citation_passed"`
	CitationValidity float64  `json:"citation_validity"`
	LanguageTotal    int      `json:"language_total"`
	LanguagePassed   int      `json:"language_passed"`
	LanguageAccuracy float64  `json:"language_accuracy"`
	Passed           bool     `json:"passed"`
}

// Run evaluates every case and applies the GA quality thresholds.
func Run(ctx context.Context, client Client, cases []Case) (Report, error) {
	if client == nil {
		return Report{}, errors.New("evaluation client is required")
	}

	if len(cases) == 0 {
		return Report{}, errors.New("evaluation dataset is empty")
	}

	report := Report{Cases: len(cases), Failures: []string{}}
	for _, testCase := range cases {
		if err := validateCase(testCase); err != nil {
			return Report{}, fmt.Errorf("case %q: %w", testCase.ID, err)
		}

		if testCase.Mode == "retrieval" || testCase.Mode == "both" {
			evaluateRetrieval(ctx, client, testCase, &report)
		}

		if testCase.Mode == "answer" || testCase.Mode == "both" {
			evaluateAnswer(ctx, client, testCase, &report)
		}
	}

	report.RetrievalRecall = ratio(report.RetrievalPassed, report.RetrievalTotal)
	report.GroundedAccuracy = ratio(report.GroundedPassed, report.GroundedTotal)
	report.RefusalAccuracy = ratio(report.RefusalPassed, report.RefusalTotal)
	report.CitationValidity = ratio(report.CitationPassed, report.CitationTotal)
	report.LanguageAccuracy = ratio(report.LanguagePassed, report.LanguageTotal)
	report.Passed = report.RetrievalTotal > 0 && report.GroundedTotal > 0 && report.RefusalTotal > 0 &&
		report.CitationTotal > 0 && report.LanguageTotal > 0 &&
		report.RetrievalRecall >= minimumRetrievalRecall &&
		report.GroundedAccuracy >= minimumGroundedAccuracy &&
		report.RefusalAccuracy >= minimumRefusalAccuracy &&
		report.CitationValidity == 1 && report.LanguageAccuracy >= minimumGroundedAccuracy

	return report, nil
}

func validateCase(testCase Case) error {
	if strings.TrimSpace(testCase.ID) == "" || strings.TrimSpace(testCase.Query) == "" {
		return errors.New("id and query are required")
	}

	if testCase.Mode != "retrieval" && testCase.Mode != "answer" && testCase.Mode != "both" {
		return errors.New("mode must be retrieval, answer, or both")
	}

	if (testCase.Mode == "retrieval" || testCase.Mode == "both") && len(testCase.KnowledgeBaseKeys) == 0 {
		return errors.New("retrieval mode requires knowledge_base_keys")
	}

	if (testCase.Mode == "answer" || testCase.Mode == "both") && strings.TrimSpace(testCase.PublicBotKey) == "" {
		return errors.New("answer mode requires public_bot_key")
	}

	if testCase.ExpectedLanguage != "fa" && testCase.ExpectedLanguage != "en" {
		return errors.New("expected_language must be fa or en")
	}

	if testCase.Answerable && len(testCase.ExpectedDocumentTitles) == 0 {
		return errors.New("answerable cases require expected_document_titles")
	}

	return nil
}

func evaluateRetrieval(ctx context.Context, client Client, testCase Case, report *Report) {
	if !testCase.Answerable {
		return
	}

	report.RetrievalTotal++

	response, err := client.Retrieve(ctx, testCase)
	if err != nil {
		report.Failures = append(report.Failures, testCase.ID+": retrieval request failed")

		return
	}

	passed := chunksContainExpectedTitle(response.Chunks, testCase.ExpectedDocumentTitles)
	if passed {
		report.RetrievalPassed++

		return
	}

	report.Failures = append(report.Failures, testCase.ID+": retrieval expectation failed")
}

func evaluateAnswer(ctx context.Context, client Client, testCase Case, report *Report) {
	report.LanguageTotal++
	if testCase.Answerable {
		report.GroundedTotal++
		report.CitationTotal++
	} else {
		report.RefusalTotal++
	}

	response, err := client.Answer(ctx, testCase)
	if err != nil {
		report.Failures = append(report.Failures, testCase.ID+": answer request failed")

		return
	}

	if testCase.Answerable {
		grounded := !response.Refusal && answerContainsTerms(response.Answer, testCase.RequiredTerms) &&
			answerExcludesTerms(response.Answer, testCase.ForbiddenTerms) &&
			citationsContainExpectedTitle(response.Citations, testCase.ExpectedDocumentTitles)
		if grounded {
			report.GroundedPassed++
		} else {
			report.Failures = append(report.Failures, testCase.ID+": grounded answer expectation failed")
		}

		if citationsAreValid(response.Answer, response.Citations) {
			report.CitationPassed++
		} else {
			report.Failures = append(report.Failures, testCase.ID+": citation validity failed")
		}
	} else {
		if response.Refusal && len(response.Citations) == 0 {
			report.RefusalPassed++
		} else {
			report.Failures = append(report.Failures, testCase.ID+": refusal expectation failed")
		}
	}

	actualLanguage := domain.ResolveAnswerLanguage("query", "fa-IR", response.Answer)
	if actualLanguage == testCase.ExpectedLanguage {
		report.LanguagePassed++
	} else {
		report.Failures = append(report.Failures, testCase.ID+": answer language failed")
	}
}

func answerExcludesTerms(answer string, forbidden []string) bool {
	normalizedAnswer := domain.NormalizePersian(answer)
	for _, term := range forbidden {
		if strings.Contains(normalizedAnswer, domain.NormalizePersian(term)) {
			return false
		}
	}

	return true
}

func chunksContainExpectedTitle(chunks []platform.RetrievalChunk, expected []string) bool {
	for _, chunk := range chunks {
		if containsFold(expected, chunk.Title) {
			return true
		}
	}

	return false
}

func citationsContainExpectedTitle(citations []platform.Citation, expected []string) bool {
	for _, citation := range citations {
		if containsFold(expected, citation.Title) {
			return true
		}
	}

	return false
}

func answerContainsTerms(answer string, required []string) bool {
	normalizedAnswer := domain.NormalizePersian(answer)
	for _, term := range required {
		if !strings.Contains(normalizedAnswer, domain.NormalizePersian(term)) {
			return false
		}
	}

	return true
}

func citationsAreValid(answer string, citations []platform.Citation) bool {
	if len(citations) == 0 {
		return false
	}

	seen := make(map[string]struct{}, len(citations))
	for _, citation := range citations {
		if strings.TrimSpace(citation.ID) == "" || strings.TrimSpace(citation.DocumentID) == "" ||
			strings.TrimSpace(citation.Title) == "" || strings.TrimSpace(citation.Quote) == "" ||
			!strings.Contains(answer, "["+citation.ID+"]") {
			return false
		}

		if _, exists := seen[citation.ID]; exists {
			return false
		}

		seen[citation.ID] = struct{}{}
	}

	return true
}

func containsFold(values []string, expected string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(expected)) {
			return true
		}
	}

	return false
}

func ratio(passed, total int) float64 {
	if total == 0 {
		return 0
	}

	return float64(passed) / float64(total)
}
