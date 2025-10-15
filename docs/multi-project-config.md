# 多项目配置说明

## 概述

Social项目支持为多个不同的项目提供独立的OAuth配置。每个项目可以有自己独立的社交媒体平台凭据和权限范围。

## 配置结构

### 1. 项目标识

每个项目通过 `server_name` 来标识，这个名称在配置文件的 `servers` 段中定义：

```yaml
servers:
  myblog:        # 项目1: 我的博客应用
    youtube:
      client_id: "myblog_youtube_client_id"
      # ...
  marketing:     # 项目2: 企业营销工具
    youtube:
      client_id: "marketing_youtube_client_id"
      # ...
  wondera:       # 项目3: 测试应用
    youtube:
      client_id: "wondera_youtube_client_id"
      # ...
```

### 2. 项目命名规范

建议使用以下命名规范：

- **项目类型 + 用途**：如 `myblog`, `marketing`, `cms`
- **公司 + 产品**：如 `acme_blog`, `acme_shop`
- **环境 + 项目**：如 `prod_blog`, `staging_blog`
- **简短描述**：如 `wondera`, `demo`, `personal`

### 3. 配置示例

#### 项目1: 个人博客 (myblog)
```yaml
servers:
  myblog:
    youtube:
      client_id: "myblog_youtube_client_id"
      client_secret: "myblog_youtube_client_secret"
      scopes:
        - "https://www.googleapis.com/auth/youtube.upload"
        - "openid"
        - "email"
    x:
      client_id: "myblog_x_client_id"
      client_secret: "myblog_x_client_secret"
      scopes:
        - "tweet.read"
        - "tweet.write"
        - "users.read"
        - "offline.access"
```

#### 项目2: 企业营销工具 (marketing)
```yaml
servers:
  marketing:
    youtube:
      client_id: "marketing_youtube_client_id"
      client_secret: "marketing_youtube_client_secret"
      scopes:
        - "https://www.googleapis.com/auth/youtube.upload"
        - "https://www.googleapis.com/auth/youtube.readonly"
        - "openid"
        - "email"
    x:
      client_id: "marketing_x_client_id"
      client_secret: "marketing_x_client_secret"
      scopes:
        - "tweet.read"
        - "tweet.write"
        - "users.read"
        - "offline.access"
        - "follows.read"
        - "follows.write"
```

## 使用方式

### 1. 前端调用

#### 开始OAuth授权
```javascript
const response = await fetch('/auth/start', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    provider: 'x',
    user_id: 'user123',
    server_name: 'myblog',  // 指定项目
    redirect_uri: 'https://myblog.com/auth/callback'
  })
});
```

#### 处理OAuth回调
```javascript
const response = await fetch('/auth/callback', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    provider: 'x',
    server_name: 'myblog',  // 必须与授权时一致
    code: 'authorization_code',
    state: 'encoded_state'
  })
});
```

#### 分享内容
```javascript
const response = await fetch('/share', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    provider: 'x',
    user_id: 'user123',
    server_name: 'myblog',  // 指定项目
    content: 'Hello World!'
  })
});
```

### 2. 配置回退机制

如果指定的 `server_name` 在配置中不存在，系统会自动回退到默认配置：

```yaml
# 默认配置（在 oauth 段中）
oauth:
  youtube:
    client_id: "default_youtube_client_id"
    client_secret: "default_youtube_client_secret"
    # ...
```

## 最佳实践

### 1. 项目隔离

- 每个项目使用独立的OAuth应用
- 不同的权限范围（scopes）
- 独立的回调URL

### 2. 命名约定

- 使用有意义的项目名称
- 避免使用特殊字符和空格
- 保持名称简短但描述性

### 3. 权限管理

- 根据项目需求配置最小必要权限
- 定期审查和更新权限范围
- 为不同环境使用不同的OAuth应用

### 4. 安全考虑

- 不要在代码中硬编码凭据
- 使用环境变量或配置文件
- 定期轮换客户端密钥

## 配置验证

系统会自动验证：

1. **服务器名称存在性**：检查 `server_name` 是否在配置中定义
2. **权限范围有效性**：验证请求的权限是否在配置的范围内
3. **回调URL一致性**：确保回调URL与OAuth应用配置一致

## 故障排除

### 常见问题

1. **服务器名称不存在**
   - 错误：`server_name "unknown" not found`
   - 解决：在配置文件中添加对应的服务器配置

2. **权限不足**
   - 错误：`insufficient scopes`
   - 解决：在配置中添加所需的权限范围

3. **回调URL不匹配**
   - 错误：`redirect_uri_mismatch`
   - 解决：确保OAuth应用配置中的回调URL与请求中的一致

### 调试技巧

1. 检查配置文件格式是否正确
2. 验证OAuth应用配置
3. 查看日志中的详细错误信息
4. 使用测试页面验证配置

## 示例场景

### 场景1: 多环境部署
```yaml
servers:
  prod_blog:     # 生产环境博客
  staging_blog:  # 测试环境博客
  dev_blog:      # 开发环境博客
```

### 场景2: 多产品线
```yaml
servers:
  main_website:  # 主网站
  mobile_app:    # 移动应用
  admin_panel:   # 管理面板
```

### 场景3: 多客户
```yaml
servers:
  client_a:      # 客户A
  client_b:      # 客户B
  internal:      # 内部工具
```
