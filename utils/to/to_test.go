package to

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_to_ArnPath(t *testing.T) {
	assert.Equal(t, "/", ArnPath(""))
	assert.Equal(t, "/bla/foo/", ArnPath("arn:aws:iam::000000:instance-profile/bla/foo/bar"))
	assert.Equal(t, "/bla/foo/", ArnPath("arn:aws:iam::000000:role/bla/foo/bar"))
	assert.Equal(t, "/", ArnPath("arn:aws:iam::000000:role/bar"))
}

func Test_to_ArnRegionAccountResource(t *testing.T) {
	r, a, res := ArnRegionAccountResource("arn:aws:lambda:<region>:<id>:function:<name>")

	assert.Equal(t, "<region>", r)
	assert.Equal(t, "<id>", a)
	assert.Equal(t, "function:<name>", res)

	r, a, res = ArnRegionAccountResource("arn:aws:iam::000000:instance-profile/bla/foo/bar")
	assert.Equal(t, "", r)
	assert.Equal(t, "000000", a)
	assert.Equal(t, "instance-profile/bla/foo/bar", res)
}
