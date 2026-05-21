package broker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/otiai10/gosseract/v2"
	"go.mongodb.org/mongo-driver/mongo"

	"mqtt-streaming-server/domain"
	"mqtt-streaming-server/repository"
	"mqtt-streaming-server/utils"
)

const (
	ocrConfidenceThreshold = 95.0
)

type BrokerHandler struct {
	photoRepository  domain.PhotoRepository
	deviceRepository domain.DeviceRepository
	ocrClient        *gosseract.Client
}

func NewBrokerHandler(db *mongo.Database, ocrClient *gosseract.Client) BrokerHandler {
	return BrokerHandler{
		photoRepository:  repository.NewPhotoRepository(db),
		deviceRepository: repository.NewDeviceRepository(db),
		ocrClient:        ocrClient,
	}
}

func (b BrokerHandler) HandlePhoto(_ mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	var deviceID string
	if topic == "ssproject/images" {
		deviceID = "camera_stream"
	} else if len(topic) > len("ssproject/images/") {
		deviceID = topic[len("ssproject/images/"):]
	} else {
		deviceID = "unknown"
	}

	ctx := context.Background()
	fmt.Println("Received message on topic:", msg.Topic())

	device, err := b.deviceRepository.GetByID(ctx, deviceID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Printf("Device ID not found: %s. Auto-registering...\n", deviceID)
			newDevice := &domain.Device{
				DeviceID:     deviceID,
				DeviceName:   "Unknown Device (" + deviceID + ")",
				DeviceStatus: "active",
			}
			if err := b.deviceRepository.Save(ctx, newDevice); err != nil {
				fmt.Printf("Failed to auto-register device: %v\n", err)
				return
			}
			device = newDevice
		} else {
			fmt.Printf("Failed to check device ID: %v\n", err)
			return
		}
	}
	fmt.Printf("Received photo from device: %s\n", device.DeviceName)

	body := msg.Payload()
	_, imageType, err := image.DecodeConfig(bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Failed to decode image: %v\n", err)
		return
	}
	fmt.Printf("Image type: %s\n", imageType)

	processingStart := time.Now()

	text, confidence, err := b.extractTextFromImage(body)
	if err != nil {
		fmt.Printf("Failed to extract text from image: %v\n", err)
		text = "OCR failed"
		confidence = 0
	}
	fmt.Printf("OCR confidence: %.2f%%\n", confidence)

	var medicalData *utils.MedicalData
	if utils.IsMedicalCertificate(text) {
		medicalData = utils.ParseMedicalCertificate(text)
		if medicalData != nil {
			fmt.Printf("Extracted medical data: %+v\n", medicalData)
		}
	}

	processingTimeMs := time.Since(processingStart).Milliseconds()
	fmt.Printf("Processing time: %dms\n", processingTimeMs)

	timestamp := time.Now().UTC()

	photo := &domain.Photo{
		ImageType:        imageType,
		Timestamp:        timestamp,
		DeviceID:         deviceID,
		Text:             text,
		OCRConfidence:    confidence,
		ProcessingTimeMs: processingTimeMs,
	}

	if confidence < ocrConfidenceThreshold {
		photo.NeedsReview = true
		photo.ReviewReason = fmt.Sprintf(
			"OCR confidence %.1f%% is below the %.0f%% threshold; extracted fields require human verification",
			confidence, ocrConfidenceThreshold,
		)
		fmt.Printf("Photo flagged for review: %s\n", photo.ReviewReason)
	}

	if medicalData != nil {
		photo.UnitateMedicala = medicalData.UnitateMedicala
		photo.AdresaUnitateMedicala = medicalData.AdresaUnitateMedicala
		photo.TelefonUnitateMedicala = medicalData.TelefonUnitateMedicala
		photo.NumarFisa = medicalData.NumarFisa
		photo.SocietateUnitate = medicalData.SocietateUnitate
		photo.AdresaAngajator = medicalData.AdresaAngajator
		photo.TelefonAngajator = medicalData.TelefonAngajator
		photo.Nume = medicalData.Nume
		photo.Prenume = medicalData.Prenume
		photo.CNP = medicalData.CNP
		photo.ProfesieFunctie = medicalData.ProfesieFunctie
		photo.LocDeMunca = medicalData.LocDeMunca
		photo.TipControl = medicalData.TipControl
		photo.ControlAngajare = medicalData.ControlAngajare
		photo.ControlPeriodic = medicalData.ControlPeriodic
		photo.ControlAdaptare = medicalData.ControlAdaptare
		photo.ControlReluare = medicalData.ControlReluare
		photo.ControlSupraveghere = medicalData.ControlSupraveghere
		photo.ControlAlte = medicalData.ControlAlte
		photo.AvizMedical = medicalData.AvizMedical
		photo.AvizApt = medicalData.AvizApt
		photo.AvizAptConditionat = medicalData.AvizAptConditionat
		photo.AvizInaptTemporar = medicalData.AvizInaptTemporar
		photo.AvizInapt = medicalData.AvizInapt
		photo.Recomandari = medicalData.Recomandari
		photo.Data = medicalData.Data
		photo.DataUrmExaminari = medicalData.DataUrmExaminari
	}

	err = b.photoRepository.Save(ctx, photo)
	if err != nil {
		fmt.Printf("Failed to insert photo into MongoDB: %v\n", err)
		return
	}

	keyName := fmt.Sprintf("photos/%d.%s", timestamp.Unix(), imageType)
	if err := utils.SaveToLocal(body, keyName); err != nil {
		fmt.Printf("Failed to save photo locally: %v\n", err)
		return
	}
	fmt.Printf("Photo saved: %s (needsReview=%v)\n", keyName, photo.NeedsReview)
}

// returns the OCR text and the average Tesseract word
func (b BrokerHandler) extractTextFromImage(imageData []byte) (string, float64, error) {
	b.ocrClient.SetImageFromBytes(imageData) // #nosec G104 -- error handled implicitly by Text()

	text, err := b.ocrClient.Text()
	if err != nil {
		return "", 0, fmt.Errorf("failed to extract text from image: %v", err)
	}

	boxes, err := b.ocrClient.GetBoundingBoxes(gosseract.RIL_WORD)
	if err != nil || len(boxes) == 0 {
		// OCR unavailable
		return text, 0, nil
	}

	var total float64
	count := 0
	for _, box := range boxes {
		if box.Confidence > 0 {
			total += float64(box.Confidence)
			count++
		}
	}
	if count == 0 {
		return text, 0, nil
	}
	return text, total / float64(count), nil
}

func (b BrokerHandler) RegisterDevice(_ mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	// topic is register/device_id
	deviceID := topic[len("register/"):]
	ctx := context.Background()
	fmt.Println("Received message on topic:", msg.Topic())
	body := msg.Payload()
	fmt.Printf("Received device registration: %s\n", body)

	var deviceName, ipAddress, port string
	var registration struct {
		Name string `json:"name"`
		IP   string `json:"ip"`
		Port string `json:"port"`
	}
	if err := json.Unmarshal(body, &registration); err == nil && registration.Name != "" {
		deviceName = registration.Name
		ipAddress = registration.IP
		port = registration.Port
	} else {
		deviceName = string(body)
	}

	// Check if device ID already exists
	_, err := b.deviceRepository.GetByID(ctx, deviceID)
	if err != nil && err != mongo.ErrNoDocuments {
		fmt.Printf("Failed to check device ID: %v\n", err)
		return
	}
	if err == mongo.ErrNoDocuments {
		// Device ID does not exist, insert it
		err = b.deviceRepository.Save(ctx, &domain.Device{
			DeviceID:     deviceID,
			DeviceName:   deviceName,
			DeviceStatus: "active",
			IPAddress:    ipAddress,
			Port:         port,
			LastSeen:     time.Now().UTC(),
		})
		if err != nil {
			fmt.Printf("Failed to insert device ID: %v\n", err)
			return
		}
		fmt.Printf("Device registered: %s (IP: %s, Port: %s)\n", deviceID, ipAddress, port)
		return
	}
	// Device ID already exists, update it
	err = b.deviceRepository.Update(ctx, deviceID, &domain.Device{
		DeviceID:     deviceID,
		DeviceName:   deviceName,
		DeviceStatus: "active",
		IPAddress:    ipAddress,
		Port:         port,
		LastSeen:     time.Now().UTC(),
	})
	if err != nil {
		fmt.Printf("Failed to update device ID: %v\n", err)
		return
	}
	fmt.Printf("Device updated: %s (IP: %s, Port: %s)\n", deviceID, ipAddress, port)
}

func (b BrokerHandler) DisconnectDevice(_ mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	// topic is device/id/device_id
	var deviceID string
	if len(topic) > len("device/id/") {
		deviceID = topic[len("device/id/"):]
	} else {
		return
	}

	ctx := context.Background()
	fmt.Println("Received message on topic:", msg.Topic())
	message := string(msg.Payload())
	fmt.Printf("Received device disconnection: %s\n", message)

	if message != "Device Disconnected" {
		fmt.Printf("Invalid disconnection message: %s\n", message)
		return
	}

	device, err := b.deviceRepository.GetByID(ctx, deviceID)
	if err != nil {
		return
	}
	if device.DeviceStatus != "active" {
		return
	}
	_ = b.deviceRepository.Update(ctx, deviceID, &domain.Device{
		DeviceID:     deviceID,
		DeviceStatus: "inactive",
		DeviceName:   device.DeviceName,
	})
}

func (b BrokerHandler) HandleCommand(_ mqtt.Client, msg mqtt.Message) {
	fmt.Println("Received command on topic:", msg.Topic())
	body := string(msg.Payload())
	fmt.Printf("Command payload: %s\n", body)
}
