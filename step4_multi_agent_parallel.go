package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/eino-ext/components/model/openai"
)

// ============ Agent 报告结构 ============

type AgentReport struct {
	AgentName string    `json:"agent_name"`
	Task      string    `json:"task"`
	Status    string    `json:"status"` // "in_progress", "completed", "failed"
	Result    string    `json:"result"`
	Duration  float64   `json:"duration_ms"`
	Timestamp time.Time `json:"timestamp"`
}

type TaskResult struct {
	Task     string      `json:"task"`
	Status   string      `json:"status"`
	Reports  []AgentReport `json:"reports"`
	Summary  string     `json:"summary"`
}

// ============ Tools ============

// ParallelTaskTool - 并行任务执行工具
type ParallelTaskTool struct{}

func (t *ParallelTaskTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "execute_parallel_tasks",
		Description: "并行执行多个任务，每个任务由不同的专业Agent处理。用于同时进行设计、编码、测试等并行工作。输入为JSON数组格式的任务列表。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"tasks": {
				Desc:     "任务列表，JSON数组格式，每个任务包含 name, agent_type, description",
				Type:     schema.String,
				Required: true,
			},
		}),
	}, nil
}

func (t *ParallelTaskTool) Run() tool.InvokableRun {
	return func(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
		var tasks []map[string]string
		if err := json.Unmarshal([]byte(arguments), &tasks); err != nil {
			return fmt.Sprintf(`{"error": "参数解析失败: %v"}`, err), nil
		}

		// 模拟并行执行（实际项目中会委派给真实Agent）
		var wg sync.WaitGroup
		results := make(chan AgentReport, len(tasks))
		
		startTime := time.Now()

		for _, task := range tasks {
			wg.Add(1)
			go func(t map[string]string) {
				defer wg.Done()
				time.Sleep(time.Duration(500+time.Now().UnixNano()%1000) * time.Millisecond) // 模拟耗时
				results <- AgentReport{
					AgentName: t["agent_type"],
					Task:      t["name"],
					Status:    "completed",
					Result:    fmt.Sprintf("✅ %s 已完成", t["name"]),
					Duration:  float64(time.Since(startTime).Milliseconds()),
					Timestamp: time.Now(),
				}
			}(task)
		}

		wg.Wait()
		close(results)

		var reports []AgentReport
		for r := range results {
			reports = append(reports, r)
		}

		summary := fmt.Sprintf("并行任务完成！共 %d 个任务，%d 个成功", len(tasks), len(reports))
		
		result := TaskResult{
			Task:    "并行开发任务",
			Status:  "completed",
			Reports: reports,
			Summary: summary,
		}
		
		data, _ := json.MarshalIndent(result, "", "  ")
		return string(data), nil
	}
}

// ReportTool - 报告生成工具
type ReportTool struct{}

func (t *ReportTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "generate_report",
		Description: "生成任务执行报告，汇总各Agent的工作成果。包含任务列表、完成状态、耗时统计和最终总结。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"task_name": {
				Desc:     "任务名称",
				Type:     schema.String,
				Required: true,
			},
			"work_summary": {
				Desc:     "工作摘要",
				Type:     schema.String,
				Required: true,
			},
		}),
	}, nil
}

func (t *ReportTool) Run() tool.InvokableRun {
	return func(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
		var params map[string]string
		if err := json.Unmarshal([]byte(arguments), &params); err != nil {
			return fmt.Sprintf(`{"error": "参数解析失败: %v"}`, err), nil
		}

		report := fmt.Sprintf(`📋 任务报告: %s

📝 工作摘要: %s

✅ 状态: 已完成
⏱️ 时间: %s

🎯 总结: 所有任务已成功完成！
`, params["task_name"], params["work_summary"], time.Now().Format("2006-01-02 15:04:05"))

		return report, nil
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
	parallelTool := &ParallelTaskTool{}
	reportTool := &ReportTool{}

	parallelInfo, _ := parallelTool.Info(ctx)
	reportInfo, _ := reportTool.Info(ctx)

	// 绑定 Tools
	err = chatModel.BindTools([]*schema.ToolInfo{parallelInfo, reportInfo})
	if err != nil {
		fmt.Printf("Failed to bind tools: %v\n", err)
		os.Exit(1)
	}

	// 创建 ToolsNode
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{parallelTool, reportTool},
	})
	if err != nil {
		fmt.Printf("Failed to create tools node: %v\n", err)
		os.Exit(1)
	}

	// 创建消息模板
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage(`你是一个项目总监，负责协调和管理多个专业Agent并行工作。

工作流程:
1. 当收到开发任务时，使用 execute_parallel_tasks 工具并行委派给多个专业Agent
2. 每个Agent完成后会汇报结果
3. 最后使用 generate_report 工具生成最终报告

专业Agent包括:
- "architect": 负责系统架构设计
- "backend_dev": 负责后端代码开发  
- "frontend_dev": 负责前端代码开发
- "test_dev": 负责编写测试用例
- "devops": 负责部署和运维配置

汇报格式:
- agent_name: Agent名称
- task: 任务名称
- status: 状态 (in_progress/completed/failed)
- result: 执行结果
- duration_ms: 执行耗时`),
		schema.MessagesPlaceholder("chat_history", true),
		schema.UserMessage("问题: {question}"),
	)

	// 测试场景
	testCases := []string{
		"开发一个用户登录模块，包含前端登录页面和后端API",
		"实现一个待办事项管理功能，包括增删改查和列表展示",
	}

	for i, query := range testCases {
		fmt.Printf("\n%s\n", "="*60)
		fmt.Printf("测试 %d: %s\n", i+1, query)
		fmt.Printf("%s\n", "="*60)

		// 渲染模板
		messages, err := template.Format(ctx, map[string]any{
			"question": query,
		})
		if err != nil {
			fmt.Printf("Failed to format template: %v\n", err)
			continue
		}

		// 构建 Chain
		chain := compose.NewChain[[]*schema.Message, []*schema.Message]()
		chain.AppendChatModel(chatModel, compose.WithNodeName("chat_model")).
			AppendToolsNode(toolsNode, compose.WithNodeName("tools"))

		agent, err := chain.Compile(ctx)
		if err != nil {
			fmt.Printf("Failed to compile chain: %v\n", err)
			continue
		}

		// 执行
		resp, err := agent.Invoke(ctx, messages)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Println("\n🤖 Agent 响应:")
		for _, msg := range resp {
			fmt.Println(msg.Content)
		}
	}
}
