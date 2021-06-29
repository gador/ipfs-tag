package main

import (
	"fmt"
	"log"
	"os"
	"strings"

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
	}

	addIPFSFile(cfg)

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

func addIPFSFile(cfg Config) {
	// Where your local node is running on localhost:5001
	sh := shell.NewShell(cfg.Host + ":" + cfg.Port)
	if !sh.IsUp() {
		fmt.Println("Gateway is not up")
		os.Exit(1)
	}

	cid, err := sh.Add(strings.NewReader("hello world!"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
	fmt.Printf("added %s", cid)
}
