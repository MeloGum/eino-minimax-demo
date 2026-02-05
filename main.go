package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/eino-ext/components/model/openai"
)

func main() {
	ctx := context.Background()

	// MiniMax API 配置
	apiKey := os.Getenv("MINIMAX_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: MINIMAX_API_KEY not set")
		os.Exit(1)
	}

	// 创建 ChatModel (使用 OpenAI 客户端 + MiniMax 端点)
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   "MiniMax-M2.1",
		APIKey:  apiKey,
		BaseURL: "https://api.minimaxi.com/v1",
	})
	if err != nil {
		fmt.Printf("Failed to create chat model: %v\n", err)
		os.Exit(1)
	}

	// 创建消息模板
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个{role}。请用{style}语气回答问题。"),
		schema.UserMessage("问题: {question}"),
	)

	// 渲染模板
	messages, err := template.Format(ctx, map[string]any{
		"role":     "助手",
		"style":    "简洁、专业",
		"question": "你好，请介绍一下你自己",
	})
	if err != nil {
		fmt.Printf("Failed to format template: %v\n", err)
		os.Exit(1)
	}

	// 调用 MiniMax API
	fmt.Println("Calling MiniMax API via Eino...")
	result, err := chatModel.Generate(ctx, messages)
	if err != nil {
		fmt.Printf("Failed to generate: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n=== Response ===")
	fmt.Println(result.Content)
}
