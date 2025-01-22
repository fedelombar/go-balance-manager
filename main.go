// main.go

package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	host := flag.String("dbhost", "db", "Database host") // Changed to 'db' for Docker
	port := flag.Int("dbport", 5432, "Database port")
	user := flag.String("dbuser", "golang_user", "Database user")
	password := flag.String("dbpass", "golang_pass", "Database password")
	dbname := flag.String("dbname", "golang_db", "Database name")
	addr := flag.String("addr", ":8080", "Server address")
	seed := flag.Bool("seed", false, "Seed predefined users")

	flag.Parse()

	// Initialize storage
	store, err := NewPostgresStore(*host, *port, *user, *password, *dbname)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer store.Close()

	// Initialize database
	if err := store.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Seed data if flag is set
	if *seed {
		if err := store.EnsurePredefinedUsers(); err != nil {
			log.Fatalf("Failed to seed users: %v", err)
		}
		fmt.Println("Predefined users seeded successfully.")
	}

	server := NewAPIServer(store)
	server.Run(*addr)
}
