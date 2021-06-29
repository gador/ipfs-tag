package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"database/sql"

	"github.com/ilyakaznacheev/cleanenv"
	shell "github.com/ipfs/go-ipfs-api"
	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
	"github.com/thatisuday/commando"
)

// config file construct
// Port and Host are used for the local or remote IPFS daemon
// Database ist the file name of the sqlite database
type Config struct {
	Port     string `yaml:"port" env:"PORT" env-default:"5001"`
	Host     string `yaml:"host" env:"HOST" env-default:"localhost"`
	Database string `yaml:"db" env:"DB" env-default:"ipfs-tag.db"`
}

var verbose bool = false

func main() {

	var add string
	var cfg Config
	commando.
		SetExecutableName("ipfs-tag").
		SetVersion("v0.0.1").
		SetDescription("IPFS-tag supplies a tag layer on top of IPFS")

	// main command
	commando.
		Register(nil).
		AddArgument("add", "add files to IPFS store", "").                      // required
		AddFlag("verbose,V", "display log information ", commando.Bool, false). // optional
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			add = args["add"].Value
			verbose, _ = flags["verbose"].GetBool()
		})
	commando.Parse(nil)

	// load config
	err := cleanenv.ReadConfig("ipfs-tag.yml", &cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading configuration file: %s", err)
	}

	// check for db file and create it, if it doesn't exist
	if _, err := os.Stat(cfg.Database); os.IsNotExist(err) {
		if verbose {
			log.Println("Database does not exist yet.")
		}

		createDataBase(cfg.Database)
		sqliteDatabase, _ := sql.Open("sqlite3", cfg.Database) // Open the created SQLite File
		defer sqliteDatabase.Close()                           // Defer Closing the database
		createTable(sqliteDatabase)
	}

	// open db for the main app
	sqliteDatabase, _ := sql.Open("sqlite3", cfg.Database) // Open the created SQLite File
	defer sqliteDatabase.Close()                           // Defer Closing the database

	// add files
	// check, if the file exist
	file, err := os.Open(add)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %s", err)
	}

	// add the file and return the cid
	cid, err := addIPFSFile(cfg, file)
	if err != nil {
		log.Printf("Error adding String to IPFS: %s", err)
	}

	// insert cid into db
	insertFile(sqliteDatabase, cid, "teststring", add)

}

// create SQLite db file
func createDataBase(dbFileName string) {
	if verbose {
		log.Println("Creating sqlite-database.db...")
	}
	file, err := os.Create(dbFileName) // Create SQLite file
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	if verbose {
		log.Println("database created")
	}
}

// returns the CID and the possible error
func addIPFSFile(cfg Config, file io.Reader) (string, error) {
	// Where your local node is running on localhost:5001
	sh := shell.NewShell(cfg.Host + ":" + cfg.Port)
	if !sh.IsUp() {
		log.Println("Gateway is not up")
		return "", errors.New("gateway not up")
	}

	//cid, err := sh.Add(strings.NewReader("hello world!"))
	cid, err := sh.Add(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		return "", errors.New("cannot add hash")
	}
	log.Printf("added %s", cid)
	return cid, nil
}

// create tables in database
func createTable(db *sql.DB) {
	createFileTableSQL := `CREATE TABLE files (
		"idFile" integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"hash" TEXT,
		"name" TEXT,
		"path" TEXT		
	  );` // SQL Statement for Create Table

	if verbose {
		log.Println("Create file table...")
	}
	statement, err := db.Prepare(createFileTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() // Execute SQL Statements
	if verbose {
		log.Println("File table created")
	}
}

// insert the hash  to the sqlite db
func insertFile(db *sql.DB, hash string, name string, path string) {
	if verbose {
		log.Println("Inserting file record for " + path)
	}
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
