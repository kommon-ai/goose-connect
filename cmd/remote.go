/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kommon-ai/agent-connect/pkg/service"
	"github.com/kommon-ai/goose-connect/pkg/config"
	"github.com/kommon-ai/goose-connect/pkg/goose"
	"github.com/spf13/cobra"
)

// remoteCmd represents the remote command
var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "リモートエージェントサーバーを起動",
	Long: `リモートエージェントサーバーを指定されたポートで起動します。
このサーバーはGooseエージェントのリモート実行を可能にし、
HTTPエンドポイントを通じてエージェントとの通信を提供します。

使用例:
  goose-connect remote --port 8080

サーバーは指定されたポートでリッスンを開始し、
エージェントタスクの実行要求を受け付けます。`,
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			log.Fatalf("Failed to get port: %v", err)
		}
		if err := config.ValidateRequiredValues(); err != nil {
			log.Fatalf("Failed to validate config: %v", err)
		}
		remoteAgent := service.NewRemoteAgentServer(goose.NewGooseAgentFactory())

		// ハンドラの作成
		mux := http.NewServeMux()

		// RemoteAgentServiceハンドラの登録
		path, handler := remoteAgent.Handler()
		mux.Handle(path, handler)

		// サーバーの設定
		srv := &http.Server{
			Addr:           fmt.Sprintf(":%d", port),
			Handler:        mux,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20, // 1MB
		}

		// サーバーの起動
		log.Printf("Starting server on %d", port)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(remoteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// remoteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// remoteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
