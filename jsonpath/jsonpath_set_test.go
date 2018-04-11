package jsonpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_JSONPath_Set_Default(t *testing.T) {
	test := map[string]interface{}{"a": "b"}
	value := map[string]interface{}{"c": "d"}

	path, err := NewPath("$")
	assert.NoError(t, err)

	setted, err := path.Set(test, value)

	assert.NoError(t, err)
	assert.Equal(t, setted, value)
}

func Test_JSONPath_Set_Simple(t *testing.T) {
	test := map[string]interface{}{"a": "b"}

	path, err := NewPath("$.a")
	assert.NoError(t, err)

	setted, err := path.Set(test, "s")
	assert.NoError(t, err)

	out, err := path.Get(setted)
	assert.NoError(t, err)
	assert.Equal(t, "s", out)
}

func Test_JSONPath_Set_Deep(t *testing.T) {
	test := map[string]interface{}{"a": "b"}
	outer := map[string]interface{}{"x": test}

	path, err := NewPath("$.x.a")
	assert.NoError(t, err)

	setted, err := path.Set(outer, "s")
	assert.NoError(t, err)

	out, err := path.Get(setted)
	assert.NoError(t, err)
	assert.Equal(t, "s", out)
}

func Test_JSONPath_Set_Create(t *testing.T) {
	test := map[string]interface{}{}

	path, err := NewPath("$.a")
	assert.NoError(t, err)

	setted, err := path.Set(test, "s")
	assert.NoError(t, err)

	out, err := path.Get(setted)
	assert.NoError(t, err)
	assert.Equal(t, "s", out)
}

func Test_JSONPath_Set_Overwrite(t *testing.T) {
	test := map[string]interface{}{"a": "b"}

	path, err := NewPath("$.a.b")
	assert.NoError(t, err)

	setted, err := path.Set(test, "s")
	assert.NoError(t, err)

	out, err := path.Get(setted)
	assert.NoError(t, err)
	assert.Equal(t, "s", out)
}
