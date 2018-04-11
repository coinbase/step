package is

import (
	"testing"
	"time"

	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

func Test_UniqueStrp(t *testing.T) {
	assert.True(t, UniqueStrp([]*string{}))
	assert.True(t, UniqueStrp([]*string{to.Strp("asd")}))
	assert.True(t, UniqueStrp([]*string{to.Strp("asd"), to.Strp("asdas")}))

	assert.False(t, UniqueStrp([]*string{nil}))
	assert.False(t, UniqueStrp([]*string{to.Strp("asd"), to.Strp("asd")}))
}

func Test_WithinTimeFrame(t *testing.T) {
	assert.True(t, WithinTimeFrame(to.Timep(time.Now()), 10*time.Second, 10*time.Second))

	assert.False(t, WithinTimeFrame(to.Timep(time.Now().Add(10*time.Minute)), 10*time.Second, 10*time.Second))
	assert.False(t, WithinTimeFrame(to.Timep(time.Now().Add(-10*time.Minute)), 10*time.Second, 10*time.Second))
}
