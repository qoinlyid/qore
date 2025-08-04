package command

import (
	"fmt"
	"strings"
)

type (
	// result defines contstraint type of the result.
	result interface {
		string | Option | []Option |
			bool
	}

	// outputInterface defines output interface.
	outputInterface interface {
		formatResponse() string
	}
)

// Output defines command output.
type Output[T result] struct {
	question string
	result   T
	context  []string
}

// newOutput creates new command output instance.
func newOutput[T result](question string, result T) *Output[T] {
	return &Output[T]{
		question: question,
		result:   result,
		context:  []string{},
	}
}

// formatResponse from output.
func (o *Output[T]) formatResponse() string {
	switch v := any(o.result).(type) {
	default:
		return "-"
	case string:
		return fmt.Sprintf("%s %s", o.question, v)
	case Option:
		return fmt.Sprintf("%s %s", o.question, v.Value)
	case []Option:
		var val []string
		for _, x := range v {
			val = append(val, x.Value)
		}
		return fmt.Sprintf("%s %s", o.question, strings.Join(val, ", "))
	}
}

// WithContext adds context into the command output.
func (o *Output[T]) WithContext(previousOutputs ...outputInterface) *Output[T] {
	for _, prev := range previousOutputs {
		if prev != nil {
			o.context = append(o.context, prev.formatResponse())
		}
	}
	return o
}

// PrintResponse prints command output.
func (o *Output[T]) PrintResponse() {
	for _, ctx := range o.context {
		fmt.Println(ctx)
	}
	fmt.Println(o.formatResponse())
}

// Value returns value of command output.
func (o *Output[T]) Value() T {
	return o.result
}
