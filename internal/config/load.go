package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tryoo0607/job-test/internal/api"
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
	pflag.Int("total-pods", 1, "total number of pods (for indexed-peer mode)")
	pflag.String("subdomain", "", "headless service subdomain (for indexed-peer mode)")
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

	if err := v.BindEnv("job-name", "JOB_NAME"); err != nil {
		panic(fmt.Errorf("bind job-name failed: %w", err))
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
	case api.Indexed, api.Peer, api.Fixed, api.Queue:
	default:
		return fmt.Errorf("invalid mode: %q", cfg.Mode)
	}

	// ✅ peer 모드에서 IP가 반드시 필요한 설계라면 체크
	// (Deployment + Headless Service로 DNS A레코드 사용 시, 자기 IP 필요)
	if cfg.Mode == api.Peer && cfg.JobName == "" {
		return fmt.Errorf("POD_IP is required in peer mode (set via fieldRef: status.podIP)")
	}

	if cfg.MaxConcurrency <= 0 {
		return fmt.Errorf("max-concurrency must be > 0")
	}

	if cfg.Mode == api.Peer {
		if cfg.JobIndex == nil {
			return fmt.Errorf("peer mode requires JOB_COMPLETION_INDEX")
		}
		if cfg.TotalPods <= 0 {
			return fmt.Errorf("peer mode requires total-pods > 0")
		}
		if cfg.Subdomain == "" {
			return fmt.Errorf("peer mode requires subdomain (headless svc)")
		}
		if cfg.JobName == "" {
			return fmt.Errorf("peer mode requires job-name")
		}
	}
	return nil
}
