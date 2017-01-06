package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	cf "github.com/aws/aws-sdk-go/service/cloudformation"
	cfi "github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

type createCSFn func(*cf.CreateChangeSetInput) (*cf.CreateChangeSetOutput, error)
type deleteCSFn func(*cf.DeleteChangeSetInput) (*cf.DeleteChangeSetOutput, error)
type descCSFn func(*cf.DescribeChangeSetInput) (*cf.DescribeChangeSetOutput, error)

type mockCSClient struct {
	cfi.CloudFormationAPI

	createCS createCSFn
	deleteCS deleteCSFn
	descCS   descCSFn
	csId     string
	csName   string
}

func (m *mockCSClient) CreateChangeSet(input *cf.CreateChangeSetInput) (*cf.CreateChangeSetOutput, error) {
	r, err := m.createCS(input)
	if r != nil {
		if *r.Id != "" && (err == nil) {
			// set change set name to indicate it was created
			m.csId = *r.Id
		}
	}
	return r, err
}

func (m *mockCSClient) DescribeChangeSet(input *cf.DescribeChangeSetInput) (*cf.DescribeChangeSetOutput, error) {
	return m.descCS(input)
}

func (m *mockCSClient) DeleteChangeSet(input *cf.DeleteChangeSetInput) (*cf.DeleteChangeSetOutput, error) {
	r, err := m.deleteCS(input)
	if err == nil && (m.csName == *input.ChangeSetName) {
		// reset change set name to indicate it was deleted
		m.csId = ""
	}
	return r, err
}

// Test failure of the AWS API call to describe the change set
func TestDescCSRetrieveStatusFailed(t *testing.T) {
	expErr := errors.New("change-set describe api failed")
	f := func(i *cf.DescribeChangeSetInput) (*cf.DescribeChangeSetOutput, error) {
		return nil, expErr
	}
	mock := &mockCSClient{descCS: f}
	r, err := describeAvailableChangeSet(mock, &cf.DescribeChangeSetInput{})
	if r != nil {
		t.Errorf("Unexpected non nil return value, ", r)
	}
	if err.Error() != expErr.Error() {
		t.Errorf("Expected (%s), Found (%s)", expErr, err)
	}
}

// Test failure when the creation of the change set to determine
// execution plan times out
func TestDescCSCreateCompleteTimeout(t *testing.T) {
	o := &cf.DescribeChangeSetOutput{
		Status:       aws.String(cf.ChangeSetStatusCreatePending),
		StatusReason: aws.String("create change set pending"),
	}
	expErr := errors.New("change set creation timed out")
	f := func(i *cf.DescribeChangeSetInput) (*cf.DescribeChangeSetOutput, error) {
		return o, nil
	}
	mock := &mockCSClient{descCS: f}
	_, err := describeAvailableChangeSet(mock, &cf.DescribeChangeSetInput{})
	if err.Error() != expErr.Error() {
		t.Errorf("Expected (%s), Found (%s)", expErr, err)
	}
}

// Test failure when a valid change set cannot be created
func TestDescCSCreateFailed(t *testing.T) {
	o := &cf.DescribeChangeSetOutput{
		Status:       aws.String(cf.ChangeSetStatusFailed),
		StatusReason: aws.String("create change set failed"),
	}
	expErr := fmt.Errorf("cannot create change set: %s", *o.StatusReason)
	f := func(i *cf.DescribeChangeSetInput) (*cf.DescribeChangeSetOutput, error) {
		return o, nil
	}
	mock := &mockCSClient{descCS: f}
	_, err := describeAvailableChangeSet(mock, &cf.DescribeChangeSetInput{})
	if err.Error() != expErr.Error() {
		t.Errorf("Expected (%s), Found (%s)", expErr, err)
	}
}

// Test successful creation of a change set from which the
// execution plan can be determined
func TestDescCSCreateComplete(t *testing.T) {
	o := &cf.DescribeChangeSetOutput{
		Status:       aws.String(cf.ChangeSetStatusCreateComplete),
		StatusReason: aws.String("change set created"),
	}
	f := func(i *cf.DescribeChangeSetInput) (*cf.DescribeChangeSetOutput, error) {
		return o, nil
	}
	mock := &mockCSClient{descCS: f}
	_, err := describeAvailableChangeSet(mock, &cf.DescribeChangeSetInput{})
	if err != nil {
		t.Errorf("Unexpected error (%s)", err)
	}
}

// Test plan command failure when the AWS API call to create the
// change set fails
func TestPlanCSCreateAPIFailed(t *testing.T) {
	expErr := errors.New("cannot create change set")
	createFn := func(i *cf.CreateChangeSetInput) (*cf.CreateChangeSetOutput, error) {
		return nil, expErr
	}

	mock := &mockCSClient{createCS: createFn}

	err := plan(mock, "", "", "", "test", true)
	if err.Error() != expErr.Error() {
		t.Errorf("Expected (%s), Found (%s)", expErr.Error(), err.Error())
	}
}

// Test retention of the change set when the change set cannot be created
// to determine the execution plan
func TestPlanCSDescFailedKeepCS(t *testing.T) {
	csName := "testcs"
	createOut := &cf.CreateChangeSetOutput{
		Id: aws.String("testcs-id"),
	}
	createFn := func(i *cf.CreateChangeSetInput) (*cf.CreateChangeSetOutput, error) {
		return createOut, nil
	}

	expErr := errors.New("change-set describe api failed")
	descFn := func(i *cf.DescribeChangeSetInput) (*cf.DescribeChangeSetOutput, error) {
		return nil, expErr
	}

	mock := &mockCSClient{createCS: createFn, descCS: descFn, csName: csName}

	err := plan(mock, "", "", "", csName, true)
	if err.Error() != expErr.Error() {
		t.Errorf("Expected (%s), Found (%s)", expErr.Error(), err.Error())
	}
	if mock.csId == "" {
		t.Errorf("Unexpected deletion of change set")
	}
}

// Test deletion of the change set when the change set cannot be created
// to determine the execution plan
func TestPlanCSDescFailedDeleteCS(t *testing.T) {
	csName := "testcs"
	createOut := &cf.CreateChangeSetOutput{
		Id: aws.String("testcs-id"),
	}
	createFn := func(i *cf.CreateChangeSetInput) (*cf.CreateChangeSetOutput, error) {
		return createOut, nil
	}

	expErr := errors.New("change-set describe api failed")
	descFn := func(i *cf.DescribeChangeSetInput) (*cf.DescribeChangeSetOutput, error) {
		return nil, expErr
	}
	deleteFn := func(i *cf.DeleteChangeSetInput) (*cf.DeleteChangeSetOutput, error) {
		return nil, nil
	}

	mock := &mockCSClient{createCS: createFn, deleteCS: deleteFn, descCS: descFn, csName: csName}

	err := plan(mock, "", "", "", csName, false)
	if err.Error() != expErr.Error() {
		t.Errorf("Expected (%s), Found (%s)", expErr.Error(), err.Error())
	}
	if mock.csId != "" {
		t.Errorf("Change set not deleted")
	}
}

// Test retention of the change set after a successful plan cmd execution
func TestPlanKeepCS(t *testing.T) {
	csName := "testcs"
	createOut := &cf.CreateChangeSetOutput{
		Id: aws.String("testcs-id"),
	}
	descOut := &cf.DescribeChangeSetOutput{
		Status:       aws.String(cf.ChangeSetStatusCreateComplete),
		StatusReason: aws.String("change set created"),
	}

	createFn := func(i *cf.CreateChangeSetInput) (*cf.CreateChangeSetOutput, error) {
		return createOut, nil
	}
	descFn := func(i *cf.DescribeChangeSetInput) (*cf.DescribeChangeSetOutput, error) {
		return descOut, nil
	}

	mock := &mockCSClient{createCS: createFn, descCS: descFn, csName: csName}

	err := plan(mock, "", "", "", csName, true)
	if err != nil {
		t.Errorf("Unexpected error (%s)", err)
	}
	if mock.csId == "" {
		t.Errorf("Unexpected deletion of change set")
	}
}

// Test deletion of the change set after a successful plan cmd execution
func TestPlanDeleteCS(t *testing.T) {
	csName := "testcs"
	createOut := &cf.CreateChangeSetOutput{
		Id: aws.String("testcs-id"),
	}
	descOut := &cf.DescribeChangeSetOutput{
		Status:       aws.String(cf.ChangeSetStatusCreateComplete),
		StatusReason: aws.String("change set created"),
	}

	createFn := func(i *cf.CreateChangeSetInput) (*cf.CreateChangeSetOutput, error) {
		return createOut, nil
	}
	descFn := func(i *cf.DescribeChangeSetInput) (*cf.DescribeChangeSetOutput, error) {
		return descOut, nil
	}
	deleteFn := func(i *cf.DeleteChangeSetInput) (*cf.DeleteChangeSetOutput, error) {
		return nil, nil
	}

	mock := &mockCSClient{createCS: createFn, deleteCS: deleteFn, descCS: descFn, csName: csName}

	err := plan(mock, "", "", "", csName, false)
	if err != nil {
		t.Errorf("Unexpected error (%s)", err)
	}
	if mock.csId != "" {
		t.Errorf("Change set not deleted")
	}
}
