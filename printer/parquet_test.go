package printer

import (
	"os"
	"testing"
	"time"

	"github.com/chalk-ai/ghz/runner"
	"github.com/stretchr/testify/assert"
)

func TestReportPrinter_WriteParquetFile(t *testing.T) {
	date := time.Now()

	report := &runner.Report{
		Name:      "test report",
		EndReason: runner.ReasonNormalEnd,
		Date:      date,
		Count:     10,
		Total:     time.Duration(1 * time.Second),
		Average:   time.Duration(100 * time.Millisecond),
		Fastest:   time.Duration(50 * time.Millisecond),
		Slowest:   time.Duration(200 * time.Millisecond),
		Rps:       10,
		Details: []runner.ResultDetail{
			{
				Timestamp: date,
				Latency:   time.Duration(100 * time.Millisecond),
				Status:    "OK",
				Error:     "",
			},
			{
				Timestamp: date.Add(time.Millisecond * 100),
				Latency:   time.Duration(150 * time.Millisecond),
				Status:    "OK",
				Error:     "",
			},
			{
				Timestamp: date.Add(time.Millisecond * 250),
				Latency:   time.Duration(200 * time.Millisecond),
				Status:    "Internal",
				Error:     "rpc error: code = Internal desc = Internal error",
			},
		},
	}

	printer := ReportPrinter{
		Out:    os.Stdout,
		Report: report,
	}

	// Write parquet file
	outputPath := "test_results.parquet"
	defer os.Remove(outputPath) // Clean up after test

	err := printer.WriteParquetFile(outputPath)
	assert.NoError(t, err)

	// Verify file exists and has content
	fileInfo, err := os.Stat(outputPath)
	assert.NoError(t, err)
	assert.True(t, fileInfo.Size() > 0, "Parquet file should not be empty")
}

func TestReportPrinter_WriteParquetSampleFile(t *testing.T) {
	date := time.Now()

	report := &runner.Report{
		Name:      "test report with samples",
		EndReason: runner.ReasonNormalEnd,
		Date:      date,
		Count:     10,
		Total:     time.Duration(1 * time.Second),
		Average:   time.Duration(100 * time.Millisecond),
		Fastest:   time.Duration(50 * time.Millisecond),
		Slowest:   time.Duration(200 * time.Millisecond),
		Rps:       10,
		SampleDetails: []runner.ResultDetail{
			{
				Timestamp:       date,
				Latency:         time.Duration(100 * time.Millisecond),
				Status:          "OK",
				Error:           "",
				RequestPayload:  `{"name":"Alice"}`,
				ResponsePayload: `{"message":"Hello Alice"}`,
			},
			{
				Timestamp:       date.Add(time.Millisecond * 100),
				Latency:         time.Duration(150 * time.Millisecond),
				Status:          "OK",
				Error:           "",
				RequestPayload:  `{"name":"Bob"}`,
				ResponsePayload: `{"message":"Hello Bob"}`,
			},
		},
	}

	printer := ReportPrinter{
		Out:    os.Stdout,
		Report: report,
	}

	// Write parquet file
	outputPath := "test_sampled_results.parquet"
	defer os.Remove(outputPath) // Clean up after test

	err := printer.WriteParquetSampleFile(outputPath)
	assert.NoError(t, err)

	// Verify file exists and has content
	fileInfo, err := os.Stat(outputPath)
	assert.NoError(t, err)
	assert.True(t, fileInfo.Size() > 0, "Parquet file should not be empty")
}

func TestReportPrinter_Print_parquet(t *testing.T) {
	date := time.Now()

	report := &runner.Report{
		Name:      "test report",
		EndReason: runner.ReasonNormalEnd,
		Date:      date,
		Count:     5,
		Total:     time.Duration(1 * time.Second),
		Details: []runner.ResultDetail{
			{
				Timestamp: date,
				Latency:   time.Duration(100 * time.Millisecond),
				Status:    "OK",
				Error:     "",
			},
		},
	}

	// Test with file output
	outputPath := "test_print_results.parquet"
	defer os.Remove(outputPath) // Clean up after test

	file, err := os.Create(outputPath)
	assert.NoError(t, err)
	defer file.Close()

	printer := ReportPrinter{
		Out:    file,
		Report: report,
	}

	err = printer.Print("parquet")
	assert.NoError(t, err)

	// Verify file exists and has content
	fileInfo, err := os.Stat(outputPath)
	assert.NoError(t, err)
	assert.True(t, fileInfo.Size() > 0, "Parquet file should not be empty")
}

func TestReportPrinter_WriteParquetFile_EmptyDetails(t *testing.T) {
	date := time.Now()

	report := &runner.Report{
		Name:      "test report with no details",
		EndReason: runner.ReasonNormalEnd,
		Date:      date,
		Count:     0,
		Total:     time.Duration(0),
		Details:   []runner.ResultDetail{},
	}

	printer := ReportPrinter{
		Out:    os.Stdout,
		Report: report,
	}

	// Write parquet file with empty details
	outputPath := "test_empty_results.parquet"
	defer os.Remove(outputPath) // Clean up after test

	err := printer.WriteParquetFile(outputPath)
	assert.NoError(t, err)
}
