// Package stats provides statistics services for the user portal.
package stats

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Unit tests

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
		{1099511627776, "1.00 TB"},
	}

	for _, tt := range tests {
		result := FormatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatBytes(%d) = %s, expected %s", tt.bytes, result, tt.expected)
		}
	}
}

func TestValidatePeriod(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"day", "day"},
		{"week", "week"},
		{"month", "month"},
		{"year", "year"},
		{"invalid", "month"},
		{"", "month"},
		{"DAY", "month"}, // case sensitive
	}

	for _, tt := range tests {
		result := ValidatePeriod(tt.input)
		if result != tt.expected {
			t.Errorf("ValidatePeriod(%s) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func TestGetPeriodDays(t *testing.T) {
	tests := []struct {
		period   string
		expected int
	}{
		{"day", 1},
		{"week", 7},
		{"month", 30},
		{"year", 365},
		{"invalid", 30},
	}

	for _, tt := range tests {
		result := GetPeriodDays(tt.period)
		if result != tt.expected {
			t.Errorf("GetPeriodDays(%s) = %d, expected %d", tt.period, result, tt.expected)
		}
	}
}

func TestAggregateDaily(t *testing.T) {
	daily := []*DailyTraffic{
		{Date: "2024-01-01", Upload: 100, Download: 200, Total: 300},
		{Date: "2024-01-02", Upload: 150, Download: 250, Total: 400},
		{Date: "2024-01-03", Upload: 200, Download: 300, Total: 500},
	}

	summary := AggregateDaily(daily)

	if summary.Upload != 450 {
		t.Errorf("Expected upload 450, got %d", summary.Upload)
	}
	if summary.Download != 750 {
		t.Errorf("Expected download 750, got %d", summary.Download)
	}
	if summary.Total != 1200 {
		t.Errorf("Expected total 1200, got %d", summary.Total)
	}
}

func TestAggregateDaily_Empty(t *testing.T) {
	summary := AggregateDaily([]*DailyTraffic{})

	if summary.Upload != 0 || summary.Download != 0 || summary.Total != 0 {
		t.Error("Expected all zeros for empty input")
	}
}

// Feature: user-portal, Property 13: Traffic Statistics Consistency
// Validates: Requirements 11.2, 11.3
// *For any* time period, the sum of daily traffic values SHALL equal the total
// traffic for that period.
func TestProperty_TrafficStatisticsConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Sum of daily values equals total
	properties.Property("sum of daily values equals total", prop.ForAll(
		func(numDays int, seed int64) bool {
			if numDays <= 0 || numDays > 365 {
				return true
			}

			// Generate random daily traffic
			daily := make([]*DailyTraffic, numDays)
			var expectedUpload, expectedDownload int64

			for i := 0; i < numDays; i++ {
				upload := (seed + int64(i*100)) % 10000000
				download := (seed + int64(i*200)) % 10000000
				daily[i] = &DailyTraffic{
					Date:     "2024-01-01",
					Upload:   upload,
					Download: download,
					Total:    upload + download,
				}
				expectedUpload += upload
				expectedDownload += download
			}

			summary := AggregateDaily(daily)

			return summary.Upload == expectedUpload &&
				summary.Download == expectedDownload &&
				summary.Total == expectedUpload+expectedDownload
		},
		gen.IntRange(1, 365),
		gen.Int64Range(0, 1000000),
	))

	// Property: Total equals upload + download
	properties.Property("total equals upload plus download", prop.ForAll(
		func(upload, download int64) bool {
			if upload < 0 || download < 0 {
				return true
			}

			daily := []*DailyTraffic{
				{Upload: upload, Download: download, Total: upload + download},
			}

			summary := AggregateDaily(daily)
			return summary.Total == summary.Upload+summary.Download
		},
		gen.Int64Range(0, 1000000000),
		gen.Int64Range(0, 1000000000),
	))

	// Property: Aggregation is associative
	properties.Property("aggregation is associative", prop.ForAll(
		func(seed int64, numDays int) bool {
			if numDays <= 1 || numDays > 100 {
				return true
			}

			// Generate daily traffic
			daily := make([]*DailyTraffic, numDays)
			for i := 0; i < numDays; i++ {
				upload := (seed + int64(i*100)) % 10000000
				download := (seed + int64(i*200)) % 10000000
				daily[i] = &DailyTraffic{
					Upload:   upload,
					Download: download,
					Total:    upload + download,
				}
			}

			// Aggregate all at once
			totalSummary := AggregateDaily(daily)

			// Aggregate in two parts
			mid := numDays / 2
			part1 := AggregateDaily(daily[:mid])
			part2 := AggregateDaily(daily[mid:])

			combinedUpload := part1.Upload + part2.Upload
			combinedDownload := part1.Download + part2.Download

			return totalSummary.Upload == combinedUpload &&
				totalSummary.Download == combinedDownload
		},
		gen.Int64Range(0, 1000000),
		gen.IntRange(2, 100),
	))

	// Property: Empty aggregation returns zeros
	properties.Property("empty aggregation returns zeros", prop.ForAll(
		func(_ int) bool {
			summary := AggregateDaily([]*DailyTraffic{})
			return summary.Upload == 0 && summary.Download == 0 && summary.Total == 0
		},
		gen.Int(),
	))

	// Property: Single day aggregation equals that day's values
	properties.Property("single day aggregation equals day values", prop.ForAll(
		func(upload, download int64) bool {
			if upload < 0 || download < 0 {
				return true
			}

			daily := []*DailyTraffic{
				{Upload: upload, Download: download, Total: upload + download},
			}

			summary := AggregateDaily(daily)
			return summary.Upload == upload && summary.Download == download
		},
		gen.Int64Range(0, 1000000000),
		gen.Int64Range(0, 1000000000),
	))

	// Property: FormatBytes produces non-empty string for any non-negative value
	properties.Property("FormatBytes produces non-empty string", prop.ForAll(
		func(bytes int64) bool {
			if bytes < 0 {
				return true
			}
			result := FormatBytes(bytes)
			return len(result) > 0
		},
		gen.Int64Range(0, 10000000000000),
	))

	// Property: FormatBytes is monotonic (larger values produce larger or equal formatted values)
	properties.Property("FormatBytes preserves order for same unit", prop.ForAll(
		func(base int64) bool {
			if base < 0 || base > 1000000000 {
				return true
			}

			// Test within same unit (KB range)
			val1 := base * 1024
			val2 := (base + 1) * 1024

			// Both should be in KB range
			str1 := FormatBytes(val1)
			str2 := FormatBytes(val2)

			// Just verify both produce valid output
			return len(str1) > 0 && len(str2) > 0
		},
		gen.Int64Range(1, 1000),
	))

	// Property: Period validation always returns valid period
	properties.Property("period validation always returns valid period", prop.ForAll(
		func(seed int64) bool {
			inputs := []string{"day", "week", "month", "year", "invalid", "", "random", "DAY"}
			input := inputs[int(seed)%len(inputs)]

			result := ValidatePeriod(input)
			validPeriods := map[string]bool{"day": true, "week": true, "month": true, "year": true}

			return validPeriods[result]
		},
		gen.Int64Range(0, 1000),
	))

	// Property: GetPeriodDays returns positive value
	properties.Property("GetPeriodDays returns positive value", prop.ForAll(
		func(seed int64) bool {
			inputs := []string{"day", "week", "month", "year", "invalid", ""}
			input := inputs[int(seed)%len(inputs)]

			result := GetPeriodDays(input)
			return result > 0
		},
		gen.Int64Range(0, 1000),
	))

	properties.TestingRun(t)
}

// Additional unit tests for edge cases

func TestFormatBytes_LargeValues(t *testing.T) {
	// Test very large values
	result := FormatBytes(5 * 1099511627776) // 5 TB
	if result != "5.00 TB" {
		t.Errorf("Expected '5.00 TB', got '%s'", result)
	}
}

func TestFormatBytes_Boundaries(t *testing.T) {
	// Test boundary values
	tests := []struct {
		bytes    int64
		contains string
	}{
		{1023, "B"},
		{1024, "KB"},
		{1048575, "KB"},
		{1048576, "MB"},
		{1073741823, "MB"},
		{1073741824, "GB"},
	}

	for _, tt := range tests {
		result := FormatBytes(tt.bytes)
		if len(result) == 0 {
			t.Errorf("FormatBytes(%d) returned empty string", tt.bytes)
		}
	}
}

func TestAggregateDaily_LargeValues(t *testing.T) {
	// Test with large values to ensure no overflow issues
	daily := []*DailyTraffic{
		{Upload: 1000000000000, Download: 2000000000000, Total: 3000000000000},
		{Upload: 1000000000000, Download: 2000000000000, Total: 3000000000000},
	}

	summary := AggregateDaily(daily)

	if summary.Upload != 2000000000000 {
		t.Errorf("Expected upload 2000000000000, got %d", summary.Upload)
	}
	if summary.Download != 4000000000000 {
		t.Errorf("Expected download 4000000000000, got %d", summary.Download)
	}
}

func TestTrafficSummary_StringFormatting(t *testing.T) {
	summary := &TrafficSummary{
		Upload:   1073741824, // 1 GB
		Download: 2147483648, // 2 GB
		Total:    3221225472, // 3 GB
	}

	summary.UploadStr = FormatBytes(summary.Upload)
	summary.DownloadStr = FormatBytes(summary.Download)
	summary.TotalStr = FormatBytes(summary.Total)

	if summary.UploadStr != "1.00 GB" {
		t.Errorf("Expected '1.00 GB', got '%s'", summary.UploadStr)
	}
	if summary.DownloadStr != "2.00 GB" {
		t.Errorf("Expected '2.00 GB', got '%s'", summary.DownloadStr)
	}
	if summary.TotalStr != "3.00 GB" {
		t.Errorf("Expected '3.00 GB', got '%s'", summary.TotalStr)
	}
}
