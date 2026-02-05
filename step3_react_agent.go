package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/eino-ext/components/model/openai"
)

// ============ Tools 定义 ============

// WeatherParams 查询天气参数
type WeatherParams struct {
	City    string `json:"city" jsonschema:"description=城市名称，如北京、上海"`
	Date    string `json:"date" jsonschema:"description=日期，格式 YYYY-MM-DD"`
}

// WeatherTool 天气查询工具
type WeatherTool struct{}

func (w *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "weather",
		Description: "查询指定城市和日期的天气情况。使用前请确认城市名称和日期。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"city": {
				Desc:     "城市名称（中文或英文）",
				Type:     schema.String,
				Required: true,
			},
			"date": {
				Desc:     "日期，格式 YYYY-MM-DD",
				Type:     schema.String,
				Required: true,
			},
		}),
	}, nil
}

func (w *WeatherTool) Run() tool.InvokableRun {
	return func(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
		var params WeatherParams
		if err := json.Unmarshal([]byte(arguments), &params); err != nil {
			return fmt.Sprintf(`{"error": "参数解析失败: %v"}`, err), nil
		}

		// Mock 天气数据
		weatherData := map[string]map[string]string{
			"北京": {
				"2026-02-05": "晴，-5°C~5°C",
				"2026-02-06": "多云，-3°C~7°C",
			},
			"上海": {
				"2026-02-05": "小雨，3°C~10°C",
				"2026-02-06": "阴，2°C~8°C",
			},
			"深圳": {
				"2026-02-05": "晴，15°C~24°C",
				"2026-02-06": "多云，16°C~25°C",
			},
		}

		cityData, ok := weatherData[params.City]
		if !ok {
			return fmt.Sprintf(`{"city": "%s", "weather": "数据未找到"}`, params.City), nil
		}

		weather, ok := cityData[params.Date]
		if !ok {
			return fmt.Sprintf(`{"city": "%s", "date": "%s", "weather": "数据未找到"}`, params.City, params.Date), nil
		}

		return fmt.Sprintf(`{"city": "%s", "date": "%s", "weather": "%s"}`, params.City, params.Date, weather), nil
	}
}

// TimeTool 获取当前时间
type TimeTool struct{}

func (t *TimeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "get_current_time",
		Description: "获取当前时间。用于回答用户关于当前时间的问题。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *TimeTool) Run() tool.InvokableRun {
	return func(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
		now := time.Now().Format("2006-01-02 15:04:05")
		return fmt.Sprintf(`{"current_time": "%s"}`, now), nil
	}
}

// ============ 主程序 ============

func main() {
	ctx := context.Background()

	// MiniMax API 配置
	apiKey := os.Getenv("MINIMAX_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: MINIMAX_API_KEY not set")
		os.Exit(1)
	}

	// 创建 ChatModel
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   "MiniMax-M2.1",
		APIKey:  apiKey,
		BaseURL: "https://api.minimaxi.com/v1",
	})
	if err != nil {
		fmt.Printf("Failed to create chat model: %v\n", err)
		os.Exit(1)
	}

	// 创建 Tools
	weatherTool := &WeatherTool{}
	timeTool := &TimeTool{}

	// 配置 Tools
	toolsConfig := compose.ToolsNodeConfig{
		InvokableTools: []tool.InvokableTool{
			weatherTool,
			timeTool,
		},
	}

	// ============ 创建 ReAct Agent ============
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig:      toolsConfig,
		MaxStep:          10, // 最多 5 轮 tool 调用
		MessageModifier: func(ctx context.Context, input []*schema.Message) []*schema.Message {
			// 添加系统提示
			res := make([]*schema.Message, 0, len(input)+1)
			res = append(res, schema.SystemMessage("你是一个智能助手。当用户询问天气或时间时，使用对应的工具来获取准确信息。"))
			res = append(res, input...)
			return res
		},
	})
	if err != nil {
		fmt.Printf("Failed to create agent: %v\n", err)
		os.Exit(1)
	}

	// ============ 测试场景 ============

	testCases := []string{
		"现在几点了？",
		"北京 2026-02-05 的天气怎么样？",
		"请告诉我上海今天的天气，再告诉我现在的时间。",
	}

	for i, query := range testCases {
		fmt.Printf("\n%s\n", "="*50)
		fmt.Printf("测试 %d: %s\n", i+1, query)
		fmt.Printf("%s\n", "="*50)

		// 调用 Agent
		resp, err := agent.Generate(ctx, []*schema.Message{
			schema.UserMessage(query),
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Println("\n最终响应:")
		fmt.Println(resp.Content)
	}

	// ============ 流式输出示例 ============
	fmt.Printf("\n%s\n", "="*50)
	fmt.Println("流式输出示例: 深圳明天的天气")
	fmt.Printf("%s\n", "="*50)

	stream, err := agent.Stream(ctx, []*schema.Message{
		schema.UserMessage("深圳 2026-02-06 的天气"),
	})
	if err != nil {
		fmt.Printf("Stream error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n流式响应:")
	for {
		msg, err := stream.Recv()
		if err != nil {
			break
		}
		if msg.Content != "" {
			fmt.Print(msg.Content)
		}
	}
	fmt.Println()
}
