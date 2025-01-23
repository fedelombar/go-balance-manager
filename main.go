package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	host := flag.String("dbhost", getEnv("DB_HOST", "db"), "Database host")
	port := flag.Int("dbport", getEnvAsInt("DB_PORT", 5432), "Database port")
	user := flag.String("dbuser", getEnv("DB_USER", "golang_user"), "Database user")
	password := flag.String("dbpass", getEnv("DB_PASS", "golang_pass"), "Database password")
	dbname := flag.String("dbname", getEnv("DB_NAME", "golang_db"), "Database name")
	addr := flag.String("addr", getEnv("APP_ADDR", ":8080"), "Server address")
	seed := flag.Bool("seed", false, "Seed predefined users")

	flag.Parse()

	store, err := NewPostgresStore(*host, *port, *user, *password, *dbname)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer store.Close()

	if err := store.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if *seed {
		if err := store.EnsurePredefinedUsers(); err != nil {
			log.Fatalf("Failed to seed users: %v", err)
		}
		fmt.Println("Predefined users seeded successfully.")
		return
	}

	server := NewAPIServer(store)
	server.Run(*addr)
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(name string, defaultVal int) int {
	if valueStr, exists := os.LookupEnv(name); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultVal
}
