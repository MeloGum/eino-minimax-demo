# Eino + MiniMax Demo

使用字节跳动 Eino 框架接入 MiniMax API 的示例项目。

## 项目迭代

| 步骤 | 文件 | 功能 | 复杂度 |
|------|------|------|--------|
| ✅ | `main.go` | 基础 ChatModel 调用 | ⭐ |
| 🔄 | `step2_agent_with_tools.go` | Agent + Tool (计算器) | ⭐⭐ |
| ⏳ | `step3_react_agent.go` | ReAct Agent | ⭐⭐⭐ |
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

## 核心代码

```go
// 创建 MiniMax ChatModel
chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
    Model:   "MiniMax-M2.1",
    APIKey:  apiKey,
    BaseURL: "https://api.minimaxi.com/v1",
})

// 绑定 Tool
chatModel.BindTools([]*schema.ToolInfo{calculatorToolInfo})

// 构建 Agent Chain
chain := compose.NewChain[...]
chain.AppendChatModel(chatModel).AppendToolsNode(toolsNode)
agent := chain.Compile(ctx)

// 调用 Agent
resp, err := agent.Invoke(ctx, messages)
```

## 目录结构

```
eino-minimax-demo/
├── main.go                    # Step 1: 基础 ChatModel
├── step2_agent_with_tools.go  # Step 2: Agent + Tools
├── step3_react_agent.go       # Step 3: ReAct Agent (待实现)
├── step4_multi_agent.go       # Step 4: Multi Agent (待实现)
├── go.mod
└── README.md
```

## 依赖

- [Eino](https://github.com/cloudwego/eino) - 字节跳动 AI 应用框架
- [Eino-Ext](https://github.com/cloudwego/eino-ext) - Eino 扩展组件

## License

MIT
