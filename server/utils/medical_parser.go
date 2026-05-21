package utils

import (
	"regexp"
	"strings"
	"time"
)

type MedicalData struct {
	UnitateMedicala        string `json:"unitate_medicala" bson:"unitate_medicala"`                 // UNITATEA MEDICALA
	AdresaUnitateMedicala  string `json:"adresa_unitate_medicala" bson:"adresa_unitate_medicala"`   // ADRESA (sus)
	TelefonUnitateMedicala string `json:"telefon_unitate_medicala" bson:"telefon_unitate_medicala"` // TEL / FAX (sus)

	NumarFisa string `json:"numar_fisa" bson:"numar_fisa"` // FISA DE APTITUDINE NR.

	SocietateUnitate string `json:"societate_unitate" bson:"societate_unitate"` // Societate, unitate, etc.
	AdresaAngajator  string `json:"adresa_angajator" bson:"adresa_angajator"`   // Adresa (jos)
	TelefonAngajator string `json:"telefon_angajator" bson:"telefon_angajator"` // Telefon / Fax (jos)

	Nume    string `json:"nume" bson:"nume"`       // NUME
	Prenume string `json:"prenume" bson:"prenume"` // PRENUME
	CNP     string `json:"cnp" bson:"cnp"`         // CNP

	ProfesieFunctie string `json:"profesie_functie" bson:"profesie_functie"` // Profesie / functie
	LocDeMunca      string `json:"loc_de_munca" bson:"loc_de_munca"`         // Locul de munca

	TipControl       string    `json:"tip_control" bson:"tip_control"`               // Angajare, Control medical periodic, etc.
	ControlAngajare  bool      `json:"control_angajare" bson:"control_angajare"`     // Angajare (checkbox)
	ControlPeriodic  bool      `json:"control_periodic" bson:"control_periodic"`     // Control medical periodic (checkbox)
	ControlAdaptare  bool      `json:"control_adaptare" bson:"control_adaptare"`     // Adaptare (checkbox)
	ControlReluare   bool      `json:"control_reluare" bson:"control_reluare"`       // Reluarea muncii (checkbox)
	ControlSupraveghere bool   `json:"control_supraveghere" bson:"control_supraveghere"` // Supraveghere speciala (checkbox)
	ControlAlte      bool      `json:"control_alte" bson:"control_alte"`             // Alte (checkbox)

	AvizMedical      string    `json:"aviz_medical" bson:"aviz_medical"`             // APT, APT CONDITIONAT, etc.
	AvizApt          bool      `json:"aviz_apt" bson:"aviz_apt"`                     // APT (checkbox)
	AvizAptConditionat bool    `json:"aviz_apt_conditionat" bson:"aviz_apt_conditionat"` // APT CONDITIONAT (checkbox)
	AvizInaptTemporar bool     `json:"aviz_inapt_temporar" bson:"aviz_inapt_temporar"`   // INAPT TEMPORAR (checkbox)
	AvizInapt        bool      `json:"aviz_inapt" bson:"aviz_inapt"`                 // INAPT (checkbox)

	Recomandari      string    `json:"recomandari" bson:"recomandari"`               // RECOMANDARI field
	Data             time.Time `json:"data" bson:"data"`                             // Data
	DataUrmExaminari time.Time `json:"data_urm_examinari" bson:"data_urm_examinari"` // Data urmatoarei examinari
}

func ParseMedicalCertificate(ocrText string) *MedicalData {
	if ocrText == "" || ocrText == "OCR failed" {
		return nil
	}

	data := &MedicalData{}

	data.UnitateMedicala = extractField(ocrText, `UNITATEA\s+MEDICALA:\s*([^\n]+)`)
	if data.UnitateMedicala == "" {
		data.UnitateMedicala = extractField(ocrText, `MEDICALA:\s*([^\n]+)`)
	}

	parts := strings.Split(ocrText, "Societate")
	topPart := parts[0]
	bottomPart := ocrText
	if len(parts) > 1 {
		bottomPart = "Societate" + parts[1] // Include "Societate" back for regex matching
	}

	data.AdresaUnitateMedicala = extractField(topPart, `ADRESA:\s*([^\n]+)`)

	data.TelefonUnitateMedicala = extractField(topPart, `TEL:\s*([^\n]+)`)

	data.NumarFisa = extractField(ocrText, `FISA\s+DE\s+APTITUDINE\s+NR\.?\s*(\d+)`)

	data.NumarFisa = extractField(ocrText, `(?i)FISA\s+DE\s+APTITUDINE\s+NR[\.:]?\s*(\d+)`)

	data.SocietateUnitate = extractMultilineField(bottomPart, `(?i)Soci[ec]tate,?\s*unitate,?\s*(?:etc[\.:]?)?\s*([^\n]+(?:\n[^\n]+)?)`)

	if data.SocietateUnitate == "" {
		data.SocietateUnitate = extractField(bottomPart, `(?i)(UNIVERSITATEA\s+(?:NATIONALA\s+DE\s+STIINTA\s+SI\s+TEHNOLOGIE\s+)?POLITEHNICA\s+(?:DIN\s+)?[A-Z]+)`)
	}

	data.AdresaAngajator = extractField(bottomPart, `(?i)Adresa[:;]?\s*([^\n]+)`)

	data.TelefonAngajator = extractField(bottomPart, `(?i)(?:Telefon|Fax)[:;]?\s*([^\n]+)`)

	data.Nume = extractField(ocrText, `(?i)NUME[:;]?\s*([A-Za-z\s]+)`)
	data.Prenume = extractField(ocrText, `(?i)PRENUME[:;]?\s*([A-Za-z\s]+)`)
	data.CNP = extractField(ocrText, `(?i)CNP[:;]?\s*(\d+)`)

	data.ProfesieFunctie = extractField(ocrText, `(?i)Profesie\s*[\/\|]\s*functie[:;]?\s*([^\n]+)`)
	
	data.LocDeMunca = extractField(ocrText, `(?i)Locul?\s+de\s+munca[:;]?\s*([^\n]+)`)

	isBoxChecked := func(text string, currentLabel, nextLabel string) bool {
		curIdx := strings.Index(text, currentLabel)
		if curIdx == -1 {
			return false
		}
		
		searchStart := curIdx + len(currentLabel)
		var gapText string
		if nextLabel != "" {
			nextIdx := strings.Index(text[searchStart:], nextLabel)
			if nextIdx == -1 {
				end := searchStart + 20
				if end > len(text) {
					end = len(text)
				}
				gapText = text[searchStart:end]
			} else {
				gapText = text[searchStart : searchStart+nextIdx]
			}
		} else {
			end := searchStart + 20
			if end > len(text) {
				end = len(text)
			}
			gapText = text[searchStart:end]
		}

		emptyBoxRegex := regexp.MustCompile(`\[\s*[\[\]\-\|]?\s*\]`)
		if emptyBoxRegex.MatchString(gapText) {
			return false // Found an empty box
		}
		
		return true
	}

	tempTop := topPart
	tempTop = strings.ReplaceAll(tempTop, "Roluarca", "Reluarea")
	tempTop = strings.ReplaceAll(tempTop, "Ane", "Alte")
	
	rowStart := strings.Index(tempTop, "Angajare")
	if rowStart != -1 {
		rowText := tempTop[rowStart:] // Work from Angajare onwards
		
		data.ControlAngajare = isBoxChecked(rowText, "Angajare", "Control")
		data.ControlPeriodic = isBoxChecked(rowText, "Control", "Adaptare")
		data.ControlAdaptare = isBoxChecked(rowText, "Adaptare", "Reluarea")
		data.ControlReluare = isBoxChecked(rowText, "Reluarea", "Supraveghere")
		data.ControlSupraveghere = isBoxChecked(rowText, "Supraveghere", "Alte")
		data.ControlAlte = isBoxChecked(rowText, "Alte", "")
	}

	if data.ControlAngajare {
		data.TipControl = "Angajare"
	} else if data.ControlPeriodic || data.ControlAdaptare {
		if data.ControlAdaptare {
			data.TipControl = "Adaptare"
		} else {
			data.TipControl = "Control medical periodic"
		}
	} else if data.ControlReluare {
		data.TipControl = "Reluarea muncii"
	} else if data.ControlSupraveghere {
		data.TipControl = "Supraveghere speciala"
	}

	tempBottom := ocrText
	tempBottom = strings.ReplaceAll(tempBottom, "ApT", "APT")
	tempBottom = strings.ReplaceAll(tempBottom, "aerconpmonat", "APT CONDITIONAT")
	
	avizStart := strings.Index(tempBottom, "AVIZ MEDICAL")
	if avizStart != -1 {
		avizText := tempBottom[avizStart:]
		
		aptIdx := -1
		aptCondIdx := -1
		inaptTempIdx := -1
		inaptIdx := -1
		
		aptRegex := regexp.MustCompile(`APT[:;\s]+`)
		loc := aptRegex.FindStringIndex(avizText)
		if loc != nil {
			aptIdx = loc[0]
			aptIdx = loc[1] // Re-assign to end of match for gap start
		} else {
			aptRegex = regexp.MustCompile(`APT`)
			loc = aptRegex.FindStringIndex(avizText)
			if loc != nil {
				aptIdx = loc[1]
			}
		}

		aptCondRegex := regexp.MustCompile(`(?i)CONDITIONAT`) // Simplified from APT CONDITIONAT
		loc = aptCondRegex.FindStringIndex(avizText)
		if loc != nil {
			aptCondIdx = loc[0]
		}
		
		if aptIdx != -1 && aptCondIdx != -1 && aptIdx < aptCondIdx {
			gap := avizText[aptIdx : aptCondIdx]
			emptyBoxRegex := regexp.MustCompile(`\[\s*[\[\]\-\|]?\s*\]`)
			data.AvizApt = !emptyBoxRegex.MatchString(gap)
		}
		
		inaptTempRegex := regexp.MustCompile(`(?i)INAPT\s*TEMPORAR`)
		loc = inaptTempRegex.FindStringIndex(avizText)
		if loc != nil {
			inaptTempIdx = loc[0]
		}
		
		if aptCondIdx != -1 {
			end := inaptTempIdx
			if end == -1 { end = len(avizText) } // Fallback
			if end > aptCondIdx {
				gap := avizText[aptCondIdx+len("APT CONDITIONAT") : end]
				emptyBoxRegex := regexp.MustCompile(`\[\s*[\[\]\-\|]?\s*\]`)
				data.AvizAptConditionat = !emptyBoxRegex.MatchString(gap)
			}
		}

		startSearch := 0
		if inaptTempIdx != -1 {
			startSearch = inaptTempIdx + len("INAPT TEMPORAR")
			
			inaptRegex := regexp.MustCompile(`(?i)INAPT[:\s]+`)
			loc = inaptRegex.FindStringIndex(avizText[startSearch:])
			if loc != nil {
				inaptIdx = startSearch + loc[0]
				
				gap := avizText[startSearch : inaptIdx]
				emptyBoxRegex := regexp.MustCompile(`\[\s*[\[\]\-\|]?\s*\]`)
				data.AvizInaptTemporar = !emptyBoxRegex.MatchString(gap)
				
				end := inaptIdx + 15
				if end > len(avizText) { end = len(avizText) }
				gapLast := avizText[inaptIdx+len("INAPT") : end]
				data.AvizInapt = !emptyBoxRegex.MatchString(gapLast)
				
			}
		}
	}

	if data.AvizApt {
		data.AvizMedical = "APT"
	} else if data.AvizAptConditionat {
		data.AvizMedical = "APT CONDITIONAT"
	} else if data.AvizInaptTemporar {
		data.AvizMedical = "INAPT TEMPORAR"
	} else if data.AvizInapt {
		data.AvizMedical = "INAPT"
	}

	dateStr := extractField(ocrText, `(?i)Data[:;]?\s*(\d{2}[\.\/\-]\d{2}[\.\/\-]\d{4})`)
	if dateStr != "" {
		normalizedDate := strings.ReplaceAll(dateStr, ".", "/")
		normalizedDate = strings.ReplaceAll(normalizedDate, "-", "/")
		if t, err := time.Parse("02/01/2006", normalizedDate); err == nil {
			data.Data = t
		}
	}

	nextDateStr := extractField(ocrText, `(?i)Data\s+urmatoarei\s+examinari[:;]?\s*(\d{2}[\.\/\-]\d{2}[\.\/\-]\d{4})`)
	if nextDateStr != "" {
		normalizedDate := strings.ReplaceAll(nextDateStr, ".", "/")
		normalizedDate = strings.ReplaceAll(normalizedDate, "-", "/")
		if t, err := time.Parse("02/01/2006", normalizedDate); err == nil {
			data.DataUrmExaminari = t
		}
	}
	
	return data
}

func extractField(text, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func extractMultilineField(text, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		result := strings.TrimSpace(matches[1])
		result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
		return result
	}
	return ""
}

func containsChecked(text, fieldName string) bool {
	patternRight := fieldName + `[:\s\._-]*\[?\s*[Xx]\s*\]?`
	if regexp.MustCompile(patternRight).MatchString(text) {
		return true
	}
	
	patternLeft := `\[?\s*[Xx]\s*\]?[:\s\._-]*` + fieldName
	if regexp.MustCompile(patternLeft).MatchString(text) {
		return true
	}

	return false
}

func IsMedicalCertificate(ocrText string) bool {
	keywords := []string{
		"MEDICINA MUNCII",
		"FISA DE APTITUDINE",
		"AVIZ MEDICAL",
		"APT",
	}
	
	matchCount := 0
	for _, keyword := range keywords {
		if strings.Contains(ocrText, keyword) {
			matchCount++
		}
	}
	
	return matchCount >= 2
}
