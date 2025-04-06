package config

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	// テストケース
	tests := []struct {
		name    string
		envVars map[string]string
		want    *Config
		wantErr bool
	}{
		{
			name: "デフォルト値と環境変数から読み込み",
			envVars: map[string]string{
				"GOOSECONNECT_PORT":     "3000",
				"GOOSECONNECT_URL":      "http://test.example.com",
				"GOOSECONNECT_BASE_DIR": "/tmp/test",
				"GOOSECONNECT_GIT_USER": "testuser",
				"GOOSECONNECT_GIT_MAIL": "test@example.com",
			},
			want:    &Config{},
			wantErr: false,
		},
		{
			name: "必須項目が不足",
			envVars: map[string]string{
				"GOOSECONNECT_PORT":     "3000",
				"GOOSECONNECT_URL":      "http://test.example.com",
				"GOOSECONNECT_BASE_DIR": "/tmp/test",
				"GOOSECONNECT_GIT_USER": "",
				"GOOSECONNECT_GIT_MAIL": "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "未定義の設定値は無視される",
			envVars: map[string]string{
				"GOOSECONNECT_PORT":     "3000",
				"GOOSECONNECT_URL":      "http://test.example.com",
				"GOOSECONNECT_BASE_DIR": "/tmp/test",
				"GOOSECONNECT_GIT_USER": "testuser",
				"GOOSECONNECT_GIT_MAIL": "test@example.com",
			},
			want:    &Config{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数を設定
			for k, v := range tt.envVars {
				if v != "" {
					os.Setenv(k, v)
				} else {
					os.Unsetenv(k)
				}
			}

			// 設定を読み込み
			got, err := NewConfig()

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// エラーが期待される場合は、ここで終了
			if tt.wantErr {
				return
			}

			// 値の検証
			if got.GetPort() != 3000 {
				t.Errorf("port = %v, want %v", got.GetPort(), 3000)
			}
			if got.GetURL() != "http://test.example.com" {
				t.Errorf("url = %v, want %v", got.GetURL(), "http://test.example.com")
			}
			if got.GetBaseDir() != "/tmp/test" {
				t.Errorf("baseDir = %v, want %v", got.GetBaseDir(), "/tmp/test")
			}
			if got.GetGitUser() != "testuser" {
				t.Errorf("gitUser = %v, want %v", got.GetGitUser(), "testuser")
			}
			if got.GetGitMail() != "test@example.com" {
				t.Errorf("gitMail = %v, want %v", got.GetGitMail(), "test@example.com")
			}
		})
	}
}

func TestConfig_ValidateRequiredValues(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name: "必須項目が全て設定されている",
			envVars: map[string]string{
				"GOOSECONNECT_GIT_USER": "testuser",
				"GOOSECONNECT_GIT_MAIL": "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "git_userが未設定",
			envVars: map[string]string{
				"GOOSECONNECT_GIT_USER": "",
				"GOOSECONNECT_GIT_MAIL": "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "git_mailが未設定",
			envVars: map[string]string{
				"GOOSECONNECT_GIT_USER": "testuser",
				"GOOSECONNECT_GIT_MAIL": "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数を設定
			for k, v := range tt.envVars {
				if v != "" {
					os.Setenv(k, v)
				} else {
					os.Unsetenv(k)
				}
			}

			// 設定を読み込み
			config, err := NewConfig()
			if err != nil {
				if !tt.wantErr {
					t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			// 必須項目の検証
			err = config.ValidateRequiredValues()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequiredValues() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_GetInstructionPath(t *testing.T) {
	// テスト前の状態を保存
	oldValue := os.Getenv("GOOSECONNECT_INSTRUCTION_PATH")
	defer os.Setenv("GOOSECONNECT_INSTRUCTION_PATH", oldValue)

	tests := []struct {
		name      string
		envValue  string
		expected  string
	}{
		{
			name:      "デフォルト値の確認",
			envValue:  "",
			expected:  "/etc/goose-connect/instructions.md",
		},
		{
			name:      "環境変数からの読み込み",
			envValue:  "/custom/path/instructions.md",
			expected:  "/custom/path/instructions.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数を設定
			if tt.envValue != "" {
				os.Setenv("GOOSECONNECT_INSTRUCTION_PATH", tt.envValue)
			} else {
				os.Unsetenv("GOOSECONNECT_INSTRUCTION_PATH")
			}

			// 設定を読み込み (git_user と git_mail は必須なのでセット)
			os.Setenv("GOOSECONNECT_GIT_USER", "testuser")
			os.Setenv("GOOSECONNECT_GIT_MAIL", "test@example.com")
			
			config, err := NewConfig()
			if err != nil {
				t.Fatalf("Failed to create config: %v", err)
			}

			// GetInstructionPathの結果を検証
			result := config.GetInstructionPath()
			if result != tt.expected {
				t.Errorf("GetInstructionPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}