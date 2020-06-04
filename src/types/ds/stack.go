package ds

type intStack struct {
	storage []int
}

type EmptyStackError struct{}

func (e EmptyStackError) Error() string {
	return "stack is empty"
}

func NewIntStackFromSlice(s []int) *intStack {
	ns := make([]int, len(s))
	copy(ns, s)
	return &intStack{ns}
}

func NewIntStack() *intStack {
	return &intStack{make([]int, 0)}
}

// IsEmpty checks if the stack is empty.
func (s *intStack) IsEmpty() bool {
	return len(s.storage) == 0
}

// Push pushes a new value onto the stack.
func (s *intStack) Push(value int) {
	s.storage = append(s.storage, value)
}

// Pop removes and return top element of stack.
// Returns false if the stack is empty.
func (s *intStack) Pop() (int, error) {
	if s.IsEmpty() {
		return 0, EmptyStackError{}
	}
	index := len(s.storage) - 1   // Get the index of the top most element.
	element := s.storage[index]   // Index into the slice and obtain the element.
	s.storage = s.storage[:index] // Remove it from the stack by slicing it off.
	return element, nil
}
