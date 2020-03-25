package mfd

type Set struct {
	set   map[string]struct{}
	order []string
}

func NewSet() Set {
	return Set{
		set:   map[string]struct{}{},
		order: []string{},
	}
}

func (s *Set) Prepend(element string) {
	if !s.Exists(element) {
		s.set[element] = struct{}{}
		s.order = append([]string{element}, s.order...)
	} else {
		for i, o := range s.order {
			if o == element {
				s.order = append([]string{o}, append(s.order[:i], s.order[i+1:]...)...)
			}
		}
	}
}

func (s *Set) Append(element string) {
	if !s.Exists(element) {
		s.set[element] = struct{}{}
		s.order = append(s.order, element)
	}
}

func (s *Set) Add(element string) {
	if !s.Exists(element) {
		s.set[element] = struct{}{}
		s.order = append([]string{element}, s.order...)
	}
}

func (s *Set) Elements() []string {
	return s.order
}

func (s *Set) Exists(element string) bool {
	_, ok := s.set[element]
	return ok
}

func (s *Set) Len() int {
	return len(s.order)
}
