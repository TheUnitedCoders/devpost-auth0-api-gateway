package slice

import mapset "github.com/deckarep/golang-set/v2"

// ConvertFuncWithSkip ...
func ConvertFuncWithSkip[From, To any](slice []From, convertFunc func(elem From) (To, bool)) []To {
	if len(slice) == 0 {
		return nil
	}

	result := make([]To, 0, len(slice))

	for _, elem := range slice {
		resElem, skip := convertFunc(elem)
		if skip {
			continue
		}

		result = append(result, resElem)
	}

	return result
}

// ConvertFunc ...
func ConvertFunc[From, To any](slice []From, convertFunc func(elem From) To) []To {
	return ConvertFuncWithSkip(
		slice,
		func(elem From) (To, bool) {
			return convertFunc(elem), false
		},
	)
}

// Merge two slices into one with deduplication.
func Merge[T comparable](s1, s2 []T) []T {
	result := make([]T, 0, len(s1)+len(s2))
	s2Set := mapset.NewThreadUnsafeSet(s2...)

	result = append(result, s2Set.ToSlice()...)

	for _, elem := range s1 {
		if s2Set.ContainsOne(elem) {
			continue
		}

		result = append(result, elem)
	}

	return result
}
