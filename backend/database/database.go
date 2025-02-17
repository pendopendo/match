package database

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

type Config struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func Init() {
	// Use DATABASE_URL from environment variable
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Default local connection string
		connStr = "postgres://postgres:postgres@localhost:5432/matchme?sslmode=disable"
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Test the connection
	err = DB.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Successfully connected to database")
	configure(DB)
	createTables()
}

func configure(db *sql.DB) {
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
}

func createTables() {
	log.Println("Starting table creation...")

	// Create users table
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Printf("Failed to create users table: %v", err)
		log.Fatal("Database initialization failed")
	}
	log.Println("Users table created/verified successfully")

	// Create profiles table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS profiles (
			user_id INTEGER PRIMARY KEY REFERENCES users(id),
			name TEXT,
			bio TEXT,
			profile_picture TEXT,
			location TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Printf("Failed to create profiles table: %v", err)
		log.Printf("This might be due to missing users table reference: %v", err)
		log.Fatal("Database initialization failed")
	}
	log.Println("Profiles table created/verified successfully")

	// Create user_bios table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS user_bios (
			user_id INTEGER PRIMARY KEY REFERENCES users(id),
			interests TEXT[],
			hobbies TEXT[],
			music_preferences TEXT[],
			food_preferences TEXT[],
			looking_for TEXT[],
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Printf("Failed to create user_bios table: %v", err)
		log.Printf("This might be due to missing users table reference or array type issues: %v", err)
		log.Fatal("Database initialization failed")
	}
	log.Println("User_bios table created/verified successfully")

	// Create connections table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS connections (
			id SERIAL PRIMARY KEY,
			user_id_1 INTEGER REFERENCES users(id),
			user_id_2 INTEGER REFERENCES users(id),
			last_message TEXT,
			last_message_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(user_id_1, user_id_2)
		)
	`)
	if err != nil {
		log.Printf("Failed to create connections table: %v", err)
		log.Printf("This might be due to missing users table reference or unique constraint issues: %v", err)
		log.Fatal("Database initialization failed")
	}
	log.Println("Connections table created/verified successfully")

	// Create messages table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			connection_id INTEGER REFERENCES connections(id),
			sender_id INTEGER REFERENCES users(id),
			content TEXT NOT NULL,
			read BOOLEAN DEFAULT false,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Printf("Failed to create messages table: %v", err)
		log.Printf("This might be due to missing connections or users table references: %v", err)
		log.Fatal("Database initialization failed")
	}
	log.Println("Messages table created/verified successfully")

	log.Println("All database tables created/verified successfully!")
}
