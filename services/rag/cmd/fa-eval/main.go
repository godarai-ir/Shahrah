package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/rag/internal/evaluation"
)

const maxEvaluationResponseBytes = 32 << 20

type evaluationHTTPClient struct {
	baseURL        string
	token          string
	knowledgeBases map[string]string
	publicBots     map[string]string
	client         *http.Client
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Persian RAG evaluation failed: %v\n", err)

		os.Exit(1)
	}
}

func run() error {
	datasetPath := flag.String("dataset", "eval/fa-ir/gold.jsonl", "Path to the Persian RAG JSONL dataset.")
	baseURL := flag.String("base-url", strings.TrimSpace(os.Getenv("RAG_EVAL_BASE_URL")), "Godar API base URL.")
	token := flag.String("token", strings.TrimSpace(os.Getenv("RAG_EVAL_TOKEN")), "Evaluation tenant API token.")
	kbMapJSON := flag.String("knowledge-base-map", strings.TrimSpace(os.Getenv("RAG_EVAL_KB_MAP")), "JSON map from dataset keys to knowledge base UUIDs.")
	botMapJSON := flag.String("bot-map", strings.TrimSpace(os.Getenv("RAG_EVAL_BOT_MAP")), "JSON map from dataset keys to public bot IDs.")
	timeout := flag.Duration("timeout", 30*time.Second, "Timeout for each API request.")

	flag.Parse()

	if strings.TrimSpace(*baseURL) == "" || strings.TrimSpace(*token) == "" {
		return errors.New("base-url and token are required")
	}

	knowledgeBases, err := decodeMap(*kbMapJSON, "knowledge-base-map")
	if err != nil {
		return err
	}

	publicBots, err := decodeMap(*botMapJSON, "bot-map")
	if err != nil {
		return err
	}

	cases, err := loadCases(*datasetPath)
	if err != nil {
		return err
	}

	client := &evaluationHTTPClient{
		baseURL: strings.TrimRight(strings.TrimSpace(*baseURL), "/"), token: strings.TrimSpace(*token),
		knowledgeBases: knowledgeBases, publicBots: publicBots, client: &http.Client{Timeout: *timeout},
	}

	report, err := evaluation.Run(context.Background(), client, cases)
	if err != nil {
		return err
	}

	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode evaluation report: %w", err)
	}

	if _, err := fmt.Fprintln(os.Stdout, string(encoded)); err != nil {
		return fmt.Errorf("write evaluation report: %w", err)
	}

	if !report.Passed {
		return errors.New("one or more Persian RAG release gates did not pass")
	}

	return nil
}

func (c *evaluationHTTPClient) Retrieve(
	ctx context.Context,
	testCase evaluation.Case,
) (platform.RetrievalResponse, error) {
	ids := make([]string, 0, len(testCase.KnowledgeBaseKeys))
	for _, key := range testCase.KnowledgeBaseKeys {
		id := strings.TrimSpace(c.knowledgeBases[key])
		if id == "" {
			return platform.RetrievalResponse{}, fmt.Errorf("knowledge base key %q is not mapped", key)
		}

		ids = append(ids, id)
	}

	payload := platform.RetrievalRequest{Query: testCase.Query, KnowledgeBaseIDs: ids}

	var response platform.RetrievalResponse
	if err := c.post(ctx, "/v1/retrieval", payload, &response); err != nil {
		return platform.RetrievalResponse{}, err
	}

	return response, nil
}

func (c *evaluationHTTPClient) Answer(
	ctx context.Context,
	testCase evaluation.Case,
) (platform.ChatbotResponse, error) {
	publicID := strings.TrimSpace(c.publicBots[testCase.PublicBotKey])
	if publicID == "" {
		return platform.ChatbotResponse{}, fmt.Errorf("public bot key %q is not mapped", testCase.PublicBotKey)
	}

	payload := platform.ChatbotResponseRequest{Message: testCase.Query}

	var response platform.ChatbotResponse
	if err := c.post(ctx, "/v1/chatbots/"+publicID+"/responses", payload, &response); err != nil {
		return platform.ChatbotResponse{}, err
	}

	return response, nil
}

func (c *evaluationHTTPClient) post(ctx context.Context, path string, payload, output any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode evaluation request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create evaluation request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Content-Type", "application/json")

	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("send evaluation request: %w", err)
	}
	defer closeResponseBody(response.Body)

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("evaluation API returned HTTP %d", response.StatusCode)
	}

	decoder := json.NewDecoder(io.LimitReader(response.Body, maxEvaluationResponseBytes+1))
	if err := decoder.Decode(output); err != nil {
		return fmt.Errorf("decode evaluation response: %w", err)
	}

	return nil
}

func loadCases(path string) (_ []evaluation.Case, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open evaluation dataset: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close evaluation dataset: %w", closeErr))
		}
	}()

	cases := make([]evaluation.Case, 0, 32)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	line := 0
	for scanner.Scan() {
		line++

		content := strings.TrimSpace(scanner.Text())
		if content == "" || strings.HasPrefix(content, "#") {
			continue
		}

		var testCase evaluation.Case
		if err := json.Unmarshal([]byte(content), &testCase); err != nil {
			return nil, fmt.Errorf("decode evaluation dataset line %d: %w", line, err)
		}

		cases = append(cases, testCase)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read evaluation dataset: %w", err)
	}

	return cases, nil
}

func decodeMap(input, name string) (map[string]string, error) {
	result := map[string]string{}
	if err := json.Unmarshal([]byte(input), &result); err != nil {
		return nil, fmt.Errorf("decode %s: %w", name, err)
	}

	return result, nil
}

func closeResponseBody(body io.Closer) {
	if err := body.Close(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Evaluation response body could not be closed.")
	}
}
