# cform

A CloudFormation utility inspired by [Terraform](https://terraform.io) and aims
to provide Terraform-like CLI functionalities.

## Status

Still developed actively for the release of the `0.1` version.

## Building cform

```sh
$ go build ./cmd/cform
$ ./cform --help
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
$ cform plan --debug \
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
