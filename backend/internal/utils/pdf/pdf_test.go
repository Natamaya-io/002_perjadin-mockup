package pdf

import (
	"testing"
	"time"

	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestGenerateSpdOverlay_ExecutesWithoutError(t *testing.T) {
	// Skip this test if template is not found (might happen in CI/limited environments)
	
	record := &models.TravelRecord{
		SPDNumber: "ID-SPJ-001",
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 0, 3),
		Employee: models.User{
			Name: "Test Employee",
			NIP:  "123456",
		},
		SuratTugasDate: time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC),
	}

	// This function tries to find the template PDF. 
	// If it's not found, it returns an error.
	// We want to at least check if the logic within the function doesn't panic.
	
	data, err := GenerateSpdOverlay(record)
	
	// We allow "template file not found" error because we are just testing the code logic here,
	// and the template might not be present in the test environment.
	if err != nil && (err.Error() == "template file not found in current, parent, or root directory") {
		t.Skip("Skipping PDF overlay test as template file was not found")
	}

	assert.NoError(t, err)
	assert.NotNil(t, data)
}
