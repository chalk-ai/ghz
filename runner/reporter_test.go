package runner

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReport_MarshalJSON(t *testing.T) {
	z, _ := time.Parse(time.RFC822Z, "02 Jan 06 15:04 -0700")
	r := &Report{
		Date:    z,
		Count:   1000,
		Total:   time.Duration(10) * time.Second,
		Average: time.Duration(500) * time.Millisecond,
		Fastest: time.Duration(10) * time.Millisecond,
		Slowest: time.Duration(1000) * time.Millisecond,
		Rps:     34567.89,
	}

	json, err := json.Marshal(&r)
	assert.NoError(t, err)

	expected := `{"date":"2006-01-02T15:04:00-07:00","options":{"insecure":false,"load-schedule":"","load-start":0,"load-end":0,"load-step":0,"load-step-duration":0,"load-max-duration":0,"concurrency-schedule":"","concurrency-start":0,"concurrency-end":0,"concurrency-step":0,"concurrency-step-duration":0,"concurrency-max-duration":0,"binary":false,"CPUs":0},"count":1000,"total":10000000000,"average":500000000,"fastest":10000000,"slowest":1000000000,"rps":34567.89,"errorDistribution":null,"statusCodeDistribution":null,"latencyDistribution":null,"histogram":null,"details":null}`
	assert.Equal(t, expected, string(json))
}

func TestReport_CorrectDetails(t *testing.T) {
	callResultsChan := make(chan *callResult)
	sampledResultsChan := make(chan *sampledResult)
	config, _ := NewConfig("call", "host")
	reporter := newReporter(callResultsChan, sampledResultsChan, config)

	go reporter.Run()

	cr1 := callResult{
		status:    "OK",
		duration:  time.Millisecond * 100,
		err:       nil,
		timestamp: time.Now(),
	}
	callResultsChan <- &cr1
	cr2 := callResult{
		status:    "DeadlineExceeded",
		duration:  time.Millisecond * 500,
		err:       context.DeadlineExceeded,
		timestamp: time.Now(),
	}
	callResultsChan <- &cr2

	close(callResultsChan)
	close(sampledResultsChan)
	<-reporter.done
	report := reporter.Finalize("stop reason", time.Second)

	assert.Equal(t, 2, len(report.Details))
	assert.Equal(t, ResultDetail{Error: "", Latency: cr1.duration, Status: cr1.status, Timestamp: cr1.timestamp}, report.Details[0])
	assert.Equal(t, ResultDetail{Error: cr2.err.Error(), Latency: cr2.duration, Status: cr2.status, Timestamp: cr2.timestamp}, report.Details[1])
}

func TestLatencies_WithoutP99_9(t *testing.T) {
	// Create 1000 latency values (0.001s to 1.0s)
	latencies := make([]float64, 1000)
	for i := 0; i < 1000; i++ {
		latencies[i] = float64(i+1) / 1000.0
	}

	// Test without p99.9
	result := Latencies(latencies, false)

	// Should have 7 percentiles (10, 25, 50, 75, 90, 95, 99)
	assert.Equal(t, 7, len(result))

	// Verify percentiles are correct
	percentiles := []float64{10, 25, 50, 75, 90, 95, 99}
	for i, ld := range result {
		assert.Equal(t, percentiles[i], ld.Percentage)
	}

	// Verify 99th percentile exists
	assert.Equal(t, float64(99), result[6].Percentage)
	// Verify p99.9 does not exist
	for _, ld := range result {
		assert.NotEqual(t, 99.9, ld.Percentage)
	}
}

func TestLatencies_WithP99_9(t *testing.T) {
	// Create 1000 latency values (0.001s to 1.0s)
	latencies := make([]float64, 1000)
	for i := 0; i < 1000; i++ {
		latencies[i] = float64(i+1) / 1000.0
	}

	// Test with p99.9
	result := Latencies(latencies, true)

	// Should have 8 percentiles (10, 25, 50, 75, 90, 95, 99, 99.9)
	assert.Equal(t, 8, len(result))

	// Verify percentiles are correct
	percentiles := []float64{10, 25, 50, 75, 90, 95, 99, 99.9}
	for i, ld := range result {
		assert.Equal(t, percentiles[i], ld.Percentage)
	}

	// Verify p99.9 exists and is correct
	assert.Equal(t, 99.9, result[7].Percentage)
	// p99.9 should be greater than p99
	assert.True(t, result[7].Latency > result[6].Latency)
}

func TestHistogram_UsesP99Label(t *testing.T) {
	// Create 1000 latency values
	latencies := make([]float64, 1000)
	for i := 0; i < 1000; i++ {
		latencies[i] = float64(i+1) / 1000.0
	}

	slowest := latencies[len(latencies)-1]
	fastest := latencies[0]
	p99Value := latencies[989] // Approximately p99

	// Test with p99 (tailPercentileValue = 99)
	result := Histogram(latencies, slowest, fastest, p99Value, 99)

	// Should have 10 buckets
	assert.Equal(t, 10, len(result))

	// Last bucket should have P99 label
	assert.Contains(t, result[9].AlternativeMark, "P99")
	assert.NotContains(t, result[9].AlternativeMark, "P99.9")
}

func TestHistogram_UsesP99_9Label(t *testing.T) {
	// Create 1000 latency values
	latencies := make([]float64, 1000)
	for i := 0; i < 1000; i++ {
		latencies[i] = float64(i+1) / 1000.0
	}

	slowest := latencies[len(latencies)-1]
	fastest := latencies[0]
	p99_9Value := latencies[998] // Approximately p99.9

	// Test with p99.9 (tailPercentileValue = 99.9)
	result := Histogram(latencies, slowest, fastest, p99_9Value, 99.9)

	// Should have 10 buckets
	assert.Equal(t, 10, len(result))

	// Last bucket should have P99.9 label
	assert.Contains(t, result[9].AlternativeMark, "P99.9")
	assert.NotContains(t, result[9].AlternativeMark, ">=P99 ")
}

func TestFinalize_WithP99_9Enabled(t *testing.T) {
	callResultsChan := make(chan *callResult)
	sampledResultsChan := make(chan *sampledResult)
	config, _ := NewConfig("call", "host", WithP99_9(true))
	reporter := newReporter(callResultsChan, sampledResultsChan, config)

	go reporter.Run()

	// Add 1000 results with latencies from 1ms to 1000ms
	for i := 0; i < 10000; i++ {
		cr := callResult{
			status:    "OK",
			duration:  time.Duration(i+1) * time.Millisecond,
			err:       nil,
			timestamp: time.Now(),
		}
		callResultsChan <- &cr
	}

	close(callResultsChan)
	close(sampledResultsChan)
	<-reporter.done
	report := reporter.Finalize("stop reason", time.Second*2)

	// Verify p99.9 flag is set
	assert.True(t, report.P99_9)

	// Verify latency distribution includes p99.9
	assert.Equal(t, 8, len(report.LatencyDistribution))

	var p99Latency, p99_9Latency time.Duration
	var hasP99, hasP99_9 bool

	for _, ld := range report.LatencyDistribution {
		if ld.Percentage == 99 {
			p99Latency = ld.Latency
			hasP99 = true
		}
		if ld.Percentage == 99.9 {
			p99_9Latency = ld.Latency
			hasP99_9 = true
		}
	}

	assert.True(t, hasP99, "Report should include p99 in latency distribution")
	assert.True(t, hasP99_9, "Report should include p99.9 in latency distribution")

	// Verify p99.9 is greater than p99
	assert.True(t, p99_9Latency > p99Latency, "p99.9 should be greater than p99")

	// With 10000 samples, p99 should be around 9900ms and p99.9 should be around 999ms
	assert.InDelta(t, 9900, p99Latency.Milliseconds(), 10, "p99 should be approximately 9900ms")
	assert.InDelta(t, 9990, p99_9Latency.Milliseconds(), 8, "p99.9 should be approximately 9990ms")

	// Verify histogram uses p99.9 for tail
	assert.NotEmpty(t, report.Histogram)
	lastBucket := report.Histogram[len(report.Histogram)-1]
	assert.Contains(t, lastBucket.AlternativeMark, "P99.9", "Histogram tail should use P99.9 label")

	// The histogram should be bounded by p99.9, not the slowest value
	// So the last bucket mark should be close to p99.9 value, not 1000ms
	assert.Less(t, lastBucket.Mark, 10.0, "Histogram should be bounded by p99.9, not slowest")
}

func TestFinalize_WithP99_9Disabled(t *testing.T) {
	callResultsChan := make(chan *callResult)
	sampledResultsChan := make(chan *sampledResult)
	config, _ := NewConfig("call", "host") // Default: p99_9 = false
	reporter := newReporter(callResultsChan, sampledResultsChan, config)

	go reporter.Run()

	// Add 1000 results with latencies from 1ms to 1000ms
	for i := 0; i < 10000; i++ {
		cr := callResult{
			status:    "OK",
			duration:  time.Duration(i+1) * time.Millisecond,
			err:       nil,
			timestamp: time.Now(),
		}
		callResultsChan <- &cr
	}

	close(callResultsChan)
	close(sampledResultsChan)
	<-reporter.done
	report := reporter.Finalize("stop reason", time.Second*2)

	// Verify p99.9 flag is not set
	assert.False(t, report.P99_9)

	// Verify latency distribution does not include p99.9
	assert.Equal(t, 7, len(report.LatencyDistribution))

	var p99Latency time.Duration
	var hasP99, hasP99_9 bool

	for _, ld := range report.LatencyDistribution {
		if ld.Percentage == 99 {
			p99Latency = ld.Latency
			hasP99 = true
		}
		if ld.Percentage == 99.9 {
			hasP99_9 = true
		}
	}

	assert.True(t, hasP99, "Report should include p99 in latency distribution")
	assert.False(t, hasP99_9, "Report should not include p99.9 in latency distribution")

	// With 1000 samples, p99 should be around 9900ms
	assert.InDelta(t, 9900, p99Latency.Milliseconds(), 10, "p99 should be approximately 9900ms")

	// Verify histogram uses p99 for tail (not p99.9)
	assert.NotEmpty(t, report.Histogram)
	lastBucket := report.Histogram[len(report.Histogram)-1]
	assert.Contains(t, lastBucket.AlternativeMark, "P99", "Histogram tail should use P99 label")
	assert.NotContains(t, lastBucket.AlternativeMark, "P99.9", "Histogram tail should not use P99.9 label")

	// The histogram should be bounded by p99, not p99.9
	// So the last bucket mark should be close to p99 value (around 9.90s)
	assert.Less(t, lastBucket.Mark, 10.0, "Histogram should be bounded by p99, not slowest")
	assert.InDelta(t, 9.90, lastBucket.Mark, 0.02, "Last bucket should be near p99 value")
}
