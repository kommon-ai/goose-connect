package goose

import (
	"github.com/kommon-ai/agent-connect/gen/proto"
	"github.com/kommon-ai/agent-go/pkg/agent"
)

// ProviderToProto は agent.Provider インターフェースを remote.ProviderInfo に変換します
func ProviderToProto(provider agent.Provider) *proto.ProviderInfo {
	if provider == nil {
		return nil
	}

	return &proto.ProviderInfo{
		ModelName:    provider.GetModelName(),
		ApiKey:       provider.GetAPIKey(),
		ProviderName: provider.GetProviderName(),
		Env:          provider.GetEnv(),
	}
}

// GitHubToProto は agent.GitHub インターフェースを remote.GitHubInfo に変換します
func GitHubToProto(github agent.GitHub) *proto.GitHubInfo {
	if github == nil {
		return nil
	}

	prNumber, _ := github.GetPRNumber()       // エラーは無視
	issueNumber, _ := github.GetIssueNumber() // エラーは無視

	// 整数オーバーフロー対策: 大きな値の場合はint32の最大値を使用
	prNumberInt32 := int32(0)
	if prNumber <= 0 {
		prNumberInt32 = 0
	} else if prNumber > 2147483647 { // int32の最大値
		prNumberInt32 = 2147483647
	} else {
		prNumberInt32 = int32(prNumber)
	}

	issueNumberInt32 := int32(0)
	if issueNumber <= 0 {
		issueNumberInt32 = 0
	} else if issueNumber > 2147483647 { // int32の最大値
		issueNumberInt32 = 2147483647
	} else {
		issueNumberInt32 = int32(issueNumber)
	}

	return &proto.GitHubInfo{
		ApiToken:    github.GetAPIToken(),
		ApiUrl:      github.GetAPIURL(),
		Repo:        github.GetRepo(),
		FullRepoUrl: github.GetFullRepoURL(),
		PrNumber:    prNumberInt32,
		IssueNumber: issueNumberInt32,
		BranchName:  github.GetBranchName(),
	}
}

// CreateMockProvider は remote.ProviderInfo からモック Provider を作成します
func CreateMockProvider(info *proto.ProviderInfo) agent.Provider {
	if info == nil {
		return nil
	}

	return &agent.NoopProvider{
		ModelName:    info.ModelName,
		APIKey:       info.ApiKey,
		ProviderName: info.ProviderName,
		Env:          info.Env,
	}
}

// CreateMockGitHub は remote.GitHubInfo からモック GitHub を作成します
func CreateMockGitHub(info *proto.GitHubInfo) agent.GitHub {
	if info == nil {
		return nil
	}

	// GitHub インターフェースを実装するモック構造体
	return &mockGitHub{
		apiToken:    info.ApiToken,
		apiURL:      info.ApiUrl,
		repo:        info.Repo,
		fullRepoURL: info.FullRepoUrl,
		prNumber:    int(info.PrNumber),
		issueNumber: int(info.IssueNumber),
		branchName:  info.BranchName,
	}
}

// mockGitHub は agent.GitHub インターフェースを実装するモック構造体です
type mockGitHub struct {
	apiToken    string
	apiURL      string
	repo        string
	fullRepoURL string
	prNumber    int
	issueNumber int
	branchName  string
}

func (g *mockGitHub) GetAPIToken() string {
	return g.apiToken
}

func (g *mockGitHub) GetAPIURL() string {
	return g.apiURL
}

func (g *mockGitHub) GetRepo() string {
	return g.repo
}

func (g *mockGitHub) GetFullRepoURL() string {
	return g.fullRepoURL
}

func (g *mockGitHub) GetPRNumber() (int, error) {
	return g.prNumber, nil
}

func (g *mockGitHub) GetIssueNumber() (int, error) {
	return g.issueNumber, nil
}

func (g *mockGitHub) GetBranchName() string {
	return g.branchName
}

// ProtoToGooseProvider は remote.ProviderInfo から GooseAPIType を抽出します
func ProtoToGooseProvider(info *proto.ProviderInfo) GooseAPIType {
	if info == nil {
		return ""
	}

	switch info.ProviderName {
	case "openai":
		return GooseAPITypeOpenAI
	case "anthropic":
		return GooseAPITypeAnthropic
	case "openrouter":
		return GooseAPITypeOpenRouter
	case "google":
		return GooseAPITypeGoogle
	case "groq":
		return GooseAPITypeGroq
	case "llamaapi":
		return GooseAPITypeLlamaAPI
	default:
		// デフォルトはOpenAI
		return GooseAPITypeOpenAI
	}
}

// ProtoToGooseGitHub は remote.GitHubInfo から GooseGitHub を作成します
func ProtoToGooseGitHub(info *proto.GitHubInfo) *GooseGitHub {
	if info == nil {
		return nil
	}

	// APIURLからホスト部分を抽出
	host := "https://github.com"
	if info.ApiUrl != "" {
		// APIURLが提供されている場合は、それをそのまま使用
		host = info.ApiUrl
	}

	return &GooseGitHub{
		InstallationToken: info.ApiToken,
		APIURL:            info.ApiUrl,
		Repo:              info.Repo,
		PRNumber:          int(info.PrNumber),
		IssueNumber:       int(info.IssueNumber),
		Host:              host,
		BranchName:        info.BranchName,
	}
}

// ProtoToGooseOptions は remote.ProviderInfo と remote.GitHubInfo から GooseOptions を作成します
func ProtoToGooseOptions(provider *proto.ProviderInfo, github *proto.GitHubInfo, instruction, sessionID string) GooseOptions {
	if provider == nil || github == nil {
		return GooseOptions{}
	}

	// Provider と GitHub オブジェクトの作成
	providerObj := CreateMockProvider(provider)
	githubObj := ProtoToGooseGitHub(github)

	// セッションIDとベースディレクトリの設定
	if sessionID == "" {
		sessionID = "remote-session"
	}

	return GooseOptions{
		SessionID:   sessionID,
		Instruction: instruction,
		Provider:    providerObj,
		GitHub:      githubObj,
	}
}

// ProtoToGooseAgent は remote.ProviderInfo と remote.GitHubInfo から GooseAgent を作成します
func ProtoToGooseAgent(provider *proto.ProviderInfo, github *proto.GitHubInfo, instruction, sessionID string) (agent.Agent, error) {
	if provider == nil || github == nil {
		return nil, nil
	}

	// GooseOptions を作成
	options := ProtoToGooseOptions(provider, github, instruction, sessionID)

	// GooseAgent を作成
	return NewGooseAgent(options)
}
