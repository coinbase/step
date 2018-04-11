package mocks

import (
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/sfn/sfniface"
	"github.com/coinbase/step/utils/to"
)

type MockSFNClient struct {
	sfniface.SFNAPI
	UpdateStateMachineResp   *sfn.UpdateStateMachineOutput
	UpdateStateMachineError  error
	StartExecutionResp       *sfn.StartExecutionOutput
	DescribeExecutionResp    *sfn.DescribeExecutionOutput
	GetExecutionHistoryResp  *sfn.GetExecutionHistoryOutput
	DescribeStateMachineResp *sfn.DescribeStateMachineOutput
}

func (m *MockSFNClient) init() {

	if m.UpdateStateMachineResp == nil {
		m.UpdateStateMachineResp = &sfn.UpdateStateMachineOutput{}
	}

	if m.StartExecutionResp == nil {
		m.StartExecutionResp = &sfn.StartExecutionOutput{}
	}

	if m.DescribeExecutionResp == nil {
		m.DescribeExecutionResp = &sfn.DescribeExecutionOutput{Status: to.Strp("SUCCEEDED")}
	}

	if m.GetExecutionHistoryResp == nil {
		m.GetExecutionHistoryResp = &sfn.GetExecutionHistoryOutput{Events: []*sfn.HistoryEvent{}}
	}
}

func (m *MockSFNClient) UpdateStateMachine(in *sfn.UpdateStateMachineInput) (*sfn.UpdateStateMachineOutput, error) {
	m.init()
	return m.UpdateStateMachineResp, m.UpdateStateMachineError
}

func (m *MockSFNClient) StartExecution(in *sfn.StartExecutionInput) (*sfn.StartExecutionOutput, error) {
	m.init()
	return m.StartExecutionResp, nil
}

func (m *MockSFNClient) DescribeExecution(in *sfn.DescribeExecutionInput) (*sfn.DescribeExecutionOutput, error) {
	m.init()
	return m.DescribeExecutionResp, nil
}

func (m *MockSFNClient) GetExecutionHistory(in *sfn.GetExecutionHistoryInput) (*sfn.GetExecutionHistoryOutput, error) {
	m.init()
	return m.GetExecutionHistoryResp, nil
}

func (m *MockSFNClient) DescribeStateMachine(in *sfn.DescribeStateMachineInput) (*sfn.DescribeStateMachineOutput, error) {
	m.init()
	return m.DescribeStateMachineResp, nil
}
