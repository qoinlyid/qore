package qore

// SliceRemoveItems removes item(s) in `src` that have in `target`.
func SliceRemoveItems[T comparable](src, target []T) []T {
	lookup := make(map[T]struct{}, len(target))
	for _, item := range target {
		lookup[item] = struct{}{}
	}
	result := src[:0]
	for _, item := range src {
		if _, found := lookup[item]; !found {
			result = append(result, item)
		}
	}
	return result
}
