package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
}

func NewConfig() (*Config, error) {
	// デフォルト値の設定
	viper.SetDefault("port", 8080)
	viper.SetDefault("url", "http://localhost:8080")
	viper.SetDefault("base_dir", "$HOME/.goose-connect")
	viper.SetDefault("git_user", "")
	viper.SetDefault("git_mail", "")

	// 環境変数の設定
	viper.AutomaticEnv()
	viper.SetEnvPrefix("GOOSECONNECT")

	// 環境変数の型変換
	viper.SetTypeByDefaultValue(true)

	// 設定ファイルの読み込み（オプション）
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
		}
		// 設定ファイルが見つからない場合は、環境変数とデフォルト値を使用
		fmt.Println("設定ファイルが見つかりません。環境変数とデフォルト値を使用します。")
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("設定の解析に失敗しました: %w", err)
	}

	if err := config.ValidateRequiredValues(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) ValidateRequiredValues() error {
	gitUser := viper.GetString("git_user")
	if gitUser == "" {
		return fmt.Errorf("git_user is required")
	}
	gitMail := viper.GetString("git_mail")
	if gitMail == "" {
		return fmt.Errorf("git_mail is required")
	}
	return nil
}

func (c *Config) GetPort() int {
	return viper.GetInt("port")
}

func (c *Config) GetURL() string {
	return viper.GetString("url")
}

func (c *Config) GetBaseDir() string {
	return viper.GetString("base_dir")
}

func (c *Config) GetGitUser() string {
	return viper.GetString("git_user")
}

func (c *Config) GetGitMail() string {
	return viper.GetString("git_mail")
}

func ValidateRequiredValues() error {
	cfg, err := NewConfig()
	if err != nil {
		return err
	}
	return cfg.ValidateRequiredValues()
}
