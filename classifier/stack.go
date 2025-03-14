package classifier

import (
	"fmt"
	"strings"
)

type Stack[T any] struct {
	items []T
	n     int
}

func (s *Stack[T]) String() string {
	ss := make([]string, len(s.items))
	for i, item := range s.items {
		ss[i] = fmt.Sprintf("%v", item)
	}
	return strings.Join(ss, ", ")
}

func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
	s.n++
}
func (s *Stack[T]) Pop() T {
	if len(s.items) == 0 {
		return *new(T)
	}
	s.n--
	item := s.items[s.n]
	s.items = s.items[:s.n]
	return item
}
func (s *Stack[T]) Top() T {
	if len(s.items) == 0 {
		return *new(T)
	}
	return s.items[s.n-1]
}

func (s *Stack[T]) Depth() int {
	return s.n
}
func (s *Stack[T]) Empty() bool {
	return s.n == 0
}
