package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mqtt-streaming-server/broker"
	"mqtt-streaming-server/ocr"
	"mqtt-streaming-server/routes"
)

func NewTLSConfig() *tls.Config {
	certpool := x509.NewCertPool()

	pemCerts, err := os.ReadFile("/run/secrets/ca.crt")
	if err != nil {
		panic(err)
	}

	certpool.AppendCertsFromPEM(pemCerts)

	cert, err := tls.LoadX509KeyPair("/run/secrets/web.crt", "/run/secrets/web.key")
	if err != nil {
		panic(err)
	}

	return &tls.Config{
		RootCAs:      certpool,
		ClientCAs:    certpool,
		Certificates: []tls.Certificate{cert},
		// #nosec G402 -- hostname verification is handled by RootCAs and service certificate setup.
		InsecureSkipVerify: false,
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := fmt.Sprintf(
		"mongodb://%s:%s@mongo-db:27017/?authSource=admin",
		os.Getenv("MONGO_INITDB_ROOT_USERNAME"),
		os.Getenv("MONGO_INITDB_ROOT_PASSWORD"),
	)

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println("Failed to connect to MongoDB:", err)
		panic(err)
	}

	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	db := mongoClient.Database("mqtt-streaming-server")
	fmt.Println("Connected to MongoDB!")

	ocrServiceURL := os.Getenv("OCR_SERVICE_URL")
	if ocrServiceURL == "" {
		ocrServiceURL = "http://ocr-sandbox:9000"
	}

	ocrClient := ocr.NewClient(ocrServiceURL)
	brokerHandler := broker.NewBrokerHandler(db, ocrClient)

	tlsconfig := NewTLSConfig()

	opts := mqtt.NewClientOptions()
	opts.AddBroker("ssl://broker:8883")
	opts.SetClientID("web").SetTLSConfig(tlsconfig)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := client.Subscribe("ssproject/images/#", 0, brokerHandler.HandlePhoto); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	if token := client.Subscribe("register/#", 0, brokerHandler.RegisterDevice); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	if token := client.Subscribe("device/id/#", 0, brokerHandler.DisconnectDevice); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	handler := routes.InitRoutes(db, client)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Println("Starting HTTP server on port 8080...")

		srv := &http.Server{
			Addr:         ":8080",
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		if err := srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	<-c
}
