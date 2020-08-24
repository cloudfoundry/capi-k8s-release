package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	RegistryUsername string
	RegistryPassword string
	Port             int
}

func Load() (*Config, error) {
	c := &Config{}
	var exists bool
	c.RegistryUsername, exists = os.LookupEnv("REGISTRY_USERNAME")
	if !exists {
		return nil, errors.New("REGISTRY_USERNAME not configured")
	}

	c.RegistryPassword, exists = os.LookupEnv("REGISTRY_PASSWORD")
	if !exists {
		return nil, errors.New("REGISTRY_PASSWORD not configured")
	}

	portStr, exists := os.LookupEnv("PORT")
	if !exists {
		portStr = "8080"
	}

	var err error
	c.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return nil, errors.New("PORT must be an integer")
	}

	return c, nil
}
