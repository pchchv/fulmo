package helpers

import "math"

// HistogramData stores the information needed to represent
// the sizes of the keys and values as a histogram.
type HistogramData struct {
	Bounds         []float64
	Count          int64
	Min            int64
	Max            int64
	Sum            int64
	CountPerBucket []int64
}

// NewHistogramData returns a new instance of HistogramData with properly initialized fields.
func NewHistogramData(bounds []float64) *HistogramData {
	return &HistogramData{
		Bounds:         bounds,
		Max:            0,
		Min:            math.MaxInt64,
		CountPerBucket: make([]int64, len(bounds)+1),
	}
}
