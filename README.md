# Eino + MiniMax Demo

使用字节跳动 Eino 框架接入 MiniMax API 的示例项目。

## 项目迭代

| 步骤 | 文件 | 功能 | 复杂度 |
|------|------|------|--------|
| ✅ | `main.go` | 基础 ChatModel 调用 | ⭐ |
| ✅ | `step2_agent_with_tools.go` | Agent + Tool (计算器) | ⭐⭐ |
| 🔄 | `step3_react_agent.go` | ReAct Agent (天气+时间工具) | ⭐⭐⭐ |
| ⏳ | `step4_multi_agent.go` | Multi Agent | ⭐⭐⭐⭐ |

## 运行示例

### Step 1: 基础 ChatModel
```bash
export MINIMAX_API_KEY="sk-cp-your-api-key"
go run main.go
```

### Step 2: Agent + Tools
```bash
export MINIMAX_API_KEY="sk-cp-your-api-key"
go run step2_agent_with_tools.go
```

### Step 3: ReAct Agent
```bash
export MINIMAX_API_KEY="sk-cp-your-api-key"
go run step3_react_agent.go
```

## 核心概念

### ReAct Agent (Step 3)
ReAct = Reasoning + Acting，通过思考-行动-观察循环解决复杂问题：

```go
// 创建 ReAct Agent
agent, err := react.NewAgent(ctx, &react.AgentConfig{
    ToolCallingModel: chatModel,
    ToolsConfig: compose.ToolsNodeConfig{
        InvokableTools: []tool.InvokableTool{weatherTool, timeTool},
    },
    MaxStep: 10,           // 最大步数
    MessageModifier: func(ctx context.Context, input []*schema.Message) []*schema.Message {
        // 修改传入模型的消息
        return append([]*schema.Message{schema.SystemMessage("你是一个智能助手")}, input...)
    },
})

// 调用 Agent
resp, err := agent.Generate(ctx, []*schema.Message{
    schema.UserMessage("北京今天的天气怎么样？"),
})

// 流式输出
stream, _ := agent.Stream(ctx, messages)
for {
    msg, _ := stream.Recv()
    fmt.Print(msg.Content)
}
```

### 核心组件

| 组件 | 说明 |
|------|------|
| `react.NewAgent()` | 创建 ReAct Agent |
| `compose.ToolsNodeConfig` | 工具配置 |
| `MaxStep` | 最大运行步数 |
| `MessageModifier` | 消息修改器 |
| `agent.Generate()` | 非流式调用 |
| `agent.Stream()` | 流式输出 |

## 目录结构

```
eino-minimax-demo/
├── main.go                   # Step 1: 基础 ChatModel
├── step2_agent_with_tools.go # Step 2: Agent + Tools (计算器)
├── step3_react_agent.go      # Step 3: ReAct Agent (天气+时间)
├── step4_multi_agent.go      # Step 4: Multi Agent (待实现)
├── go.mod
└── README.md
```

## 依赖

- [Eino](https://github.com/cloudwego/eino) - 字节跳动 AI 应用框架
- [Eino-Ext](https://github.com/cloudwego/eino-ext) - Eino 扩展组件

## License

MIT
