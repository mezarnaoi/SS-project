package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"mqtt-streaming-server/domain"
)

type reportRepository struct {
	db *mongo.Database
}

func NewReportRepository(db *mongo.Database) *reportRepository {
	return &reportRepository{db: db}
}

func isMedicalFilter() bson.M {
	return bson.M{"$or": bson.A{
		bson.M{"aviz_medical": bson.M{"$ne": ""}},
		bson.M{"tip_control": bson.M{"$ne": ""}},
		bson.M{"numar_fisa": bson.M{"$ne": ""}},
		bson.M{"cnp": bson.M{"$ne": ""}},
	}}
}

func (repo *reportRepository) GetSummary(ctx context.Context) (*domain.ReportSummary, error) {
	col := repo.db.Collection("photos")
	now := time.Now().UTC()

	lastMonthStart := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.UTC)
	lastMonthEnd := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	thisMonthStart := lastMonthEnd
	thisMonthEnd := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	nextMonthStart := thisMonthEnd
	nextMonthEnd := time.Date(now.Year(), now.Month()+2, 1, 0, 0, 0, 0, time.UTC)

	medFilter := isMedicalFilter()
	summary := &domain.ReportSummary{}

	// Total checks: all photos recognised as medical certificates
	total, err := col.CountDocuments(ctx, medFilter)
	if err != nil {
		return nil, err
	}
	summary.TotalChecks = int(total)

	lastMonthFilter := bson.M{
		"$and": bson.A{
			medFilter,
			bson.M{"timestamp": bson.M{"$gte": lastMonthStart, "$lt": lastMonthEnd}},
		},
	}
	lastMonth, err := col.CountDocuments(ctx, lastMonthFilter)
	if err != nil {
		return nil, err
	}
	summary.ChecksLastMonth = int(lastMonth)

	thisMonthFilter := bson.M{
		"$and": bson.A{
			medFilter,
			bson.M{"timestamp": bson.M{"$gte": thisMonthStart, "$lt": thisMonthEnd}},
		},
	}
	thisMonth, err := col.CountDocuments(ctx, thisMonthFilter)
	if err != nil {
		return nil, err
	}
	summary.ChecksThisMonth = int(thisMonth)

	// Expiration alerts for next month
	expAlerts, err := repo.GetExpirationAlerts(ctx, 0)
	if err != nil {
		return nil, err
	}
	for _, a := range expAlerts {
		if !a.NextExamDate.Before(nextMonthStart) && a.NextExamDate.Before(nextMonthEnd) {
			summary.ExpiringNextMonth = append(summary.ExpiringNextMonth, a)
		}
	}
	if summary.ExpiringNextMonth == nil {
		summary.ExpiringNextMonth = []domain.ExpirationAlert{}
	}

	for _, pair := range []struct {
		value string
	}{
		{"APT"}, {"APT CONDITIONAT"}, {"INAPT TEMPORAR"}, {"INAPT"},
	} {
		n, err := col.CountDocuments(ctx, bson.M{"aviz_medical": pair.value})
		if err != nil {
			return nil, err
		}
		switch pair.value {
		case "APT":
			summary.FitCount = int(n)
		case "APT CONDITIONAT":
			summary.ConditionallyFitCount = int(n)
		case "INAPT TEMPORAR":
			summary.TemporarilyUnfitCount = int(n)
		case "INAPT":
			summary.UnfitCount = int(n)
		}
	}

	metrics, err := repo.GetPerformanceMetrics(ctx)
	if err != nil {
		return nil, err
	}
	summary.OCRSuccessRate = metrics.OCRSuccessRate
	summary.NeedsReviewCount = metrics.NeedsReviewCount
	summary.AvgProcessingTimeMs = metrics.AvgProcessingTimeMs

	return summary, nil
}

func (repo *reportRepository) GetExpirationAlerts(ctx context.Context, daysAhead int) ([]domain.ExpirationAlert, error) {
	col := repo.db.Collection("photos")
	now := time.Now().UTC()

	var end time.Time
	if daysAhead <= 0 {
		end = now.AddDate(1, 0, 0)
	} else {
		end = now.AddDate(0, 0, daysAhead)
	}

	cursor, err := col.Find(ctx, bson.M{
		"data_urm_examinari": bson.M{
			"$gte": now,
			"$lte": end,
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var alerts []domain.ExpirationAlert
	for cursor.Next(ctx) {
		var p domain.Photo
		if err := cursor.Decode(&p); err != nil {
			return nil, err
		}
		if err := DecryptPhotoFields(&p); err != nil {
			return nil, err
		}
		daysUntil := int(p.DataUrmExaminari.Sub(now).Hours() / 24)
		alerts = append(alerts, domain.ExpirationAlert{
			LastName:        p.Nume,
			FirstName:       p.Prenume,
			Company:         p.SocietateUnitate,
			Workplace:       p.LocDeMunca,
			NextExamDate:    p.DataUrmExaminari,
			DaysUntilExpiry: daysUntil,
			MedicalOpinion:  p.AvizMedical,
		})
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	if alerts == nil {
		alerts = []domain.ExpirationAlert{}
	}
	return alerts, nil
}

func (repo *reportRepository) GetAnonymizedRecords(ctx context.Context, start, end time.Time) ([]domain.AnonymizedRecord, error) {
	col := repo.db.Collection("photos")

	filter := bson.M{}
	if !start.IsZero() {
		filter["timestamp"] = bson.M{"$gte": start, "$lte": end}
	}

	cursor, err := col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []domain.AnonymizedRecord
	for cursor.Next(ctx) {
		var p domain.Photo
		if err := cursor.Decode(&p); err != nil {
			return nil, err
		}
		if err := DecryptPhotoFields(&p); err != nil {
			return nil, err
		}
		records = append(records, domain.AnonymizedRecord{
			ControlType:    p.TipControl,
			MedicalOpinion: p.AvizMedical,
			Company:        p.SocietateUnitate,
			Workplace:      p.LocDeMunca,
			JobTitle:       p.ProfesieFunctie,
			ExamDate:       p.Data,
			NextExamDate:   p.DataUrmExaminari,
			OCRConfidence:  p.OCRConfidence,
		})
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	if records == nil {
		records = []domain.AnonymizedRecord{}
	}
	return records, nil
}

func (repo *reportRepository) GetPerformanceMetrics(ctx context.Context) (*domain.PerformanceMetrics, error) {
	col := repo.db.Collection("photos")

	total, err := col.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	medicalCerts, err := col.CountDocuments(ctx, isMedicalFilter())
	if err != nil {
		return nil, err
	}

	const confidenceThreshold = 95.0
	flagged, err := col.CountDocuments(ctx, bson.M{"ocr_confidence": bson.M{"$lt": confidenceThreshold}})
	if err != nil {
		return nil, err
	}

	autoProcessed, err := col.CountDocuments(ctx, bson.M{"ocr_confidence": bson.M{"$gte": confidenceThreshold}})
	if err != nil {
		return nil, err
	}

	reviewed, err := col.CountDocuments(ctx, bson.M{"reviewed_by": bson.M{"$exists": true, "$ne": ""}})
	if err != nil {
		return nil, err
	}

	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.M{
			"_id":            nil,
			"avg_confidence": bson.M{"$avg": "$ocr_confidence"},
			"avg_latency":    bson.M{"$avg": bson.M{"$cond": bson.A{bson.M{"$gt": bson.A{"$processing_time_ms", 0}}, "$processing_time_ms", nil}}},
			"min_latency":    bson.M{"$min": bson.M{"$cond": bson.A{bson.M{"$gt": bson.A{"$processing_time_ms", 0}}, "$processing_time_ms", nil}}},
			"max_latency":    bson.M{"$max": bson.M{"$cond": bson.A{bson.M{"$gt": bson.A{"$processing_time_ms", 0}}, "$processing_time_ms", nil}}},
		}}},
	}
	cur, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var avgConf float64
	var avgLatency, minLatency, maxLatency int64
	if cur.Next(ctx) {
		var result struct {
			AvgConfidence float64 `bson:"avg_confidence"`
			AvgLatency    float64 `bson:"avg_latency"`
			MinLatency    int64   `bson:"min_latency"`
			MaxLatency    int64   `bson:"max_latency"`
		}
		if err := cur.Decode(&result); err == nil {
			avgConf = result.AvgConfidence
			avgLatency = int64(result.AvgLatency)
			minLatency = result.MinLatency
			maxLatency = result.MaxLatency
		}
	}

	var successRate float64
	if total > 0 {
		successRate = float64(autoProcessed) / float64(total) * 100
	}

	return &domain.PerformanceMetrics{
		TotalProcessed:      int(total),
		OCRSuccessCount:     int(autoProcessed),
		OCRSuccessRate:      successRate,
		NeedsReviewCount:    int(flagged),
		ReviewedCount:       int(reviewed),
		AvgOCRConfidence:    avgConf,
		MedicalCertCount:    int(medicalCerts),
		AvgProcessingTimeMs: avgLatency,
		MinProcessingTimeMs: minLatency,
		MaxProcessingTimeMs: maxLatency,
	}, nil
}
