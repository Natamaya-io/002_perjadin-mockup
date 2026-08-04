package record

import (
	"context"
	"sort"
	"strings"

	"github.com/kemnaker/perjadin-backend/internal/models"
)

func (h *Handler) getGlobalRecordIndex(record *models.TravelRecord) int {
	return record.SequenceNumber
}

func (h *Handler) getGroupRecordsSorted(spdNumber string) []models.TravelRecord {
	records, err := h.svc.GetRecords(context.Background(), map[string]interface{}{"spd": spdNumber})
	if err != nil {
		return []models.TravelRecord{}
	}

	// [ANTI-OOM BYPASS]: Fetch full data for each record to recover stripped documentation images (Files/Base64)
	for i := range records {
		fullRec, err := h.svc.GetRecordByID(context.Background(), records[i].ID)
		if err == nil && fullRec != nil {
			records[i] = *fullRec
		}
	}

	sort.Slice(records, func(i, j int) bool {
		nameI := strings.ToLower(strings.TrimSpace(records[i].Employee.Name))
		nameJ := strings.ToLower(strings.TrimSpace(records[j].Employee.Name))

		getPriority := func(name string) int {
			if strings.Contains(name, "auditya hermawan") {
				return 1
			}
			if strings.Contains(name, "mochamad gufron") {
				return 2
			}
			if strings.Contains(name, "muhammad isa") {
				return 3
			}
			return 4
		}

		pI := getPriority(nameI)
		pJ := getPriority(nameJ)

		if pI != pJ {
			return pI < pJ
		}

		nipI := strings.TrimSpace(records[i].Employee.NIP)
		nipJ := strings.TrimSpace(records[j].Employee.NIP)

		hasNIPI := nipI != "" && nipI != "-"
		hasNIPJ := nipJ != "" && nipJ != "-"

		if hasNIPI && !hasNIPJ {
			return true
		} else if !hasNIPI && hasNIPJ {
			return false
		}
		return records[i].CreatedAt.Unix() < records[j].CreatedAt.Unix()
	})

	return records
}
