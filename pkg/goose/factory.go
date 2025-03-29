package goose

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v57/github"
	"github.com/kommon-ai/agent-connect/gen/proto"
	"github.com/kommon-ai/agent-go/pkg/agent"
)

func addLabel(client *github.Client, org, repo string, prNumber int, label string) error {
	_, _, err := client.Issues.AddLabelsToIssue(context.Background(), org, repo, prNumber, []string{label})
	return err
}

func removeLabel(client *github.Client, org, repo string, prNumber int, label string) error {
	_, err := client.Issues.RemoveLabelForIssue(context.Background(), org, repo, prNumber, label)
	return err
}

type GooseAgentFactory struct {
	beforeFunc func(msg *proto.ExecuteTaskRequest) error
	afterFunc  func(msg *proto.ExecuteTaskRequest) error
}

func prOrIssueNumber(gh *proto.GitHubInfo) (int, error) {
	if gh.PrNumber > 0 {
		return int(gh.PrNumber), nil
	}
	if gh.IssueNumber > 0 {
		return int(gh.IssueNumber), nil
	}
	return -1, fmt.Errorf("no PR or issue number found")

}

func splitRepo(repo string) (string, string) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func NewGooseAgentFactory() *GooseAgentFactory {
	return &GooseAgentFactory{
		beforeFunc: func(msg *proto.ExecuteTaskRequest) error {
			githubClient := github.NewTokenClient(context.Background(), msg.Github.ApiToken)
			num, err := prOrIssueNumber(msg.Github)
			if err != nil {
				return err
			}
			org, repo := splitRepo(msg.Github.GetRepo())
			return addLabel(githubClient, org, repo, num, "goose-running")
		},
		afterFunc: func(msg *proto.ExecuteTaskRequest) error {
			githubClient := github.NewTokenClient(context.Background(), msg.Github.ApiToken)
			num, err := prOrIssueNumber(msg.Github)
			if err != nil {
				return err
			}
			org, repo := splitRepo(msg.Github.GetRepo())
			return removeLabel(githubClient, org, repo, num, "goose-running")
		},
	}
}

func (f *GooseAgentFactory) NewAgentFactory() func(msg *proto.ExecuteTaskRequest) (agent.Agent, error) {
	return func(msg *proto.ExecuteTaskRequest) (agent.Agent, error) {
		gooseAgent, err := ProtoToGooseAgent(msg.Provider, msg.Github, msg.Instruction, msg.SessionId)
		if err != nil {
			return nil, err
		}
		return gooseAgent, nil
	}
}

func (f *GooseAgentFactory) GetAfterTaskExecutionFunc() func(msg *proto.ExecuteTaskRequest) error {
	if f.afterFunc == nil {
		return func(msg *proto.ExecuteTaskRequest) error {
			return fmt.Errorf("afterFunc is not set")
		}
	}
	return f.afterFunc
}

func (f *GooseAgentFactory) GetBeforeTaskExecutionFunc() func(msg *proto.ExecuteTaskRequest) error {
	if f.beforeFunc == nil {
		return func(msg *proto.ExecuteTaskRequest) error {
			return fmt.Errorf("beforeFunc is not set")
		}
	}
	return f.beforeFunc
}

func (f *GooseAgentFactory) SetAfterTaskExecutionFunc(afterFunc func(msg *proto.ExecuteTaskRequest) error) error {
	f.afterFunc = afterFunc
	return nil
}

func (f *GooseAgentFactory) SetBeforeTaskExecutionFunc(beforeFunc func(msg *proto.ExecuteTaskRequest) error) error {
	f.beforeFunc = beforeFunc
	return nil
}
