package jsonpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_JSONPath_NotFound(t *testing.T) {
	test := map[string]interface{}{}

	path, err := NewPath("$.a")
	assert.NoError(t, err)

	_, err = path.Get(test)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "Not Found")
}

func Test_JSONPath_Get_Default(t *testing.T) {
	test := map[string]interface{}{"a": "b"}

	path, err := NewPath("$")
	assert.NoError(t, err)

	out, err := path.Get(test)
	assert.NoError(t, err)
	assert.Equal(t, out, test)
}

func Test_JSONPath_Get_Simple(t *testing.T) {
	test := map[string]interface{}{"a": "b"}

	path, err := NewPath("$.a")
	assert.NoError(t, err)

	out, err := path.Get(test)
	assert.NoError(t, err)
	assert.Equal(t, out, "b")
}

func Test_JSONPath_Get_Deep(t *testing.T) {
	test := map[string]interface{}{"a": "b"}
	outer := map[string]interface{}{"x": test}

	path, err := NewPath("$.x.a")
	assert.NoError(t, err)

	out, err := path.Get(outer)
	assert.NoError(t, err)
	assert.Equal(t, out, "b")
}

func Test_JSONPath_GetMap(t *testing.T) {
	test := map[string]interface{}{"a": "b"}
	outer := map[string]interface{}{"x": test}

	path, err := NewPath("$.x")
	assert.NoError(t, err)

	out, err := path.GetMap(outer)
	assert.NoError(t, err)
	assert.Equal(t, out, test)
}

func Test_JSONPath_GetMap_Error(t *testing.T) {
	test := map[string]interface{}{"a": "b"}
	outer := map[string]interface{}{"x": test}

	path, err := NewPath("$.x.a")
	assert.NoError(t, err)

	_, err = path.GetMap(outer)
	assert.Equal(t, err.Error(), "GetMap Error: must return map")
}

func Test_JSONPath_GetTime(t *testing.T) {
	test := "2006-01-02T15:04:05Z"
	outer := map[string]interface{}{"x": test}

	path, err := NewPath("$.x")
	assert.NoError(t, err)

	out, err := path.GetTime(outer)
	assert.NoError(t, err)
	assert.Equal(t, out.Year(), 2006)
}

func Test_JSONPath_GetBool(t *testing.T) {
	test := true
	outer := map[string]interface{}{"x": test}

	path, err := NewPath("$.x")
	assert.NoError(t, err)

	out, err := path.GetBool(outer)
	assert.NoError(t, err)
	assert.Equal(t, *out, test)
}

func Test_JSONPath_GetNumber(t *testing.T) {
	test := 1.2
	outer := map[string]interface{}{"x": test}

	path, err := NewPath("$.x")
	assert.NoError(t, err)

	out, err := path.GetNumber(outer)
	assert.NoError(t, err)
	assert.Equal(t, *out, test)
}

func Test_JSONPath_GetString(t *testing.T) {
	test := "String"
	outer := map[string]interface{}{"x": test}

	path, err := NewPath("$.x")
	assert.NoError(t, err)

	out, err := path.GetString(outer)
	assert.NoError(t, err)
	assert.Equal(t, *out, test)
}

func Test_JSONPath_GetSplice(t *testing.T) {
	test := []interface{}{1,2,3}
	outer := map[string]interface{}{"x": test}

	path, err := NewPath("$.x")
	assert.NoError(t, err)

	out, err := path.GetSlice(outer)
	assert.NoError(t, err)
	assert.Equal(t, out, test)

}
