package printer_test

import (
	"fmt"
	"os"

	"github.com/chalk-ai/ghz/printer"
	"github.com/chalk-ai/ghz/runner"
)

// ExampleReportPrinter_WriteParquetFile demonstrates how to save report details to a parquet file.
func ExampleReportPrinter_WriteParquetFile() {
	// Run the benchmark
	report, err := runner.Run(
		"helloworld.Greeter.SayHello",
		"localhost:50051",
		runner.WithProtoFile("greeter.proto", []string{}),
		runner.WithDataFromFile("data.json"),
		runner.WithInsecure(true),
	)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Create printer
	printer := printer.ReportPrinter{
		Out:    os.Stdout,
		Report: report,
	}

	// Save results to parquet file
	if err := printer.WriteParquetFile("results.parquet"); err != nil {
		fmt.Println("Error writing parquet file:", err.Error())
		os.Exit(1)
	}

	fmt.Println("Results saved to results.parquet")
}

// ExampleReportPrinter_WriteParquetSampleFile demonstrates how to save sampled results to a parquet file.
// This is useful when you've enabled data sampling with WithDataSamplingRate.
func ExampleReportPrinter_WriteParquetSampleFile() {
	// Run the benchmark with data sampling enabled
	report, err := runner.Run(
		"helloworld.Greeter.SayHello",
		"localhost:50051",
		runner.WithProtoFile("greeter.proto", []string{}),
		runner.WithDataFromFile("data.json"),
		runner.WithInsecure(true),
		runner.WithDataSamplingRate(0.1), // Sample 10% of requests
	)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Create printer
	printer := printer.ReportPrinter{
		Out:    os.Stdout,
		Report: report,
	}

	// Save sampled results with payloads to parquet file
	if err := printer.WriteParquetSampleFile("sampled_results.parquet"); err != nil {
		fmt.Println("Error writing parquet file:", err.Error())
		os.Exit(1)
	}

	fmt.Println("Sampled results saved to sampled_results.parquet")
}

// ExampleReportPrinter_Print_parquet demonstrates using Print method with parquet format.
func ExampleReportPrinter_Print_parquet() {
	// Run the benchmark
	report, err := runner.Run(
		"helloworld.Greeter.SayHello",
		"localhost:50051",
		runner.WithProtoFile("greeter.proto", []string{}),
		runner.WithDataFromFile("data.json"),
		runner.WithInsecure(true),
	)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Open a file for output
	file, err := os.Create("results.parquet")
	if err != nil {
		fmt.Println("Error creating file:", err.Error())
		os.Exit(1)
	}
	defer file.Close()

	// Create printer with file as output
	printer := printer.ReportPrinter{
		Out:    file,
		Report: report,
	}

	// Use Print method with parquet format
	if err := printer.Print("parquet"); err != nil {
		fmt.Println("Error writing parquet:", err.Error())
		os.Exit(1)
	}

	fmt.Println("Results saved to results.parquet")
}
