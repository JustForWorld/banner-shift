package config

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env-default:"dev"`
	HTTPServer `yaml:"http_server"`
	PostgreSQL `yaml:"postgres"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type PostgreSQL struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5432"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DB       string `yaml:"db" env-default:"postgres"`
}

type User struct {
	Username string `yaml:"username"`
	Tag      int64  `yaml:"tag"`
	Role     string `yaml:"role"`
}

func MustLoad() (*Config, *User) {
	// TODO: remove default value
	configPath := flag.String("config", "./config/local.yaml", "for initiate configuration")
	userPath := flag.String("user", "./config/mock/user.yaml", "for initiate default user")
	flag.Parse()

	if *configPath == "" {
		log.Fatal(`"config" is not set`)
	}

	// check if file exists
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", *configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(*configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	var usr User
	if err := cleanenv.ReadConfig(*userPath, &usr); err != nil {
		log.Fatalf("cannot read user config: %s", err)
	}

	return &cfg, &usr
}
