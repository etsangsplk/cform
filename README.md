# cform

A CloudFormation utility inspired by [Terraform](https://terraform.io) and aims
to provide Terraform-like CLI functionalities.

## Status

Still actively developed for the `0.1` release.

## Building cform

```sh
$ go build ./cmd/cform
$ ./cform --help
```
## AWS configuration

Setup the `~/.aws/credentials` and `~/.aws/config` files and sets the
environment variables:

```sh
$ export AWS_PROFILE=<profile> && export AWS_REGION=us-east-1
```

## Supported commands

### cform merge

This command can be used to merge multiple CloudFormation templates into a 
single template. This is can useful when organising a large CloudFormation 
stack template into multiple files. Run the following for help -

```sh
$ cform merge --help
```

An example on how a large CloudFormation template can be organised in multiple 
templates can be found in the [cfn-hugo](https://github.com/isubuz/cfn-hugo)
project.

### cform plan

This command can be used to display an execution plan for any changes to a 
CloudFormation template i.e. it displays the changes to any new or existing
AWS resources based on the change to the template similar to the functionality
provided by the `terraform plan` command. E.g -

```sh
$ ./cform plan --debug \
    --template-src examples \
    --stack-name test-stack \
    --keep-change-set \

DEBU[0000] created new output file for template          template-out=/var/folders/j5/4433kz115732274b3p7l6n380000gn/T/cform543365852
DEBU[0000] generating new change set name                change-set-name=cs-20170107233101
DEBU[0000] created change set to determine plan          change-set-arn=arn:aws:cloudformation:us-east-1:663481583451:changeSet/cs-20170107233101/29170942-bf52-4c13-93e5-44239e1cc060
DEBU[0001] waiting for change set to be created...       change-set-status=CREATE_IN_PROGRESS
DEBU[0001] waiting for change set to be created...       change-set-status=CREATE_IN_PROGRESS
DEBU[0002] waiting for change set to be created...       change-set-status=CREATE_IN_PROGRESS
Bucket1 (AWS::S3::Bucket)
        action         : Modify
        physical-id    : b1.isubuz.com
        replacement    : True

Bucket2 (AWS::S3::Bucket)
        action         : Modify
        physical-id    : b2.isubuz.com
        replacement    : True

```

### cform apply

This command is similar to the `terraform apply` command and creates or updates
a CloudFormation stack. The stack events are displayed in the CLI as and when
the events occur. E.g. -

```sh
$ ./cform apply --debug \
    --template-src ~/examples \
    --stack-name test-stack

DEBU[0000] created new output file for template          template-out=/var/folders/j5/4433kz115732274b3p7l6n380000gn/T/cform261840516
DEBU[0000] stack exists; running update mode             stack-name=test-stack
2017-01-25 11:06:45 +0000 UTC   UPDATE_IN_PROGRESS      AWS::CloudFormation::Stack      test-stack                  User Initiated
2017-01-25 11:06:48 +0000 UTC   CREATE_IN_PROGRESS      AWS::S3::Bucket                 Bucketb5
2017-01-25 11:06:49 +0000 UTC   CREATE_IN_PROGRESS      AWS::S3::Bucket                 Bucketb44
2017-01-25 11:06:50 +0000 UTC   CREATE_IN_PROGRESS      AWS::S3::Bucket                 Bucketb5                Resource creation Initiated
2017-01-25 11:06:50 +0000 UTC   CREATE_IN_PROGRESS      AWS::S3::Bucket                 Bucketb44               Resource creation Initiated
2017-01-25 11:07:10 +0000 UTC   CREATE_COMPLETE         AWS::S3::Bucket                 Bucketb5
2017-01-25 11:07:10 +0000 UTC   CREATE_COMPLETE         AWS::S3::Bucket                 Bucketb44
2017-01-25 11:07:13 +0000 UTC   UPDATE_COMPLETE_CLEANUP_IN_PROGRESS     AWS::CloudFormation::Stack      test-stack
2017-01-25 11:07:15 +0000 UTC   DELETE_IN_PROGRESS      AWS::S3::Bucket                 Bucketb4
2017-01-25 11:07:37 +0000 UTC   DELETE_COMPLETE         AWS::S3::Bucket                 Bucketb4
2017-01-25 11:07:37 +0000 UTC   UPDATE_COMPLETE         AWS::CloudFormation::Stack      test-stack
```

Note that unlike the AWS CloudFormation console, it shows only the events for the 
current operation (update or create) and prints them in chronological order
(which is reversed the CloudFormation console)

## Limitations

### Intrinsic function short names

The YAML parser fails to parse the templates correctly when the template 
contains CloudFormation intrinsic functions in their short form i.e. using the
`!` character with the function name. Multi-line strings aren't parsed correctly
in such cases as the YAML parser treats the `!` character to be special. 

Until this is resolved, use the `Fn::*` syntax. E.g. for specifying the 
userdata for an EC2 instance, use the following -

```yaml
UserData:
  Fn::Base64:
    Fn::Sub: |
      #!/bin/bash -xe
      yum update -y aws-cfn-bootstrap
      /opt/aws/bin/cfn-init -v --stack ${AWS::StackName} --resource LaunchConfig --configsets wordpress_install --region ${AWS::Region}
      /opt/aws/bin/cfn-signal -e $? --stack ${AWS::StackName} --resource WebServerGroup --region ${AWS::Region}
```

Note that for `!Ref`, the alternate is `Ref` and not `Fn::Ref`.

## Assumptions

The reader which reads the YAML files expects the structure of the YAML file to
be same as that of a CloudFormation template i.e. top level attributes like
`Resources`, `Parameters` and attributes of these top level ones (if any) are 
dictionaries. E.g.

```yaml
Description: Some stack

Resources:
  Resource1: 
    attr1: value1
```

Note that `Resource1` is a dictionary.
