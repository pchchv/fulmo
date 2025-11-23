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

// Clear reset the histogram.
// Helpful in situations where we need to reset the metrics.
func (histogram *HistogramData) Clear() {
	if histogram == nil {
		return
	}

	histogram.Count = 0
	histogram.CountPerBucket = make([]int64, len(histogram.Bounds)+1)
	histogram.Sum = 0
	histogram.Max = 0
	histogram.Min = math.MaxInt64
}

func (histogram *HistogramData) Copy() *HistogramData {
	if histogram == nil {
		return nil
	}

	return &HistogramData{
		Bounds:         append([]float64{}, histogram.Bounds...),
		CountPerBucket: append([]int64{}, histogram.CountPerBucket...),
		Count:          histogram.Count,
		Min:            histogram.Min,
		Max:            histogram.Max,
		Sum:            histogram.Sum,
	}
}
