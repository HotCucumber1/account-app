package config

import (
	"time"

	"github.com/pkg/errors"
)

type Env struct {
	App      App           `envconfig:"app"`
	GRPC     GRPC          `envconfig:"grpc"`
	HTTP     HTTP          `envconfig:"http"`
	MySQL    MySQL         `envconfig:"mysql"`
	AMQP     AMQP          `envconfig:"amqp"`
	Relay    Relay         `envconfig:"relay"`
	Account  AccountClient `envconfig:"account"`
	Session  SessionClient `envconfig:"session"`
	Token    Token         `envconfig:"token"`
	Password Password      `envconfig:"password"`

	Registration Registration `envconfig:"registration"`
}

type App struct {
	Debug        bool          `envconfig:"debug"`
	GraceTimeout time.Duration `envconfig:"grace_timeout" default:"30s"`
}

type GRPC struct {
	Address string `envconfig:"address" default:":8081"`
}

type HTTP struct {
	Address string `envconfig:"address" default:":8082"`
}

type MySQL struct {
	Address               string        `envconfig:"address" default:"localhost:3306"`
	Database              string        `envconfig:"database" default:"auth"`
	User                  string        `envconfig:"user" default:"auth"`
	Password              string        `envconfig:"password" default:"auth"`
	MaxConnections        int           `envconfig:"max_connections" default:"10"`
	ConnectionMaxLifeTime time.Duration `envconfig:"connection_max_lifetime" default:"30m"`
	ConnectionMaxIdleTime time.Duration `envconfig:"connection_max_idle_time" default:"5m"`
	ConnectTimeout        time.Duration `envconfig:"connect_timeout" default:"5s"`
	ReadTimeout           time.Duration `envconfig:"read_timeout" default:"10s"`
	WriteTimeout          time.Duration `envconfig:"write_timeout" default:"10s"`
}

type AMQP struct {
	URL            string        `envconfig:"url" default:"amqp://local:local@localhost:5672/"`
	ConnectTimeout time.Duration `envconfig:"connect_timeout" default:"5s"`
}

type Relay struct {
	Interval       time.Duration `envconfig:"interval" default:"1s"`
	BatchSize      int           `envconfig:"batch_size" default:"100"`
	PublishTimeout time.Duration `envconfig:"publish_timeout" default:"10s"`
	LockTimeout    time.Duration `envconfig:"lock_timeout" default:"1s"`
}

type AccountClient struct {
	GRPCAddress string        `envconfig:"grpc_address" default:"account-service:8081"`
	Timeout     time.Duration `envconfig:"timeout" default:"5s"`
}

type SessionClient struct {
	GRPCAddress string        `envconfig:"grpc_address" default:"session-service:8081"`
	Timeout     time.Duration `envconfig:"timeout" default:"5s"`
}

type Token struct {
	AccessTTL time.Duration `envconfig:"access_ttl" default:"15m"`
}

type Registration struct {
	Deadline         time.Duration `envconfig:"deadline" default:"10s"`
	RecoveryInterval time.Duration `envconfig:"recovery_interval" default:"1s"`
	RecoveryGrace    time.Duration `envconfig:"recovery_grace" default:"5s"`
	RecoveryBatch    int           `envconfig:"recovery_batch" default:"100"`
}

type Password struct {
	BcryptCost int `envconfig:"bcrypt_cost" default:"10"`
}

func (c Env) Validate() error {
	if c.App.GraceTimeout <= 0 {
		return errors.New("grace timeout must be positive")
	}
	if c.MySQL.ConnectTimeout <= 0 || c.MySQL.ReadTimeout <= 0 || c.MySQL.WriteTimeout <= 0 || c.MySQL.ConnectionMaxLifeTime < 0 || c.MySQL.ConnectionMaxIdleTime < 0 {
		return errors.New("MySQL timeouts must be positive; connection lifetimes must not be negative")
	}
	if c.MySQL.MaxConnections <= 0 {
		return errors.New("connection count must be positive")
	}
	if c.MySQL.Address == "" || c.MySQL.Database == "" || c.MySQL.User == "" || c.HTTP.Address == "" || c.GRPC.Address == "" {
		return errors.New("database and server configuration must not be empty")
	}
	if c.AMQP.URL == "" || c.AMQP.ConnectTimeout <= 0 {
		return errors.New("broker configuration must be set")
	}
	if c.Relay.Interval <= 0 || c.Relay.PublishTimeout <= 0 || c.Relay.LockTimeout < 0 || c.Relay.BatchSize <= 0 {
		return errors.New("relay settings must be positive")
	}
	if c.Account.GRPCAddress == "" || c.Session.GRPCAddress == "" || c.Account.Timeout <= 0 || c.Session.Timeout <= 0 {
		return errors.New("account and session clients must be configured")
	}
	if c.Registration.Deadline <= 0 || c.Registration.RecoveryInterval <= 0 || c.Registration.RecoveryGrace < 0 || c.Registration.RecoveryBatch <= 0 {
		return errors.New("registration transaction settings must be positive")
	}
	if c.Token.AccessTTL <= 0 || c.Password.BcryptCost <= 0 {
		return errors.New("token and password settings must be positive")
	}
	return nil
}
