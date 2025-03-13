package mustache

type Stack[T any] struct {
	items []T
	n     int
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

type Cloner[T any] interface {
	Clone() T
}

type ClonableStack[T Cloner[T]] struct {
	Stack[T]
}

func (s *ClonableStack[T]) Clone() *ClonableStack[T] {
	stack := &ClonableStack[T]{}
	for _, item := range s.items {
		// TODO Does this need to be cloned?
		stack.Push(item.Clone())
	}
	return stack
}
