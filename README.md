# cfn-tmpl

Merge multiple CloudFormation templates into a single template. This helps to
structure the templates into separate files.

## Usage

Currently this tool supports only a `merge` operation to merge multiple 
CloudFormation YAML files in one. 

Run the following to use the tool:

```
$ go build
$ cfn-tmpl merge --help
```

## Cloudformation YAML syntax

The YAML parser used to parse the YAML templates requires that the templates
contain valid YAML. Unfortunately when the shorthand form of the Cloudformation
intrinsic functions are used, the parser does not parse the templates correctly.

Hence for the `cfn-tmpl` tool to work, the templates mustn't contain the 
short form function syntax i.e. use `Fn::Ref` instead of `!Ref`. E.g. for 
specifying the userdata for an EC2 instance, use the following `Fn::*` syntax -

```yaml
UserData:
  Fn::Base64:
    Fn::Sub: |
      #!/bin/bash -xe
      yum update -y aws-cfn-bootstrap
      /opt/aws/bin/cfn-init -v --stack ${AWS::StackName} --resource LaunchConfig --configsets wordpress_install --region ${AWS::Region}
      /opt/aws/bin/cfn-signal -e $? --stack ${AWS::StackName} --resource WebServerGroup --region ${AWS::Region}
```

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
