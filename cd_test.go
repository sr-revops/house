package brickhouse

import (
	"fmt"
	"testing"
)

func TestCompare(t *testing.T) {
	d1 := map[string]interface{}{
		"1": map[string]interface{}{
			"a": "foo",
			"b": "bar",
			"c": "baz",
		},
		"2": map[string]interface{}{
			"a": "foo",
			"b": "bar",
			"c": "baz",
		},
		"3": map[string]interface{}{
			"a": "foo",
			"b": "bar",
			"c": "baz",
		},
	}
	d2 := map[string]interface{}{
		"1": map[string]interface{}{
			"a": "zoo",
			"b": "bar",
			"c": "baz",
		},
		"3": map[string]interface{}{
			"b": "bar",
			"c": "baz",
			"d": "zoo",
		},
		"4": map[string]interface{}{
			"a": "foo",
			"b": "bar",
			"c": "baz",
		},
	}

	changes := Compare(d1, d2)
	if len(changes) != 5 {
		t.Error("failed")
	} else {
		fmt.Println(changes)
	}
}
