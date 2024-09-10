package kth

import (
	"cmp"
	"math"
	"sort"
)

// FloydRivest implements the Floyd-Rivest selection algorithm to find the k-th smallest elements.
// It typically makes fewer comparisons than other selection algorithms by narrowing the search range
// based on order statistics estimates before partitioning.
func FloydRivest(data sort.Interface, k int) {
	n := data.Len()
	if k < 1 || k > n {
		return
	}
	floydRivest(data, 0, n-1, k-1)
}

// rangeNarrowingThreshold represents the size above which we narrow the search range
// using order statistics estimates before partitioning.
const rangeNarrowingThreshold = 600

// The Floyd-Rivest algorithm maintains two core invariants:
//  1. After each iteration, elements known to be less than the k-th element
//     are to its left, and elements known to be greater are to its right
//  2. The k-th element is always within our search bounds [left, right]
//
// The algorithm combines two strategies with proven optimality:
// - Range narrowing based on order statistics for large arrays
// - Efficient partitioning for reduced ranges
func floydRivest(data sort.Interface, left, right, k int) {
	// Loop invariant: k-th element is within [left, right]
	for right > left {
		size := right - left

		// For large arrays, attempt to narrow the search range
		// This is a heuristic that can fail with pathological data distributions
		// but the algorithm remains correct due to the outer loop's invariants
		if size > rangeNarrowingThreshold {
			n := size + 1
			i := k - left + 1

			// Statistical parameters for range estimation
			z := math.Log(float64(n))
			s := 0.5 * math.Exp(2*z/3)

			// Standard deviation for range width
			sd := 0.5 * math.Sqrt(z*s*(float64(n)-s)/float64(n))

			if i < n/2 {
				sd *= -1.0
			}

			// Estimate new boundaries
			// These could be wrong for non-uniform distributions
			// but the algorithm will still converge correctly
			newLeft := max(left, int(float64(k)-float64(i)*s/float64(n)+sd))
			newRight := min(right, int(float64(k)+float64(n-i)*s/float64(n)+sd))

			floydRivest(data, newLeft, newRight, k)
		}

		// Partitioning section
		// Core invariants during partitioning:
		// 1. Elements strictly < pivot end up left of final pivot position
		// 2. Elements strictly > pivot end up right of final pivot position
		// 3. Elements = pivot cluster around final pivot position
		i, j := left, right

		// Initial pivot selection and positioning
		// Using the larger value as pivot creates an important property:
		// - When we encounter elements equal to the original pivot value,
		//   they are guaranteed to be strictly less than our chosen pivot
		// - This removes ambiguity in the partitioning process and
		//   prevents infinite loops when many duplicates are present
		data.Swap(left, k)
		swap := data.Less(left, right)
		pivot := right
		if swap {
			data.Swap(left, right)
			pivot = left
		}

		// Main partitioning loop
		// Initial swap before moving pointers serves multiple purposes:
		// 1. Elements at i,j have been compared to pivot but not yet placed
		// 2. Swapping before moving pointers ensures we don't skip elements
		// 3. This ordering guarantees progress even when many elements equal pivot
		for i < j {
			data.Swap(i, j)
			i++
			j--

			// Forward and backward scans maintain strict invariants:
			// - All elements before i are strictly < pivot_value
			// - All elements after j are strictly > pivot_value
			// - Elements between i and j are yet to be classified
			for data.Less(i, pivot) {
				i++
			}
			for data.Less(pivot, j) {
				j--
			}
		}

		// Final pivot positioning requires careful handling because:
		// 1. When pointers cross (i ≥ j), we have three regions:
		//    [left...j] < pivot_value
		//    [j...i] = pivot_value
		//    [i...right] > pivot_value
		// 2. The pivot_value itself is still at either left or right
		// 3. It must end up at the boundary between < and = regions
		if swap {
			data.Swap(left, j)
		} else {
			j++
			data.Swap(right, j)
		}

		// Range reduction implements the key efficiency of selection:
		// 1. After partitioning, j is the exact count of elements ≤ pivot_value
		// 2. This gives us perfect information for reducing the search range:
		//    - If k ≤ j: k-th element must be in left portion
		//    - If k > j: k-th element must be in right portion
		// 3. The reduction is optimal because:
		//    - We never discard elements that could be the k-th element
		//    - We always discard elements that cannot be the k-th element
		if j <= k {
			left = j + 1
		}
		if k <= j {
			right = j - 1
		}
	}
}

// FloydRivestOrdered is a specialized version of FloydRivest that works with slices of
// ordered types (i.e. types that implement the cmp.Ordered interface).
func FloydRivestOrdered[T cmp.Ordered](data []T, k int) {
	n := len(data)
	if k < 1 || k > n {
		return
	}
	floydRivestOrdered(data, 0, n-1, k-1)
}

func floydRivestOrdered[T cmp.Ordered](data []T, left, right, k int) {
	for right > left {
		size := right - left

		if size > rangeNarrowingThreshold {
			n := size + 1
			i := k - left + 1

			z := math.Log(float64(n))
			s := 0.5 * math.Exp(2*z/3)
			sd := 0.5 * math.Sqrt(z*s*(float64(n)-s)/float64(n))

			if i < n/2 {
				sd *= -1.0
			}

			newLeft := max(left, int(float64(k)-float64(i)*s/float64(n)+sd))
			newRight := min(right, int(float64(k)+float64(n-i)*s/float64(n)+sd))

			floydRivestOrdered(data, newLeft, newRight, k)
		}

		i, j := left, right

		// Initial pivot selection and positioning
		data[left], data[k] = data[k], data[left]
		swap := data[left] < data[right]
		pivot := right
		if swap {
			data[left], data[right] = data[right], data[left]
			pivot = left
		}

		for i < j {
			data[i], data[j] = data[j], data[i]
			i++
			j--

			for data[i] < data[pivot] {
				i++
			}
			for data[pivot] < data[j] {
				j--
			}
		}

		if swap {
			data[left], data[j] = data[j], data[left]
		} else {
			j++
			data[right], data[j] = data[j], data[right]
		}

		if j <= k {
			left = j + 1
		}
		if k <= j {
			right = j - 1
		}
	}
}

// FloydRivestFunc is a generic version of FloydRivest that allows the caller to provide
// a custom comparison function to determine the order of elements.
func FloydRivestFunc[E any](data []E, k int, less func(a, b E) bool) {
	n := len(data)
	if k < 1 || k > n {
		return
	}
	floydRivestFunc(data, 0, n-1, k-1, less)
}

func floydRivestFunc[E any](data []E, left, right, k int, less func(a, b E) bool) {
	for right > left {
		size := right - left

		if size > rangeNarrowingThreshold {
			n := size + 1
			i := k - left + 1

			z := math.Log(float64(n))
			s := 0.5 * math.Exp(2*z/3)
			sd := 0.5 * math.Sqrt(z*s*(float64(n)-s)/float64(n))

			if i < n/2 {
				sd *= -1.0
			}

			newLeft := max(left, int(float64(k)-float64(i)*s/float64(n)+sd))
			newRight := min(right, int(float64(k)+float64(n-i)*s/float64(n)+sd))

			floydRivestFunc(data, newLeft, newRight, k, less)
		}

		i, j := left, right

		// Initial pivot selection and positioning
		data[left], data[k] = data[k], data[left]
		swap := less(data[left], data[right])
		pivot := right
		if swap {
			data[left], data[right] = data[right], data[left]
			pivot = left
		}

		for i < j {
			data[i], data[j] = data[j], data[i]
			i++
			j--

			for less(data[i], data[pivot]) {
				i++
			}
			for less(data[pivot], data[j]) {
				j--
			}
		}

		if swap {
			data[left], data[j] = data[j], data[left]
		} else {
			j++
			data[right], data[j] = data[j], data[right]
		}

		if j <= k {
			left = j + 1
		}
		if k <= j {
			right = j - 1
		}
	}
}
