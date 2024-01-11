package config

type Config struct {
	Name        string      `json:"name"`
	RedisConfig RedisConfig `json:"redis_config"`
}

type RedisConfig struct {
	Addr        string `json:"addr"`
	Network     string `json:"network"`
	TimeOut     int64  `json:"timeout"` // millisecond
	MaxIdle     int    `json:"max_idle"`
	IdleTimeout int64  `json:"idle_timeout"`
	MaxActive   int    `json:"max_active"`
	Wait        bool   `json:"wait"`
}
