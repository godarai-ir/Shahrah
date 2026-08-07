package metadata

import (
	"encoding/json"
	"testing"
)

func TestObjectTypedAccessors(t *testing.T) {
	object := Wrap(map[string]interface{}{
		"name":    "image",
		"enabled": true,
		"count":   json.Number("3"),
	})

	name, nameOK := object.String(Key("name"))
	enabled, enabledOK := object.Bool(Key("enabled"))
	count, countOK := object.Int(Key("count"))

	if !nameOK || name != "image" || !enabledOK || !enabled || !countOK || count != 3 {
		t.Fatalf("unexpected typed values: name=%q enabled=%t count=%d", name, enabled, count)
	}
}

func TestObjectCloneValuesDoesNotMutateSource(t *testing.T) {
	object := Wrap(map[string]interface{}{"name": "original"})
	cloned := object.CloneValues()
	cloned["name"] = "changed"

	name, ok := object.String(Key("name"))
	if !ok || name != "original" {
		t.Fatalf("source name = %q, want original", name)
	}
}
