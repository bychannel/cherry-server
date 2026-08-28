# 玩家注册登录流程分析与优化建议

> 生成日期：2026-08-28
> 项目：cherry-server

---

## 一、当前整体流程

### 1.1 系统架构图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              客户端                                      │
└─────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│    Web      │────▶│   Center    │     │    Gate     │     │    Game     │
│  (HTTP API) │     │  (帐号服务)  │     │   (网关)     │     │  (游戏逻辑)  │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │                   │
       │                   │                   │                   │
       ▼                   ▼                   ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  第三方 SDK  │     │  DB/Cache   │     │  Session    │     │   Player    │
│ (登录验证)   │     │  (数据存储)  │     │  (会话管理)  │     │   (玩家)     │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
```

### 1.2 注册登录时序图

#### 注册流程 (开发模式)

```
客户端                    Web节点                  Center节点                DB/Cache
   │                         │                         │                         │
   │── GET /register ───────>│                         │                         │
   │                         │── RegisterDevAccount ──>│                         │
   │                         │                         │── 检查帐号是否存在 ─────>│
   │                         │                         │<── 不存在 ──────────────│
   │                         │                         │── 创建DevAccountTable─>│
   │                         │                         │── 写入缓存 ───────────>│
   │                         │<── code.OK ─────────────│<── OK ────────────────│
   │<── 注册成功 ────────────│                         │                         │
```

#### 登录流程 (开发模式)

```
客户端                    Web节点                  第三方SDK                 Center节点                DB/Cache
   │                         │                         │                         │                         │
   │── GET /login ─────────>│                         │                         │                         │
   │                         │── sdkInvoke.Login() ──>│                         │                         │
   │                         │                         │── GetDevAccount() ────>│                         │
   │                         │                         │                         │── 验证帐号密码 ───────>│
   │                         │                         │                         │<── accountId ──────────│
   │                         │<── {open_id: accountId}│<── accountId ──────────│                         │
   │                         │                         │                         │                         │
   │                         │── 生成Token ────────────│                         │                         │
   │<── base64Token ────────│                         │                         │                         │
```

#### Gate 验证流程

```
客户端                    Gate节点                  Center节点
   │                         │                         │
   │── gate.user.login ────>│                         │
   │                         │── DecodeToken() ───────│
   │                         │── GetUID() ───────────>│
   │                         │<── uid ───────────────│
   │                         │                         │
   │                         │── pomelo.Bind(uid,sid) │
   │                         │── 踢掉旧连接(如有) ────│
   │<── LoginResponse ──────│                         │
```

---

## 二、注册流程详解

### 2.1 流程步骤

| 步骤 | 组件 | 操作 | 文件位置 |
|------|------|------|----------|
| 1 | Web | 接收注册请求 `GET /register?account=&password=` | `nodes/web/controller/loginController.go:27` |
| 2 | Web→Center | RPC调用 `RegisterDevAccount` | `internal/rpc/center/center.go:50` |
| 3 | Center | 验证帐密格式 (长度3-18) | `nodes/center/module/account/actor_account.go:34-44` |
| 4 | Center | 检查帐号是否存在 | `nodes/center/db/dev_account_table.go:26` |
| 5 | Center | 创建 `DevAccountTable` → 写入缓存 | `nodes/center/db/dev_account_table.go:31-39` |

### 2.2 核心代码

**Web 层 (loginController.go:27-33)**:
```go
func (p *LoginController) register(c *cherryGin.Context) {
    accountName := c.GetString("account", "", true)
    password := c.GetString("password", "", true)

    statusCode := rpcCenter.RegisterDevAccount(p.App, accountName, password, c.ClientIP())
    code.RenderResult(c, statusCode)
}
```

**Center 层 (actor_account.go:29-47)**:
```go
func (p *ActorAccount) registerDevAccount(req *pb.DevRegister) int32 {
    accountName := req.AccountName
    password := req.Password

    if strings.TrimSpace(accountName) == "" || strings.TrimSpace(password) == "" {
        return code.LoginError
    }

    if len(accountName) < 3 || len(accountName) > 18 {
        return code.LoginError
    }

    if len(password) < 3 || len(password) > 18 {
        return code.LoginError
    }

    return db.DevAccountRegister(accountName, password, req.Ip)
}
```

**数据层 (dev_account_table.go:25-43)**:
```go
func DevAccountRegister(accountName, password, ip string) int32 {
    devAccount, _ := DevAccountWithName(accountName)
    if devAccount != nil {
        return code.AccountNameIsExist
    }

    devAccountTable := &DevAccountTable{
        AccountId:   guid.Next(),
        AccountName: accountName,
        Password:    password,        // ⚠️ 密码明文存储
        CreateIP:    ip,
        CreateTime:  cherryTime.Now().Unix(),
    }

    devAccountCache.Put(accountName, devAccountTable)
    // TODO 保存db              // ⚠️ 数据未持久化

    return code.OK
}
```

---

## 三、登录流程详解

### 3.1 开发模式登录 (DevSDK)

| 步骤 | 组件 | 操作 | 文件位置 |
|------|------|------|----------|
| 1 | Web | 接收登录请求 `GET /login?pid=&account=&password=` | `loginController.go:37` |
| 2 | Web | 根据 pid 获取 SdkConfig | `loginController.go:46` |
| 3 | Web→SDK | 调用 `sdkInvoke.Login()` | `loginController.go:64` |
| 4 | DevSDK | 调用 `GetDevAccount` 验证帐密 | `dev_sdk.go:31` |
| 5 | Center | 验证帐密，返回 accountId 作为 openId | `actor_account.go:50-59` |
| 6 | Web | 生成 Token (`pid:openId:timestamp` + MD5) | `token.go:25-34` |
| 7 | Web | 返回 base64 Token 给客户端 | `loginController.go:88-89` |

### 3.2 Token 生成机制

**Token 结构 (token.go:18-23)**:
```go
type Token struct {
    PID       int32  `json:"pid"`
    OpenID    string `json:"open_id"`
    Timestamp int64  `json:"tt"`
    Hash      string `json:"hash"`
}
```

**Token 生成 (token.go:25-39)**:
```go
func New(pid int32, openId string, appKey string) *Token {
    token := &Token{
        PID:       pid,
        OpenID:    openId,
        Timestamp: cherryTime.Now().ToMillisecond(),
    }

    token.Hash = BuildHash(token, appKey)
    return token
}

func (t *Token) ToBase64() string {
    bytes, _ := json.Marshal(t)
    return cherryCrypto.Base64Encode(string(bytes))
}

func BuildHash(t *Token, appKey string) string {
    value := fmt.Sprintf(hashFormat, t.PID, t.OpenID, t.Timestamp)
    return cherryCrypto.MD5(value + appKey)
}
```

### 3.3 Gate 验证流程

| 步骤 | Gate | 操作 | 文件位置 |
|------|------|------|----------|
| 1 | Gate | 接收客户端 `gate.user.login` | `actor_agent.go:49` |
| 2 | Gate | 解码并验证 Token | `actor_agent.go:56, 103-119` |
| 3 | Gate | 检查 pid 对应的 SdkConfig | `actor_agent.go:63` |
| 4 | Gate→Center | 调用 `GetUID` 获取 uid | `actor_agent.go:70` |
| 5 | Gate | 绑定 uid ↔ sid (`pomelo.Bind`) | `actor_agent.go:76` |
| 6 | Gate | 踢掉同一 uid 的旧连接 | `actor_agent.go:84-86` |
| 7 | Gate | 跨节点通知其他 Gate 踢人 | `actor_agent.go:122-134` |
| 8 | Gate | 存储 session 数据 (serverId, pid, openId) | `actor_agent.go:90-92` |

---

## 四、数据流总览

### 4.1 数据存储位置

| 数据 | 存储位置 | 持久化 | 备注 |
|------|----------|--------|------|
| 开发帐号 (DevAccountTable) | `devAccountCache` (内存) | ❌ 未持久化 | `nodes/center/db/cache.go` |
| 用户绑定 (UserBindTable) | `userBindCache` (内存) | ❌ 未持久化 | 绑定 openId ↔ uid |
| Token | 客户端存储 | - | Base64 编码 |
| Session | Gate 内存 | ❌ 连接断开丢失 | sid ↔ uid 映射 |

### 4.2 关键缓存

**dev_account_table.go:45-52**:
```go
func DevAccountWithName(accountName string) (*DevAccountTable, error) {
    val, found := devAccountCache.GetIfPresent(accountName)
    if found == false {
        return nil, cherryError.Error("account not found")
    }

    return val.(*DevAccountTable), nil
}
```

**节点启动时预加载 (dev_account_table.go:55-72)**:
```go
func loadDevAccount() {
    // 演示用，直接手工构建几个帐号
    for i := 1; i <= 10; i++ {
        index := cherryString.ToString(i)

        devAccount := &DevAccountTable{
            AccountId:   guid.Next(),
            AccountName: "test" + index,
            Password:    "test" + index,
            CreateIP:    "127.0.0.1",
            CreateTime:  cherryTime.Now().ToMillisecond(),
        }

        devAccountCache.Put(devAccount.AccountName, devAccount)
    }

    cherryLogger.Info("preload DevAccountTable")
}
```

---

## 五、问题汇总

### 5.1 🔴 高优先级 (安全性/数据完整性)

| # | 问题 | 位置 | 风险等级 | 详细说明 |
|---|------|------|----------|----------|
| 1 | **密码明文存储** | `dev_account_table.go:34` | 🔴 严重 | 密码未加密，数据库泄露 = 全部帐密泄露 |
| 2 | **注册数据未持久化** | `dev_account_table.go:40` | 🔴 严重 | 服务重启后注册数据全部丢失 |
| 3 | **Token 无重放保护** | `token.go` | 🔴 严重 | 可截获 Token 后无限期使用 |
| 4 | **登录成功无事件通知** | `loginController.go` | 🟠 中等 | 无法记录登录时间/IP/设备信息 |

### 5.2 🟡 中优先级 (架构/可扩展性)

| # | 问题 | 位置 | 影响 |
|---|------|------|------|
| 5 | **Token 有效期计算 Bug** | `token.go:63-64` | token 永不过期 |
| 6 | **中心节点单点依赖** | `rpc/center/center.go` | Center 挂了登录完全失败 |
| 7 | **登录流程与 SDK 紧耦合** | `loginController.go:64` | 新 SDK 需要修改控制器 |
| 8 | **无登录会话管理** | `actor_agent.go` | 无法查询/强制下线/踢出用户 |
| 9 | **无登录失败限制** | `dev_sdk.go` | 可暴力破解密码 |

### 5.3 🟢 低优先级 (功能完善)

| # | 问题 | 位置 | 说明 |
|---|------|------|------|
| 10 | **无登出接口** | - | 无法主动使 Token 失效 |
| 11 | **无登录日志** | - | 审计/风控困难 |
| 12 | **QuickSDK PayCallback 空实现** | `quick_sdk.go` | 支付回调未完成 |

---

## 六、具体问题详解

### 6.1 Token 有效期 Bug (严重)

**问题位置**: `token.go:62-69`

```go
func Validate(token *Token, appKey string) (int32, bool) {
    now := cherryTime.Now()
    now.AddDays(tokenExpiredDay)  // ⚠️ 返回值未使用!

    if token.Timestamp > now.ToMillisecond() {  // ⚠️ 比较逻辑反了
        cherryLogger.Warnf("token is expired, token = %s", token)
        return code.AccountTokenValidateFail, false
    }

    newHash := BuildHash(token, appKey)
    if newHash != token.Hash {
        cherryLogger.Warnf("hash validate fail. newHash = %s, token = %s", token)
        return code.AccountTokenValidateFail, false
    }

    return code.OK, true
}
```

**问题分析**:
1. `now.AddDays(tokenExpiredDay)` 返回新的时间对象，但未赋值给任何变量
2. 条件判断 `token.Timestamp > now.ToMillisecond()` 逻辑相反，应该是检查当前时间是否超过过期时间

**正确逻辑**:
```go
func Validate(token *Token, appKey string) (int32, bool) {
    now := cherryTime.Now()

    // 检查 Token 是否过期
    // 过期时间 = token.Timestamp + 3天(毫秒)
    expiry := token.Timestamp + (tokenExpiredDay * 24 * 60 * 60 * 1000)
    if now.ToMillisecond() > expiry {
        cherryLogger.Warnf("token is expired, token = %s", token)
        return code.AccountTokenValidateFail, false
    }

    newHash := BuildHash(token, appKey)
    if newHash != token.Hash {
        cherryLogger.Warnf("hash validate fail. newHash = %s, token = %s", token)
        return code.AccountTokenValidateFail, false
    }

    return code.OK, true
}
```

**影响**: Token 设置了 3 天有效期，但因为 Bug 导致**永远不会过期**。

---

### 6.2 密码明文存储

**问题位置**: `dev_account_table.go:31-38`

```go
devAccountTable := &DevAccountTable{
    AccountId:   guid.Next(),
    AccountName: accountName,
    Password:    password,        // ⚠️ 直接存储明文密码
    CreateIP:    ip,
    CreateTime:  cherryTime.Now().Unix(),
}
```

**风险**:
- 数据库泄露后所有用户密码曝光
- 无法防止彩虹表攻击

**建议方案**:
```go
import "golang.org/x/crypto/bcrypt"

// 存储时哈希
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
devAccountTable.Password = string(hashedPassword)

// 验证时比较
err := bcrypt.CompareHashAndPassword([]byte storedHash, []byte(password))
```

---

### 6.3 注册数据未持久化

**问题位置**: `dev_account_table.go:39-40`

```go
devAccountCache.Put(accountName, devAccountTable)
// TODO 保存db              // ⚠️ 标记了但未实现
```

**影响**:
- 服务重启后所有注册的帐号丢失
- 只能使用预置的 test1-test10 帐号

**建议**: 实现 GORM 写入数据库

---

### 6.4 无登录失败限流

**问题位置**: `dev_sdk.go:21-39`

```go
func (p devSdk) Login(_ *data.SdkRow, params Params, callback Callback) {
    accountName, _ := params.GetString("account")
    password, _ := params.GetString("password")

    if accountName == "" || password == "" {
        // ...
    }

    accountId := rpcCenter.GetDevAccount(p.app, accountName, password)
    if accountId < 1 {
        callback(code.LoginError, nil)
        return
    }

    callback(code.OK, map[string]string{
        "open_id": cherryString.ToString(accountId),
    })
}
```

**风险**: 可进行无限次的密码暴力猜测

**建议方案**: 使用 Redis 存储登录失败计数
```
key: login:fail:{account}
TTL: 15分钟
value: 失败次数
限制: 5次/15分钟
```

---

### 6.5 Center 节点单点依赖

**问题位置**: `rpc/center/center.go:105-111`

```go
func GetCenterNodeID(app cfacade.IApplication) string {
    list := app.Discovery().ListByType(centerType)
    if len(list) > 0 {
        return list[0].GetNodeID()
    }
    return ""
}
```

**风险**: 
- 只获取第一个 Center 节点
- Center 挂了，所有依赖它的服务（登录、获取 UID）全部失败

**建议**: 实现负载均衡或主备切换

---

## 七、优化建议优先级

### 第一阶段：数据安全 (紧急)

| 优先级 | 优化项 | 工作量 | 预期效果 |
|--------|--------|--------|----------|
| P0 | 修复 Token 过期判断 Bug | 小 | 修复安全逻辑 |
| P0 | 密码哈希存储 | 中 | 防止密码泄露 |
| P1 | 注册数据持久化 | 中 | 防止数据丢失 |

### 第二阶段：安全性增强

| 优先级 | 优化项 | 工作量 | 预期效果 |
|--------|--------|--------|----------|
| P1 | 添加登录失败限流 | 中 | 防止暴力破解 |
| P2 | Token 重放保护 | 大 | 使用一次性 Token |
| P2 | 登录事件记录 | 中 | 审计日志 |

### 第三阶段：架构优化

| 优先级 | 优化项 | 工作量 | 预期效果 |
|--------|--------|--------|----------|
| P2 | 登出机制 / Token 主动失效 | 中 | 会话管理 |
| P2 | SDK 插件化解耦 | 中 | 便于扩展 |
| P3 | Center 节点高可用 | 大 | 提高稳定性 |

---

## 八、相关文件清单

### 核心文件

| 文件 | 职责 |
|------|------|
| `nodes/web/controller/loginController.go` | Web 层登录注册接口 |
| `internal/token/token.go` | Token 生成与验证 |
| `nodes/center/module/account/actor_account.go` | Center 帐号 Actor |
| `nodes/center/db/dev_account_table.go` | 开发帐号数据表 |
| `nodes/gate/actor_agent.go` | Gate 会话管理 |
| `nodes/gate/route.go` | 消息路由规则 |
| `nodes/web/sdk/dev_sdk.go` | 开发模式 SDK |
| `nodes/web/sdk/quick_sdk.go` | Quick SDK |
| `internal/rpc/center/center.go` | Center RPC 封装 |

### 协议文件

| 文件 | 消息 |
|------|------|
| `internal/protocol/login.proto` | 登录相关协议定义 |
| `internal/pb/login.pb.go` | 生成的 Go 代码 |
| `internal/protocol/rpc.proto` | RPC 协议定义 |

### 配置文件

| 文件 | 用途 |
|------|------|
| `config/cluster.json` | 集群节点配置 |
| `config/data/` | 游戏配表 |

---

## 九、附录

### A. 常量定义

```go
// internal/token/token.go
const (
    hashFormat      = "pid:%d,openid:%s,timestamp:%d"
    tokenExpiredDay = 3
)

// nodes/web/sdk/sdk.go
const (
    DevMode  int32 = 1 // 开发模式
    QuickSDK int32 = 2 // quick sdk
)
```

### B. 错误码

| 错误码 | 定义 | 位置 |
|--------|------|------|
| OK | 成功 | `internal/code/code.go` |
| LoginError | 登录错误 | 同上 |
| AccountAuthFail | 帐号认证失败 | 同上 |
| AccountTokenValidateFail | Token 验证失败 | 同上 |
| AccountBindFail | 帐号绑定失败 | 同上 |
| AccountNameIsExist | 帐号已存在 | 同上 |
| PIDError | pid 错误 | 同上 |
| PlayerDenyLogin | 玩家禁止登录 | 同上 |
| PlayerDuplicateLogin | 玩家重复登录 | 同上 |
