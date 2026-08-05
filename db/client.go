package db

import (
	"context"
	"log"

	"googledominator-backend/config"
)

type DBClient struct {
	Prisma *PrismaClient
	IsConnected bool
}

var Instance *DBClient

func InitDB(cfg *config.Config) (*DBClient, error) {
	client := NewClient()
	
	log.Println("[INFO] Connecting to PostgreSQL via Prisma Client...")
	err := client.Connect()
	if err != nil {
		log.Printf("[WARNING] Could not connect to PostgreSQL database: %v", err)
		log.Println("[INFO] Operating in fallback mode with mock data seeding enabled")
		Instance = &DBClient{
			Prisma:      client,
			IsConnected: false,
		}
		return Instance, nil
	}

	log.Println("[INFO] Successfully connected to PostgreSQL via Prisma!")
	Instance = &DBClient{
		Prisma:      client,
		IsConnected: true,
	}

	// Seed initial data into PostgreSQL if empty
	SeedDB(client)

	return Instance, nil
}

func (d *DBClient) Disconnect() {
	if d.Prisma != nil && d.IsConnected {
		if err := d.Prisma.Disconnect(); err != nil {
			log.Printf("[ERROR] Failed to disconnect Prisma client: %v", err)
		} else {
			log.Println("[INFO] Prisma client disconnected safely")
		}
	}
}

func (d *DBClient) Context() context.Context {
	return context.Background()
}
