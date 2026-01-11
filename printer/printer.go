package printer

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/alecthomas/template"
	"github.com/chalk-ai/ghz/runner"
	"github.com/parquet-go/parquet-go"
)

const (
	barChar = "∎"
)

// ParquetResultDetail is the Parquet-optimized version of ResultDetail
type ParquetResultDetail struct {
	Timestamp       int64  `parquet:"timestamp,timestamp(millisecond)"`
	Latency         int64  `parquet:"latency"`
	Error           string `parquet:"error,optional,dict"`
	Status          string `parquet:"status,dict"`
	RequestPayload  string `parquet:"request_payload,optional"`
	ResponsePayload string `parquet:"response_payload,optional"`
}

// ReportPrinter is used for printing the report
type ReportPrinter struct{
	Out    io.Writer
	Report *runner.Report
}

// Print the report using the given format
// If format is "csv" detailed listing is printer in csv format.
// Otherwise the summary of results is printed.
//
// Supported Format:
//
//	summary
//	csv
//	json
//	pretty
//	html
//	influx-summary
//	influx-details
func (rp *ReportPrinter) Print(format string) error {
	if format == "" {
		format = "summary"
	}

	switch format {
	case "summary", "csv":
		outputTmpl := defaultTmpl
		if format == "csv" {
			outputTmpl = csvTmpl
		}
		buf := &bytes.Buffer{}
		templ := template.Must(template.New("tmpl").Funcs(tmplFuncMap).Parse(outputTmpl))
		if err := templ.Execute(buf, *rp.Report); err != nil {
			return err
		}

		return rp.print(buf.String())
	case "json", "pretty":
		rep, err := json.Marshal(*rp.Report)
		if err != nil {
			return err
		}

		if format == "pretty" {
			var out bytes.Buffer
			err = json.Indent(&out, rep, "", "  ")
			if err != nil {
				return err
			}
			rep = out.Bytes()
		}
		return rp.print(string(rep))
	case "html":
		buf := &bytes.Buffer{}

		// Create a temporary report with SampleDetails for embedding
		reportForHTML := *rp.Report
		reportForHTML.Details = reportForHTML.SampleDetails

		// Debug: print counts
		fmt.Fprintf(os.Stderr, "DEBUG Printer: Original Details=%d, SampleDetails=%d\n",
			len(rp.Report.Details), len(rp.Report.SampleDetails))

		templ := template.Must(template.New("tmpl").Funcs(tmplFuncMap).Parse(htmlTmpl))
		if err := templ.Execute(buf, reportForHTML); err != nil {
			return err
		}
		return rp.print(buf.String())
	case "influx-summary":
		return rp.printInfluxLine()
	case "influx-details":
		return rp.printInfluxDetails()
	case "prometheus":
		return rp.printPrometheus()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func (rp *ReportPrinter) print(s string) error {
	_, err := fmt.Fprint(rp.Out, s)
	return err
}

// toParquetFormat converts ResultDetail slice to ParquetResultDetail slice
func toParquetFormat(details []runner.ResultDetail) []ParquetResultDetail {
	result := make([]ParquetResultDetail, len(details))
	for i, d := range details {
		result[i] = ParquetResultDetail{
			Timestamp:       d.Timestamp.UnixMilli(),
			Latency:         d.Latency.Nanoseconds(),
			Error:           d.Error,
			Status:          d.Status,
			RequestPayload:  d.RequestPayload,
			ResponsePayload: d.ResponsePayload,
		}
	}
	return result
}

// generateParquetBytes generates parquet bytes from ResultDetail slice
func generateParquetBytes(details []runner.ResultDetail) ([]byte, error) {
	if len(details) == 0 {
		return []byte{}, nil
	}

	// Convert to Parquet format
	parquetData := toParquetFormat(details)

	// Create buffer
	buf := new(bytes.Buffer)

	// Create Parquet writer with Snappy compression
	writer := parquet.NewGenericWriter[ParquetResultDetail](buf,
		parquet.Compression(&parquet.Snappy),
	)

	// Write data
	if _, err := writer.Write(parquetData); err != nil {
		return nil, fmt.Errorf("failed to write parquet data: %w", err)
	}

	// Close writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close parquet writer: %w", err)
	}

	return buf.Bytes(), nil
}

// parquetify converts ResultDetail slice to base64-encoded parquet
func parquetify(details []runner.ResultDetail) (string, error) {
	parquetBytes, err := generateParquetBytes(details)
	if err != nil {
		return "", err
	}

	// Base64 encode
	return base64.StdEncoding.EncodeToString(parquetBytes), nil
}

// writeParquetFile writes ResultDetail slice to a parquet file
func writeParquetFile(outputPath string, details []runner.ResultDetail) error {
	// Determine parquet file path
	var parquetPath string
	if strings.HasSuffix(outputPath, ".html") {
		parquetPath = strings.TrimSuffix(outputPath, ".html") + ".parquet"
	} else {
		parquetPath = outputPath + ".parquet"
	}

	// Generate parquet bytes
	parquetBytes, err := generateParquetBytes(details)
	if err != nil {
		return fmt.Errorf("failed to generate parquet file: %w", err)
	}

	// Write to file
	if err := os.WriteFile(parquetPath, parquetBytes, 0644); err != nil {
		return fmt.Errorf("failed to write parquet file: %w", err)
	}

	return nil
}

var tmplFuncMap = template.FuncMap{
	"formatMilli":      formatMilli,
	"formatSeconds":    formatSeconds,
	"histogram":        histogram,
	"jsonify":          jsonify,
	"parquetify":       parquetify,
	"formatMark":       formatMarkMs,
	"formatPercent":    formatPercent,
	"formatStatusCode": formatStatusCode,
	"formatErrorDist":  formatErrorDist,
	"formatDate":       formatDate,
	"formatNanoUnit":   formatNanoUnit,
}

func jsonify(v interface{}, pretty bool) string {
	d, _ := json.Marshal(v)
	if !pretty {
		return string(d)
	}

	var out bytes.Buffer
	err := json.Indent(&out, d, "", "  ")
	if err != nil {
		return string(d)
	}

	return out.String()
}

func formatNanoUnit(d time.Duration) string {
	v := d.Nanoseconds()
	if v < 10000 {
		return fmt.Sprintf("%+v ns", v)
	}

	valMs := float64(v) / 1000000.0
	if valMs < 1000 {
		return fmt.Sprintf("%4.2f ms", valMs)
	}

	return fmt.Sprintf("%4.2f s", float64(valMs)/1000.0)
}

func formatMilli(duration float64) string {
	return fmt.Sprintf("%4.2f", duration*1000)
}

func formatDate(d time.Time) string {
	return d.Format("Mon Jan _2 2006 @ 15:04:05")
}

func formatSeconds(duration float64) string {
	return fmt.Sprintf("%4.2f", duration)
}

func formatPercent(num int, total uint64) string {
	p := float64(num) / float64(total)
	return fmt.Sprintf("%.2f", p*100)
}

func histogram(buckets []runner.Bucket) string {
	maxMark := 0.0
	maxCount := 0
	for _, b := range buckets {
		if v := b.Mark; v > maxMark {
			maxMark = v
		}
		if v := b.Count; v > maxCount {
			maxCount = v
		}
	}

	formatMark := func(mark float64) string {
		return fmt.Sprintf("%.3f", mark*1000)
	}
	formatCount := func(count int) string {
		return fmt.Sprintf("%v", count)
	}

	var maxMarkLen int
	if len(buckets) != 0 {
		maxMarkLen = max(len(formatMark(maxMark)), len(buckets[len(buckets)-1].AlternativeMark))
	} else {
		maxMarkLen = len(formatMark(maxMark))
	}
	maxCountLen := len(formatCount(maxCount))
	res := new(bytes.Buffer)
	for i := 0; i < len(buckets); i++ {
		// Normalize bar lengths.
		var barLen int
		if maxCount > 0 {
			barLen = (buckets[i].Count*40 + maxCount/2) / maxCount
		}
		markStr := buckets[i].AlternativeMark
		countStr := formatCount(buckets[i].Count)
		res.WriteString(fmt.Sprintf(
			"  %s%s [%v]%s |%v\n",
			markStr,
			strings.Repeat(" ", maxMarkLen-len(markStr)),
			countStr,
			strings.Repeat(" ", maxCountLen-len(countStr)),
			strings.Repeat(barChar, barLen),
		))
	}

	return res.String()
}

func formatMarkMs(m float64) string {
	m = m * 1000.0

	if m < 1 {
		return fmt.Sprintf("'%4.4f ms'", m)
	}

	return fmt.Sprintf("'%4.2f ms'", m)
}

func formatStatusCode(statusCodeDist map[string]int) string {
	padding := 3
	buf := &bytes.Buffer{}
	w := tabwriter.NewWriter(buf, 0, 0, padding, ' ', 0)
	for status, count := range statusCodeDist {
		// bytes.Buffer can be assumed to not fail on write
		_, _ = fmt.Fprintf(w, "  [%+s]\t%+v responses\t\n", status, count)
	}
	// bytes.Buffer can be assumed to not fail on write
	_ = w.Flush()
	return buf.String()
}

func formatErrorDist(errDist map[string]int) string {
	padding := 3
	buf := &bytes.Buffer{}
	w := tabwriter.NewWriter(buf, 0, 0, padding, ' ', 0)
	for status, count := range errDist {
		// bytes.Buffer can be assumed to not fail on write
		_, _ = fmt.Fprintf(w, "  [%+v]\t%+s\t\n", count, status)
	}
	// bytes.Buffer can be assumed to not fail on write
	_ = w.Flush()
	return buf.String()
}
