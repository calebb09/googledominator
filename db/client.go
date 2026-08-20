package db

import (
	"context"
	"fmt"
	"log"

	"googledominator-backend/config"
)

type DBClient struct {
	Prisma      *PrismaClient
	IsConnected bool
}

var Instance *DBClient

func InitDB(cfg *config.Config) (*DBClient, error) {
	client := NewClient(WithDatasourceURL(cfg.DatabaseURL))

	log.Println("[INFO] Connecting to PostgreSQL via Prisma Client...")
	err := client.Connect()
	if err != nil {
		return nil, fmt.Errorf("connect to PostgreSQL: %w", err)
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
