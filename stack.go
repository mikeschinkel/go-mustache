package mustache

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}
func (s *Stack[T]) Pop() T {
	if len(s.items) == 0 {
		return *new(T)
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item
}
func (s *Stack[T]) Top() T {
	if len(s.items) == 0 {
		return *new(T)
	}
	return s.items[len(s.items)-1]
}

func (s *Stack[T]) Depth() int {
	return len(s.items)
}
func (s *Stack[T]) Empty() bool {
	return len(s.items) == 0
}

type Cloner[T any] interface {
	Clone() T
}

type ClonableStack[T Cloner[T]] struct {
	Stack[T]
}

func (s *ClonableStack[T]) Clone() *ClonableStack[T] {
	stack := &ClonableStack[T]{}
	for _, item := range s.items {
		stack.Push(item.Clone())
	}
	return stack
}
