package lib

type StringSet struct {
	items map[string]bool
}

func (s *StringSet) initializeSet() {
	if s.items == nil {
		s.items = make(map[string]bool)
	}
}

func (s *StringSet) Add(item string) {
	s.initializeSet()
	_, found := s.items[item]
	if !found {
		s.items[item] = true
	}
}

func (s *StringSet) Remove(item string) {
	s.initializeSet()
	delete(s.items, item)
}

func (s *StringSet) GetItems() map[string]bool {
	s.initializeSet()
	return s.items
}

func (s *StringSet) Array() []string {
	s.initializeSet()
	var array []string
	for item, _ := range s.items {
		array = append(array, item)
	}
	return array
}
