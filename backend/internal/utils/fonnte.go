package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/kemnaker/perjadin-backend/internal/config"
)

type FonntePayload struct {
	Target      string `json:"target"`
	Message     string `json:"message"`
	Delay       string `json:"delay"`
	CountryCode string `json:"countryCode"`
	Typing      bool   `json:"typing"`
}

type waJob struct {
	cfg     *config.Config
	target  string
	message string
}

var (
	// Antrean pesan WA untuk dieksekusi di background tanpa membebani memori Goroutine
	waQueue = make(chan waJob, 5000)
)

func init() {
	// Menjalankan single background worker saat aplikasi menyala
	go fonnteWorker()
}

func fonnteWorker() {
	var lastSent time.Time
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	for job := range waQueue {
		// 1. Thread-safe Global Rate Limiting (Anti-Ban Utama)
		elapsed := time.Since(lastSent)
		minDelay := 3 * time.Second
		if elapsed < minDelay {
			time.Sleep(minDelay - elapsed)
		}

		// 2. Jitter / Jeda Acak (Human-like behavior)
		jitter := time.Duration(rand.Intn(3)+1) * time.Second
		time.Sleep(jitter)

		url := "https://api.fonnte.com/send"

		payload := FonntePayload{
			Target:      job.target,
			Message:     job.message,
			Delay:       "2-5",
			Typing:      true,
			CountryCode: "62",
		}

		jsonPayload, err := json.Marshal(payload)
		if err == nil {
			req, errReq := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonPayload))
			if errReq == nil {
				req.Header.Set("Authorization", job.cfg.Fonnte.Token)
				req.Header.Set("Content-Type", "application/json")

				resp, errResp := client.Do(req)
				lastSent = time.Now()
				
				if errResp == nil {
					resp.Body.Close()
				} else {
					log.Printf("Background Worker: Gagal mengirim WhatsApp ke %s: %v", job.target, errResp)
				}
			}
		}
	}
}

// cleanPhoneNumber membersihkan karakter non-angka agar sesuai format yang aman
func cleanPhoneNumber(phone string) string {
	reg := regexp.MustCompile("[^0-9]+")
	cleaned := reg.ReplaceAllString(phone, "")

	// Fonnte menggunakan parameter CountryCode: "62", jadi "08" adalah format standar yang baik.
	if strings.HasPrefix(cleaned, "62") {
		cleaned = "0" + strings.TrimPrefix(cleaned, "62")
	}
	return cleaned
}

func SendWhatsAppMessage(cfg *config.Config, target string, message string) error {
	if cfg.Fonnte.Token == "" {
		return fmt.Errorf("Fonnte token is not configured")
	}

	target = cleanPhoneNumber(target)
	if target == "" || len(target) < 9 {
		return fmt.Errorf("invalid or empty target phone number")
	}

	// Kirim job ke antrean tanpa memblokir thread HTTP utama
	select {
	case waQueue <- waJob{cfg: cfg, target: target, message: message}:
		return nil
	default:
		// Jika antrean penuh (lebih dari 5000 pesan tertunda)
		return fmt.Errorf("whatsapp queue is full, dropping message to prevent memory leak")
	}
}
