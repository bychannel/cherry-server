# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是一个基于 cherry 框架的多节点游戏服务端示例项目，采用分布式架构，通过 NATS 实现节点间通信。

## 环境要求

- Go >= 1.18
- NATS >= 2.0

## 常用命令

### 编译
```bash
make.bat              # Windows 编译所有节点到 bin/ 目录
```

### 启动 NATS
```bash
# Docker 方式（推荐）
start_nats.bat          # Windows 一键启动 NATS 容器
docker stop nats-server # 停止容器

# 或直接运行 nats-server
nats-server
```

### 启动节点
所有节点从 `nodes/main.go` 启动，使用 urfave/cli 框架：
```bash
# 在 nodes 目录下执行
go run main.go master --path=../../config/cluster.json --node=gc-master
go run main.go center --path=../../config/cluster.json --node=gc-center
go run main.go web --path=../../config/cluster.json --node=gc-web-1
go run main.go gate --path=../../config/cluster.json --node=gc-gate-1
go run main.go game --path=../../config/cluster.json --node=10001
```

### 生成 Protobuf 代码
```bash
build_js_protocol.bat   # 生成 JavaScript 协议代码到 nodes/web/static/
build_go_protocol.bat  # 生成 Go 协议代码到 internal/pb/
```

### 运行测试机器人
```bash
# 位于 robot_client/main.go，通过 TCP 连接 gate 进行压测
```

## 架构概览

### 节点类型

| 节点 | 目录 | 职责 |
|------|------|------|
| master | nodes/master/ | 服务发现，基于 NATS 构建 |
| center | nodes/center/ | 帐号服务、全局唯一性业务 |
| web | nodes/web/ | HTTP 接口（注册、区服列表、SDK 登录/支付回调） |
| gate | nodes/gate/ | 对外网关，管理客户端连接、消息路由 |
| game | nodes/game/ | 游戏逻辑业务，可多节点部署 |

### 目录结构

```
internal/           # 内部业务逻辑
  code/            # 业务状态码定义
  component/       # 组件（如 check_center 启动检查）
  constant/         # 常量定义
  data/            # 配表结构定义（读取 config/data/）
  event/            # 游戏事件（玩家创建/登录/登出）
  guid/             # 全局 ID 生成
  pb/               # Protobuf 生成的 Go 代码
  protocol/         # Protobuf 协议定义
  rpc/              # 跨节点 RPC 函数封装
  session_key/      # Session 相关常量
  token/            # 登录 Token 逻辑
  types/            # 自定义类型封装（序列化/反序列化）

nodes/              # 分布式节点
  center/           # center 节点（含 db、module）
  game/             # game 节点（含 db、module/player）
  gate/             # gate 节点（含 actor_agent）
  master/           # master 节点
  web/              # web 节点（含 controller、sdk）
  main.go           # 所有节点的统一入口
```

### 通信模式

- **节点发现**: master 节点基于 NATS 提供服务发现
- **跨节点调用**: 通过 cherry 框架的 ActorSystem 进行 RPC 调用
  - 格式: `节点ID.节点类型.函数名`
  - 示例: `rpcCenter.GetUID()` → 路由到 center 节点的 account actor
- **消息路由**: gate 节点根据路由规则将客户端消息转发到对应 game 节点

### Protobuf 协议

协议定义在 `internal/protocol/`，编译后生成到 `internal/pb/`。主要消息：
- `login.proto` - 登录相关
- `player.proto` - 玩家相关
- `base_type.proto` - 基础类型
- `rpc.proto` - RPC 相关

### 配表配置

策划配表位于 `config/data/`，读取后封装为 Go 结构体在 `internal/data/` 中使用。
