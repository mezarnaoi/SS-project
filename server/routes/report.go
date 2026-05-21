package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"mqtt-streaming-server/domain"
	"mqtt-streaming-server/repository"
)

type ReportController struct {
	ReportRepository domain.ReportRepository
}

func InitReportRoutes(db *mongo.Database, mux *http.ServeMux) {
	ctrl := &ReportController{
		ReportRepository: repository.NewReportRepository(db),
	}

	mux.Handle("/reports/summary", withAuth(http.HandlerFunc(ctrl.GetSummary)))
	mux.Handle("/reports/expiring", withAuth(http.HandlerFunc(ctrl.GetExpirationAlerts)))
	mux.Handle("/reports/anonymized", withAuth(http.HandlerFunc(ctrl.GetAnonymizedRecords)))
	mux.Handle("/reports/performance", withAuth(http.HandlerFunc(ctrl.GetPerformanceMetrics)))
}

func (c *ReportController) GetSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	summary, err := c.ReportRepository.GetSummary(r.Context())
	if err != nil {
		http.Error(w, "Failed to generate summary report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary) // #nosec G104
}

func (c *ReportController) GetExpirationAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	daysAhead := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			daysAhead = parsed
		}
	}

	alerts, err := c.ReportRepository.GetExpirationAlerts(r.Context(), daysAhead)
	if err != nil {
		http.Error(w, "Failed to fetch expiration alerts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts) // #nosec G104
}

func (c *ReportController) GetAnonymizedRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var start, end time.Time

	if s := r.URL.Query().Get("start"); s != "" {
		if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
			start = time.Unix(ts, 0).UTC()
		}
	}
	if e := r.URL.Query().Get("end"); e != "" {
		if ts, err := strconv.ParseInt(e, 10, 64); err == nil {
			end = time.Unix(ts, 0).UTC()
		}
	}

	records, err := c.ReportRepository.GetAnonymizedRecords(r.Context(), start, end)
	if err != nil {
		http.Error(w, "Failed to fetch anonymized records", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records) // #nosec G104
}

func (c *ReportController) GetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics, err := c.ReportRepository.GetPerformanceMetrics(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch performance metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics) // #nosec G104
}
