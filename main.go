package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/opensearch-project/opensearch-go/v4"

	// "github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"

	"github.com/gin-gonic/gin"
	"github.com/tanush-128/openzo_backend/search/config"
	handlers "github.com/tanush-128/openzo_backend/search/internal/api"
	"github.com/tanush-128/openzo_backend/search/internal/pb"
	"github.com/tanush-128/openzo_backend/search/internal/service"
	// "google.golang.org/grpc"
)

var UserClient pb.UserServiceClient

type User2 struct {
}

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to load config: %w", err))
	}

	client, err := opensearchapi.NewClient(
		opensearchapi.Config{
			Client: opensearch.Config{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // For testing only. Use certificate for validation.
				},
				Addresses: []string{cfg.OpenSearchURL},
				Username:  cfg.OpenSearchUsername, // For testing only. Don't store credentials in code.
				Password:  cfg.OpenSearchPassword,
			},
		},
	)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	searchService := service.NewSearchService(client)
	// searchService.SearchStoresByPincode("123456", "Think bigg")
	handlers := handlers.NewSearchHandler(searchService)

	go consumeKafkaTopicStores(client)
	go consumeKafkaTopicProducts(client)

	// Initialize HTTP server with Gin
	router := gin.Default()
	// handler := handlers.NewHandler(&searchService)

	router.GET("ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.GET("/stores/:pincode", handlers.SearchStoresByPincode)
	router.GET("/products/:pincode", handlers.SearchProductsByPincode)
	// // router.Use(middlewares.JwtMiddleware(c))

	// // router.Use(middlewares.JwtMiddleware)

	router.Run(fmt.Sprintf(":%s", cfg.HTTPPort))

}

type Notification struct {
	Message  string `json:"message"`
	FCMToken string `json:"fcm_token"`
	Data     string `json:"data,omitempty"`
	Topic    string `json:"topic,omitempty"`
}

func consumeKafkaTopicStores(client *opensearchapi.Client) {
	conf := ReadConfig()
	topic := "stores"
	conf["group.id"] = "go-group-1"
	conf["auto.offset.reset"] = "earliest"

	for {
		consumer, err := kafka.NewConsumer(&conf)
		if err != nil {
			log.Printf("Error creating consumer: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		err = consumer.SubscribeTopics([]string{topic}, nil)
		if err != nil {
			log.Printf("Error subscribing  to new  topic: %v. Retrying in 5 seconds...", err)
			consumer.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		run := true
		for run {
			e := consumer.Poll(1000)
			switch ev := e.(type) {
			case *kafka.Message:
				ctx := context.Background()
				_, err := client.Index(ctx, opensearchapi.IndexReq{
					Index:      "stores-index",
					DocumentID: string(ev.Key),
					Body:       strings.NewReader(string(ev.Value)),
				})

				if err != nil {
					log.Fatalf("Error indexing document: %v", err)
				}
				log.Printf("Message on %s: %s\n", ev.TopicPartition, string(ev.Value))

			case kafka.Error:
				log.Printf("Error: %v", ev)
				run = false
			}
		}

		log.Println("Consumer disconnected. Reconnecting in 5 seconds...")
		consumer.Close()
		time.Sleep(5 * time.Second)
	}
}

func consumeKafkaTopicProducts(client *opensearchapi.Client) {
	conf := ReadConfig()
	topic := "products"
	conf["group.id"] = "go-group-1"
	conf["auto.offset.reset"] = "earliest"

	for {
		consumer, err := kafka.NewConsumer(&conf)
		if err != nil {
			log.Printf("Error creating consumer: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		err = consumer.SubscribeTopics([]string{topic}, nil)
		if err != nil {
			log.Printf("Error subscribing to topic: %v. Retrying in 5 seconds...", err)
			consumer.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		run := true
		for run {
			e := consumer.Poll(1000)
			switch ev := e.(type) {
			case *kafka.Message:
				ctx := context.Background()
				_, err := client.Index(ctx, opensearchapi.IndexReq{
					Index:      "products-index",
					DocumentID: string(ev.Key),
					Body:       strings.NewReader(string(ev.Value)),
				})
				print("hello")
				if err != nil {

					log.Printf("Error indexing document: %v", err)
				}
				log.Printf("Message on %s: %s\n", ev.TopicPartition, string(ev.Value))

			case kafka.Error:
				log.Printf("Error: %v", ev)
				run = false
			}
		}

		log.Println("Consumer disconnected. Reconnecting in 1 second...")
		consumer.Close()
		time.Sleep(1 * time.Second)
	}
}
