package service

import "math"

// ZScoreResult holds the result of a z-score calculation.
type ZScoreResult struct {
	ZScore        float64
	Mean          float64
	StdDev        float64
	CurrentVolume int32
	Status        ZScoreStatus
	ChangePercent float64
}

// ZScoreStatus represents the classification of a z-score value.
type ZScoreStatus int

const (
	ZScoreStatusSpike  ZScoreStatus = 1
	ZScoreStatusRising ZScoreStatus = 2
	ZScoreStatusStable ZScoreStatus = 3
)

// ZScoreService calculates z-scores for topic volume data.
type ZScoreService struct{}

// NewZScoreService creates a new ZScoreService.
func NewZScoreService() *ZScoreService {
	return &ZScoreService{}
}

// Calculate computes z-score and status from historical volumes and the current volume.
// Returns nil if there are fewer than 2 historical data points (insufficient baseline).
func (s *ZScoreService) Calculate(historicalVolumes []int32, currentVolume int32) *ZScoreResult {
	if len(historicalVolumes) < 2 {
		return nil
	}

	mean := calcMean(historicalVolumes)
	stddev := calcStdDev(historicalVolumes, mean)

	var z float64
	if stddev == 0 {
		z = 0
	} else {
		z = (float64(currentVolume) - mean) / stddev
	}

	var changePercent float64
	if mean != 0 {
		changePercent = ((float64(currentVolume) - mean) / mean) * 100
	}

	status := ZScoreStatusStable
	if z >= 3.0 {
		status = ZScoreStatusSpike
	} else if z >= 2.0 {
		status = ZScoreStatusRising
	}

	return &ZScoreResult{
		ZScore:        z,
		Mean:          mean,
		StdDev:        stddev,
		CurrentVolume: currentVolume,
		Status:        status,
		ChangePercent: changePercent,
	}
}

func calcMean(values []int32) float64 {
	var sum float64
	for _, v := range values {
		sum += float64(v)
	}
	return sum / float64(len(values))
}

func calcStdDev(values []int32, mean float64) float64 {
	var sumSquares float64
	for _, v := range values {
		diff := float64(v) - mean
		sumSquares += diff * diff
	}
	return math.Sqrt(sumSquares / float64(len(values)))
}
