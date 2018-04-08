package deployer

import (
	"testing"

	"github.com/coinbase/step/aws/mocks"
	"github.com/coinbase/step/machine"
	"github.com/coinbase/step/utils/to"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"
)

func Test_Release_Basic_Fuzz(t *testing.T) {
	for i := 0; i < 20; i++ {
		f := fuzz.New()
		var release Release
		f.Fuzz(&release) // myInt gets a random value.

		assertNoPanic(t, &release)
	}
}

func Test_Release_ValidSM_Fuzz(t *testing.T) {
	for i := 0; i < 20; i++ {
		f := fuzz.New()
		var release Release
		f.Fuzz(&release) // myInt gets a random value.

		release.StateMachineJSON = to.Strp(machine.EmptyStateMachine)
		assertNoPanic(t, &release)
	}
}

func assertNoPanic(t *testing.T, release *Release) {
	state_machine := createTestStateMachine(t, mocks.MockAwsClients())

	_, err := state_machine.ExecuteToMap(release)
	if err != nil {
		assert.NotRegexp(t, "Panic", err.Error())
	}

	le := state_machine.LastOutput()
	assert.NotRegexp(t, "Panic", le)
}
