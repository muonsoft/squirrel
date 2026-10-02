package squirrel

// immutableList owns its nodes while values retain their original ownership.
// Appending shares the preceding chain, so construction remains linear even for
// long queries. Rendering materializes one slice in the original append order.
type immutableList[T any] struct {
	tail   *immutableNode[T]
	length int
}

type immutableNode[T any] struct {
	previous *immutableNode[T]
	value    T
}

func appendPersistent[T any](items immutableList[T], values ...T) immutableList[T] {
	for _, value := range values {
		items.tail = &immutableNode[T]{previous: items.tail, value: value}
		items.length++
	}
	return items
}

func (items immutableList[T]) slice() []T {
	if items.length == 0 {
		return nil
	}
	result := make([]T, items.length)
	end := items.length
	for node := items.tail; node != nil; node = node.previous {
		end--
		result[end] = node.value
	}
	return result
}
