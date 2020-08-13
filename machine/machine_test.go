package machine

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/coinbase/step/utils/to"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func loadFixture(file string, t *testing.T) *StateMachine {
	example_machine, err := ParseFile(file)
	assert.NoError(t, err)
	return example_machine
}

func execute(json []byte, input interface{}, t *testing.T) (map[string]interface{}, error) {
	example_machine, err := FromJSON(json)
	assert.NoError(t, err)
	example_machine.SetDefaultHandler()

	exec, err := example_machine.Execute(input)

	return exec.Output, err
}

func executeFixture(file string, input map[string]interface{}, t *testing.T) map[string]interface{} {
	example_machine := loadFixture(file, t)

	exec, err := example_machine.Execute(input)

	assert.NoError(t, err)

	return exec.Output
}

//////
// TESTS
//////

func Test_Machine_EmptyStateMachinePassExample(t *testing.T) {
	_, err := execute([]byte(EmptyStateMachine), make(map[string]interface{}), t)
	assert.NoError(t, err)
}

func Test_Machine_SimplePassExample_With_Execute(t *testing.T) {
	json := []byte(`
  {
      "StartAt": "start",
      "States": {
        "start": {
          "Type": "Pass",
          "Result": "b",
          "ResultPath": "$.a",
          "End": true
        }
    }
  }`)

	output, err := execute(json, make(map[string]interface{}), t)
	assert.NoError(t, err)
	assert.Equal(t, output["a"], "b")

	output, err = execute(json, "{}", t)
	assert.NoError(t, err)
	assert.Equal(t, output["a"], "b")

	output, err = execute(json, to.Strp("{}"), t)
	assert.NoError(t, err)
	assert.Equal(t, output["a"], "b")
}

func Test_Machine_ErrorUnknownState(t *testing.T) {
	example_machine := loadFixture("../examples/bad_unknown_state.json", t)
	_, err := example_machine.Execute(make(map[string]interface{}))

	assert.Error(t, err)
	assert.Regexp(t, "Unknown State", err.Error())
}

func Test_Machine_MarshallAllTypes(t *testing.T) {
	file := "../examples/all_types.json"
	sm, err := ParseFile(file)
	assert.NoError(t, err)

	sm.SetDefaultHandler()
	assert.NoError(t, sm.Validate())

	marshalled_json, err := json.Marshal(sm)
	assert.NoError(t, err)

	raw_json, err := ioutil.ReadFile(file)
	assert.NoError(t, err)

	assert.JSONEq(t, string(raw_json), string(marshalled_json))
}

func Test_Machine_Execute_Simple_Chain(t *testing.T) {
	stateMachine, err := FromJSON([]byte(`{
      "Comment": "Calculate Vector Multiplication on Const",
      "StartAt": "AssignCoords",
      "States": {
        "AssignCoords": {
          "Type": "TaskFn",
          "Next": "Multiply"
        },
        "Multiply": {
          "Type": "TaskFn",
          "End": true
        }
      }
    }`))

	assert.Nil(t, err)

	type Vector struct {
		Coords []float64
	}

	stateMachine.SetTaskHandler("AssignCoords", func (context context.Context, input interface{})(Vector, error){
		vector := Vector{}
		spaceSize := input.(map[string]interface{})["Input"].(map[string]interface{})["data"].(map[string]interface{})["spaceSize"].(float64)
		vector.Coords = make([]float64, int(spaceSize))
		for i:=0; i < int(spaceSize); i++ {
            vector.Coords[i] = float64(i + 1)
		}
		return vector, nil
	})

	stateMachine.SetTaskHandler("Multiply", func (context context.Context, input interface{})(Vector, error){
		vector := Vector{}
		const multiplyFactor = 4
		mapstructure.Decode(input.(map[string]interface{})["Input"], &vector)
		for i := 0; i < len(vector.Coords); i++ {
			vector.Coords[i] *= multiplyFactor
		}
		return vector, nil
	})

	testLambda := "test_execute_vector_const_multiply"
	stateMachine.SetResource(&testLambda)

	input := "{\"data\": {\"spaceSize\": 5 } }"
	executionRes, executionErr := stateMachine.Execute(input)

	assert.NotNil(t, executionRes)
	assert.Equal(t,"{\n \"Coords\": [\n  4,\n  8,\n  12,\n  16,\n  20\n ]\n}", executionRes.OutputJSON)
	assert.Nil(t, executionErr)
}

func Test_Machine_Execute_With_Parallel_State(t *testing.T) {
	stateMachine, err := FromJSON([]byte(`{
      "Comment": "Triangle Calculation",
      "StartAt": "StartCalculation",
      "States": {
        "StartCalculation": {
          "Type": "Pass",
          "Next": "CalculateTriangleAngles"
        },
        "CalculateTriangleAngles": {
          "Type": "Parallel",
          "Next": "Summarize",
          "Branches": [
            {
              "StartAt": "CalculateAlpha",
              "States": {
                "CalculateAlpha": {
                  "Type": "TaskFn",
                  "Parameters": {
                     "angle": "alpha"
                   },
                  "Resource": "host1:app1:func1",
                  "End": true
                }
              }
            },
            {
              "StartAt": "CalculateBeta",
              "States": {
                "CalculateBeta": {
                  "Type": "TaskFn",
                  "Resource": "host1:app2:func1",
                  "End": true
                }
              }
            },
            {
              "StartAt": "CalculateGamma",
              "States": {
                "CalculateGamma": {
                  "Type": "TaskFn",
                  "Resource": "host1:app3:func1",
                  "End": true
                }
              }
            }
          ]
        },
        "Summarize": {
            "Type": "TaskFn",
            "End": true
        }
      }
    }`)) //
	type Angle struct {
		name string
		value float64
	}

    type Triangle struct {
    	a float64
    	b float64
    	c float64
    	alpha Angle
    	beta Angle
    	gamma Angle
	}

	assert.Nil(t, err)

	calculateAngle := func(context context.Context, input interface{})(Angle, error){
		angle := Angle{}
		return angle, nil
	}

	branches := stateMachine.States["CalculateTriangleAngles"].(*ParallelState)
	branches.Branches[0].SetTaskHandler("CalculateAlpha", calculateAngle)
	branches.Branches[1].SetTaskHandler("CalculateBeta", calculateAngle)
	branches.Branches[2].SetTaskHandler("CalculateGamma", calculateAngle)

	stateMachine.SetTaskHandler("Summarize", func(context context.Context, input interface{})(Triangle, error){
		triangle := Triangle{}
		return triangle, nil
	})

	testLambda := "test_execute_triangle_angles_calculation"
	stateMachine.SetResource(&testLambda)

	input := "{\"triangle\": {\"a\": 3,  \"b\": 4, \"c\": 5} }"
	executionRes, executionErr := stateMachine.Execute(input)
	assert.Nil(t, executionErr)
	assert.NotNil(t, executionRes)

	/*stateMachine.SetTaskHandler("StartCalculation", func (context context.Context, input interface{})(CalculationData, error){
		data := CalculationData{}
		x := input.(map[string]interface{})["Input"].(map[string]interface{})["data"].(map[string]interface{})["X"].(float64)
		data.value = x
		data.initial = 0
		if x > 0 {
			data.initial = 1
		}
		return data, nil
	})

	branches := stateMachine.States["CalculateVals"].(*ParallelState)
	branches.Branches[0].SetTaskHandler("CalculateOffset", _calculateCoeff)
	branches.Branches[1].SetTaskHandler("CalculateCoeff", _calculateOffset)
	stateMachine.SetTaskHandler("CalculateY", _calculateResult)

	testLambda := "test"
	stateMachine.SetResource(&testLambda)

	input := "{\"data\": {\"X\": 100 } }"
	executionRes, executionErr := stateMachine.Execute(input)
	assert.NotNil(t, executionErr)
	assert.NotNil(t, executionRes)*/
}

/// private types and fucntions
/*
type CalculationData struct{
	value float64
	initial float64
}

type CalculationResult struct {
	value float64
}

func _prepareCalculationData(context context.Context, input interface{})(CalculationData, error){
	data := CalculationData{}
	x := input.(map[string]interface{})["Input"].(map[string]interface{})["data"].(map[string]interface{})["X"].(float64)
	data.value = x
	data.initial = 0
	if x > 0 {
		data.initial = 1
	}
	return data, nil
}

func _calculateCoeff (context context.Context, input interface{})(CalculationResult, error){
	result := CalculationResult {}
	//input["Input"]
	return result, nil
}

func _calculateOffset (context context.Context, input interface{})(CalculationResult, error){
	result := CalculationResult {}
	//input["Input"]
	return result, nil
}

func _calculateResult (context context.Context, input interface{})(CalculationResult, error){
	result := CalculationResult {}
	//input["Input"]
	return result, nil
}*/