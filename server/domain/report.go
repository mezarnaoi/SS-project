package domain

import (
	"context"
	"time"
)

type ReportSummary struct {
	TotalChecks            int               `json:"total_checks"`
	ChecksLastMonth        int               `json:"checks_last_month"`
	ChecksThisMonth        int               `json:"checks_this_month"`
	ExpiringNextMonth      []ExpirationAlert `json:"expiring_next_month"`
	FitCount               int               `json:"fit_count"`
	ConditionallyFitCount  int               `json:"conditionally_fit_count"`
	TemporarilyUnfitCount  int               `json:"temporarily_unfit_count"`
	UnfitCount             int               `json:"unfit_count"`
	OCRSuccessRate         float64           `json:"ocr_success_rate"`
	NeedsReviewCount       int               `json:"needs_review_count"`
	AvgProcessingTimeMs    int64             `json:"avg_processing_time_ms"`
}

type ExpirationAlert struct {
	LastName        string    `json:"last_name"`
	FirstName       string    `json:"first_name"`
	Company         string    `json:"company"`
	Workplace       string    `json:"workplace"`
	NextExamDate    time.Time `json:"next_exam_date"`
	DaysUntilExpiry int       `json:"days_until_expiry"`
	MedicalOpinion  string    `json:"medical_opinion"`
}

type AnonymizedRecord struct {
	ControlType    string    `json:"control_type"`
	MedicalOpinion string    `json:"medical_opinion"`
	Company        string    `json:"company"`
	Workplace      string    `json:"workplace"`
	JobTitle       string    `json:"job_title"`
	ExamDate       time.Time `json:"exam_date"`
	NextExamDate   time.Time `json:"next_exam_date"`
	OCRConfidence  float64   `json:"ocr_confidence"`
}

type PerformanceMetrics struct {
	TotalProcessed      int     `json:"total_processed"`
	OCRSuccessCount     int     `json:"ocr_success_count"`
	OCRSuccessRate      float64 `json:"ocr_success_rate"`
	NeedsReviewCount    int     `json:"needs_review_count"`
	ReviewedCount       int     `json:"reviewed_count"`
	AvgOCRConfidence    float64 `json:"avg_ocr_confidence"`
	MedicalCertCount    int     `json:"medical_cert_count"`
	AvgProcessingTimeMs int64   `json:"avg_processing_time_ms"`
	MinProcessingTimeMs int64   `json:"min_processing_time_ms"`
	MaxProcessingTimeMs int64   `json:"max_processing_time_ms"`
}

type ReportRepository interface {
	GetSummary(ctx context.Context) (*ReportSummary, error)
	GetExpirationAlerts(ctx context.Context, daysAhead int) ([]ExpirationAlert, error)
	GetAnonymizedRecords(ctx context.Context, start, end time.Time) ([]AnonymizedRecord, error)
	GetPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error)
}