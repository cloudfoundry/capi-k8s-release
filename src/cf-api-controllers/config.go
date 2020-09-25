package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	cfAPIHost          string
	uaaEndpoint        string
	uaaClientName      string
	uaaClientSecret    string
	workloadsNamespace string
}

func LoadConfig() (*Config, error) {
	c := &Config{}

	if c.cfAPIHost = os.Getenv("CF_API_HOST"); c.cfAPIHost == "" {
		return nil, envNotSetErr("CF_API_HOST")
	}

	if c.uaaEndpoint = os.Getenv("UAA_ENDPOINT"); c.uaaEndpoint == "" {
		return nil, envNotSetErr("UAA_ENDPOINT")
	}

	if c.uaaClientName = os.Getenv("UAA_CLIENT_NAME"); c.uaaClientName == "" {
		return nil, envNotSetErr("UAA_CLIENT_NAME")
	}

	if c.workloadsNamespace = os.Getenv("WORKLOADS_NAMESPACE"); c.workloadsNamespace == "" {
		return nil, envNotSetErr("WORKLOADS_NAMESPACE")
	}

	var err error
	c.uaaClientSecret, err = c.fetchUaaClientSecret()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) CFAPIHost() string {
	return c.cfAPIHost
}

func (c *Config) UAAEndpoint() string {
	return c.uaaEndpoint
}

func (c *Config) UAAClientName() string {
	return c.uaaClientName
}

func (c *Config) UAAClientSecret() string {
	return c.uaaClientSecret
}

func (c *Config) WorkloadsNamespace() string {
	return c.workloadsNamespace
}

func (c *Config) fetchUaaClientSecret() (string, error) {
	secretFile := os.Getenv("UAA_CLIENT_SECRET_FILE")
	if secretFile == "" {
		secretEnv := os.Getenv("UAA_CLIENT_SECRET")
		if secretEnv == "" {
			return "", errors.New("`UAA_CLIENT_SECRET_FILE` or `UAA_CLIENT_SECRET` environment variable must be set")
		}

		return secretEnv, nil
	}

	secretBytes, err := ioutil.ReadFile(secretFile)
	if err != nil {
		return "", err
	}

	return string(secretBytes), nil
}

func envNotSetErr(e string) error {
	return errors.New(fmt.Sprintf("`%s` environment variable must be set", e))
}
