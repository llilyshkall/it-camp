package utils

import (
	"testing"
)

func TestParseExcelFromBytes(t *testing.T) {
	// Тест с пустыми данными
	_, err := ParseExcelFromBytes([]byte{})
	if err == nil {
		t.Error("Expected error for empty data, got nil")
	}

	// Тест с некорректными данными (не Excel)
	_, err = ParseExcelFromBytes([]byte("not an excel file"))
	if err == nil {
		t.Error("Expected error for invalid Excel data, got nil")
	}
}
