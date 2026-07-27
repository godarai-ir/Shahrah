package video

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type adapterContractTunnel struct {
	last platform.TunnelRequest
}

func (t *adapterContractTunnel) Proxy(_ context.Context, request platform.TunnelRequest) (*platform.TunnelResponse, error) {
	t.last = request
	return &platform.TunnelResponse{StatusCode: 200, Body: []byte(`{"id":"job-1","status":"queued"}`)}, nil
}
func (*adapterContractTunnel) Stream(context.Context, platform.TunnelRequest) (platform.TunnelStream, error) {
	return nil, nil
}
func (*adapterContractTunnel) Healthy(context.Context) error { return nil }

func TestProviderAdapterSubmitContracts(t *testing.T) {
	cases := []struct {
		provider string
		path     string
		dep      platform.Deployment
	}{
		{provider: "openai", path: "/videos"},
		{provider: "azure-open-ai", path: "/videos"},
		{provider: "xai", path: "/videos/generations"},
		{provider: "deepinfra", path: "/v1/inference/provider-model"},
		{provider: "together-ai", path: "/videos"},
		{provider: "google-gemini", path: "/models/provider-model:predictLongRunning"},
		{
			provider: "google-vertex",
			path:     "/v1/projects/project-1/locations/us-central1/publishers/google/models/provider-model:predictLongRunning",
			dep:      platform.Deployment{ProjectID: "project-1", RegionName: "us-central1"},
		},
		{
			provider: "aws-bedrock",
			path:     "/model/provider-model/async-invoke",
			dep: platform.Deployment{
				RegionName: "us-west-2",
				ModelInfo:  map[string]interface{}{"video_output_s3_uri": "s3://video-output/prefix"},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.provider, func(t *testing.T) {
			tunnel := &adapterContractTunnel{}
			adapter, err := NewAdapterRegistry(tunnel).Adapter(testCase.provider)
			if err != nil {
				t.Fatal(err)
			}
			dep := testCase.dep
			dep.Provider = testCase.provider
			dep.Model = "provider-model"
			result, err := adapter.Submit(context.Background(), "request-1", "tenant-1", dep, platform.VideoCreateRequest{
				Model: "logical-model", Prompt: "test prompt", Seconds: 4, Count: 1, Resolution: "720p",
			}, "")
			if err != nil {
				t.Fatal(err)
			}
			if result.ProviderJobID != "job-1" {
				t.Fatalf("unexpected provider job id %q", result.ProviderJobID)
			}
			if got := tunnel.last.RequestMetadata["provider_endpoint_path"]; got != testCase.path {
				t.Fatalf("unexpected endpoint path %v", got)
			}
			if got := tunnel.last.RequestMetadata["provider_http_method"]; got != "POST" {
				t.Fatalf("unexpected HTTP method %v", got)
			}
			if testCase.provider == "aws-bedrock" {
				var body map[string]interface{}
				if err := json.Unmarshal(tunnel.last.Body, &body); err != nil {
					t.Fatal(err)
				}
				if body["clientRequestToken"] != "request-1" {
					t.Fatalf("missing Bedrock idempotency token: %#v", body)
				}
				if !strings.Contains(string(tunnel.last.Body), "s3://video-output/prefix") {
					t.Fatalf("missing Bedrock S3 output config: %s", tunnel.last.Body)
				}
			}
		})
	}
}

func TestVertexPollingUsesFetchPredictOperationContract(t *testing.T) {
	tunnel := &adapterContractTunnel{}
	adapter, err := NewAdapterRegistry(tunnel).Adapter("google-vertex")
	if err != nil {
		t.Fatal(err)
	}
	operation := "projects/project-1/locations/us-central1/publishers/google/models/veo/operations/op-1"
	_, err = adapter.GetStatus(context.Background(), "request-1", "tenant-1", platform.Deployment{
		Provider: "google-vertex", Model: "veo", ProjectID: "project-1", RegionName: "us-central1",
	}, operation)
	if err != nil {
		t.Fatal(err)
	}
	if got := tunnel.last.RequestMetadata["provider_http_method"]; got != "POST" {
		t.Fatalf("Vertex polling must use POST, got %v", got)
	}
	if !strings.HasSuffix(tunnel.last.RequestMetadata["provider_endpoint_path"].(string), ":fetchPredictOperation") ||
		!strings.Contains(string(tunnel.last.Body), operation) {
		t.Fatalf("unexpected Vertex polling request: path=%v body=%s", tunnel.last.RequestMetadata["provider_endpoint_path"], tunnel.last.Body)
	}
}

func TestParseProviderResultNormalizesCompletionAndUsage(t *testing.T) {
	body, _ := json.Marshal(map[string]interface{}{
		"id": "job-1", "status": "completed", "duration": 4.5, "resolution": "720p", "n": 2,
		"outputs": []interface{}{
			map[string]interface{}{"video_url": "https://provider.example/one.mp4"},
			map[string]interface{}{"video_url": "https://provider.example/two.mp4"},
		},
	})

	result := parseProviderResult(body)
	if result.Status != platform.VideoGenerationSucceeded || result.ProviderJobID != "job-1" {
		t.Fatalf("unexpected normalized result: %#v", result)
	}
	if len(result.Artifacts) != 2 || len(result.ActualUsage) != 1 {
		t.Fatalf("expected two outputs and one usage measure: %#v", result)
	}
	if result.ActualUsage[0].Amount != "9" || result.ActualUsage[0].Resolution != "720p" {
		t.Fatalf("unexpected usage: %#v", result.ActualUsage[0])
	}
}

func TestExplicitOutputSecondsAreAlreadyAggregate(t *testing.T) {
	raw := map[string]interface{}{
		"usage": map[string]interface{}{"output_seconds": 7.25, "output_count": 2.0},
	}
	measures := extractVideoUsage(raw, []platform.VideoArtifact{{ID: "1"}, {ID: "2"}})
	if len(measures) != 1 || measures[0].Amount != "7.25" {
		t.Fatalf("explicit output seconds must not be multiplied: %#v", measures)
	}
}

func TestUsageFromArtifactMetadataGroupsResolutionExactly(t *testing.T) {
	measures := usageFromArtifactMetadata([]platform.VideoArtifact{
		{Duration: "2.25", Resolution: "540p"},
		{Duration: "3.75", Resolution: "540p"},
		{Duration: "4", Resolution: "720p"},
	}, "priority")
	if len(measures) != 2 {
		t.Fatalf("expected two resolution groups: %#v", measures)
	}
	if measures[0].Resolution != "540p" || measures[0].Amount != "6" || measures[0].ServiceTier != "priority" {
		t.Fatalf("unexpected 540p measure: %#v", measures[0])
	}
	if measures[1].Resolution != "720p" || measures[1].Amount != "4" {
		t.Fatalf("unexpected 720p measure: %#v", measures[1])
	}
}

func TestUsageFromArtifactMetadataRequiresEveryDuration(t *testing.T) {
	measures := usageFromArtifactMetadata([]platform.VideoArtifact{
		{Duration: "2", Resolution: "720p"},
		{Resolution: "720p"},
	}, "standard")
	if measures != nil {
		t.Fatalf("partial metadata must not produce synthetic usage: %#v", measures)
	}
}
