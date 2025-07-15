package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

func main() {
	var (
		databaseURI   = flag.String("uri", "", "Database URI (required)")
		migrationsDir = flag.String("dir", "migrations", "Migrations directory")
	)
	flag.Parse()

	if *databaseURI == "" {
		log.Fatal("Database URI is required. Use -uri flag")
	}

	// Подключаемся к базе данных
	db, err := sql.Open("postgres", *databaseURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully")

	// Читаем и выполняем миграции
	migrationFile := filepath.Join(*migrationsDir, "001_init_schema.sql")

	fmt.Printf("Executing migration: %s\n", migrationFile)

	content, err := os.ReadFile(migrationFile)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	// Выполняем SQL
	_, err = db.Exec(string(content))
	if err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	fmt.Println("Migration completed successfully!")
}
