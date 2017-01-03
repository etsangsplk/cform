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
provided by the `terraform plan` command. Run the followin for help -

```sh
$ cform plan --help
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
