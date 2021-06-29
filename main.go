package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"database/sql"

	"github.com/ilyakaznacheev/cleanenv"
	shell "github.com/ipfs/go-ipfs-api"
	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type Config struct {
	Port     string `yaml:"port" env:"PORT" env-default:"5001"`
	Host     string `yaml:"host" env:"HOST" env-default:"localhost"`
	Database string `yaml:"db" env:"DB" env-default:"ipfs-tag.db"`
}

func main() {
	var cfg Config

	err := cleanenv.ReadConfig("ipfs-tag.yml", &cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading configuration file: %s", err)
	}

	if _, err := os.Stat(cfg.Database); os.IsNotExist(err) {
		log.Println("Database does not exist yet.")
		createDataBase(cfg.Database)
		sqliteDatabase, _ := sql.Open("sqlite3", cfg.Database) // Open the created SQLite File
		defer sqliteDatabase.Close()                           // Defer Closing the database
		createTable(sqliteDatabase)
	}
	sqliteDatabase, _ := sql.Open("sqlite3", cfg.Database) // Open the created SQLite File
	defer sqliteDatabase.Close()                           // Defer Closing the database
	cid, err := addIPFSFile(cfg)
	if err != nil {
		log.Printf("Error adding String to IPFS: %s", err)
	}
	insertFile(sqliteDatabase, cid, "teststring", "/root/test")

}

func createDataBase(dbFileName string) {
	log.Println("Creating sqlite-database.db...")
	file, err := os.Create(dbFileName) // Create SQLite file
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	log.Println("database created")
}

// returns the CID and the possible error
func addIPFSFile(cfg Config) (string, error) {
	// Where your local node is running on localhost:5001
	sh := shell.NewShell(cfg.Host + ":" + cfg.Port)
	if !sh.IsUp() {
		log.Println("Gateway is not up")
		return "", errors.New("gateway not up")
	}

	cid, err := sh.Add(strings.NewReader("hello world!"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		return "", errors.New("cannot add hash")
	}
	log.Printf("added %s", cid)
	return cid, nil
}

func createTable(db *sql.DB) {
	createFileTableSQL := `CREATE TABLE files (
		"idFile" integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"hash" TEXT,
		"name" TEXT,
		"path" TEXT		
	  );` // SQL Statement for Create Table

	log.Println("Create file table...")
	statement, err := db.Prepare(createFileTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() // Execute SQL Statements
	log.Println("File table created")
}

// We are passing db reference connection from main to our method with other parameters
func insertFile(db *sql.DB, hash string, name string, path string) {
	log.Println("Inserting file record ...")
	insertFileSQL := `INSERT INTO files(hash, name, path) VALUES (?, ?, ?)`
	statement, err := db.Prepare(insertFileSQL) // Prepare statement. This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(hash, name, path)
	if err != nil {
		log.Fatalln(err.Error())
	}
}
