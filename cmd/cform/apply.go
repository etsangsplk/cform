package main

import (
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/isubuz/cform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/spf13/cobra"
)

var APPLY_STATUS_FMT = "%-25s\t%-20s\t%-30s\t%-20s\t%s\n"

var applyCmdFlags struct {
	// Name of the CloudFormation stack
	stackName string

	// Stack configuration file containing details of the stack. This file
	// contains the parameters which are usually passed to the `cloudformation`
	// CLI command
	stackConfigFile string
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Create or update a CloudFormation stack",
	Run: func(cmd *cobra.Command, args []string) {
		if err := mergeFromDir(rootCmdFlags.tmplSrc, rootCmdFlags.tmplOut); err != nil {
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
		if err := apply(svc, string(tmpl), planCmdFlags.stackConfigFile, planCmdFlags.stackName); err != nil {
			os.Exit(-1)
		}
	},
}

func apply(svc cloudformationiface.CloudFormationAPI, tmpl, stackConfigFile, stackName string) error {
	var stackExists bool
	descInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}

	// Check if stack exists
	_, err := svc.DescribeStacks(descInput)
	if err != nil {
		awsErr := err.(awserr.Error)

		// ValidationError indicates a stack not found error.
		// TODO can a constant be used?
		if awsErr.Code() == "ValidationError" {
			log.WithField("stack-name", stackName).Debug("stack does not exist; running create mode")
			stackExists = false
		} else {
			log.WithError(err).Error("unknown error encountered")
			return err
		}
	} else {
		log.WithField("stack-name", stackName).Debug("stack exists; running update mode")
		stackExists = true
	}

	// Get the timestamp of the last stack event
	p0 := &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stackName),
	}
	resp, err := svc.DescribeStackEvents(p0)
	if err != nil {
		return err
	}
	ts := *resp.StackEvents[0].Timestamp

	if !stackExists {
		p := &cloudformation.CreateStackInput{
			StackName:    aws.String(stackName),
			TemplateBody: aws.String(tmpl),
		}
		_, err := svc.CreateStack(p)
		if err != nil {
			log.WithError(err).Error("cannot create stack")
			return err
		}
	} else {
		p := &cloudformation.UpdateStackInput{
			StackName:    aws.String(stackName),
			TemplateBody: aws.String(tmpl),
		}
		_, err := svc.UpdateStack(p)
		if err != nil {
			log.WithError(err).Error("cannot update stack")
			return err
		}
	}
	if err = cform.PrintStackEventsDuringOperation(svc, stackName, ts, os.Stdout); err != nil {
		log.WithError(err).Error("cannot print stack events")
		return err
	}
	return nil
}

func init() {
	applyCmd.Flags().StringVar(&planCmdFlags.stackName, "stack-name", "", "Name of the CloudFormation stack")
	applyCmd.Flags().StringVar(&planCmdFlags.stackConfigFile, "stack-config", "", "Path to stack config file")

	rootCmd.AddCommand(applyCmd)
}
