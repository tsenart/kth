// The pdqselect package implements an adaptive selection algorithm that finds
// the k-th smallest elements in an ordered data structure. It is based on Go's pdqsort
// implementation, which is a hybrid sorting algorithm that combines quicksort, insertion sort,
// heapsort and other pattern defeating techniques to achieve optimal performance on a wide range of data.
package pdqselect

import (
	"cmp"
	"math/bits"
	"sort"
)

// Select swaps elements in the data provided so that the first k elements
// (i.e. the elements occuping indices 0, 1, ..., k-1) are the smallest k elements
// in the data. It doesn't guarantee any particular order among the smallest k elements,
// only that they are the smallest k elements in the data.
//
// Select implements Hoare's Selection Algorithm and runs in O(n) time, so it
// is asymptotically faster than sorting or other heap-like implementations for
// finding the smallest k elements in a data structure.
//
// It's an adaptation of Go's internal pdqsort implementation, which makes it adaptive
// to bad data patterns like already sorted data, duplicate elements, and more.
func Select(data sort.Interface, k int) {
	n := data.Len()
	if k < 1 || k > n {
		return
	}
	pdqselect(data, 0, n, k-1, bits.Len(uint(n)))
}

// Ordered is a specialized version of Select that works with slices of
// ordered types (i.e. types that implement the cmp.Ordered interface).
func Ordered[T cmp.Ordered](data []T, k int) {
	n := len(data)
	if k < 1 || k > n {
		return
	}
	pdqselectOrdered(data, 0, n, k-1, bits.Len(uint(n)))
}

// Func is a generic version of Select that allows the caller to provide
// a custom comparison function to determine the order of elements.
func Func[E any](data []E, k int, cmp func(i, j E) int) {
	n := len(data)
	if k < 1 || k > n {
		return
	}
	pdqselectFunc(data, 0, n, k-1, bits.Len(uint(n)), cmp)
}

func pdqselect(data sort.Interface, a, b, k, limit int) {
	if k == 0 { // Fast path; just find the minimum and place it in a
		mn := a
		for i := a; i < b; i++ {
			if data.Less(i, mn) {
				mn = i
			}
		}
		if mn != a {
			data.Swap(mn, a)
		}
		return
	}

	if hi := b - 1; k == hi { // Fast path; just find the maximum and place it in b-1
		mx := a
		for i := a + 1; i < b; i++ {
			if data.Less(mx, i) {
				mx = i
			}
		}
		if mx != hi {
			data.Swap(mx, hi)
		}
		return
	}

	const maxInsertion = 12

	var (
		wasBalanced    = true
		wasPartitioned = true
	)

	for {
		length := b - a

		if length <= maxInsertion {
			insertionSort(data, a, b)
			return
		}

		// Fall back to heap select if too many bad choices were made.
		if limit == 0 {
			heapSelect(data, a, b, k)
			return
		}

		// Break patterns if the last partitioning was imbalanced
		if !wasBalanced {
			breakPatterns(data, a, b)
			limit--
		}

		pivot, hint := choosePivot(data, a, b)
		if hint == decreasingHint {
			reverseRange(data, a, b)
			// The chosen pivot was pivot-a elements after the start of the array.
			// After reversing it is pivot-a elements before the end of the array.
			// The idea came from Rust's implementation.
			pivot = (b - 1) - (pivot - a)
			hint = increasingHint
		}

		// Check if the slice is likely already sorted
		if wasBalanced && wasPartitioned && hint == increasingHint {
			if partialInsertionSort(data, a, b) {
				return
			}
		}

		// Probably the slice contains many duplicate elements, partition the slice into
		// elements equal to and elements greater than the pivot.
		if a > 0 && !data.Less(a-1, pivot) {
			mid := partitionEqual(data, a, b, pivot)
			if k < mid {
				return
			}
			a = mid
			continue
		}

		mid, alreadyPartitioned := partition(data, a, b, pivot)
		if k == mid {
			return
		}

		wasPartitioned = alreadyPartitioned
		leftLen, rightLen := mid-a, b-mid
		balanceThreshold := length / 8

		if k < mid {
			wasBalanced = leftLen >= balanceThreshold
			b = mid
		} else { // k < mid
			wasBalanced = rightLen >= balanceThreshold
			a = mid + 1
		}
	}
}

func pdqselectOrdered[T cmp.Ordered](data []T, a, b, k, limit int) {
	if k == 0 { // Fast path; just find the minimum and place it in a
		mn := a
		for i := a + 1; i < b; i++ {
			if data[i] < data[mn] {
				mn = i
			}
		}
		data[a], data[mn] = data[mn], data[a]
		return
	}

	if hi := b - 1; k == hi { // Fast path; just find the maximum and place it in b-1
		mx := a
		for i := a + 1; i < b; i++ {
			if data[i] > data[mx] {
				mx = i
			}
		}
		data[hi], data[mx] = data[mx], data[hi]
		return
	}

	const maxInsertion = 12

	var (
		wasBalanced    = true
		wasPartitioned = true
	)

	for {
		length := b - a

		if length <= maxInsertion {
			insertionSortOrdered(data, a, b)
			return
		}

		// Fall back to heap select if too many bad choices were made.
		if limit == 0 {
			heapSelectOrdered(data, a, b, k)
			return
		}

		// Break patterns if the last partitioning was imbalanced
		if !wasBalanced {
			breakPatternsOrdered(data, a, b)
			limit--
		}

		pivot, hint := choosePivotOrdered(data, a, b)
		if hint == decreasingHint {
			reverseRangeOrdered(data, a, b)
			// The chosen pivot was pivot-a elements after the start of the array.
			// After reversing it is pivot-a elements before the end of the array.
			// The idea came from Rust's implementation.
			pivot = (b - 1) - (pivot - a)
			hint = increasingHint
		}

		// Check if the slice is likely already sorted
		if wasBalanced && wasPartitioned && hint == increasingHint {
			if partialInsertionSortOrdered(data, a, b) {
				return
			}
		}

		// Probably the slice contains many duplicate elements, partition the slice into
		// elements equal to and elements greater than the pivot.
		if a > 0 && data[a-1] >= data[pivot] {
			mid := partitionEqualOrdered(data, a, b, pivot)
			if k < mid {
				return
			}
			a = mid
			continue
		}

		mid, alreadyPartitioned := partitionOrdered(data, a, b, pivot)
		if k == mid {
			return
		}

		wasPartitioned = alreadyPartitioned
		leftLen, rightLen := mid-a, b-mid
		balanceThreshold := length / 8

		if k < mid {
			wasBalanced = leftLen >= balanceThreshold
			b = mid
		} else { // k < mid
			wasBalanced = rightLen >= balanceThreshold
			a = mid + 1
		}
	}
}

func pdqselectFunc[E any](data []E, a, b, k, limit int, cmp func(a, b E) int) {
	if k == 0 { // Fast path; just find the minimum and place it in a
		mn := a
		for i := a + 1; i < b; i++ {
			if cmp(data[i], data[mn]) < 0 {
				mn = i
			}
		}
		if mn != a {
			data[a], data[mn] = data[mn], data[a]
		}
		return
	}

	if hi := b - 1; k == hi { // Fast path; just find the maximum
		mx := a
		for i := a + 1; i < b; i++ {
			if cmp(data[i], data[mx]) > 0 {
				mx = i
			}
		}
		if mx != hi {
			data[hi], data[mx] = data[mx], data[hi]
		}
		return
	}

	const maxInsertion = 12

	var (
		wasBalanced    = true
		wasPartitioned = true
	)

	for {
		length := b - a

		if length <= maxInsertion {
			insertionSortCmpFunc(data, a, b, cmp)
			return
		}

		// Fall back to heap select if too many bad choices were made.
		if limit == 0 {
			heapSelectFunc(data, a, b, k, cmp)
			return
		}

		// Break patterns if the last partitioning was imbalanced
		if !wasBalanced {
			breakPatternsCmpFunc(data, a, b, cmp)
			limit--
		}

		pivot, hint := choosePivotCmpFunc(data, a, b, cmp)
		if hint == decreasingHint {
			reverseRangeCmpFunc(data, a, b, cmp)
			// The chosen pivot was pivot-a elements after the start of the array.
			// After reversing it is pivot-a elements before the end of the array.
			// The idea came from Rust's implementation.
			pivot = (b - 1) - (pivot - a)
			hint = increasingHint
		}

		// Check if the slice is likely already sorted
		if wasBalanced && wasPartitioned && hint == increasingHint {
			if partialInsertionSortCmpFunc(data, a, b, cmp) {
				return
			}
		}

		// Probably the slice contains many duplicate elements, partition the slice into
		// elements equal to and elements greater than the pivot.
		if a > 0 && cmp(data[a-1], data[pivot]) >= 0 {
			mid := partitionEqualCmpFunc(data, a, b, pivot, cmp)
			if k < mid {
				return
			}
			a = mid
			continue
		}

		mid, alreadyPartitioned := partitionCmpFunc(data, a, b, pivot, cmp)
		if k == mid {
			return
		}

		wasPartitioned = alreadyPartitioned
		leftLen, rightLen := mid-a, b-mid
		balanceThreshold := length / 8

		if k < mid {
			wasBalanced = leftLen >= balanceThreshold
			b = mid
		} else { // k < mid
			wasBalanced = rightLen >= balanceThreshold
			a = mid + 1
		}
	}
}

func heapSelect(data sort.Interface, a, b, k int) {
	n := b - a
	hi := k + 1

	// Build max-heap of first k elements
	for i := k / 2; i >= 0; i-- {
		siftDown(data, i, hi, a)
	}

	// Process remaining elements
	for i := hi; i < n; i++ {
		j := a + i
		if data.Less(j, a) {
			data.Swap(a, j)
			siftDown(data, 0, hi, a)
		}
	}

	// Place the k-th element into its final place
	data.Swap(a, a+k)
}

func heapSelectOrdered[T cmp.Ordered](data []T, a, b, k int) {
	n := b - a
	hi := k + 1

	// Build max-heap of first k elements
	for i := k / 2; i >= 0; i-- {
		siftDownOrdered(data, i, hi, a)
	}

	// Process remaining elements
	for i := hi; i < n; i++ {
		j := a + i
		if data[j] < data[a] {
			data[a], data[j] = data[j], data[a]
			siftDownOrdered(data, 0, hi, a)
		}
	}

	// Place the k-th element into its final place
	data[a], data[a+k] = data[a+k], data[a]
}

func heapSelectFunc[E any](data []E, a, b, k int, cmp func(a, b E) int) {
	n := b - a
	hi := k + 1

	// Build max-heap of first k elements
	for i := k / 2; i >= 0; i-- {
		siftDownCmpFunc(data, i, hi, a, cmp)
	}

	// Process remaining elements
	for i := hi; i < n; i++ {
		j := a + i
		if cmp(data[j], data[a]) < 0 {
			data[a], data[j] = data[j], data[a]
			siftDownCmpFunc(data, 0, hi, a, cmp)
		}
	}

	// Place the k-th element into its final place
	data[a], data[a+k] = data[a+k], data[a]
}
