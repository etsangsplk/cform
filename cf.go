package cform

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	cf "github.com/aws/aws-sdk-go/service/cloudformation"
	cfi "github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

// String format used to print the stack event status
const APPLY_STATUS_FMT = "%-25s\t%-20s\t%-30s\t%-20s\t%s\n"

// PageStackEvent represents a stack events in a specific page in all the
// pages returned by `DescribeStackEventsPages`.
type PageStackEvent struct {
	// Page number index in which the event was found
	pageNum int
	// Timestamp of the most recent event found in the previous page
	lastPageLatestEventTs time.Time
	// The CloudFormation stack event
	event *cf.StackEvent
}

func (e PageStackEvent) String() string {
	event := e.event
	return fmt.Sprintf(APPLY_STATUS_FMT, event.Timestamp.Format("2006-01-02 15:04:05 -0700 MST"),
		*event.ResourceStatus, *event.ResourceType, *event.LogicalResourceId, DerefString(event.ResourceStatusReason, ""))
}

// GetStackEventsAfterTime returns all the events that happened after the input
// time. This pages through stack event pages and stops when the last event
// which happened after the input time is found.
//
// The events are returned in reverse chronological order.
func GetStackEventsAfterTime(svc cfi.CloudFormationAPI, stackName string, ts time.Time) ([]PageStackEvent, error) {
	var events []PageStackEvent
	var newestEvent *cf.StackEvent

	pageNum := 0

	fn := func(page *cf.DescribeStackEventsOutput, lastPage bool) bool {
		seen := false
		for _, event := range page.StackEvents {
			if event.Timestamp.After(ts) {
				events = append(events, PageStackEvent{pageNum, ts, event})
			} else {
				seen = true
				break
			}
		}

		if pageNum == 0 {
			newestEvent = page.StackEvents[0]
		}
		pageNum++

		shouldContinue := !lastPage && !seen
		if !shouldContinue {
			ts = *newestEvent.Timestamp
		}
		return shouldContinue
	}

	p1 := &cf.DescribeStackEventsInput{StackName: aws.String(stackName)}
	if err := svc.DescribeStackEventsPages(p1, fn); err != nil {
		return nil, fmt.Errorf("Failed to fetch stack event pages: %s", err.Error())
	}
	return events, nil
}

// PrintStackEventsDuringOperation prints the most recent events happening
// during an stack operation after the input time.
// The events are writer using the input writer and stops when the stack
// operation is complete.
//
// TODO Add support for timeouts if the stack operation never completes
// TODO Handle scenario when the stack is already in "_COMPLETE" state
func PrintStackEventsDuringOperation(svc cfi.CloudFormationAPI, stackName string, lastEventTs time.Time, writer io.Writer) error {
	opComplete := false

	for !opComplete {
		events, err := GetStackEventsAfterTime(svc, stackName, lastEventTs)
		if err != nil {
			return fmt.Errorf("Failed to get stack events after time %s: ", lastEventTs.String, err)
		}

		// Note that events are returned in reverse chronological order.
		// Iterate through the array in reverse and print the oldest event first
		for i := len(events) - 1; i >= 0; i-- {
			_, err := writer.Write([]byte(events[i].String()))
			if err != nil {
				return fmt.Errorf("Failed to print stack event: %s", err.Error())
			}
		}

		// Set the timestamp to the timestamp of the newest event seen
		if len(events) > 0 {
			lastEventTs = *events[0].event.Timestamp
		}

		dp := &cf.DescribeStacksInput{StackName: aws.String(stackName)}
		r, err := svc.DescribeStacks(dp)
		if err != nil {
			return fmt.Errorf("Failed to fetch stack status: %s", err.Error())
		}
		stack := r.Stacks[0]
		stackStatus := *stack.StackStatus

		// Check if the current stack status is in some "COMPLETE" state.
		// This state could be rollback/create/delete completion and
		// indicates the recent set of operations have completed.
		if !strings.HasSuffix(stackStatus, "_COMPLETE") {
			// Sleep for sometime to avoid calls with no new stack events
			time.Sleep(3000 * time.Millisecond)
		} else {
			opComplete = true
		}
	}

	return nil
}

// DerefString checks if the input string pointer is not nil and can be
// dereferenced. If not, it returns the input default value.
func DerefString(strPtr *string, dft string) string {
	if strPtr == nil {
		return dft
	}
	return *strPtr
}
