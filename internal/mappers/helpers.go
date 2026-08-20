package mappers

func mapItems[S any, T any](items []S, mapItem func(S) T) []T {
	mapped := make([]T, len(items))
	for i, item := range items {
		mapped[i] = mapItem(item)
	}
	return mapped
}
