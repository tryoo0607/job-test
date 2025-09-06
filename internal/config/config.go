package config

import "time"

type Config struct {
	HTTPPort       string        `mapstructure:"http-port"`
	Mode           string        `mapstructure:"mode"`
	MaxConcurrency int           `mapstructure:"max-concurrency"`
	RetryMax       int           `mapstructure:"retry-max"`
	RetryBackoff   time.Duration `mapstructure:"retry-backoff"`
	Items          []string      `mapstructure:"items"`
	ItemsFile      string        `mapstructure:"items-file"`
	QueueURL       string        `mapstructure:"queue-url"`
	QueueKey       string        `mapstructure:"queue-key"`
	InputDir       string        `mapstructure:"input-dir"`
	OutputDir      string        `mapstructure:"output-dir"`
	JobIndex       *int          `mapstructure:"job-index"`
	TotalPods      int           `mapstructure:"total-pods"`
	Subdomain      string        `mapstructure:"subdomain"` // 헤드리스 서비스 이름
	HoldFor        time.Duration `mapstructure:"hold-for"`
}
