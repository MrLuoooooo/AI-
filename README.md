# AI 视觉对话助手

基于 **Eino 框架** + **Go** + **阿里云百炼** 的多模态实时视觉对话助手。

打开摄像头与麦克风，AI 实时观察画面、识别语音、给予自然回应。

## 技术栈

| 层 | 技术 |
|---|------|
| AI 框架 | [Eino](https://github.com/cloudwego/eino) (字节开源) |
| 后端 | Go 1.23 + Gin + Fx |
| 多模态 | 阿里云百炼 qwen-vl-plus |
| 通信 | WebSocket (帧+语音) |
| 前端 | 原生 HTML5 + Web Speech API |
| 部署 | Docker Compose (双容器) |

## 快速开始

```bash
# 1. 配置 API Key
cp .env.example .env
# 编辑 .env 填入百炼 API Key

# 2. 启动
docker compose up -d

# 3. 打开浏览器
http://localhost:8080
```

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
