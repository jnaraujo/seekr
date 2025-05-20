package env

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type EnvSchema struct {
	OllamaAPIUrl string `envconfig:"OLLAMA_API_URL" required:"true,url"`
}

var Env EnvSchema

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	err = envconfig.Process("", &Env)
	if err != nil {
		panic(err)
	}
}
