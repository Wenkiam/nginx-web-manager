package utils

type Set[E any] struct {
	container map[any]bool
}
type Consumer[E any] func(o E)

type Filter[E any] func(o E) bool

func NewSet[E any]() *Set[E] {
	return &Set[E]{
		make(map[any]bool),
	}
}
func SetOf[E any](slice []E) *Set[E] {
	set := &Set[E]{
		make(map[any]bool, len(slice)),
	}
	set.AddAll(slice)
	return set
}
