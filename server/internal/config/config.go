package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Mongo  MongoConfig  `yaml:"mongo"`
	JWT    JWTConfig    `yaml:"jwt"`
	Game   GameConfig   `yaml:"game"`
	Admin  AdminConfig  `yaml:"admin"`
}

type ServerConfig struct {
	Addr    string `yaml:"addr"`
	Mode    string `yaml:"mode"`
	WebDist string `yaml:"web_dist"`
}

type MongoConfig struct {
	URI            string `yaml:"uri"`
	Database       string `yaml:"database"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type GameConfig struct {
	Mock           bool   `yaml:"mock"`
	SignSecret     string `yaml:"sign_secret"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

type AdminConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Load 读取 yaml 配置，敏感项可通过环境变量覆盖。
func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(raw, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.applyEnv()
	cfg.applyDefaults()
	return cfg, nil
}

func (c *Config) applyEnv() {
	overrides := map[string]*string{
		"GM_SERVER_ADDR":   &c.Server.Addr,
		"GM_MONGO_URI":     &c.Mongo.URI,
		"GM_MONGO_DB":      &c.Mongo.Database,
		"GM_JWT_SECRET":    &c.JWT.Secret,
		"GM_GAME_SIGN_KEY": &c.Game.SignSecret,
	}
	for env, field := range overrides {
		if v := os.Getenv(env); v != "" {
			*field = v
		}
	}
}

func (c *Config) applyDefaults() {
	if c.Server.Addr == "" {
		c.Server.Addr = ":8080"
	}
	if c.Server.Mode == "" {
		c.Server.Mode = "debug"
	}
	if c.Mongo.URI == "" {
		c.Mongo.URI = "mongodb://127.0.0.1:27017"
	}
	if c.Mongo.Database == "" {
		c.Mongo.Database = "slg_gm"
	}
	if c.Mongo.TimeoutSeconds <= 0 {
		c.Mongo.TimeoutSeconds = 5
	}
	if c.JWT.ExpireHours <= 0 {
		c.JWT.ExpireHours = 24
	}
	if c.Game.TimeoutSeconds <= 0 {
		c.Game.TimeoutSeconds = 5
	}
}
