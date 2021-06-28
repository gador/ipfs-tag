package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
	shell "github.com/ipfs/go-ipfs-api"
)

type Config struct {
	Port string `yaml:"port" env:"PORT" env-default:"5001"`
	Host string `yaml:"host" env:"HOST" env-default:"localhost"`
}

func main() {
	var cfg Config
	err := cleanenv.ReadConfig("ipfs-tag.yml", &cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading configuration file: %s", err)
	}

	// Where your local node is running on localhost:5001
	sh := shell.NewShell(cfg.Host + ":" + cfg.Port)
	cid, err := sh.Add(strings.NewReader("hello world!"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
	fmt.Printf("added %s", cid)
}
