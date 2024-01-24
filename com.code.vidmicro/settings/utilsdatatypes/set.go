package utilsdatatypes

type Set struct {
	data map[string]bool
}

// NewSet creates a new Set
func NewSet() *Set {
	return &Set{
		data: make(map[string]bool),
	}
}

// Add adds an element to the Set
func (set *Set) Add(element string) {
	set.data[element] = true
}

// Contains checks if an element is present in the Set
func (set *Set) Contains(element string) bool {
	return set.data[element]
}

// Remove removes an element from the Set
func (set *Set) Remove(element string) {
	delete(set.data, element)
}

// Size returns the size of the Set
func (set *Set) Size() int {
	return len(set.data)
}
