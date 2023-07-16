package utils

func (set *Set[E]) Add(ele E) bool {
	_, ok := set.container[ele]
	if ok {
		return false
	}
	set.container[ele] = true
	return true
}

func (set *Set[E]) Remove(ele any) bool {
	_, ok := set.container[ele]
	if !ok {
		return false
	}
	delete(set.container, ele)
	return true
}

func (set *Set[E]) AddAll(elements []E) {
	for _, e := range elements {
		set.Add(e)
	}
}

func (set *Set[E]) Contains(e any) bool {
	_, ok := set.container[e]
	return ok
}
func (set *Set[E]) Clear() {
	set.container = make(map[any]bool)
}

func (set *Set[E]) RemoveIf(filter Filter[E]) {
	for k := range set.container {
		e := k.(E)
		if filter(e) {
			set.Remove(e)
		}
	}
}

func (set *Set[E]) Foreach(f Consumer[E]) {
	for k := range set.container {
		f(k.(E))
	}
}
