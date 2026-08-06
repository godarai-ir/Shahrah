package billing

// EntitlementSource identifies the precedence tier of an entitlement value.
type EntitlementSource uint8

const (
	EntitlementSourceDefault EntitlementSource = iota
	EntitlementSourcePlan
	EntitlementSourceOverride
)

// EntitlementValue is an unresolved entitlement value loaded from persistence.
type EntitlementValue struct {
	Key    string
	Value  interface{}
	Source EntitlementSource
}

// ResolveEntitlements applies domain precedence independently of persistence.
func ResolveEntitlements(values []EntitlementValue) (map[string]interface{}, map[string]interface{}) {
	features := make(map[string]interface{}, len(values))
	overrides := make(map[string]interface{})
	precedence := make(map[string]EntitlementSource, len(values))

	for _, value := range values {
		current, exists := precedence[value.Key]
		if exists && current > value.Source {
			continue
		}

		features[value.Key] = value.Value
		precedence[value.Key] = value.Source

		if value.Source == EntitlementSourceOverride {
			overrides[value.Key] = value.Value
		} else {
			delete(overrides, value.Key)
		}
	}

	return features, overrides
}
