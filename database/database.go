package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() {
	var err error

	if os.Getenv("GO_ENV") == "local" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	// Get the environment variables
	user := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")

	// Connection string
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", user, password, dbname, host, port)

	// Connect to the database
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully connected to the database")
}

func CreateEnumType() {
	createEnumSQL := `
    DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'status') THEN
            CREATE TYPE status AS ENUM ('draft', 'publish');
        END IF;
    END
    $$;
    `
	_, err := db.Exec(createEnumSQL)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Enum type created successfully")
}

func CreateTables() {
	createTableContentSQL := `
    CREATE TABLE IF NOT EXISTS content (
        id SERIAL PRIMARY KEY,
        title TEXT NOT NULL,
        content TEXT NOT NULL,
		status status NOT NULL DEFAULT 'draft',
        publish_date TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
    );
    `

	_, err := db.Exec(createTableContentSQL)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("content Table created successfully")

	createTableTagSQL := `
    CREATE TABLE IF NOT EXISTS tag (
        id SERIAL PRIMARY KEY,
        label TEXT UNIQUE NOT NULL
    );
    `

	_, err = db.Exec(createTableTagSQL)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("tag Table created successfully")

	createTableContentTagSQL := `
    CREATE TABLE IF NOT EXISTS content_tag (
        content_id INT NOT NULL,
        tag_id INT NOT NULL,
        PRIMARY KEY (content_id, tag_id),
        FOREIGN KEY (content_id) REFERENCES content (id) ON DELETE CASCADE,
        FOREIGN KEY (tag_id) REFERENCES tag (id) ON DELETE CASCADE
    );
    `

	_, err = db.Exec(createTableContentTagSQL)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("content_tag Table created successfully")
}
