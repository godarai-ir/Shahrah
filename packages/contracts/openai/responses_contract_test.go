package openai

import (
	"testing"
)

func TestResponsesContractInventoryIsCompleteAndUnique(t *testing.T) {
	t.Parallel()

	assertUniqueStrings(t, "top-level fields", ResponsesTopLevelFields())
	assertUniqueStrings(t, "streaming events", ResponsesStreamingEventTypes())
	assertUniqueStrings(t, "statuses", ResponsesStatuses())
	assertUniqueStrings(t, "terminal streaming events", ResponsesTerminalStreamingEventTypes())
	assertUniqueStrings(t, "compacted response fields", ResponsesCompactedFields())
	assertUniqueStrings(t, "input token count fields", ResponsesInputTokenCountFields())
	assertUniqueStrings(t, "input items page fields", ResponsesInputItemsPageFields())
	assertUniqueStrings(t, "operations", ResponsesOperations())

	assertLength(t, "top-level fields", ResponsesTopLevelFields(), 35)
	assertLength(t, "top-level field contracts", ResponsesFieldContracts(), 35)
	assertLength(t, "output variants", ResponsesOutputVariants(), 28)
	assertLength(t, "streaming events", ResponsesStreamingEventTypes(), 53)
	assertLength(t, "statuses", ResponsesStatuses(), 6)
	assertLength(t, "terminal streaming events", ResponsesTerminalStreamingEventTypes(), 4)
	assertLength(t, "response item variants", ResponsesItemVariants(), 29)
	assertLength(t, "compacted response fields", ResponsesCompactedFields(), 5)
	assertLength(t, "input token count fields", ResponsesInputTokenCountFields(), 2)
	assertLength(t, "input items page fields", ResponsesInputItemsPageFields(), 5)
	assertLength(t, "operations", ResponsesOperations(), 7)

	var (
		variantNames    = make([]string, 0, len(ResponsesOutputVariants()))
		variantSDKNames = make([]string, 0, len(ResponsesOutputVariants()))
		variantTypes    = make([]string, 0, len(ResponsesOutputVariants()))
	)

	for _, variant := range ResponsesOutputVariants() {
		variantNames = append(variantNames, variant.Name)
		variantSDKNames = append(variantSDKNames, variant.SDKName)
		variantTypes = append(variantTypes, variant.Type)
	}

	assertUniqueStrings(t, "output variant names", variantNames)
	assertUniqueStrings(t, "output variant SDK names", variantSDKNames)
	assertUniqueStrings(t, "output variant types", variantTypes)

	fieldContractNames := make([]string, 0, len(ResponsesFieldContracts()))
	for _, field := range ResponsesFieldContracts() {
		fieldContractNames = append(fieldContractNames, field.Name)

		if field.TypeExpression == "" {
			t.Errorf("field %q has no type expression", field.Name)
		}
	}

	assertUniqueStrings(t, "top-level field contract names", fieldContractNames)
	assertSlicesEqual(t, ResponsesTopLevelFields(), fieldContractNames)

	itemVariantNames := make([]string, 0, len(ResponsesItemVariants()))
	for _, variant := range ResponsesItemVariants() {
		itemVariantNames = append(itemVariantNames, variant.Name)
	}

	assertUniqueStrings(t, "response item variant names", itemVariantNames)
}

func TestResponsesTerminalEventsBelongToStreamingInventory(t *testing.T) {
	t.Parallel()

	all := make(map[string]struct{}, len(ResponsesStreamingEventTypes()))
	for _, eventType := range ResponsesStreamingEventTypes() {
		all[eventType] = struct{}{}
	}

	for _, terminalType := range ResponsesTerminalStreamingEventTypes() {
		_, ok := all[terminalType]
		if !ok {
			t.Errorf("terminal event %q is absent from the complete event inventory", terminalType)
		}
	}
}

func TestResponsesContractAccessorsReturnDefensiveCopies(t *testing.T) {
	t.Parallel()

	fields := ResponsesTopLevelFields()
	fields[0] = "mutated"

	if ResponsesTopLevelFields()[0] != "id" {
		t.Fatal("top-level field accessor returned shared mutable state")
	}

	variants := ResponsesOutputVariants()
	variants[0].Type = "mutated"

	if ResponsesOutputVariants()[0].Type != "message" {
		t.Fatal("output variant accessor returned shared mutable state")
	}

	events := ResponsesStreamingEventTypes()
	events[0] = "mutated"

	if ResponsesStreamingEventTypes()[0] != "response.created" {
		t.Fatal("streaming event accessor returned shared mutable state")
	}

	operations := ResponsesOperations()
	operations[0] = "mutated"

	if ResponsesOperations()[0] != "POST /responses" {
		t.Fatal("operation accessor returned shared mutable state")
	}
}

func assertUniqueStrings(t *testing.T, label string, values []string) {
	t.Helper()

	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			t.Fatalf("%s contains an empty value", label)
		}

		_, exists := seen[value]
		if exists {
			t.Errorf("%s contains duplicate %q", label, value)
		}

		seen[value] = struct{}{}
	}
}

func assertLength[T any](t *testing.T, label string, values []T, expected int) {
	t.Helper()

	if len(values) != expected {
		t.Fatalf("expected %d %s, got %d", expected, label, len(values))
	}
}

func assertSlicesEqual(t *testing.T, expected, actual []string) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}

	for index := range expected {
		if expected[index] != actual[index] {
			t.Fatalf("expected %v, got %v", expected, actual)
		}
	}
}
