# Eino + MiniMax Demo

使用字节跳动 Eino 框架接入 MiniMax API 的示例项目。

## 前置条件

- Go 1.21+
- MiniMax API Key

## 安装

```bash
go mod tidy
```

## 配置

设置环境变量：

```bash
export MINIMAX_API_KEY="sk-cp-your-api-key-here"
```

## 运行

```bash
go run main.go
```

## 项目结构

```
eino-minimax-demo/
├── go.mod          # Go 模块配置
├── main.go         # 主程序入口
└── README.md       # 说明文档
```

## 依赖

- [Eino](https://github.com/cloudwego/eino) - 字节跳动 AI 应用框架
- [Eino-Ext](https://github.com/cloudwego/eino-ext) - Eino 扩展组件

## License

MIT
