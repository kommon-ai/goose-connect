package goose

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kommon-ai/agent-go/pkg/agent"
	"github.com/kommon-ai/goose-connect/pkg/config"
)

type GooseAPIType string

const (
	GooseAPITypeOpenRouter GooseAPIType = "openrouter"
	GooseAPITypeOpenAI     GooseAPIType = "openai"
	GooseAPITypeAnthropic  GooseAPIType = "anthropic"
	GooseAPITypeGoogle     GooseAPIType = "google"
	GooseAPITypeGroq       GooseAPIType = "groq"
	GooseAPITypeLlamaAPI   GooseAPIType = "llamaapi"
)

// GooseEnv implements the AgentEnv interface for Goose
type GooseEnv struct {
	APIKey              string
	Model               string
	Provider            string
	Repo                string
	InstallationToken   string
	BaseDir             string
	SessionID           string
	InstructionFIlePath string
	ScriptFIlePath      string
	EnvFilePath         string
	BranchName          string
}

func (e *GooseEnv) GetEnv() map[string]string {
	return map[string]string{
		GetAPIKeyEnv(e.Provider): e.APIKey,
		"GOOSE_PROVIDER":         e.Provider,
		"GOOSE_MODEL":            e.Model,
		"GITHUB_TOKEN":           e.InstallationToken,
		"REPO":                   e.Repo,
		"BASE_DIR":               e.BaseDir,
		"SESSION_ID":             e.SessionID,
		"INSTRUCTION_FILE_PATH":  e.InstructionFIlePath,
		"SCRIPT_FILE_PATH":       e.ScriptFIlePath,
		"ENV_FILE_PATH":          e.EnvFilePath,
		"PR_BRANCH":              e.BranchName,
	}
}

func (e *GooseEnv) GetRequiredEnv() []string {
	return []string{
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
}

func (e *GooseEnv) ValidateRequiredEnv() error {
	env := e.GetEnv()
	for _, k := range e.GetRequiredEnv() {
		if v, ok := env[k]; !ok || v == "" {
			return fmt.Errorf("environment variable %s is not set", k)
		}
	}
	return nil
}

type GooseGitHub struct {
	InstallationToken string
	APIURL            string
	Repo              string
	PRNumber          int
	IssueNumber       int
	Host              string // ex: https://github.com
	BranchName        string
}

func (g GooseGitHub) GetAPIToken() string {
	return g.InstallationToken
}

func (g GooseGitHub) GetAPIURL() string {
	return g.APIURL
}

func (g GooseGitHub) GetRepo() string {
	return g.Repo
}

func (g GooseGitHub) GetFullRepoURL() string {
	return g.Host + "/" + g.Repo
}

func (g GooseGitHub) GetFullRepoURLWithCredential() (string, error) {
	apiURL := g.GetFullRepoURL()
	url, err := url.Parse(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse API URL: %w", err)
	}
	return fmt.Sprintf("%s://:%s@%s/%s", url.Scheme, g.InstallationToken, url.Host, url.Path), nil
}

func (g GooseGitHub) GetPRNumber() (int, error) {
	return g.PRNumber, nil
}

func (g GooseGitHub) GetIssueNumber() (int, error) {
	return g.IssueNumber, nil
}

func (g GooseGitHub) GetBranchName() string {
	return g.BranchName
}

func (g *GooseGitHub) ImportEventURL(eventURL string) error {
	// eventURL は https://github.com/org/repo/issues/123 のような形式
	// これを PRNumber と IssueNumber に分解する
	parsedURL, err := url.Parse(eventURL)
	if err != nil {
		return fmt.Errorf("failed to parse event URL: %w", err)
	}
	splitPath := strings.Split(parsedURL.Path, "/")

	// パスの先頭は空文字になるため、実際のパスは [1:] から始まる
	// 例: "/org/repo/issues/123" -> ["", "org", "repo", "issues", "123"]
	if len(splitPath) < 5 {
		return fmt.Errorf("invalid event URL: %s", eventURL)
	}

	// インデックスを調整: 先頭が空文字のため、実際のパスは1つずれる
	if len(splitPath) >= 5 && splitPath[3] == "pull" {
		prNumber, err := strconv.Atoi(splitPath[4])
		if err != nil {
			return fmt.Errorf("failed to convert PR number to int: %w", err)
		}
		g.PRNumber = prNumber
	} else if len(splitPath) >= 5 && splitPath[3] == "issues" {
		issueNumber, err := strconv.Atoi(splitPath[4])
		if err != nil {
			return fmt.Errorf("failed to convert issue number to int: %w", err)
		}
		g.IssueNumber = issueNumber
	} else {
		return fmt.Errorf("invalid event URL: %s", eventURL)
	}
	return nil
}

// GooseAgent implements the agent interface for Goose
type GooseAgent struct {
	Opts    GooseOptions
	Env     agent.AgentEnv
	baseDir string
	cfg     *config.Config
}

type GooseOptions struct {
	SessionID   string
	Instruction string
	Provider    agent.Provider
	GitHub      agent.GitHub
}

// GetProvider returns the Provider interface
func (a *GooseAgent) GetProvider() agent.Provider {
	return a.Opts.Provider
}

// GetEnv returns the agent.AgentEnv interface
func (a *GooseAgent) GetEnv() agent.AgentEnv {
	return a.Env
}

// NewGooseAgent creates a new Goose agent
func NewGooseAgent(opts GooseOptions) (agent.Agent, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}
	if opts.SessionID == "" {
		return nil, fmt.Errorf("session ID is required for Goose agent")
	}
	opts.SessionID = strings.ReplaceAll(opts.SessionID, "/", "-")
	var baseDir string
	if cfg.GetBaseDir() != "" {
		baseDir = cfg.GetBaseDir()
	} else {
		baseDir = fmt.Sprintf("%s/.config/goose-connect", os.Getenv("HOME"))
	}
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}
	if opts.GitHub.GetAPIToken() == "" {
		return nil, fmt.Errorf("GitHub API token is required")
	}

	sessionDir := filepath.Join(baseDir, opts.SessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	agent := &GooseAgent{
		Opts: opts,
		Env: &GooseEnv{
			APIKey:              opts.Provider.GetAPIKey(),
			Model:               opts.Provider.GetModelName(),
			Provider:            opts.Provider.GetProviderName(),
			Repo:                opts.GitHub.GetRepo(),
			InstallationToken:   opts.GitHub.GetAPIToken(),
			BranchName:          opts.GitHub.GetBranchName(),
			BaseDir:             baseDir,
			SessionID:           opts.SessionID,
			InstructionFIlePath: filepath.Join(sessionDir, "instruction"),
			ScriptFIlePath:      filepath.Join(sessionDir, "goose-execute.sh"),
			EnvFilePath:         filepath.Join(sessionDir, "env"),
		},
		baseDir: baseDir,
		cfg:     cfg,
	}

	return agent, nil
}

func (a *GooseAgent) GetAgentEndpoint() string {
	orgID := strings.Split(a.Opts.GitHub.GetRepo(), "/")[0]
	return fmt.Sprintf("http://goose-agent-%s.kommon.svc.cluster.local", orgID)
}

func (a *GooseAgent) sessionDir() string {
	return filepath.Join(a.baseDir, a.Opts.SessionID)
}

func (a *GooseAgent) getInstructionScript(input string) string {
	instruction := []string{
		`あなたはソフトウェア開発のプロフェッショナルです。アーキテクチャ構成を検討したり、コードを記述することが得意です。`,
		`言語のランタイムやパッケージマネージャは、mise を経由して使用してください。 mise exec -- の後に続けると実行することができます。必要に応じて mise 経由で言語等をインストールしてください。`,
	}

	instruction = append(instruction,
		`Makefile を確認し、lint, test, build などのコマンドが存在するか確認して、存在した場合はコミット前にそれらを実行して、通るまで修正を繰り返してください。`,
		`CIが存在するか確認して、存在する場合は結果をプッシュごとに確認してください。`,
		`進捗は適宜issueやPRにコメントを投稿してください。 LLM の出力は私は見ません。 issue, PR のコメント頼りです。何卒お願いいたします。`,
		`以下、この変更に関連する情報を提示します。適宜利用してください。`,
		fmt.Sprintf(`Session ID: %s`, a.GetSessionID()),
		`Session ID の末尾にある番号はissueやPRの番号です。他にも関連するissue/PRがある場合は、それらも参照してください。`,
		`対象について言及のない場合は、issueやPRに関連する処理を行うと解釈してください。`,
		fmt.Sprintf(`リポジトリ: https://github.com/%s`, a.Opts.GitHub.GetRepo()),
		`memory-bank を使用できます。作業開始前後に memory-bank を使用して、作業内容を記憶してください。`,
		`memory-bank を使用する際は、まず session ごとにプロジェクトとし、最終的な知見をリポジトリグローバルの memory-bank に蓄積してください。`,
		`実装の際には、まず始めに sequential-thinking を使用して実装方針を検討してください。`,
		`git のコミットは、メソッド単位、または数十行を目安に、粒度を小さく細かくコミットしてください。`,
	)

	// PRブランチ情報を追加（指定されていれば）
	instruction = append(instruction,
		`以下が要求されたプロンプトです。`,
		`---`,
		input,
		`---`,
	)

	return strings.Join(instruction, "\n")
}

func createFile(text string, filePath string, perm os.FileMode) (*os.File, error) {
	f, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	_, err = f.WriteString(text)
	if err != nil {
		return nil, fmt.Errorf("failed to write to file: %w", err)
	}

	if chmodErr := os.Chmod(filePath, perm); chmodErr != nil {
		return nil, fmt.Errorf("failed to chmod file: %w", chmodErr)
	}
	log.Printf("Created file: %s", filePath)

	return f, nil
}

// m.key の名前のファイルを作成する
func createFiles(m map[string]string) error {
	for k, v := range m {
		f, err := createFile(v, k, 0644)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer f.Close()
	}
	return nil
}

// Execute sends a command to Goose
func (a *GooseAgent) Execute(ctx context.Context, input string) (string, error) {
	agentEnv := a.GetEnv()
	gooseEnv, ok := agentEnv.(*GooseEnv)
	if !ok {
		return "", fmt.Errorf("failed to cast agentEnv to GooseEnv")
	}
	if finalizeErr := FinalizeEnvFile(gooseEnv.EnvFilePath, gooseEnv); finalizeErr != nil {
		return "", fmt.Errorf("failed to finalize env file: %w", finalizeErr)
	}
	if err := createFiles(map[string]string{
		gooseEnv.InstructionFIlePath: a.getInstructionScript(input),
		gooseEnv.ScriptFIlePath:      a.getExecutionScriptFile(gooseEnv.EnvFilePath),
	}); err != nil {
		return "", fmt.Errorf("failed to create files: %w", err)
	}
	// #nosec G204 -- This is a controlled environment where we create the script
	cmd := exec.CommandContext(ctx, "bash", gooseEnv.ScriptFIlePath)
	log.Printf("Executing command: %v", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to execute command: %v", err)
		return "", fmt.Errorf("failed to execute command: %w", err)
	}
	log.Printf("Command output: %s", out)
	return string(out), nil
}

func GetAPIKeyEnv(provider string) string {
	switch provider {
	case "openai":
		return "OPENAI_API_KEY"
	case "anthropic":
		return "ANTHROPIC_API_KEY"
	case "groq":
		return "GROQ_API_KEY"
	case "openrouter":
		return "OPENROUTER_API_KEY"
	case "google":
		return "GOOGLE_API_KEY"
	}
	return ""
}
func (a *GooseAgent) GetAPIKeyEnv() string {
	return fmt.Sprintf("%s=%s", GetAPIKeyEnv(a.Opts.Provider.GetProviderName()), a.Opts.Provider.GetAPIKey())
}

// createExecutionScriptFile は実行用のシェルスクリプトを生成し、一時ファイルとして保存します
// 引数の instructionPath は指示ファイルのパスです
// 戻り値は生成されたスクリプトファイルとエラーです
func (a *GooseAgent) getExecutionScriptFile(envFilePath string) string {
	gituser := a.cfg.GetGitUser()
	gitmail := a.cfg.GetGitMail()
	script := fmt.Sprintf(`#!/bin/bash
source %s
SESSION_DIR=$BASE_DIR/$SESSION_ID
REPO_URL=https://x-oauth-token:${GITHUB_TOKEN}@github.com/$REPO
mkdir -p $SESSION_DIR
if [ -d "$SESSION_DIR/repo" ]; then
  cd $SESSION_DIR/repo
  git remote remove origin
  git remote add origin $REPO_URL
  git fetch origin
else
  mkdir -p $SESSION_DIR
  git clone $REPO_URL $SESSION_DIR/repo
  cd $SESSION_DIR/repo
fi

git config --global user.email "%s"
git config --global user.name "%s"

# PRブランチが指定されている場合はそのブランチをチェックアウト
if [ -n "$PR_BRANCH" ]; then
  echo "Checking out PR branch: $PR_BRANCH"
  git checkout $PR_BRANCH || git checkout -b $PR_BRANCH origin/$PR_BRANCH
fi

run_goose() {
  RESUME=$1
  goose run --name $SESSION_ID $RESUME \
    --with-builtin "developer" \
    --with-extension "GITHUB_PERSONAL_ACCESS_TOKEN=$GITHUB_TOKEN mise exec -- npx -y @modelcontextprotocol/server-github" \
    --with-extension "MEMORY_BANK_ROOT=$HOME/.kommon/memory mise exec -- npx -y @allpepper/memory-bank-mcp" \
	--with-extension "mise exec -- npx -y @modelcontextprotocol/server-sequential-thinking" \
    --instructions $INSTRUCTION_FILE_PATH
  return $?
}
run_goose -r || run_goose
wait

`, envFilePath, gitmail, gituser)
	return script
}

// GetSessionID returns the current session ID
func (a *GooseAgent) GetSessionID() string {
	return a.Opts.SessionID
}

// Clean removes all resources associated with the session
func (a *GooseAgent) Clean() error {
	sessionDir := a.sessionDir()

	log.Printf("Cleaning up session directory: %s", sessionDir)

	// Check if directory exists
	if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
		log.Printf("Session directory does not exist: %s", sessionDir)
		return nil
	}

	// Remove the entire session directory
	if err := os.RemoveAll(sessionDir); err != nil {
		return fmt.Errorf("failed to remove session directory: %w", err)
	}

	log.Printf("Successfully removed session directory: %s", sessionDir)
	return nil
}

func FinalizeEnvFile(filePath string, agentEnv agent.AgentEnv) error {
	if err := agentEnv.ValidateRequiredEnv(); err != nil {
		return fmt.Errorf("failed to validate required env: %w", err)
	}
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()
	env := agentEnv.GetEnv()
	for k, v := range env {
		_, err = f.WriteString(fmt.Sprintf("export %s=\"%s\"\n", k, v))
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}
	return nil
}
