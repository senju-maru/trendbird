package service

import (
	"math"
	"testing"
)

func TestZScoreService_Calculate_InsufficientData(t *testing.T) {
	svc := NewZScoreService()

	tests := []struct {
		name    string
		volumes []int32
	}{
		{"nil slice", nil},
		{"empty slice", []int32{}},
		{"single point", []int32{100}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.Calculate(tt.volumes, 100)
			if result != nil {
				t.Errorf("expected nil for insufficient data, got %+v", result)
			}
		})
	}
}

func TestZScoreService_Calculate_AllSameValues(t *testing.T) {
	svc := NewZScoreService()
	result := svc.Calculate([]int32{50, 50, 50, 50}, 50)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ZScore != 0 {
		t.Errorf("expected z-score 0 for zero stddev, got %f", result.ZScore)
	}
	if result.Status != ZScoreStatusStable {
		t.Errorf("expected Stable status, got %d", result.Status)
	}
	if result.ChangePercent != 0 {
		t.Errorf("expected 0%% change, got %f", result.ChangePercent)
	}
}

func TestZScoreService_Calculate_SpikeStatus(t *testing.T) {
	svc := NewZScoreService()

	// Historical: mean=100, stddev=10 (approx)
	// Current: 140 → z = (140 - 100) / 10 = 4.0 → Spike
	historical := []int32{90, 110, 90, 110, 90, 110, 90, 110}
	result := svc.Calculate(historical, 140)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Status != ZScoreStatusSpike {
		t.Errorf("expected Spike status, got %d", result.Status)
	}
	if result.ZScore < 3.0 {
		t.Errorf("expected z-score >= 3.0, got %f", result.ZScore)
	}
}

func TestZScoreService_Calculate_RisingStatus(t *testing.T) {
	svc := NewZScoreService()

	// Historical: mean=100, stddev=10
	// Current: 125 → z = (125 - 100) / 10 = 2.5 → Rising
	historical := []int32{90, 110, 90, 110, 90, 110, 90, 110}
	result := svc.Calculate(historical, 125)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Status != ZScoreStatusRising {
		t.Errorf("expected Rising status, got %d", result.Status)
	}
	if result.ZScore < 2.0 || result.ZScore >= 3.0 {
		t.Errorf("expected z-score in [2.0, 3.0), got %f", result.ZScore)
	}
}

func TestZScoreService_Calculate_StableStatus(t *testing.T) {
	svc := NewZScoreService()

	// Historical: mean=100, stddev=10
	// Current: 105 → z = (105 - 100) / 10 = 0.5 → Stable
	historical := []int32{90, 110, 90, 110, 90, 110, 90, 110}
	result := svc.Calculate(historical, 105)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Status != ZScoreStatusStable {
		t.Errorf("expected Stable status, got %d", result.Status)
	}
	if result.ZScore >= 2.0 {
		t.Errorf("expected z-score < 2.0, got %f", result.ZScore)
	}
}

func TestZScoreService_Calculate_ChangePercent(t *testing.T) {
	svc := NewZScoreService()

	// Historical: mean=100
	// Current: 150 → change = ((150 - 100) / 100) * 100 = 50%
	historical := []int32{80, 120, 80, 120}
	result := svc.Calculate(historical, 150)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if math.Abs(result.ChangePercent-50.0) > 0.01 {
		t.Errorf("expected ~50%% change, got %f", result.ChangePercent)
	}
	if math.Abs(result.Mean-100.0) > 0.01 {
		t.Errorf("expected mean ~100, got %f", result.Mean)
	}
}

func TestZScoreService_Calculate_ZeroMean(t *testing.T) {
	svc := NewZScoreService()

	// All zeros → mean=0, stddev=0 → z=0, change=0
	result := svc.Calculate([]int32{0, 0, 0}, 0)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ZScore != 0 {
		t.Errorf("expected z-score 0, got %f", result.ZScore)
	}
	if result.ChangePercent != 0 {
		t.Errorf("expected 0%% change, got %f", result.ChangePercent)
	}
}
