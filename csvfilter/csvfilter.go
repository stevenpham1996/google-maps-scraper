package csvfilter

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/gosom/google-maps-scraper/gmaps"
	"github.com/gosom/scrapemate"
)

// FilteredCsvWriter is a CSV writer that only writes specified fields
type FilteredCsvWriter struct {
	writer *csv.Writer
	fields map[string]bool // Map of fields to include
}

// NewFilteredCsvWriter creates a new filtered CSV writer
func NewFilteredCsvWriter(writer *csv.Writer, fields string) *FilteredCsvWriter {
	fieldMap := make(map[string]bool)
	
	// If fields is empty, include all fields
	if fields == "" {
		return &FilteredCsvWriter{
			writer: writer,
			fields: fieldMap,
		}
	}
	
	// Parse the fields string and create a map for O(1) lookup
	for _, field := range strings.Split(fields, ",") {
		fieldMap[strings.TrimSpace(field)] = true
	}
	
	return &FilteredCsvWriter{
		writer: writer,
		fields: fieldMap,
	}
}

// Run processes the results and writes them to CSV
func (w *FilteredCsvWriter) Run(ctx context.Context, in <-chan scrapemate.Result) error {
	// Write headers only once
	var headersWritten bool
	
	for result := range in {
		// Check if we have an Entry
		entry, ok := result.Data.(*gmaps.Entry)
		if !ok {
			// Try to handle a slice of entries
			entries, err := asSlice(result.Data)
			if err != nil {
				return fmt.Errorf("invalid data type: %T", result.Data)
			}
			
			for _, entry := range entries {
				if err := w.writeEntry(entry, &headersWritten); err != nil {
					return err
				}
			}
			continue
		}
		
		// Handle single entry
		if err := w.writeEntry(entry, &headersWritten); err != nil {
			return err
		}
	}
	
	w.writer.Flush()
	return nil
}

// writeEntry writes a single entry to the CSV file
func (w *FilteredCsvWriter) writeEntry(entry *gmaps.Entry, headersWritten *bool) error {
	// Write headers if not written yet
	if !*headersWritten {
		headers := w.filterHeaders(entry.CsvHeaders())
		if err := w.writer.Write(headers); err != nil {
			return err
		}
		*headersWritten = true
	}
	
	// Write the row with filtered fields
	row := w.filterRow(entry.CsvHeaders(), entry.CsvRow())
	if err := w.writer.Write(row); err != nil {
		return err
	}
	
	return nil
}

// filterHeaders filters the headers based on the fields map
func (w *FilteredCsvWriter) filterHeaders(headers []string) []string {
	// If no fields specified, return all headers
	if len(w.fields) == 0 {
		return headers
	}
	
	filteredHeaders := make([]string, 0)
	for _, header := range headers {
		if w.fields[strings.ToLower(header)] {
			filteredHeaders = append(filteredHeaders, header)
		}
	}
	
	return filteredHeaders
}

// filterRow filters the row based on the fields map
func (w *FilteredCsvWriter) filterRow(headers []string, row []string) []string {
	// If no fields specified, return all values
	if len(w.fields) == 0 {
		return row
	}
	
	filteredRow := make([]string, 0)
	for i, header := range headers {
		if w.fields[strings.ToLower(header)] {
			filteredRow = append(filteredRow, row[i])
		}
	}
	
	return filteredRow
}

// asSlice attempts to convert the data to a slice of *gmaps.Entry
func asSlice(data interface{}) ([]*gmaps.Entry, error) {
	// Try to cast to a slice of *gmaps.Entry
	if entries, ok := data.([]*gmaps.Entry); ok {
		return entries, nil
	}
	
	// Try to cast to a slice of interface{}
	if slice, ok := data.([]interface{}); ok {
		result := make([]*gmaps.Entry, 0, len(slice))
		for _, item := range slice {
			if entry, ok := item.(*gmaps.Entry); ok {
				result = append(result, entry)
			}
		}
		if len(result) > 0 {
			return result, nil
		}
	}
	
	return nil, fmt.Errorf("cannot convert %T to []*gmaps.Entry", data)
}
