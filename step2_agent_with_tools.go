package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/eino-ext/components/model/openai"
)

// ============ Tool 定义 ============

// CalculatorParams 计算器参数
type CalculatorParams struct {
	A        float64  `json:"a" jsonschema:"description=第一个数字"`
	B        float64  `json:"b" jsonschema:"description=第二个数字"`
	Operator string  `json:"operator" jsonschema:"description=运算符: add, sub, mul, div"`
}

// Calculator 计算器工具
type Calculator struct{}

func (c *Calculator) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "calculator",
		Description: "执行基本数学计算（加、减、乘、除）。例如：计算 10 + 5，计算 100 * 0.5",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"a": {
				Desc:     "第一个数字",
				Type:     schema.Float,
				Required: true,
			},
			"b": {
				Desc:     "第二个数字",
				Type:     schema.Float,
				Required: true,
			},
			"operator": {
				Desc:     "运算符：add（加）、sub（减）、mul（乘）、div（除）",
				Type:     schema.String,
				Required: true,
			},
		}),
	}, nil
}

func (c *Calculator) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params CalculatorParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return fmt.Sprintf(`{"error": "参数解析失败: %v"}`, err), nil
	}

	var result float64
	switch params.Operator {
	case "add", "+":
		result = params.A + params.B
	case "sub", "-":
		result = params.A - params.B
	case "mul", "*":
		result = params.A * params.B
	case "div", "/":
		if params.B == 0 {
			return `{"error": "除数不能为零"}`, nil
		}
		result = params.A / params.B
	default:
		return fmt.Sprintf(`{"error": "不支持的运算符: %s"}`, params.Operator), nil
	}

	return fmt.Sprintf(`{"result": %.2f}`, result), nil
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

	// 创建 Tool
	calcTool := &Calculator{}
	calcInfo, err := calcTool.Info(ctx)
	if err != nil {
		fmt.Printf("Failed to get tool info: %v\n", err)
		os.Exit(1)
	}

	// 绑定工具到 ChatModel
	err = chatModel.BindTools([]*schema.ToolInfo{calcInfo})
	if err != nil {
		fmt.Printf("Failed to bind tools: %v\n", err)
		os.Exit(1)
	}

	// 创建 ToolsNode
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{calcTool},
	})
	if err != nil {
		fmt.Printf("Failed to create tools node: %v\n", err)
		os.Exit(1)
	}

	// 创建消息模板
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个数学助手。当用户提出计算问题时，使用 calculator 工具来计算并返回结果。"),
		schema.MessagesPlaceholder("chat_history", true),
		schema.UserMessage("问题: {question}"),
	)

	// 渲染模板
	messages, err := template.Format(ctx, map[string]any{
		"question": "计算 100 + 200，再计算 50 * 3",
	})
	if err != nil {
		fmt.Printf("Failed to format template: %v\n", err)
		os.Exit(1)
	}

	// ============ Step 1: 直接调用 ChatModel ============
	fmt.Println("=== Step 1: ChatModel 直接调用 ===")
	result, err := chatModel.Generate(ctx, messages)
	if err != nil {
		fmt.Printf("Failed to generate: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("ChatModel 响应:\n%s\n\n", result.Content)

	// ============ Step 2: 使用 ToolsNode 调用 Tool ============
	fmt.Println("=== Step 2: ChatModel + ToolsNode ===")

	// 构建 Chain: ChatModel -> ToolsNode
	chain := compose.NewChain[[]*schema.Message, []*schema.Message]()
	chain.AppendChatModel(chatModel, compose.WithNodeName("chat_model")).
		AppendToolsNode(toolsNode, compose.WithNodeName("tools"))

	// 编译 Chain
	agent, err := chain.Compile(ctx)
	if err != nil {
		fmt.Printf("Failed to compile chain: %v\n", err)
		os.Exit(1)
	}

	// 运行 Agent
	resp, err := agent.Invoke(ctx, messages)
	if err != nil {
		fmt.Printf("Failed to invoke agent: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Agent 响应:")
	for _, msg := range resp {
		fmt.Println(msg.Content)
	}
}
