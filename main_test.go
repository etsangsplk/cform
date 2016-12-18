package main

import (
	"strings"
	"testing"
)

type inMemoryReader struct {
	yamls [][]byte
	idx   int
}

func newInMemoryReader(yamls []string) (*inMemoryReader, error) {
	r := &inMemoryReader{}

	for _, str := range yamls {
		r.yamls = append(r.yamls, []byte(str))
	}
	r.idx = -1

	return r, nil
}

func (r *inMemoryReader) Next() ([]byte, error) {
	return r.yamls[r.idx], nil
}

func (r *inMemoryReader) HasNext() bool {
	r.idx++
	return r.idx < len(r.yamls)
}

func format(s string) string {
	// Replace tab char by 4 spaces so that the curated YAML string does not
	// contain mixed indentation.
	return strings.Replace(strings.Trim(s, "\n"), "\t", "    ", -1)
}

func testResult(t *testing.T, expected, actual string) {
	// TODO
	// Replace the spaces to make the comparisons possible. Marshalling and
	// unmarshalling the yaml somehow effects the space identation. Figure out
	// how to avoid this.
	expected = strings.Replace(expected, " ", "", -1)
	actual = strings.Replace(actual, " ", "", -1)

	if expected != actual {
		t.Errorf("Expected %s, Found %s", expected, actual)
	}
}

func TestSingleYamlMerge(t *testing.T) {
	var d = format(`
	a: 
	  a1: 123
	b:
	  b1: 1
	  b2: 2
	`)
	r, err := newInMemoryReader([]string{d})
	if err != nil {
		t.Errorf("Failed to create reader", err)
		return
	}

	b, err := MergeYaml(r)
	if err != nil {
		t.Errorf("Failed to merge yaml", err)
		return
	}

	testResult(t, d, string(b))
}

func TestYamlMergeWithOverlap(t *testing.T) {
	var d1 = format(`
	a:
	  a1: 1
	b:
	  b1: 1
	c:
	  c1: 1
	`)

	var d2 = format(`
	a:
	  a2: 2
	b:
	  b2: 2
	d:
	  d1: 1
	`)

	var d = format(`
	a:
	  a1: 1
	  a2: 2
	b:
	  b1: 1
	  b2: 2
	c:
	  c1: 1
	d:
	  d1: 1
	`)

	r, err := newInMemoryReader([]string{d1, d2})
	if err != nil {
		t.Errorf("Failed to create reader", err)
		return
	}

	b, err := MergeYaml(r)
	if err != nil {
		t.Errorf("Failed to merge yaml", err)
		return
	}

	testResult(t, d, string(b))
}

func TestYamlMergeWithNoOverlap(t *testing.T) {
	var d1 = format(`
	a:
	  a1: 1
	  a2: 2
	c:
	  c1: 1
	`)

	var d2 = format(`
	b:
	  b1: 1
	  b2: 2
	d:
	  d1: 1
	`)

	var d = format(`
	a:
	  a1: 1
	  a2: 2
	b:
	  b1: 1
	  b2: 2
	c:
	  c1: 1
	d:
	  d1: 1
	`)

	r, err := newInMemoryReader([]string{d1, d2})
	if err != nil {
		t.Errorf("Failed to create reader", err)
		return
	}

	b, err := MergeYaml(r)
	if err != nil {
		t.Errorf("Failed to merge yaml", err)
		return
	}

	testResult(t, d, string(b))
}

func TestYamlMergeWithStrings(t *testing.T) {
	var d1 = format(`
	a:
	  a1: 1
	b:
	  b1:
	    Fn::Sub: "foo.${variable}"
	c:
	  c1: 1
	`)

	var d2 = format(`
	a:
	  a2: 2
	b:
	  b2: 2
	  b3: "hello"
	d:
	  d1: "d1"
	`)

	var d = format(`
	a:
	  a1: 1
	  a2: 2
	b:
	  b1:
	    Fn::Sub: "foo.${variable}"
	  b2: 2
	  b3: "hello"
	c:
	  c1: 1
	d:
	  d1: "d1"
	`)

	r, err := newInMemoryReader([]string{d1, d2})
	if err != nil {
		t.Errorf("Failed to create reader", err)
		return
	}

	b, err := MergeYaml(r)
	if err != nil {
		t.Errorf("Failed to merge yaml", err)
		return
	}

	testResult(t, d, string(b))
}

func TestYamlMergeWithMultiLineStrings(t *testing.T) {
	var d1 = format(`
	b:
	  b1:
	    Fn::Sub: |
		  This is a line with a ${variable}
		  This is another line
	c:
	  c1: 1`)

	var d2 = format(`
	a:
	  a1: "2"
	  a2: |
	    This is a line
		This is the second line
		This is another line
	b:
	  b2: "hello"
	`)

	var d = format(`
	a:
	  a1: "2"
	  a2: |
	    This is a line
		This is the second line
		This is another line
	b:
	  b1:
	    Fn::Sub: |
		  This is a line with a ${variable}
		  This is another line
	  b2: "hello"
	c:
	  c1: 1
	`)

	r, err := newInMemoryReader([]string{d1, d2})
	if err != nil {
		t.Errorf("Failed to create reader", err)
		return
	}

	b, err := MergeYaml(r)
	if err != nil {
		t.Errorf("Failed to merge yaml", err)
		return
	}

	testResult(t, d, string(b))
}
