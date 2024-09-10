package kth

import (
	"cmp"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestSelect(t *testing.T) {
	testCases := []struct {
		name  string
		input []int
		k     int
	}{
		{"Small sorted", []int{1, 2, 3, 4, 5}, 3},
		{"Small reversed", []int{5, 4, 3, 2, 1}, 3},
		{"Medium random", []int{3, 7, 2, 1, 4, 6, 5, 8, 9}, 5},
		{"Large random", []int{15, 3, 9, 8, 5, 2, 7, 1, 6, 13, 11, 12, 10, 4, 14}, 8},
		{"All equal", []int{1, 1, 1, 1, 1}, 3},
		{"Mostly equal", []int{2, 2, 2, 2, 1, 2, 2, 3, 2, 2}, 6},
		{"Single element", []int{42}, 1},
		{"Two elements", []int{2, 1}, 1},
	}

	for _, tc := range testCases {
		t.Run("PDQSelect/"+tc.name, func(t *testing.T) {
			testSelect(t, tc.input, 0, len(tc.input), tc.k, "PDQSelect", func(input []int, a, b, k int) {
				PDQSelect(sort.IntSlice(input), k)
			})
		})

		t.Run("PDQSelectOrdered/"+tc.name, func(t *testing.T) {
			testSelect(t, tc.input, 0, len(tc.input), tc.k, "PDQSelectOrdered", func(input []int, a, b, k int) {
				PDQSelectOrdered(input, k)
			})
		})

		t.Run("PDQSelectFunc/"+tc.name, func(t *testing.T) {
			testSelect(t, tc.input, 0, len(tc.input), tc.k, "PDQSelectFunc", func(input []int, a, b, k int) {
				PDQSelectFunc(input, k, cmp.Less)
			})
		})

		t.Run("FloydRivest/"+tc.name, func(t *testing.T) {
			testSelect(t, tc.input, 0, len(tc.input), tc.k, "FloydRivest", func(input []int, a, b, k int) {
				FloydRivest(sort.IntSlice(input), k)
			})
		})

		t.Run("FloydRivestOrdered/"+tc.name, func(t *testing.T) {
			testSelect(t, tc.input, 0, len(tc.input), tc.k, "FloydRivestOrdered", func(input []int, a, b, k int) {
				FloydRivestOrdered(input, k)
			})
		})

		t.Run("FloydRivestFunc/"+tc.name, func(t *testing.T) {
			testSelect(t, tc.input, 0, len(tc.input), tc.k, "FloydRivestFunc", func(input []int, a, b, k int) {
				FloydRivestFunc(input, k, cmp.Less)
			})
		})
	}
}

func FuzzSelect(f *testing.F) {
	f.Add(encodeInts(1, 4), uint16(1), uint16(0), uint16(2))
	f.Add(encodeInts(1, 4, 2), uint16(2), uint16(0), uint16(3))
	f.Add(encodeInts(1, 4, 2, 1), uint16(2), uint16(1), uint16(4))
	f.Add(encodeInts(1, 2, 3, 4, 5), uint16(3), uint16(0), uint16(5))
	f.Add(encodeInts(5, 4, 3, 2, 1), uint16(2), uint16(1), uint16(4))
	f.Add(encodeInts(1, 1, 1, 1, 1), uint16(1), uint16(0), uint16(5))
	f.Add(encodeInts(1, 4, 7, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1), uint16(7), uint16(3), uint16(12))
	f.Add(encodeInts(254, 4, 7, 2, 0, 0, 0, 255, 0, 0, 0, 0, 0, 0, 0, 253), uint16(7), uint16(0), uint16(16))
	f.Add(encodeInts(0, 0, 0, 0, 0, 0, 0, 255, 0, 0, 0, 0, 0, 0, 0, 253, 0, 0, 0, 0, 0, 0), uint16(0), uint16(20), uint16(12))

	now := time.Now().UnixNano()
	rng := rand.New(rand.NewPCG(uint64(now), uint64(now>>32)))

	for _, dist := range []Distribution{
		UniformDist,
		NormalDist,
		ZipfDist,
		ConstantDist,
		BimodalDist,
	} {
		for _, order := range []Ordering{
			RandomOrder,
			SortedOrder,
			ReversedOrder,
			MostlySorted,
			PushFrontOrder,
			PushMiddleOrder,
		} {
			for _, size := range []int{10, 100, 1000} {
				data := genDistribution(rng, size, dist)
				applyOrdering(rng, data, order)
				encodedData := encodeInts(data...)
				f.Add(encodedData, uint16(size/2), uint16(0), uint16(size))
				f.Add(encodedData, uint16(1), uint16(0), uint16(size))
				f.Add(encodedData, uint16(size), uint16(0), uint16(size))
			}
		}
	}

	f.Fuzz(func(t *testing.T, data []byte, k, a, b uint16) {
		if len(data)%4 != 0 {
			return // Skip if data length is not a multiple of 4
		}

		// Convert byte slice to int slice
		input := decodeInts(data)

		if len(input) == 0 {
			return // Skip empty slices
		}

		k = k % uint16(len(input))
		if k == 0 {
			k++
		}

		testSelect(t, input, 0, len(input), int(k), "PDQSelect", func(slice []int, a, b, k int) {
			PDQSelect(sort.IntSlice(slice), k)
		})

		testSelect(t, input, 0, len(input), int(k), "PDQSelectFunc", func(slice []int, a, b, k int) {
			PDQSelectFunc(slice, k, cmp.Less)
		})

		testSelect(t, input, 0, len(input), int(k), "PDQSelectOrdered", func(slice []int, a, b, k int) {
			PDQSelectOrdered(slice, k)
		})

		testSelect(t, input, 0, len(input), int(k), "pdqselect", func(slice []int, a, b, k int) {
			pdqselect(sort.IntSlice(slice), 0, len(slice), k-1, 0)
		})

		testSelect(t, input, 0, len(input), int(k), "pdqselectOrdered", func(slice []int, a, b, k int) {
			pdqselectOrdered(slice, 0, len(slice), k-1, 0)
		})

		testSelect(t, input, 0, len(input), int(k), "pdqselectFunc", func(slice []int, a, b, k int) {
			pdqselectFunc(slice, 0, len(slice), k-1, 0, cmp.Less)
		})

		testSelect(t, input, 0, len(input), int(k), "FloydRivest", func(slice []int, a, b, k int) {
			FloydRivest(sort.IntSlice(slice), k)
		})

		testSelect(t, input, 0, len(input), int(k), "FloydRivestOrdered", func(slice []int, a, b, k int) {
			FloydRivestOrdered(slice, k)
		})

		testSelect(t, input, 0, len(input), int(k), "FloydRivestFunc", func(slice []int, a, b, k int) {
			FloydRivestFunc(slice, k, cmp.Less)
		})

		testSelect(t, input, 0, len(input), int(k), "floydRivestSelect", func(slice []int, a, b, k int) {
			floydRivest(sort.IntSlice(slice), 0, len(slice)-1, k-1)
		})

		testSelect(t, input, 0, len(input), int(k), "floydRivestOrdered", func(slice []int, a, b, k int) {
			floydRivestOrdered(slice, 0, len(slice)-1, k-1)
		})

		testSelect(t, input, 0, len(input), int(k), "floydRivestFunc", func(slice []int, a, b, k int) {
			floydRivestFunc(slice, 0, len(slice)-1, k-1, cmp.Less)
		})

		// Ensure a, b, and k are within bounds
		a = a % uint16(len(input))
		b = b % uint16(len(input))
		if a > b {
			a, b = b, a // Ensure a < b
		} else if a == b {
			b++ // Ensure b is at least 1 greater than a
		}

		n := b - a
		k %= n
		if k == 0 {
			k++
		}

		testSelect(t, input, int(a), int(b), int(k), "heapSelect", func(slice []int, a, b, k int) {
			heapSelect(sort.IntSlice(slice), a, b, k-1)
		})

		testSelect(t, input, int(a), int(b), int(k), "heapSelectOrdered", func(slice []int, a, b, k int) {
			heapSelectOrdered(slice, a, b, k-1)
		})

		testSelect(t, input, int(a), int(b), int(k), "heapSelectFunc", func(slice []int, a, b, k int) {
			heapSelectFunc(slice, a, b, k-1, cmp.Less)
		})
	})
}

func encodeInts(ints ...int) []byte {
	buf := make([]byte, len(ints)*4)
	for i, v := range ints {
		binary.BigEndian.PutUint32(buf[i*4:], uint32(v))
	}
	return buf
}

func decodeInts(data []byte) []int {
	ints := make([]int, len(data)/4)
	for i := range ints {
		ints[i] = int(binary.BigEndian.Uint32(data[i*4:]))
	}
	return ints
}

func testSelect(t *testing.T, input []int, a, b, k int, name string, selectFunc func([]int, int, int, int)) {
	t.Helper()

	// Create a copy for sorting
	sorted := make([]int, len(input))
	copy(sorted, input)
	slices.Sort(sorted[a:b])

	// Create another copy for selecting
	output := make([]int, len(input))
	copy(output, input)

	// Run pdqselect
	selectFunc(output, a, b, k)

	// Assert that the kth element is the expected one
	if output[a+k-1] != sorted[a+k-1] {
		t.Errorf("%s(a=%d, b=%d, k=%d, n=%d): k-th element (%d) does not match sorted input (%d)\ninput:  %v\nsorted: %v\noutput: %v",
			name, a, b, k, b-a, output[a+k-1], sorted[a+k-1], input, sorted, output)
	}

	// Get the first k elements, sort them, and compare with sorted slice
	firstK := make([]int, k)
	copy(firstK, output[a:a+k])
	slices.Sort(firstK)
	for i := range firstK {
		if firstK[i] != sorted[a+i] {
			t.Errorf("%s(a=%d, b=%d, k=%d, n=%d): sorted output element at index %d (%d) does not match sorted input (%d)\ninput:  %v\nsorted: %v\noutput: %v\nfirstK: %v",
				name, a, b, k, b-a, i, firstK[i], sorted[a+i], input, sorted, output, firstK)
		}
	}

	// Check if all elements before and including k are smaller or equal, and all elements after k are larger or equal
	for i := a; i < a+k; i++ {
		if output[i] > sorted[a+k-1] {
			t.Errorf("%s(a=%d, b=%d, k=%d, n=%d): element at index %d (%d) is larger than k-th element (%d)\ninput:  %v\nsorted: %v\noutput: %v",
				name, a, b, k, b-a, i, output[i], sorted[a+k-1], input, sorted, output)
		}
	}

	for i := a + k - 1; i < b; i++ {
		if output[i] < output[a+k-1] {
			t.Errorf("%s(a=%d, b=%d, k=%d, n=%d): element at index %d (%d) is smaller than k-th element (%d)\ninput:  %v\nsorted: %v\noutput: %v",
				name, a, b, k, b-a, i, output[i], output[a+k-1], input, sorted, output)
		}
	}
}

func BenchmarkSelect(b *testing.B) {
	rng := rand.New(rand.NewPCG(42, 42)) // Deterministic random number generator

	// Test parameters
	const n = 10_000_000
	ks := []int{1, 100, n / 2, n - 100, n - 1}
	distributions := []Distribution{
		UniformDist,
		NormalDist,
		ZipfDist,
		ConstantDist,
		BimodalDist,
	}
	orderings := []Ordering{
		RandomOrder,
		SortedOrder,
		ReversedOrder,
		MostlySorted,
		PushFrontOrder,
		PushMiddleOrder,
	}

	type benchCase struct {
		name string
		fn   func([]int, int)
	}

	cases := []benchCase{
		// Sortint
		{"PDQSort", func(data []int, _ int) { sort.Ints(data) }},
		{"PDQSortOrdered", func(data []int, _ int) { slices.Sort(data) }},
		{"PDQSortFunc", func(data []int, _ int) { slices.SortFunc(data, cmp.Compare) }},
		// Selection
		{"PDQSelect", func(data []int, k int) { PDQSelect(sort.IntSlice(data), k) }},
		{"PDQSelectOrdered", func(data []int, k int) { PDQSelectOrdered(data, k) }},
		{"PDQSelectFunc", func(data []int, k int) { PDQSelectFunc(data, k, cmp.Less) }},
		{"FloydRivestSelect", func(data []int, k int) { FloydRivest(sort.IntSlice(data), k) }},
		{"FloydRivestSelectOrdered", func(data []int, k int) { FloydRivestOrdered(data, k) }},
		{"FloydRivestSelectFunc", func(data []int, k int) { FloydRivestFunc(data, k, cmp.Less) }},
		// Partial sorting
		{"PDQPartialSort", func(data []int, k int) {
			PDQSelect(sort.IntSlice(data), k)
			sort.Ints(data[:k])
		}},
		{"PDQPartialSortOrdered", func(data []int, k int) {
			PDQSelectOrdered(data, k)
			slices.Sort(data[:k])
		}},
		{"PDQPartialSortFunc", func(data []int, k int) {
			PDQSelectFunc(data, k, cmp.Less)
			slices.SortFunc(data[:k], cmp.Compare)
		}},
		{"FloydRivestPartialSort", func(data []int, k int) {
			FloydRivest(sort.IntSlice(data), k)
			sort.Ints(data[:k])
		}},
		{"FloydRivestPartialSortOrdered", func(data []int, k int) {
			FloydRivestOrdered(data, k)
			slices.Sort(data[:k])
		}},
		{"FloydRivestPartialSortFunc", func(data []int, k int) {
			FloydRivestFunc(data, k, cmp.Less)
			slices.SortFunc(data[:k], cmp.Compare)
		}},
	}

	// Main benchmark loops
	for _, k := range ks {
		for _, dist := range distributions {
			for _, order := range orderings {
				data := genDistribution(rng, n, dist)
				applyOrdering(rng, data, order)

				for _, bc := range cases {
					name := fmt.Sprintf("fn=%s/n=%d/k=%d/dist=%s/order=%s", bc.name, n, k, dist, order)
					b.Run(name, func(b *testing.B) {
						dataCopy := make([]int, len(data))

						b.ReportAllocs()
						b.ResetTimer()
						for i := 0; i < b.N; i++ {
							copy(dataCopy, data)
							bc.fn(dataCopy, k)
						}
					})
				}
			}
		}
	}
}

func TestGenDistribution(t *testing.T) {
	now := time.Now().UnixNano()
	seeds := []uint64{uint64(now), uint64(now >> 32)}
	rng := rand.New(rand.NewPCG(seeds[0], seeds[1]))

	t.Logf("Seeds: %v", seeds)

	type check struct {
		name string
		fn   func([]int) error
	}

	tests := []struct {
		dist    Distribution
		must    check
		mustNot []check
	}{
		{UniformDist, check{"uniform", checkUniform}, []check{
			{"normal", checkNormal},
			{"zipf", checkZipf},
			{"constant", checkConstant},
			{"bimodal", checkBimodal},
		}},
		{NormalDist, check{"normal", checkNormal}, []check{
			{"zipf", checkZipf},
			{"constant", checkConstant},
			{"bimodal", checkBimodal},
			{"uniform", checkUniform},
		}},
		{ZipfDist, check{"zipf", checkZipf}, []check{
			{"normal", checkNormal},
			{"constant", checkConstant},
			{"bimodal", checkBimodal},
			{"uniform", checkUniform},
		}},
		{ConstantDist, check{"constant", checkConstant}, []check{
			{"normal", checkNormal},
			{"zipf", checkZipf},
			{"bimodal", checkBimodal},
			{"uniform", checkUniform},
		}},
		{BimodalDist, check{"bimodal", checkBimodal}, []check{
			{"normal", checkNormal},
			{"zipf", checkZipf},
			{"constant", checkConstant},
			{"uniform", checkUniform},
		}},
	}

	const size = 10000
	for _, tt := range tests {
		name := fmt.Sprintf("size=%d/dist=%s", size, tt.dist)
		t.Run(name, func(t *testing.T) {
			dist := genDistribution(rng, size, tt.dist)

			var sb strings.Builder
			plotDistribution(&sb, string(tt.dist), dist)
			t.Log(sb.String())

			if err := tt.must.fn(dist); err != nil {
				t.Error(err)
			}

			for _, mustNot := range tt.mustNot {
				if err := mustNot.fn(dist); err == nil {
					t.Errorf("%v distribution incorrectly passed %v check", tt.dist, mustNot.name)
				}
			}
		})
	}
}

type (
	Distribution string
	Ordering     string
)

const (
	UniformDist  Distribution = "uniform"
	NormalDist   Distribution = "normal"
	ZipfDist     Distribution = "zipf"
	ConstantDist Distribution = "constant"
	BimodalDist  Distribution = "bimodal"
)

const (
	RandomOrder     Ordering = "random"
	SortedOrder     Ordering = "sorted"
	ReversedOrder   Ordering = "reversed"
	MostlySorted    Ordering = "mostly_sorted"
	PushFrontOrder  Ordering = "push_front"
	PushMiddleOrder Ordering = "push_middle"
)

func plotDistribution(w io.Writer, name string, dist []int) {
	// Use 20 bins
	const bins = 20
	min, max := slices.Min(dist), slices.Max(dist)
	range_ := max - min
	if range_ == 0 {
		range_ = 1 // Prevent division by zero for constant distribution
	}

	// Count frequencies in bins
	counts := make([]int, bins)
	for _, v := range dist {
		bin := int(float64(v-min) / float64(range_) * float64(bins-1))
		if bin == bins {
			bin-- // Handle edge case of maximum value
		}
		counts[bin]++
	}

	// Find max count for scaling
	maxCount := 0
	for _, c := range counts {
		if c > maxCount {
			maxCount = c
		}
	}

	// Plot distribution
	const width = 40
	fmt.Fprintf(w, "%s Distribution (n=%d):\n", name, len(dist))
	fmt.Fprintln(w, "bin count")
	fmt.Fprintln(w, "-----------------")
	for i, count := range counts {
		binStart := min + (range_*i)/bins
		bars := int(float64(count) / float64(maxCount) * width)
		fmt.Fprintf(w, "%5d %5d %s\n", binStart, count, strings.Repeat("█", bars))
	}
}

func checkUniform(data []int) error {
	// Calculate histogram
	const bins = 20
	counts := make([]float64, bins)
	min, max := slices.Min(data), slices.Max(data)
	range_ := max - min

	if range_ == 0 {
		return fmt.Errorf("all values are identical - not uniform")
	}

	// Fill histogram
	for _, v := range data {
		bin := int(float64(v-min) / float64(range_) * float64(bins-1))
		if bin >= bins {
			bin = bins - 1
		}
		counts[bin]++
	}

	// Normalize counts to get density
	total := float64(len(data))
	for i := range counts {
		counts[i] /= total
	}

	// Calculate mean density
	mean := 1.0 / float64(bins)

	// Check for peaks or valleys
	// For uniform distribution, no bin should deviate too far from mean
	maxPeak := 2.0 * mean
	minValley := 0.5 * mean

	peaks := 0
	valleys := 0
	for _, density := range counts {
		if density > maxPeak {
			peaks++
		}
		if density < minValley {
			valleys++
		}
	}

	// Uniform distribution should have few significant peaks or valleys
	if peaks > 1 || valleys > 1 {
		return fmt.Errorf("distribution has too many peaks (%d) or valleys (%d)", peaks, valleys)
	}

	return nil
}

func checkNormal(data []int) error {
	n := float64(len(data))
	mean := 0.0
	for _, v := range data {
		mean += float64(v)
	}
	mean /= n

	variance := 0.0
	for _, v := range data {
		diff := float64(v) - mean
		variance += diff * diff
	}
	variance /= n
	stdDev := math.Sqrt(variance)

	within1Sigma := 0
	within2Sigma := 0
	within3Sigma := 0
	skewness := 0.0
	kurtosis := 0.0

	for _, v := range data {
		z := (float64(v) - mean) / stdDev
		if math.Abs(z) <= 1.0 {
			within1Sigma++
		}
		if math.Abs(z) <= 2.0 {
			within2Sigma++
		}
		if math.Abs(z) <= 3.0 {
			within3Sigma++
		}

		z3 := z * z * z
		z4 := z3 * z
		skewness += z3
		kurtosis += z4
	}

	p1 := float64(within1Sigma) / n
	p2 := float64(within2Sigma) / n
	p3 := float64(within3Sigma) / n

	skewness /= n
	kurtosis = kurtosis/n - 3.0

	if math.Abs(p1-0.68) > 0.05 {
		return fmt.Errorf("68%% rule violated: %.2f", p1)
	}
	if math.Abs(p2-0.95) > 0.05 {
		return fmt.Errorf("95%% rule violated: %.2f", p2)
	}
	if math.Abs(p3-0.997) > 0.05 {
		return fmt.Errorf("99.7%% rule violated: %.2f", p3)
	}
	if math.Abs(skewness) > 0.5 {
		return fmt.Errorf("skewness too high: %.2f", skewness)
	}
	if math.Abs(kurtosis) > 2.0 {
		return fmt.Errorf("excess kurtosis too high: %.2f", kurtosis)
	}

	return nil
}

func checkZipf(data []int) error {
	type pair struct {
		freq int
		rank int
	}

	// Count frequencies
	freq := make(map[int]int)
	for _, v := range data {
		freq[v]++
	}

	// Convert to ranked pairs
	ranks := make([]pair, 0, len(freq))
	for _, f := range freq {
		ranks = append(ranks, pair{freq: f})
	}
	slices.SortFunc(ranks, func(a, b pair) int {
		return cmp.Compare(b.freq, a.freq) // descending
	})

	if len(ranks) < 2 {
		return fmt.Errorf("insufficient distinct values for Zipf distribution")
	}

	// Check the initial frequency drop-off
	initialDropOff := float64(ranks[0].freq) / float64(ranks[1].freq)
	if initialDropOff < 1.5 {
		return fmt.Errorf("initial frequency drop-off %.2f too small for Zipf distribution", initialDropOff)
	}

	// Convert to log-log coordinates
	points := make([][2]float64, len(ranks))
	for i := range ranks {
		points[i] = [2]float64{
			math.Log(float64(i + 1)),         // log(rank)
			math.Log(float64(ranks[i].freq)), // log(frequency)
		}
	}

	// Linear regression
	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(points))

	for _, p := range points {
		sumX += p[0]
		sumY += p[1]
		sumXY += p[0] * p[1]
		sumX2 += p[0] * p[0]
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// Calculate R² to measure fit
	meanY := sumY / n
	var ssTotal, ssResidual float64

	for _, p := range points {
		fitted := slope*p[0] + intercept
		ssResidual += (p[1] - fitted) * (p[1] - fitted)
		ssTotal += (p[1] - meanY) * (p[1] - meanY)
	}

	r2 := 1 - (ssResidual / ssTotal)

	if slope >= 0 {
		return fmt.Errorf("slope %.2f is not negative", slope)
	}

	// Adjust R² threshold based on sample size
	minR2 := 0.9
	if len(data) > 100 {
		minR2 = math.Max(0.85, 0.9-0.05*math.Log10(float64(len(data))/100))
	}

	if r2 < minR2 {
		return fmt.Errorf("R² value %.2f indicates poor power law fit", r2)
	}

	// Allow for a wider range of slopes
	if slope < -3.0 || slope > -0.3 {
		return fmt.Errorf("slope %.2f outside typical Zipf-like range (-3.0 to -0.3)", slope)
	}

	return nil
}

func checkConstant(data []int) error {
	first := data[0]
	for i, v := range data {
		if v != first {
			return fmt.Errorf("value at index %d differs: %d != %d", i, v, first)
		}
	}
	return nil
}

func checkBimodal(data []int) error {
	sorted := make([]float64, len(data))
	for i, v := range data {
		sorted[i] = float64(v)
	}
	slices.Sort(sorted)

	// Calculate mean and standard deviation
	n := float64(len(sorted))
	mean := 0.0
	for _, v := range sorted {
		mean += v
	}
	mean /= n

	variance := 0.0
	for _, v := range sorted {
		diff := v - mean
		variance += diff * diff
	}
	variance /= n
	std := math.Sqrt(variance)

	// Check if all values are the same (constant distribution)
	if std == 0 {
		return fmt.Errorf("found 0 peaks: constant distribution")
	}

	// Use Scott's rule for bandwidth selection
	bandwidth := 1.06 * std * math.Pow(n, -1.0/5.0)

	// Estimate density at regular intervals
	intervals := 100
	min, max := sorted[0], sorted[len(sorted)-1]
	range_ := max - min
	densities := make([]float64, intervals)

	for i := range densities {
		x := min + (float64(i)/float64(intervals-1))*range_
		densities[i] = kernelDensity(x, sorted, bandwidth)
	}

	// Find peaks (local maxima)
	peaks := []int{}
	for i := 1; i < len(densities)-1; i++ {
		if densities[i] > densities[i-1] && densities[i] > densities[i+1] {
			// Check if it's a significant peak (at least 20% of max density)
			if densities[i] > 0.2*slices.Max(densities) {
				peaks = append(peaks, i)
			}
		}
	}

	// Merge peaks that are too close
	if len(peaks) > 0 {
		distinctPeaks := []int{peaks[0]}
		minPeakDistance := intervals / 5 // At least 20% of range apart

		for i := 1; i < len(peaks); i++ {
			if peaks[i]-distinctPeaks[len(distinctPeaks)-1] > minPeakDistance {
				distinctPeaks = append(distinctPeaks, peaks[i])
			}
		}

		if len(distinctPeaks) != 2 {
			return fmt.Errorf("found %d significant peaks, expected 2", len(distinctPeaks))
		}

		// Check that peaks are of similar height (within 50% of each other)
		peak1 := densities[distinctPeaks[0]]
		peak2 := densities[distinctPeaks[1]]
		ratio := peak1 / peak2
		if ratio < 0.5 || ratio > 2.0 {
			return fmt.Errorf("peak heights too different: ratio %.2f", ratio)
		}

		// Check that there's a significant valley between peaks
		valleyPoint := (distinctPeaks[0] + distinctPeaks[1]) / 2
		valleyHeight := densities[valleyPoint]
		minPeakHeight := math.Min(peak1, peak2)
		if valleyHeight > 0.7*minPeakHeight {
			return fmt.Errorf("valley not deep enough between peaks")
		}
	} else {
		return fmt.Errorf("no significant peaks found")
	}

	return nil
}

func kernelDensity(x float64, data []float64, bandwidth float64) float64 {
	density := 0.0
	n := float64(len(data))

	for _, xi := range data {
		z := (x - xi) / bandwidth
		density += math.Exp(-0.5 * z * z)
	}

	return density / (bandwidth * math.Sqrt(2*math.Pi) * n)
}

func genDistribution(rng *rand.Rand, size int, dist Distribution) []int {
	slice := make([]int, size)

	switch dist {
	case UniformDist:
		for i := range slice {
			slice[i] = rng.IntN(size)
		}

	case NormalDist:
		mean := size / 2
		stdDev := float64(size) / 6.0
		for i := range slice {
			slice[i] = int(math.Round(rng.NormFloat64()*stdDev + float64(mean)))
		}

	case ZipfDist:
		zipf := rand.NewZipf(rng, 1.5, 1.0, uint64(size-1))
		for i := range slice {
			slice[i] = int(zipf.Uint64())
		}

	case ConstantDist:
		val := rng.Int()
		for i := range slice {
			slice[i] = val
		}

	case BimodalDist:
		peak1 := size / 4
		peak2 := 3 * size / 4
		stdDev := float64(size) / 16.0
		for i := range slice {
			peak := peak1
			if rng.Float64() >= 0.5 {
				peak = peak2
			}
			slice[i] = int(math.Round(rng.NormFloat64()*stdDev + float64(peak)))
		}

	default:
		panic("unknown distribution")
	}

	return slice
}

func applyOrdering(rng *rand.Rand, slice []int, order Ordering) {
	switch order {
	case RandomOrder:
		rng.Shuffle(len(slice), func(i, j int) {
			slice[i], slice[j] = slice[j], slice[i]
		})

	case SortedOrder:
		slices.Sort(slice)

	case ReversedOrder:
		slices.SortFunc(slice, func(a, b int) int { return cmp.Compare(b, a) })

	case MostlySorted:
		slices.Sort(slice)
		// Shuffle about 10% of the elements
		swaps := len(slice) / 10
		for i := 0; i < swaps; i++ {
			j := rng.IntN(len(slice))
			k := rng.IntN(len(slice))
			slice[j], slice[k] = slice[k], slice[j]
		}

	case PushFrontOrder:
		// Sort first to establish relative ordering
		temp := make([]int, len(slice))
		copy(temp, slice)
		slices.Sort(temp)

		// Move smallest to end, shift everything else left
		smallest := temp[0]
		copy(slice, temp[1:]) // Copy all but smallest
		slice[len(slice)-1] = smallest

	case PushMiddleOrder:
		// Sort first to establish relative ordering
		temp := make([]int, len(slice))
		copy(temp, slice)
		slices.Sort(temp)

		// Move middle value to end, preserving order of others
		mid := len(temp) / 2
		midVal := temp[mid]
		copy(slice, temp[:mid])         // Copy before middle
		copy(slice[mid:], temp[mid+1:]) // Copy after middle
		slice[len(slice)-1] = midVal

	default:
		panic("unknown ordering")
	}
}
