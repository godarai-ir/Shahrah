package main

import (
	"testing"
)

func TestAuditNormalizesRouteParametersAndReportsMissing(t *testing.T) {
	official := []string{
		"POST /responses",
		"GET /responses/{response_id}",
		"POST /responses/{response_id}/cancel?beta=true",
	}
	routes := `
v1.Post("/responses", handler)
v1.Get("/responses/:id", handler)
`

	report := audit(official, routes)
	if report.ContractReady {
		t.Fatal("expected incomplete route matrix")
	}

	if report.ImplementedCount != 2 || len(report.Missing) != 1 ||
		report.Missing[0] != "POST /responses/{}/cancel" {
		t.Fatalf("unexpected report: %+v", report)
	}

	if report.LiveValidated {
		t.Fatal("contract audit must never claim live validation")
	}
}

func TestAuditScopesTheOfficialBaselineToResponsesOperations(t *testing.T) {
	official := []string{
		"POST /embeddings",
		"GET /models",
		"POST /responses",
		"GET /responses/{response_id}",
	}
	routes := `v1.Post("/responses", handler)
v1.Get("/responses/:id", handler)`

	report := audit(official, routes)
	if !report.ContractReady || len(report.Missing) != 0 {
		t.Fatalf("expected contract-ready report: %+v", report)
	}

	if report.OfficialCount != 2 || report.ImplementedCount != 2 {
		t.Fatalf("expected responses-only counts: %+v", report)
	}

	if report.ProductionReady || report.WebSocketReady {
		t.Fatalf("static REST audit must not claim websocket or production readiness: %+v", report)
	}
}
