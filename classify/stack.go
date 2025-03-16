package classify

import (
	"fmt"
	"strings"
)

// Stack is a generic stack implementation that can store any type. It is used in
// this package by the TemplateClassifier to track nested tags during template
// classification.
type Stack[T any] struct {
	items []T
	n     int
}

// String returns a comma-separated string representation of all items in the
// stack for use in error messages and test output.
func (s *Stack[T]) String() string {
	ss := make([]string, len(s.items))
	for i, item := range s.items {
		ss[i] = fmt.Sprintf("%v", item)
	}
	return strings.Join(ss, ", ")
}

// Push adds an item to the top of the stack.
func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
	s.n++
}

// Pop removes and returns the item at the top of the stack. If the stack is
// empty, returns a zero value of type T.
func (s *Stack[T]) Pop() T {
	if len(s.items) == 0 {
		return *new(T)
	}
	s.n--
	item := s.items[s.n]
	s.items = s.items[:s.n]
	return item
}

// Top returns the item at the top of the stack without removing it. If the stack
// is empty, returns a zero value of type T.
func (s *Stack[T]) Top() T {
	if len(s.items) == 0 {
		return *new(T)
	}
	return s.items[s.n-1]
}

// Depth returns the number of items in the stack.
func (s *Stack[T]) Depth() int {
	return s.n
}

// Empty returns true if the stack contains no items.
func (s *Stack[T]) Empty() bool {
	return s.n == 0
}
