package es

import (
	"fmt"
	"strconv"
	"time"

	"go.k6.io/k6/output"
)

// Config is the config for the template collector
type Config struct {
	Address        string
	PushInterval   time.Duration
	Username       string
	Password       string
	Index          string
	SnifferEnabled bool
	MaxBulkSize    int
}

// NewConfig creates a new Config instance from the provided output.Params
func NewConfig(p output.Params) (Config, error) {
	cfg := Config{
		Address:        "http://127.0.0.1:9200",
		Username:       "",
		Password:       "",
		Index:          "",
		SnifferEnabled: false,
		PushInterval:   1 * time.Second,
		MaxBulkSize:    2048,
	}

	for k, v := range p.Environment {
		switch k {
		case "K6_ES_PUSH_INTERVAL":
			var err error
			cfg.PushInterval, err = time.ParseDuration(v)
			if err != nil {
				return cfg, fmt.Errorf("error parsing environment variable 'K6_ES_PUSH_INTERVAL': %w", err)
			}

		case "K6_ES_ADDRESS":
			cfg.Address = v

		case "K6_ES_USERNAME":
			cfg.Username = v

		case "K6_ES_PASSWORD":
			cfg.Password = v

		case "K6_ES_INDEX":
			cfg.Index = v

		case "K6_ES_ENABLE_SNIFFER":
			cfg.SnifferEnabled = true

		case "K6_ES_MAX_BULKSIZE":
			var err error
			cfg.MaxBulkSize, err = strconv.Atoi(v)
			if err != nil {
				return cfg, fmt.Errorf("error parsing environment variable 'K6_ES_MAX_BULKSIZE': %w", err)
			}
		}
	}

	return cfg, nil
}
