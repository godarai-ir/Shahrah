// Package metadata provides typed access to schema-dynamic metadata at adapter boundaries.
package metadata

import (
	"encoding/json"
	"strconv"
)

// Key names a documented field in a dynamic metadata object.
type Key string

// Object wraps schema-dynamic metadata and centralizes safe type conversion.
type Object struct {
	values map[string]interface{}
}

// Wrap creates a typed view over values. Mutations through Set update the supplied map.
func Wrap(values map[string]interface{}) Object {
	if values == nil {
		values = make(map[string]interface{})
	}

	return Object{values: values}
}

// Get returns a raw value for boundary code that cannot use a narrower accessor.
func (o Object) Get(key Key) (interface{}, bool) {
	value, ok := o.values[string(key)]

	return value, ok
}

// String returns a string value.
func (o Object) String(key Key) (string, bool) {
	value, ok := o.Get(key)
	if !ok {
		return "", false
	}

	text, ok := value.(string)

	return text, ok
}

// Bool returns a boolean value.
func (o Object) Bool(key Key) (bool, bool) {
	value, ok := o.Get(key)
	if !ok {
		return false, false
	}

	boolean, ok := value.(bool)

	return boolean, ok
}

// Int returns an integer from the common JSON and Go numeric representations.
func (o Object) Int(key Key) (int, bool) {
	value, ok := o.Get(key)
	if !ok {
		return 0, false
	}

	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	case json.Number:
		parsed, err := strconv.ParseInt(typed.String(), 10, 64)
		if err != nil {
			return 0, false
		}

		return int(parsed), true
	default:
		return 0, false
	}
}

// Set assigns a value to a documented metadata key.
func (o Object) Set(key Key, value interface{}) {
	o.values[string(key)] = value
}

// Values returns the wrapped map for serialization at a boundary.
func (o Object) Values() map[string]interface{} {
	return o.values
}

// CloneValues returns an independent shallow copy for result DTOs.
func (o Object) CloneValues() map[string]interface{} {
	cloned := make(map[string]interface{}, len(o.values))
	for key, value := range o.values {
		cloned[key] = value
	}

	return cloned
}
