package jsonpath

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_JSONPath_Parse_Path(t *testing.T) {
	out, err := ParsePathString("$")
	assert.NoError(t, err)
	assert.Equal(t, len(out), 0)
}

func Test_JSONPath_Parse_PathLong(t *testing.T) {
	out, err := ParsePathString("$.a.b.c")

	assert.NoError(t, err)

	assert.Equal(t, len(out), 3)

	assert.Equal(t, out[0], "a")
	assert.Equal(t, out[1], "b")
	assert.Equal(t, out[2], "c")
}

func Test_JSONPath_NewPath(t *testing.T) {
	path, err := NewPath("$.a.b.c")

	assert.NoError(t, err)

	assert.Equal(t, len(path.path), 3)

	assert.Equal(t, path.path[0], "a")
	assert.Equal(t, path.path[1], "b")
	assert.Equal(t, path.path[2], "c")
}

type testPathStruct struct {
	Input Path
}

func Test_JSONPath_Parsing(t *testing.T) {
	raw := []byte(`"$.a.b.c"`)

	var pathstr Path
	err := json.Unmarshal(raw, &pathstr)

	assert.NoError(t, err)

	assert.Equal(t, len(pathstr.path), 3)

	assert.Equal(t, pathstr.path[0], "a")
	assert.Equal(t, pathstr.path[1], "b")
	assert.Equal(t, pathstr.path[2], "c")
}
