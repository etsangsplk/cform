package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/spf13/cobra"
)

var planCmdFlags struct {
	// Name of the CloudFormation stack
	stackName string

	// Stack configuration file containing details of the stack. This file
	// contains the parameters which are usually passed to the `cloudformation`
	// CLI command.
	stackConfigFile string

	// The name of the temporary change set to be created which is used to
	// determine the execution plan. By default a random timestamped name is
	// generated.
	changeSetName string

	// If true, the change set created to determine the execution plan will be
	// retained.
	keepChangeSet bool
}

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Show execution plan",
	PreRun: func(cmd *cobra.Command, args []string) {
		// Set random change set name if not passed
		if planCmdFlags.changeSetName == "" {
			t := time.Now()
			planCmdFlags.changeSetName = fmt.Sprintf("cs-%s", t.Format("20060102150405"))
			log.WithField("change-set-name", planCmdFlags.changeSetName).Debug("generating new change set name")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// TODO Check if stack exists
		if err := mergeFromDir(rootCmdFlags.tmplSrc, rootCmdFlags.tmplOut, rootCmdFlags.tmplOverwrite); err != nil {
			os.Exit(-1)
		}

		tmpl, err := ioutil.ReadFile(rootCmdFlags.tmplOut)
		if err != nil {
			log.WithError(err).Error("cannot read merged template file")
			os.Exit(-1)
		}

		sess, err := session.NewSession()
		if err != nil {
			log.WithError(err).Error("failed to create session")
			os.Exit(-1)
		}

		svc := cloudformation.New(sess)
		if err := plan(svc, string(tmpl), planCmdFlags.stackConfigFile, planCmdFlags.stackName,
			planCmdFlags.changeSetName, planCmdFlags.keepChangeSet); err != nil {
			os.Exit(-1)
		}
	},
}

// plan creates a new change set using the input template and returns the
// execution plan based on information retrieved from the change set.
func plan(svc cloudformationiface.CloudFormationAPI, tmpl, stackConfigFile, stackName, changeSetName string, keepChangeSet bool) error {
	changeSetCreated := false
	defer func() {
		if changeSetCreated && !keepChangeSet {
			// Delete the change set
			delInput := &cloudformation.DeleteChangeSetInput{
				ChangeSetName: aws.String(changeSetName),
				StackName:     aws.String(stackName),
			}
			if _, err := svc.DeleteChangeSet(delInput); err != nil {
				log.WithField("change-set-name", changeSetName).Error("cannot delete change set")
			} else {
				log.WithField("change-set-name", changeSetName).Debug("deleted change set")
			}
		}
	}()

	// Create the change set
	createInput := &cloudformation.CreateChangeSetInput{
		ChangeSetName: aws.String(changeSetName),
		StackName:     aws.String(stackName),
		TemplateBody:  aws.String(tmpl),
	}

	createResp, err := svc.CreateChangeSet(createInput)
	if err != nil {
		log.WithError(err).Error("cannot create change set to determine plan")
		return err
	}
	changeSetCreated = true
	log.WithField("change-set-arn", *createResp.Id).Debug("created change set to determine plan")

	descInput := &cloudformation.DescribeChangeSetInput{
		ChangeSetName: aws.String(changeSetName),
		StackName:     aws.String(stackName),
	}
	descResp, err := describeAvailableChangeSet(svc, descInput)
	if err != nil {
		log.WithError(err).Error("cannot retrieve change set status")
		return err
	}
	printChangeSetChanges(descResp)

	return nil
}

// printChangeSetChanges prints the changes to any new or existing resources
// to standard out.
// TODO need to decide on a better format for the execution plan
func printChangeSetChanges(status *cloudformation.DescribeChangeSetOutput) {
	for _, change := range status.Changes {
		rs := change.ResourceChange
		fmt.Printf("%s (%s - %s): %s\n", *rs.LogicalResourceId, *rs.ResourceType, *rs.PhysicalResourceId, *rs.Action)
	}
}

// describeAvailableChangeSet waits for the change set to be created and then
// returns the status of the change set.
func describeAvailableChangeSet(svc cloudformationiface.CloudFormationAPI, input *cloudformation.DescribeChangeSetInput) (*cloudformation.DescribeChangeSetOutput, error) {
	timeout := time.After(10 * time.Second)
	tick := time.Tick(500 * time.Millisecond)

	for {
		select {
		case <-timeout:
			err := errors.New("change set creation timed out")
			log.Error(err)
			return nil, err
		case <-tick:
			resp, err := svc.DescribeChangeSet(input)
			if err != nil {
				log.WithError(err).Error("cannot retrieve change set status")
				return nil, err
			}

			status := *resp.Status

			if status == cloudformation.ChangeSetStatusFailed {
				return nil, fmt.Errorf("cannot create change set: %s", *resp.StatusReason)
			}

			if status == cloudformation.ChangeSetStatusCreateComplete {
				return resp, nil
			}

			log.WithField("change-set-status", status).Debug("waiting for change set to be created...")
		}
	}
}

func init() {
	planCmd.Flags().StringVar(&planCmdFlags.stackName, "stack-name", "", "Name of the CloudFormation stack")
	planCmd.Flags().StringVar(&planCmdFlags.stackConfigFile, "stack-config", "", "Path to stack config file")
	planCmd.Flags().StringVar(&planCmdFlags.changeSetName, "change-set-name", "", "Name of the change set")
	planCmd.Flags().BoolVar(&planCmdFlags.keepChangeSet, "keep-change-set", false, "Retain the change set created to prepare the plan")

	rootCmd.AddCommand(planCmd)
}
