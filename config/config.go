package config

import (
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/rs/zerolog"

	"github.com/AndersonQ/gokafka/constants"
)

// Config - environment variables are parsed to this struct
type Config struct {
	AppName    string `env:"APP_NAME" envDefault:"gokafka"`
	ServerPort int    `env:"PORT" envDefault:"8000"`
	Env        string `env:"ENV" envDefault:"env not set"`
	LogLevel   string `env:"LOG_LEVEL" envDefault:"debug"`
	LogOutput  string `env:"LOG_OUTPUT" envDefault:"console"`

	// RequestTimeout the timeout for the incoming request
	RequestTimeout time.Duration `env:"REQUEST_TIMEOUT" envDefault:"5s"`
	// ShutdownTimeout the time the sever will wait server.Shutdown to return
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"6s"`

	ClientID     string `env:"CLIENT_ID,required"`
	ClientSecret string `env:"CLIENT_SECRET,required"`
	TokenURL     string `env:"TOKEN_URL,required"`
	KafkaServer  string `env:"KAFKA_SERVER,required"`
	KafkaGroupID string `env:"KAFKA_GROUP_ID,required"`
}

// Parse environment variables, returns (guess what?) an error if an error occurs
func Parse() (Config, error) {
	confs := Config{}
	err := env.Parse(&confs)
	return confs, err
}

// Logger returns a initialised zerolog.Logger
func (c Config) Logger() zerolog.Logger {
	logLevelOk := true
	logLevel, err := zerolog.ParseLevel(c.LogLevel)
	if err != nil {
		logLevel = zerolog.InfoLevel
		logLevelOk = false
	}

	zerolog.SetGlobalLevel(logLevel)
	zerolog.TimestampFieldName = constants.FieldTimestamp

	host, _ := os.Hostname()
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str(constants.FieldApplication, c.AppName).
		Str(constants.FieldHost, host).
		Str(constants.FieldEnvironment, c.Env).
		Logger()

	if strings.ToUpper(c.LogOutput) == "CONSOLE" {
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	if !logLevelOk {
		logger.Warn().Err(err).Msgf("%s is not a valid zerolog log level, defaulting to info", c.LogLevel)
	}

	return logger
}
