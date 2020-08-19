package machine

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"strings"
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

	marshalledJson, err := json.Marshal(sm)
	assert.NoError(t, err)

	rawJson, err := ioutil.ReadFile(file)
	assert.NoError(t, err)

	assert.JSONEq(t, string(rawJson), string(marshalledJson))
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
                  "End": true
                }
              }
            },
            {
              "StartAt": "CalculateBeta",
              "States": {
                "CalculateBeta": {
                  "Type": "TaskFn",
                  "End": true
                }
              }
            },
            {
              "StartAt": "CalculateGamma",
              "States": {
                "CalculateGamma": {
                  "Type": "TaskFn",
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
		Name string
		Value float64
	}

    type TriangleSides struct {
    	A float64
    	B float64
    	C float64
	}

	type TriangleAngles struct {
		Alpha Angle
		Beta Angle
		Gamma Angle
	}

	assert.Nil(t, err)

	calculateAngle := func(context context.Context, input interface{})(Angle, error){
		// TODO: UMV: Effective Input from Parameters does not work
		// TODO: UMV: Therefore I have to understand what type of task here by TaskName (CalculateAlpha, ..Beta, ..Gamma)
		taskName := strings.ToLower(input.(map[string]interface{})["Task"].(string))
		sides := TriangleSides{}
		triangleSidesRaw := input.(map[string]interface{})["Input"].(map[string]interface{})["triangle"]
		mapstructure.Decode(triangleSidesRaw, &sides)
		angle := Angle{}
		aSquare := sides.A * sides.A
		bSquare := sides.B * sides.B
		cSquare := sides.C * sides.C
		if strings.Contains(taskName, "alpha"){
            angle.Name = "alpha"
			angle.Value = math.Acos((bSquare + cSquare - aSquare)/(2 * sides.B * sides.C)) * (180 / math.Pi)
		} else if strings.Contains(taskName, "beta") {
			angle.Name = "beta"
			angle.Value = math.Acos((aSquare + cSquare - bSquare)/(2 * sides.A * sides.C)) * (180 / math.Pi)
		} else if strings.Contains(taskName, "gamma") {
			angle.Name = "gamma"
			angle.Value = math.Acos((aSquare + bSquare - cSquare)/(2 * sides.A * sides.B)) * (180 / math.Pi)
		} else {
			return angle, fmt.Errorf("unable to understand what angle to calculate")
		}
		return angle, nil
	}

	testAlphaAngleCalcLambda := "test_execute_lambda_alpha_calc"
	testBetaAngleCalcLambda := "test_execute_lambda_beta_calc"
	testGammaAngleCalcLambda := "test_execute_lambda_gamma_calc"

	branches := stateMachine.States["CalculateTriangleAngles"].(*ParallelState)
	branches.Branches[0].SetTaskHandler("CalculateAlpha", calculateAngle)
	branches.Branches[0].SetResource(&testAlphaAngleCalcLambda)
	branches.Branches[1].SetTaskHandler("CalculateBeta", calculateAngle)
	branches.Branches[1].SetResource(&testBetaAngleCalcLambda)
	branches.Branches[2].SetTaskHandler("CalculateGamma", calculateAngle)
	branches.Branches[2].SetResource(&testGammaAngleCalcLambda)

	stateMachine.SetTaskHandler("Summarize", func(context context.Context, input interface{})(TriangleAngles, error){
		triangle := TriangleAngles{}
		// todo: umv: check that orders the same as expected
		mapstructure.Decode(input.(map[string]interface{})["Input"].([]interface{})[0], &triangle.Alpha)
		mapstructure.Decode(input.(map[string]interface{})["Input"].([]interface{})[1], &triangle.Beta)
		mapstructure.Decode(input.(map[string]interface{})["Input"].([]interface{})[2], &triangle.Gamma)
		return triangle, nil
	})

	testLambda := "test_execute_triangle_angles_calculation"
	stateMachine.SetResource(&testLambda)

	input := "{\"triangle\": {\"a\": 3,  \"b\": 4, \"c\": 5} }"
	executionRes, executionErr := stateMachine.Execute(input)
	assert.Nil(t, executionErr)
	assert.NotNil(t, executionRes)
	assert.Equal(t, "{\n \"Alpha\": {\n  \"Name\": \"alpha\",\n  \"Value\": 36.86989764584401\n },\n " +
		                        "\"Beta\": {\n  \"Name\": \"beta\",\n  \"Value\": 53.13010235415599\n },\n " +
		                        "\"Gamma\": {\n  \"Name\": \"gamma\",\n  \"Value\": 90\n }\n}",
		         executionRes.OutputJSON)

}

func Test_Machine_Execute_With_Map_State(t *testing.T){
	stateMachine, err := FromJSON([]byte(`{
      "Comment": "Calculate sum with subsequent category selection",
      "StartAt": "ProcessGrades",
      "States": {
        "ProcessGrades": {
          "Type": "Map",
          "Next": "Avg",
          "ItemsPath": "$.marks",
          "Iterator": {
            "StartAt": "Scale",
            "States": {
              "Scale": {
                "Type": "TaskFn",
                "End": true
              }
            }
          },
          "ResultPath": "$.allMarks"
        },
        "Avg": {
          "Type": "TaskFn",
          "Next": "SelectGrade"
        },
        "SelectGrade": {
          "Type": "Choice",
          "Choices": [
            {
              "Variable": "$.Avg",
              "NumericLessThan" : 60,
              "Next": "CGrade"
            },
            {
              "And": [
                {
                  "Variable": "$.Avg",
                  "NumericGreaterThanEquals" : 60
                },
                {
                  "Variable": "$.Avg",
                  "NumericLessThan" : 80
                }
              ],
              "Next": "BGrade"
            },
            {
              "Variable": "$.Avg",
              "NumericGreaterThanEquals" : 80,
              "Next": "AGrade"
            }
          ]
        },
        "AGrade": {
          "Type": "TaskFn",
          "End": true
        },
        "BGrade": {
          "Type": "TaskFn",
          "End": true
        },
        "CGrade": {
          "Type": "TaskFn",
          "End": true
        }
      }
    }`))

	type StudentMark struct {
		Subject string
		Mark float64
	}

	type AvgMark struct {
		Avg float64
	}

	type StudentGrade struct {
		Grade string
	}

	gradeStateMachineResource := "text_execute_machine_student_grade_define"
	stateMachine.SetResource(&gradeStateMachineResource)

	scaleStateMachine := stateMachine.States["ProcessGrades"].(*MapState).Iterator
	scaleResource := "test_state_machine_scale_grades"
	scaleStateMachine.SetResource(&scaleResource)

	scaleStateMachine.SetTaskHandler("Scale", func(context context.Context, input interface{})(StudentMark, error){
		mark :=StudentMark{}
		mapstructure.Decode(input.(map[string]interface{})["Input"], &mark)
        mark.Mark *= 100 / 5
		return mark, nil
	})

	stateMachine.SetTaskHandler("Avg", func(context context.Context, input interface{})(AvgMark, error){
		var marks []StudentMark
		rawMarks := input.(map[string]interface{})["Input"].(map[string]interface{})["allMarks"].([]interface{})
		mapstructure.Decode(rawMarks, &marks)
		avg := AvgMark{}
		var total float64
		total = 0
		for _, m := range marks{
			total += m.Mark
		}
		avg.Avg = total / float64(len(marks))
		return avg, nil
	})

	stateMachine.SetTaskHandler("AGrade", func(context context.Context, input interface{})(StudentGrade, error){
		grade := StudentGrade{}
		grade.Grade = "Very good student"
		return grade, nil
	})

	stateMachine.SetTaskHandler("BGrade", func(context context.Context, input interface{})(StudentGrade, error){
		grade := StudentGrade{}
		grade.Grade = "Good student"
		return grade, nil
	})

	stateMachine.SetTaskHandler("CGrade", func(context context.Context, input interface{})(StudentGrade, error){
		grade := StudentGrade{}
		grade.Grade = "So-so student"
		return grade, nil
	})

	assert.Nil(t, err)
	err = stateMachine.Validate()
	assert.Nil(t, err)
	input := "{\"marks\": [ {\"subject\": \"math\", \"mark\": 4},  " +
		                   "{\"subject\": \"physics\", \"mark\": 5}, " +
		                   "{\"subject\": \"chemistry\", \"mark\": 5} ] }"
	executionRes, executionErr := stateMachine.Execute(input)
	assert.Nil(t, executionErr)
	assert.NotNil(t, executionRes)
	assert.Equal(t, "{\n \"Grade\": \"Very good student\"\n}", executionRes.OutputJSON)
}