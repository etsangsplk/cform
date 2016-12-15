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
