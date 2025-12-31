package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:",squash"`
	Feishu FeishuConfig `mapstructure:",squash"`
	LLM    LLMConfig    `mapstructure:",squash"`
}

type ServerConfig struct {
	Port string `mapstructure:"PORT"`
}

type FeishuConfig struct {
	AppID             string `mapstructure:"FEISHU_APP_ID"`
	AppSecret         string `mapstructure:"FEISHU_APP_SECRET"`
	EncryptKey        string `mapstructure:"FEISHU_ENCRYPT_KEY"`
	VerificationToken string `mapstructure:"FEISHU_VERIFICATION_TOKEN"`
}

type LLMConfig struct {
	Provider  string `mapstructure:"LLM_PROVIDER"`
	APIKey    string `mapstructure:"LLM_API_KEY"`
	APIURL    string `mapstructure:"LLM_API_URL"`
	ModelName string `mapstructure:"LLM_MODEL_NAME"` // e.g. "deepseek-chat", "gpt-4o"
}

var AppConfig *Config

func Init() {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: .env file not found, relying on environment variables: %v", err)
	}

	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}

    // Set default values if needed
    if AppConfig.Server.Port == "" {
        AppConfig.Server.Port = "8080"
    }
}
