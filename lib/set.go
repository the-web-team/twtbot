package lib

type Set struct {
	items map[interface{}]bool
}

func (s *Set) initializeSet() {
	if s.items == nil {
		s.items = make(map[interface{}]bool)
	}
}

func (s *Set) Add(item interface{}) {
	s.initializeSet()
	_, found := s.items[item]
	if !found {
		s.items[item] = true
	}
}

func (s *Set) Remove(item interface{}) {
	s.initializeSet()
	delete(s.items, item)
}

func (s *Set) GetItems() map[interface{}]bool {
	s.initializeSet()
	return s.items
}

func (s *Set) Array() []interface{} {
	s.initializeSet()
	var array []interface{}
	for item, _ := range s.items {
		array = append(array, item)
	}
	return array
}
