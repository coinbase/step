package to

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CompactJSONStr(t *testing.T) {
	assert.Equal(t, `{}`, CompactJSONStr(Strp("{}")))
	assert.Equal(t, `{"a":"b"}`, CompactJSONStr(Strp("{\n  \"a\": \"b\"\n}")))
}

func Test_PrettyJSONStr(t *testing.T) {
	assert.Equal(t, `{}`, PrettyJSONStr(Strp("{}")))
	assert.Equal(t, "{\n \"a\": \"b\"\n}", PrettyJSONStr(Strp(`{"a":"b"}`)))
}

func Test_AByte(t *testing.T) {
	raw, err := AByte(nil)
	assert.NoError(t, err)
	assert.Equal(t, raw, []byte(""))

	var str *string
	raw, err = AByte(str)
	assert.NoError(t, err)
	assert.Equal(t, raw, []byte(""))

	raw, err = AByte(Strp("asd"))
	assert.NoError(t, err)
	assert.Equal(t, raw, []byte("asd"))

	raw, err = AByte("asd")
	assert.NoError(t, err)
	assert.Equal(t, raw, []byte("asd"))

	raw, err = AByte(struct{ Name string }{"asd"})
	assert.NoError(t, err)
	assert.Equal(t, raw, []byte(`{"Name":"asd"}`))
}
