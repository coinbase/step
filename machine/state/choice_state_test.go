package state

import (
	"testing"

	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

// Comparison Operations
func Test_ChoiceState_AllBasicChoices(t *testing.T) {
	// Primary Purpose of this test is to hit every comparison operator

	state := parseChoiceState([]byte(`{
		"Choices": [
		  {
		    "Variable": "$.valueseq",
		    "Next": "PassStringEquals",
		    "StringEquals": "seq"
		  },
		  {
		    "Variable": "$.valueslt",
		    "Next": "PassStringLessThan",
		    "StringLessThan": "slt"
		  },
		  {
		    "Variable": "$.valuesgt",
		    "Next": "PassStringGreaterThan",
		    "StringGreaterThan": "sgt"
		  },
		  {
		    "Variable": "$.valuesleq",
		    "Next": "PassStringLessThanEquals",
		    "StringLessThanEquals": "sleq"
		  },
		  {
		    "Variable": "$.valuesgeq",
		    "Next": "PassStringGreaterThanEquals",
		    "StringGreaterThanEquals": "sgeq"
		  },
		  {
		    "Variable": "$.valueneq",
		    "Next": "PassNumericEquals",
		    "NumericEquals": 0
		  },
		  {
		    "Variable": "$.valuenlt",
		    "Next": "PassNumericLessThan",
		    "NumericLessThan": 0
		  },
		  {
		    "Variable": "$.valuengt",
		    "Next": "PassNumericGreaterThan",
		    "NumericGreaterThan": 0
		  },
		  {
		    "Variable": "$.valuenleq",
		    "Next": "PassNumericLessThanEquals",
		    "NumericLessThanEquals": 0
		  },
		  {
		    "Variable": "$.valuengeq",
		    "Next": "PassNumericGreaterThanEquals",
		    "NumericGreaterThanEquals": 0
		  },
		  {
		    "Variable": "$.valuebeq",
		    "Next": "PassBooleanEquals",
		    "BooleanEquals": true
		  },
		  {
		    "Variable": "$.valueteq",
		    "Next": "PassTimestampEquals",
		    "TimestampEquals": "2007-01-02T15:04:05Z"
		  },
		  {
		    "Variable": "$.valuetlt",
		    "Next": "PassTimestampLessThan",
		    "TimestampLessThan": "2007-01-02T15:04:05Z"
		  },
		  {
		    "Variable": "$.valuetgt",
		    "Next": "PassTimestampGreaterThan",
		    "TimestampGreaterThan": "2007-01-02T15:04:05Z"
		  },
		  {
		    "Variable": "$.valuetleq",
		    "Next": "PassTimestampLessThanEquals",
		    "TimestampLessThanEquals": "2007-01-02T15:04:05Z"
		  },
		  {
		    "Variable": "$.valuetgeq",
		    "Next": "PassTimestampGreaterThanEquals",
		    "TimestampGreaterThanEquals": "2007-01-02T15:04:05Z"
		  }
		],
		"Default": "Fail"
	}`), t)

	// Default
	testState(state, stateTestData{
		Next: to.Strp("Fail"),
	}, t)

	// Strings
	testState(state, stateTestData{
		Input: map[string]interface{}{"valueseq": "seq"},
		Next:  to.Strp("PassStringEquals"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valueslt": "alt"},
		Next:  to.Strp("PassStringLessThan"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuesgt": "zgt"},
		Next:  to.Strp("PassStringGreaterThan"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuesleq": "aleq"},
		Next:  to.Strp("PassStringLessThanEquals"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuesgeq": "zgeq"},
		Next:  to.Strp("PassStringGreaterThanEquals"),
	}, t)

	// Negative Strings

	testState(state, stateTestData{
		Input: map[string]interface{}{"valueseq": "noop"},
		Next:  to.Strp("Fail"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valueslt": "zlt"},
		Next:  to.Strp("Fail"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuesgt": "agt"},
		Next:  to.Strp("Fail"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuesleq": "zleq"},
		Next:  to.Strp("Fail"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuesgeq": "ageq"},
		Next:  to.Strp("Fail"),
	}, t)

	// Numbers

	testState(state, stateTestData{
		Input: map[string]interface{}{"valueneq": 0},
		Next:  to.Strp("PassNumericEquals"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuenlt": -1},
		Next:  to.Strp("PassNumericLessThan"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuengt": 1},
		Next:  to.Strp("PassNumericGreaterThan"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuenleq": -1},
		Next:  to.Strp("PassNumericLessThanEquals"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuengeq": 1},
		Next:  to.Strp("PassNumericGreaterThanEquals"),
	}, t)

	// Negative Numbers

	testState(state, stateTestData{
		Input: map[string]interface{}{"valueneq": -1},
		Next:  to.Strp("Fail"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuenlt": 1},
		Next:  to.Strp("Fail"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuengt": -1},
		Next:  to.Strp("Fail"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuenleq": 1},
		Next:  to.Strp("Fail"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuengeq": -1},
		Next:  to.Strp("Fail"),
	}, t)

	// Boolean
	testState(state, stateTestData{
		Input: map[string]interface{}{"valuebeq": true},
		Next:  to.Strp("PassBooleanEquals"),
	}, t)

	// Boolean Negative
	testState(state, stateTestData{
		Input: map[string]interface{}{"valuebeq": false},
		Next:  to.Strp("Fail"),
	}, t)

	// Timestamps

	testState(state, stateTestData{
		Input: map[string]interface{}{"valueteq": "2007-01-02T15:04:05Z"},
		Next:  to.Strp("PassTimestampEquals"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuetlt": "2006-01-02T15:04:05Z"},
		Next:  to.Strp("PassTimestampLessThan"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuetgt": "2008-01-02T15:04:05Z"},
		Next:  to.Strp("PassTimestampGreaterThan"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuetleq": "2006-01-02T15:04:05Z"},
		Next:  to.Strp("PassTimestampLessThanEquals"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuetgeq": "2008-01-02T15:04:05Z"},
		Next:  to.Strp("PassTimestampGreaterThanEquals"),
	}, t)

	// Timestamps negative

	testState(state, stateTestData{
		Input: map[string]interface{}{"valueteq": "2006-01-02T15:04:05Z"},
		Next:  to.Strp("Fail"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuetlt": "2008-01-02T15:04:05Z"},
		Next:  to.Strp("Fail"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuetgt": "2006-01-02T15:04:05Z"},
		Next:  to.Strp("Fail"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuetleq": "2008-01-02T15:04:05Z"},
		Next:  to.Strp("Fail"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"valuetgeq": "2006-01-02T15:04:05Z"},
		Next:  to.Strp("Fail"),
	}, t)
}

func Test_ChoiceState_StringEquals(t *testing.T) {
	state := parseChoiceState([]byte(`{
		"Choices": [
			{
				"Variable": "$.value",
				"StringEquals": "public",
				"Next": "Pass"
			}
		],
		"Default": "Fail"
	}`), t)

	testState(state, stateTestData{
		Next: to.Strp("Fail"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"value": "public"},
		Next:  to.Strp("Pass"),
	}, t)
}

// Logical Comparisons

func Test_ChoiceState_Not(t *testing.T) {
	state := parseChoiceState([]byte(`{
		"Choices": [
			{
				"Not": {
					"Variable": "$.value",
					"StringEquals": "public"
				},
				"Next": "Pass"
			}
		],
		"Default": "Fail"
	}`), t)

	// Default
	testState(state, stateTestData{
		Next: to.Strp("Pass"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"value": "nope"},
		Next:  to.Strp("Pass"),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"value": "public"},
		Next:  to.Strp("Fail"),
	}, t)
}

func Test_ChoiceState_And(t *testing.T) {
	state := parseChoiceState([]byte(`{
		"Choices": [
			{
				"And" : [{
					"Variable": "$.value1",
					"StringEquals": "one"
				},{
					"Variable": "$.value2",
					"StringEquals": "two"
				}],

				"Next": "Pass"
			}
		],
		"Default": "Fail"
	}`), t)

	// Default
	testState(state, stateTestData{
		Next: to.Strp("Fail"),
	}, t)

	// both bad
	testState(state, stateTestData{
		Input: map[string]interface{}{"value1": "noop", "value2": "noop"},
		Next:  to.Strp("Fail"),
	}, t)

	// second bad
	testState(state, stateTestData{
		Input: map[string]interface{}{"value1": "one", "value2": "twot"},
		Next:  to.Strp("Fail"),
	}, t)

	// first bad
	testState(state, stateTestData{
		Input: map[string]interface{}{"value1": "onee", "value2": "two"},
		Next:  to.Strp("Fail"),
	}, t)

	// both good
	testState(state, stateTestData{
		Input: map[string]interface{}{"value1": "one", "value2": "two"},
		Next:  to.Strp("Pass"),
	}, t)
}

func Test_ChoiceState_OR(t *testing.T) {
	state := parseChoiceState([]byte(`{
		"Choices": [
			{
				"Or" : [{
					"Variable": "$.value1",
					"StringEquals": "one"
				},{
					"Variable": "$.value2",
					"StringEquals": "two"
				}],

				"Next": "Pass"
			}
		],
		"Default": "Fail"
	}`), t)

	// Default
	testState(state, stateTestData{
		Next: to.Strp("Fail"),
	}, t)

	// both bad
	testState(state, stateTestData{
		Input: map[string]interface{}{"value1": "noop", "value2": "noop"},
		Next:  to.Strp("Fail"),
	}, t)

	// second bad
	testState(state, stateTestData{
		Input: map[string]interface{}{"value1": "one", "value2": "twot"},
		Next:  to.Strp("Pass"),
	}, t)

	// first bad
	testState(state, stateTestData{
		Input: map[string]interface{}{"value1": "onee", "value2": "two"},
		Next:  to.Strp("Pass"),
	}, t)

	// both good
	testState(state, stateTestData{
		Input: map[string]interface{}{"value1": "one", "value2": "two"},
		Next:  to.Strp("Pass"),
	}, t)
}

// Validations

func Test_ChoiceState_NotAllowed2ComparisonOperators(t *testing.T) {
	state := parseChoiceState([]byte(`{"Default": "Fail", "Choices": [
	{
		"StringEquals": "Private",
		"NumericEquals": 0,
		"Next": "Public"
	}
	]}`), t)

	err := state.Validate()

	assert.Error(t, err)
	assert.Regexp(t, "Not Exactly One comparison Operator", err.Error())
}

func Test_ChoiceState_NotAllowed0ComparisonOperators(t *testing.T) {
	state := parseChoiceState([]byte(`{"Default": "Fail", "Choices": [
	{
		"Next": "Public"
	}
	]}`), t)

	err := state.Validate()
	assert.Error(t, err)
	assert.Regexp(t, "Not Exactly One comparison Operator", err.Error())
}

func Test_ChoiceState_NotAllowed2EmbeddedComparisonOperators(t *testing.T) {
	state := parseChoiceState([]byte(`{"Default": "Fail", "Choices": [
	{
		"And": [
			{
				"StringEquals": "Private",
				"NumericEquals": 0
			}
		],
		"Next": "Public"
	}
	]}`), t)

	err := state.Validate()
	assert.Error(t, err)
	assert.Regexp(t, "Not Exactly One comparison Operator", err.Error())
}

func Test_ChoiceState_NotAllowed2DeeplyEmbeddedComparisonOperators(t *testing.T) {
	state := parseChoiceState([]byte(`{"Default": "Fail", "Choices": [
	{
		"And": [
			{
				"Not": {
					"StringEquals": "Private",
					"NumericEquals": 0
				}
			}
		],
		"Next": "Public"
	}
	]}`), t)

	err := state.Validate()
	assert.Error(t, err)
	assert.Regexp(t, "Not Exactly One comparison Operator", err.Error())
}
