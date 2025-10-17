package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const dbFile = "entries.db"

type Entry struct {
	BusinessName string
	Name         string
	Email        string
	Phone        string
	Message      string
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "read" {
		readEntries()
		return
	}

	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fs := http.FileServer(http.Dir("images"))
	http.Handle("/images/", http.StripPrefix("/images/", fs))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/submit", handleSubmit(db))
	log.Println("Starting server on :443")
	log.Fatal(http.ListenAndServeTLS(":443", "certs/cert.pem", "certs/key.pem", nil))
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS entries (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"business_name" TEXT,
		"name" TEXT,
		"email" TEXT,
		"phone" TEXT,
		"message" TEXT
	);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func handleSubmit(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		entry := Entry{
			BusinessName: r.FormValue("business_name"),
			Name:         r.FormValue("name"),
			Email:        r.FormValue("email"),
			Phone:        r.FormValue("phone"),
			Message:      r.FormValue("message"),
		}

		if err := addEntry(db, entry); err != nil {
			http.Error(w, "Error saving entry", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Entry saved successfully")
	}
}

func addEntry(db *sql.DB, entry Entry) error {
	insertSQL := `INSERT INTO entries(business_name, name, email, phone, message) VALUES (?, ?, ?, ?, ?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec(entry.BusinessName, entry.Name, entry.Email, entry.Phone, entry.Message)
	return err
}

func readEntries() {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT business_name, name, email, phone, message FROM entries")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var businessName, name, email, phone, message string
		if err := rows.Scan(&businessName, &name, &email, &phone, &message); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Business Name: %s, Name: %s, Email: %s, Phone: %s, Message: %s\n", businessName, name, email, phone, message)
	}
}
