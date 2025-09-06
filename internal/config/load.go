package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var v *viper.Viper

func Load() (*Config, error) {

	loadFlag()
	loadEnv()

	cfg, err := unmarshalConfig()
	if err != nil {
		return nil, err
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadFlag() {
	// 1) 플래그 정의
	pflag.String("http-port", "8080", "http port")
	pflag.String("mode", "indexed", "indexed|fixed|queue")
	pflag.Int("max-concurrency", 1, "goroutine workers")
	pflag.Int("retry-max", 3, "max retries")
	pflag.Duration("retry-backoff", time.Second, "initial backoff (e.g. 1s)")
	pflag.StringSlice("items", []string{}, "items (comma sep)")
	pflag.String("items-file", "", "items file")
	pflag.String("queue-url", "", "redis/rabbit url")
	pflag.String("queue-key", "tasks", "queue key")
	pflag.String("input-dir", "/data/inputs", "input dir")
	pflag.String("output-dir", "/data/outputs", "output dir")
	// job-index는 env전용이라 플래그 생략
	pflag.Parse()
}

func loadEnv() {
	v = viper.New()
	// v.SetEnvPrefix("APP")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// 플래그 바인딩 (플래그 > env > default)
	if err := v.BindPFlags(pflag.CommandLine); err != nil {
		panic(fmt.Errorf("bind flags failed: %w", err))
	}

	// K8s Indexed Job 환경변수 바인딩
	if err := v.BindEnv("job-index", "JOB_COMPLETION_INDEX"); err != nil {
		panic(fmt.Errorf("bind job-index failed: %w", err))
	}
}

func unmarshalConfig() (*Config, error) {
	var cfg Config

	// decoder custom
	// time.Duration과 []string 변환 위해 사용
	decode := func(c *mapstructure.DecoderConfig) {
		c.TagName = "mapstructure"
		c.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		)
	}

	// 설정한 custom decorder로 Config 구조체에 맵핑해줌
	if err := v.Unmarshal(&cfg, decode); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Config 값 검증
func validate(cfg *Config) error {
	switch cfg.Mode {
	case "indexed", "fixed", "queue":
	default:
		return fmt.Errorf("invalid mode: %q", cfg.Mode)
	}
	if cfg.MaxConcurrency <= 0 {
		return fmt.Errorf("max-concurrency must be > 0")
	}
	return nil
}
