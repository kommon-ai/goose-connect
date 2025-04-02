package goose

import (
        "fmt"
        "os"
        "strings"
        "testing"

        "github.com/kommon-ai/agent-go/pkg/agent"
        "github.com/kommon-ai/goose-connect/pkg/config"
)

// MockProvider はテスト用のProvider実装です
type MockProvider struct {
        env          map[string]string
        providerName string
        modelName    string
        apiKey       string
}

func (m *MockProvider) GetEnv() map[string]string {
        return m.env
}

func (m *MockProvider) GetProviderName() string {
        return m.providerName
}

func (m *MockProvider) GetModelName() string {
        return m.modelName
}

func (m *MockProvider) GetAPIKey() string {
        return m.apiKey
}

// TestGooseEnvGetEnv はGooseEnv.GetEnvメソッドをテストします
func TestGooseEnvGetEnv(t *testing.T) {
        // テスト用のGooseEnvを作成
        env := &GooseEnv{
                APIKey:              "test-api-key",
                Model:               "test-model",
                Provider:            "openai",
                Repo:                "test-owner/test-repo",
                InstallationToken:   "test-token",
                BaseDir:             "/tmp/test-base-dir",
                SessionID:           "test-session-id",
                InstructionFIlePath: "/tmp/test-instruction.md",
                ScriptFIlePath:      "/tmp/test-script.sh",
                EnvFilePath:         "/tmp/test-env.sh",
                BranchName:          "test-branch",
        }

        // GetEnvメソッドを実行
        result := env.GetEnv()

        // 期待される環境変数が含まれているか検証
        expectedValues := map[string]string{
                "OPENAI_API_KEY":        "test-api-key",
                "GOOSE_PROVIDER":        "openai",
                "GOOSE_MODEL":           "test-model",
                "GITHUB_TOKEN":          "test-token",
                "REPO":                  "test-owner/test-repo",
                "BASE_DIR":              "/tmp/test-base-dir",
                "SESSION_ID":            "test-session-id",
                "INSTRUCTION_FILE_PATH": "/tmp/test-instruction.md",
                "SCRIPT_FILE_PATH":      "/tmp/test-script.sh",
                "ENV_FILE_PATH":         "/tmp/test-env.sh",
                "PR_BRANCH":             "test-branch",
        }

        for key, expectedValue := range expectedValues {
                if value, ok := result[key]; !ok || value != expectedValue {
                        t.Errorf("Expected %s=%s, got %s=%s", key, expectedValue, key, value)
                }
        }

        // 期待される数の環境変数が含まれているか検証
        if len(result) != len(expectedValues) {
                t.Errorf("Expected %d environment variables, got %d", len(expectedValues), len(result))
        }
}

// TestGooseEnvGetRequiredEnv はGooseEnv.GetRequiredEnvメソッドをテストします
func TestGooseEnvGetRequiredEnv(t *testing.T) {
        // テスト用のGooseEnvを作成
        env := &GooseEnv{}

        // GetRequiredEnvメソッドを実行
        result := env.GetRequiredEnv()

        // 期待される必須環境変数が含まれているか検証
        expectedRequiredEnvs := []string{
                "GOOSE_PROVIDER",
                "GOOSE_MODEL",
                "GITHUB_TOKEN",
                "REPO",
                "BASE_DIR",
                "SESSION_ID",
                "INSTRUCTION_FILE_PATH",
                "SCRIPT_FILE_PATH",
                "ENV_FILE_PATH",
        }

        if len(result) != len(expectedRequiredEnvs) {
                t.Errorf("Expected %d required environment variables, got %d", len(expectedRequiredEnvs), len(result))
        }

        // すべての期待される必須環境変数が含まれているか検証
        for _, expectedEnv := range expectedRequiredEnvs {
                found := false
                for _, actualEnv := range result {
                        if actualEnv == expectedEnv {
                                found = true
                                break
                        }
                }
                if !found {
                        t.Errorf("Required environment variable %s not found in result", expectedEnv)
                }
        }
}

// TestGooseEnvValidateRequiredEnv はGooseEnv.ValidateRequiredEnvメソッドをテストします
func TestGooseEnvValidateRequiredEnv(t *testing.T) {
        testCases := []struct {
                name        string
                env         *GooseEnv
                expectError bool
        }{
                {
                        name: "All required environment variables set",
                        env: &GooseEnv{
                                APIKey:              "test-api-key",
                                Model:               "test-model",
                                Provider:            "openai",
                                Repo:                "test-owner/test-repo",
                                InstallationToken:   "test-token",
                                BaseDir:             "/tmp/test-base-dir",
                                SessionID:           "test-session-id",
                                InstructionFIlePath: "/tmp/test-instruction.md",
                                ScriptFIlePath:      "/tmp/test-script.sh",
                                EnvFilePath:         "/tmp/test-env.sh",
                        },
                        expectError: false,
                },
                {
                        name: "Missing required environment variables",
                        env: &GooseEnv{
                                APIKey:    "test-api-key",
                                Model:     "test-model",
                                Provider:  "openai",
                                SessionID: "test-session-id",
                        },
                        expectError: true,
                },
                {
                        name: "Empty required environment variables",
                        env: &GooseEnv{
                                APIKey:              "test-api-key",
                                Model:               "",
                                Provider:            "",
                                Repo:                "test-owner/test-repo",
                                InstallationToken:   "test-token",
                                BaseDir:             "/tmp/test-base-dir",
                                SessionID:           "test-session-id",
                                InstructionFIlePath: "/tmp/test-instruction.md",
                                ScriptFIlePath:      "/tmp/test-script.sh",
                                EnvFilePath:         "/tmp/test-env.sh",
                        },
                        expectError: true,
                },
        }

        for _, tc := range testCases {
                t.Run(tc.name, func(t *testing.T) {
                        // ValidateRequiredEnvメソッドを実行
                        err := tc.env.ValidateRequiredEnv()

                        // エラーの有無を検証
                        if tc.expectError && err == nil {
                                t.Errorf("Expected error, but got nil")
                        }
                        if !tc.expectError && err != nil {
                                t.Errorf("Expected no error, but got: %v", err)
                        }
                })
        }
}

func TestCreateFiles(t *testing.T) {
        // テスト用の一時ディレクトリを作成
        tempDir, err := os.MkdirTemp("", "goose-test-createfiles-*")
        if err != nil {
                t.Fatalf("Failed to create temp directory: %v", err)
        }
        defer os.RemoveAll(tempDir)

        // テスト用のファイル名とコンテンツを準備
        testFiles := map[string]string{
                fmt.Sprintf("%s/file1.txt", tempDir): "This is file 1 content",
                fmt.Sprintf("%s/file2.txt", tempDir): "This is file 2 content",
        }

        // createFilesを実行
        err = createFiles(testFiles)
        if err != nil {
                t.Fatalf("createFiles failed: %v", err)
        }

        // 結果を検証
        for filename, expectedContent := range testFiles {
                // ファイルが存在するか確認
                if _, err := os.Stat(filename); os.IsNotExist(err) {
                        t.Errorf("Expected file %s in result map, but it was not found", filename)
                }

                // ファイルの内容を確認
                content, err := os.ReadFile(filename)
                if err != nil {
                        t.Errorf("Failed to read file %s: %v", filename, err)
                        continue
                }

                if string(content) != expectedContent {
                        t.Errorf("File %s content mismatch. Expected: %s, Got: %s", filename, expectedContent, string(content))
                }
        }
}

func TestGetEnvFile(t *testing.T) {
        testCases := []struct {
                name           string
                provider       agent.Provider
                installToken   string
                expectedValues []string
        }{
                {
                        name: "OpenAI Provider",
                        provider: &MockProvider{
                                env: map[string]string{
                                        "TEST_ENV1": "value1",
                                        "TEST_ENV2": "value2",
                                },
                                providerName: "openai",
                                modelName:    "gpt-4",
                                apiKey:       "sk-test-openai-key",
                        },
                        installToken: "github-token-123",
                        expectedValues: []string{
                                "export OPENAI_API_KEY=sk-test-openai-key",
                                "export GOOSE_PROVIDER=openai",
                                "export GOOSE_MODEL=gpt-4",
                                "export GITHUB_TOKEN=github-token-123",
                                "export TEST_ENV1=value1",
                                "export TEST_ENV2=value2",
                        },
                },
                {
                        name: "Anthropic Provider",
                        provider: &MockProvider{
                                env: map[string]string{
                                        "CUSTOM_ENV": "custom-value",
                                },
                                providerName: "anthropic",
                                modelName:    "claude-3-sonnet",
                                apiKey:       "sk-test-anthropic-key",
                        },
                        installToken: "github-token-456",
                        expectedValues: []string{
                                "export ANTHROPIC_API_KEY=sk-test-anthropic-key",
                                "export GOOSE_PROVIDER=anthropic",
                                "export GOOSE_MODEL=claude-3-sonnet",
                                "export GITHUB_TOKEN=github-token-456",
                                "export CUSTOM_ENV=custom-value",
                        },
                },
        }

        for _, tc := range testCases {
                t.Run(tc.name, func(t *testing.T) {
                        // 直接GooseEnvを作成してテスト
                        env := &GooseEnv{
                                APIKey:              tc.provider.GetAPIKey(),
                                Model:               tc.provider.GetModelName(),
                                Provider:            tc.provider.GetProviderName(),
                                Repo:                "test-repo",
                                InstallationToken:   tc.installToken,
                                BaseDir:             "/tmp/test-base-dir",
                                SessionID:           "test-session-id",
                                InstructionFIlePath: "/tmp/test-instruction.md",
                                ScriptFIlePath:      "/tmp/test-script.sh",
                                EnvFilePath:         "/tmp/test-env.sh",
                        }

                        // 環境変数を追加
                        envMap := env.GetEnv()
                        for k, v := range tc.provider.GetEnv() {
                                envMap[k] = v
                        }

                        // 環境変数を文字列に変換
                        var envContent strings.Builder
                        for k, v := range envMap {
                                envContent.WriteString(fmt.Sprintf("export %s=%s\n", k, v))
                        }

                        // 期待される環境変数が含まれているか検証
                        for _, expectedValue := range tc.expectedValues {
                                if !strings.Contains(envContent.String(), expectedValue) {
                                        t.Errorf("Env file doesn't contain expected value: %s", expectedValue)
                                }
                        }
                })
        }
}

func TestGetAPIKeyEnv(t *testing.T) {
        testCases := []struct {
                provider string
                expected string
        }{
                {
                        provider: "openai",
                        expected: "OPENAI_API_KEY",
                },
                {
                        provider: "anthropic",
                        expected: "ANTHROPIC_API_KEY",
                },
                {
                        provider: "groq",
                        expected: "GROQ_API_KEY",
                },
                {
                        provider: "openrouter",
                        expected: "OPENROUTER_API_KEY",
                },
                {
                        provider: "google",
                        expected: "GOOGLE_API_KEY",
                },
                {
                        provider: "unknown",
                        expected: "",
                },
        }

        for _, tc := range testCases {
                t.Run(tc.provider, func(t *testing.T) {
                        result := GetAPIKeyEnv(tc.provider)
                        if result != tc.expected {
                                t.Errorf("GetAPIKeyEnv(%s) = %s, expected %s", tc.provider, result, tc.expected)
                        }
                })
        }
}

func TestAgentGetAPIKeyEnv(t *testing.T) {
        testCases := []struct {
                name     string
                provider agent.Provider
                expected string
        }{
                {
                        name: "OpenAI Provider",
                        provider: &MockProvider{
                                providerName: "openai",
                                apiKey:       "sk-test-openai-key",
                        },
                        expected: "OPENAI_API_KEY=sk-test-openai-key",
                },
                {
                        name: "Anthropic Provider",
                        provider: &MockProvider{
                                providerName: "anthropic",
                                apiKey:       "sk-test-anthropic-key",
                        },
                        expected: "ANTHROPIC_API_KEY=sk-test-anthropic-key",
                },
                {
                        name: "Unknown Provider",
                        provider: &MockProvider{
                                providerName: "unknown",
                                apiKey:       "test-key",
                        },
                        expected: "=test-key",
                },
        }

        for _, tc := range testCases {
                t.Run(tc.name, func(t *testing.T) {
                        agent := &GooseAgent{
                                Opts: GooseOptions{
                                        Provider: tc.provider,
                                },
                        }
                        result := agent.GetAPIKeyEnv()
                        if result != tc.expected {
                                t.Errorf("agent.GetAPIKeyEnv() = %s, expected %s", result, tc.expected)
                        }
                })
        }
}

func TestImportEventURL(t *testing.T) {
        testCases := []struct {
                name          string
                eventURL      string
                expectedPR    int
                expectedIssue int
                expectError   bool
        }{
                {
                        name:          "Valid PR URL",
                        eventURL:      "https://github.com/org/repo/pull/123",
                        expectedPR:    123,
                        expectedIssue: 0,
                        expectError:   false,
                },
                {
                        name:          "Valid Issue URL",
                        eventURL:      "https://github.com/org/repo/issues/456",
                        expectedPR:    0,
                        expectedIssue: 456,
                        expectError:   false,
                },
                {
                        name:          "Invalid URL format",
                        eventURL:      "https://github.com/org/repo",
                        expectedPR:    0,
                        expectedIssue: 0,
                        expectError:   true,
                },
                {
                        name:          "Invalid event type",
                        eventURL:      "https://github.com/org/repo/unknown/789",
                        expectedPR:    0,
                        expectedIssue: 0,
                        expectError:   true,
                },
        }

        for _, tc := range testCases {
                t.Run(tc.name, func(t *testing.T) {
                        g := &GooseGitHub{}
                        err := g.ImportEventURL(tc.eventURL)

                        if tc.expectError {
                                if err == nil {
                                        t.Errorf("Expected error but got nil")
                                }
                        } else {
                                if err != nil {
                                        t.Errorf("Unexpected error: %v", err)
                                }

                                if g.PRNumber != tc.expectedPR {
                                        t.Errorf("Expected PR number %d, got %d", tc.expectedPR, g.PRNumber)
                                }

                                if g.IssueNumber != tc.expectedIssue {
                                        t.Errorf("Expected Issue number %d, got %d", tc.expectedIssue, g.IssueNumber)
                                }
                        }
                })
        }
}

// TestGetExecutionScriptFile tests the getExecutionScriptFile method
func TestGetExecutionScriptFile(t *testing.T) {
        // Create a temporary directory for testing
        tempDir, err := os.MkdirTemp("", "goose-test")
        if err != nil {
                t.Fatalf("Failed to create temp dir: %v", err)
        }
        defer os.RemoveAll(tempDir)

        // Create a mock config
        mockCfg := &config.Config{}

        // Create a mock agent
        agent := &GooseAgent{
                baseDir: tempDir,
                cfg:     mockCfg,
        }

        // Test the getExecutionScriptFile function
        scriptContent := agent.getExecutionScriptFile("/path/to/env/file")

        // Verify that the script content is not empty
        if scriptContent == "" {
                t.Errorf("Expected non-empty script content, got empty string")
        }

        // Check if the script contains expected content
        if !strings.Contains(scriptContent, "#!/bin/bash") {
                t.Errorf("Script content does not contain expected header")
        }

        // Check if the script contains the run_goose function
        if !strings.Contains(scriptContent, "run_goose()") {
                t.Errorf("Script content does not contain run_goose function")
        }
}
