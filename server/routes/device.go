package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/mongo"

	"mqtt-streaming-server/domain"
	"mqtt-streaming-server/repository"
)

type DeviceController struct {
	DeviceRepository domain.DeviceRepository
	mqttClient       mqtt.Client
}

func InitDeviceRoutes(db *mongo.Database, mqttClient mqtt.Client, mux *http.ServeMux) {
	deviceController := &DeviceController{
		DeviceRepository: repository.NewDeviceRepository(db),
		mqttClient:       mqttClient,
	}

	// TODO: Implement authentication - See docs/AUTH_IMPLEMENTATION.md
	mux.Handle("/devices", withAuth(withAnyPage(http.HandlerFunc(deviceController.GetDevices), "devices")))
	mux.Handle("/devices/switch", withAuth(withAdminOnly(http.HandlerFunc(deviceController.SwitchDeviceMode))))
	mux.Handle("/devices/command", withAuth(withAnyPage(http.HandlerFunc(deviceController.SendCommand), "devices")))
}

func (ctlr DeviceController) SwitchDeviceMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role, _ := r.Context().Value("role").(string)
	pages, _ := r.Context().Value("pages").([]string)
	canAccess := role == "admin"
	if !canAccess {
		for _, page := range pages {
			if page == "devices" {
				canAccess = true
				break
			}
		}
	}
	if !canAccess {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var device struct {
		ID   string `json:"id"`
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	topic := fmt.Sprintf("setup/%s", device.ID)
	if token := ctlr.mqttClient.Publish(topic, 0, false, "start "+device.Mode); token.Wait() && token.Error() != nil {
		http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (ctlr DeviceController) GetDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role, _ := r.Context().Value("role").(string)
	if role != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	// Fetch devices from the database
	devices, err := ctlr.DeviceRepository.GetAllDevices(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch devices", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices) // #nosec G104 -- encoding errors on http.ResponseWriter are non-actionable
}

func (ctlr DeviceController) SendCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		DeviceID string `json:"device_id"`
		Command  string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate command
	validCommands := map[string]bool{
		"CAPTURE":    true,
		"START-LIVE": true,
		"STOP-LIVE":  true,
	}
	if !validCommands[request.Command] {
		http.Error(w, "Invalid command. Must be CAPTURE, START-LIVE, or STOP-LIVE", http.StatusBadRequest)
		return
	}

	// Publish command to MQTT topic ssproject/commands
	topic := "ssproject/commands"
	payload := request.Command
	if token := ctlr.mqttClient.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
		http.Error(w, "Failed to publish command", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{ // #nosec G104 -- encoding errors on http.ResponseWriter are non-actionable
		"status":  "success",
		"message": fmt.Sprintf("Command %s sent to device %s", request.Command, request.DeviceID),
	})
}
