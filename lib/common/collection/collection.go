package collection

func Map[T, R any](items []T, f func(T) R) []R {
	result := make([]R, 0, len(items))
	for _, item := range items {
		result = append(result, f(item))
	}
	return result
}

func MapErr[T, R any](items []T, f func(T) (R, error)) ([]R, error) {
	result := make([]R, 0, len(items))
	for _, item := range items {
		mapped, err := f(item)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}
	return result, nil
}

func Filter[T any](items []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(items))
	for _, item := range items {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

func Reduce[T, A any](items []T, initial A, f func(A, T) A) A {
	accumulator := initial
	for _, item := range items {
		accumulator = f(accumulator, item)
	}
	return accumulator
}

func GroupBy[T any, K comparable](items []T, key func(T) K) map[K][]T {
	return Reduce(items, make(map[K][]T), func(groups map[K][]T, item T) map[K][]T {
		k := key(item)
		groups[k] = append(groups[k], item)
		return groups
	})
}

func KeyBy[T any, K comparable](items []T, key func(T) K) map[K]T {
	return Reduce(items, make(map[K]T, len(items)), func(index map[K]T, item T) map[K]T {
		index[key(item)] = item
		return index
	})
}
