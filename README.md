# AI 视觉对话助手

基于 **Eino 框架** + **Go** + **阿里云百炼** 的多模态实时视觉对话助手。

打开摄像头与麦克风，AI 实时观察画面、识别语音、给予自然回应。

## 第三方依赖

| 依赖 | 用途 | 许可证 |
|------|------|--------|
| [Eino](https://github.com/cloudwego/eino) v0.8 | AI Agent 图编排框架（字节开源） | Apache 2.0 |
| [Gin](https://github.com/gin-gonic/gin) v1.10 | HTTP 路由与中间件 | MIT |
| [Fx](https://github.com/uber-go/fx) v1.20 | 依赖注入框架（Uber） | MIT |
| [Zap](https://github.com/uber-go/zap) v1.27 | 结构化日志（Uber） | MIT |
| [Viper](https://github.com/spf13/viper) v1.21 | 配置管理 | MIT |
| [Gorilla WebSocket](https://github.com/gorilla/websocket) v1.5 | WebSocket 协议 | BSD 2-Clause |
| [Lumberjack](https://github.com/natefinch/lumberjack) v2.2 | 日志文件轮转 | MIT |
| [阿里云百炼 DashScope API](https://help.aliyun.com/zh/model-studio/) | 多模态大模型 qwen-vl-plus | 商业 |

## 代码复用声明

本项目后端基础设施代码（`internal/config/`、`internal/server/`、`internal/logger/`、`internal/callback/`、`internal/component/openaimodel/`）引用自本人过往项目 [GoAgentPro](https://github.com/MrLuoooooo/MrLuoooooagent.git)，在此基础上进行了以下原创开发：

| 模块 | 来源 | 说明 |
|------|------|------|
| `internal/graph/vision.go` | 原创 | 基于 Eino 框架构建的视觉多模态对话图 |
| `internal/component/qianwenmodel/` | 原创 | 适配阿里百炼原生 DashScope API 的 ChatModel 实现 |
| `internal/component/frame/` | 原创 | 自适应帧采样器（时间间隔+帧差哈希+缩放压缩） |
| `internal/component/vad/` | 原创 | 纯 Go 能量阈值语音端点检测 |
| `internal/handler/vision.go` | 原创 | WebSocket 实时视觉对话处理器 |
| `internal/service/vision.go` | 原创 | 会话管理、历史追踪、成本计量 |
| `internal/model/vision.go` | 原创 | 视觉对话数据模型 |
| `web/` | 原创 | 前端视觉对话界面（HTML5 + Web Speech API + Canvas） |
| `internal/config/` | 复用 | 基于 GoAgentPro，新增 VisionConfig 配置段 |
| `internal/server/` | 复用 | 基于 GoAgentPro，精简为视觉助手专用路由和依赖注入 |
| `internal/logger/` | 复用 | 基于 GoAgentPro，Zap + Lumberjack 日志方案 |
| `internal/callback/` | 复用 | 基于 GoAgentPro，Eino 全局回调日志 |
| `internal/component/openaimodel/` | 复用 | 基于 GoAgentPro，保留以备 OpenAI 兼容模式 |

## 项目结构

```
├── cmd/server/main.go         # 入口
├── internal/
│   ├── config/                # 配置管理 (Viper)
│   ├── server/                # HTTP 服务 + Fx DI
│   ├── graph/vision.go        # Eino 视觉对话图
│   ├── service/vision.go      # 会话编排
│   ├── handler/vision.go      # WebSocket 处理
│   ├── component/
│   │   ├── qianwenmodel/      # 百炼多模态适配
│   │   ├── frame/             # 自适应帧采样器
│   │   └── vad/               # 能量阈值 VAD
│   └── model/                 # 数据模型
├── web/                       # 前端页面
│   ├── index.html             # 视觉对话界面
│   ├── nginx.conf             # Nginx 反代配置
│   └── Dockerfile             # 前端容器
├── configs/config.yaml        # 应用配置
├── docker-compose.yml         # 双容器编排
├── Dockerfile                 # 后端容器
└── DESIGN.md                  # 设计文档
```

## 设计原则

- **Eino Graph**：Lambda(多模态组装) → ChatModel(视觉推理)
- **LSP 合规**：ModelManager nil-safe、ProcessFrame/Stream 行为一致
- **SOLID**：依赖倒置(Fx DI)、接口隔离(qianwenmodel 实现 ChatModel)
- **成本控制**：帧差采样、VAD 静默过滤、历史滑动窗口
