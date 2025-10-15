# 配置管理说明

## 概述

本项目采用**配置文件管理**方式处理多服务配置，相比数据库管理具有以下优势：

- ✅ **性能更好**: 启动时加载，内存访问快
- ✅ **复杂度低**: 简单YAML配置，易于维护
- ✅ **版本控制**: 可以Git管理配置变更
- ✅ **备份简单**: 文件备份比数据库备份简单
- ✅ **多环境支持**: 不同环境使用不同配置文件

## 配置文件结构

### 环境特定配置
```
config.yaml          # 默认配置
config.dev.yaml      # 开发环境配置
config.staging.yaml  # 测试环境配置
config.prod.yaml     # 生产环境配置
```

### 配置优先级
1. **环境变量** (最高优先级)
2. **环境特定配置文件** (如 config.dev.yaml)
3. **默认配置文件** (config.yaml)
4. **代码默认值** (最低优先级)

## 环境变量支持

### 服务器配置
```bash
export SERVER_PORT=8080
export SERVER_BASE_URL=http://localhost:8080
```

### Redis配置
```bash
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=your_password
export REDIS_DB=0
```

### 环境设置
```bash
export ENVIRONMENT=development  # development, staging, production
export GIN_MODE=release
```

## 多服务配置

### 配置文件示例
```yaml
# 默认OAuth配置
oauth:
  youtube:
    client_id: "default_youtube_client_id"
    client_secret: "default_youtube_client_secret"
    scopes:
      - "https://www.googleapis.com/auth/youtube.upload"

# 多项目配置
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

  marketing:
    youtube:
      client_id: "marketing_youtube_client_id"
      client_secret: "marketing_youtube_client_secret"
      scopes:
        - "https://www.googleapis.com/auth/youtube.upload"
        - "https://www.googleapis.com/auth/youtube.readonly"
```

## 配置管理工具

### 验证配置
```bash
# 验证当前配置
go run cmd/config/main.go -validate

# 验证特定环境配置
go run cmd/config/main.go -validate -env production
```

### 查看配置
```bash
# 查看当前配置
go run cmd/config/main.go -show

# 以JSON格式查看
go run cmd/config/main.go -show -format json
```

## 代码改进

### 1. 消除硬编码

#### 之前 (硬编码)
```go
// 硬编码的OAuth端点
endpoint: oauth2.Endpoint{
    AuthURL:  "https://twitter.com/i/oauth2/authorize",
    TokenURL: "https://api.twitter.com/2/oauth2/token",
}
```

#### 现在 (使用常量)
```go
// 使用常量定义
endpoint: oauth2.Endpoint{
    AuthURL:  XAuthURL,
    TokenURL: XTokenURL,
}
```

### 2. 动态URL生成

#### 之前 (硬编码URL)
```html
<input value="https://test-pubproject.wondera.io/static/callback.html">
```

#### 现在 (动态生成)
```javascript
// 动态获取当前域名
const redirectUri = `${window.location.origin}/static/callback.html`;
```

### 3. 环境感知配置

#### 之前 (固定配置)
```go
viper.SetDefault("server.base_url", "http://localhost:8080")
```

#### 现在 (环境感知)
```go
// 根据环境设置默认值
if IsDevelopment() {
    viper.SetDefault("server.base_url", "http://localhost:8080")
} else {
    viper.SetDefault("server.base_url", "https://api.example.com")
}
```

## 部署建议

### 开发环境
```bash
export ENVIRONMENT=development
go run main.go
```

### 生产环境
```bash
export ENVIRONMENT=production
export SERVER_BASE_URL=https://api.yourdomain.com
export REDIS_ADDR=redis.yourdomain.com:6379
export REDIS_PASSWORD=your_secure_password
./main
```

### Docker部署
```dockerfile
# 使用环境变量覆盖配置
ENV ENVIRONMENT=production
ENV SERVER_BASE_URL=https://api.yourdomain.com
ENV REDIS_ADDR=redis:6379
```

## 安全建议

### 1. 敏感信息管理
- 生产环境的Client Secret使用环境变量
- 配置文件不要提交到版本控制
- 使用配置管理工具 (如 Kubernetes Secrets)

### 2. 配置验证
- 启动时验证配置完整性
- 检查生产环境配置警告
- 定期审查配置安全性

### 3. 访问控制
- 限制配置文件访问权限
- 使用不同的Redis数据库隔离
- 实施API访问限制

## 故障排除

### 常见问题

1. **配置加载失败**
   ```bash
   # 检查配置文件是否存在
   ls -la config*.yaml

   # 验证配置文件格式
   go run cmd/config/main.go -validate
   ```

2. **环境变量不生效**
   ```bash
   # 检查环境变量
   env | grep SERVER

   # 重新加载环境变量
   source ~/.bashrc
   ```

3. **多服务配置不工作**
   ```bash
   # 检查服务器配置
   go run cmd/config/main.go -show -format json | jq '.servers'
   ```

### 调试技巧

1. **启用详细日志**
   ```bash
   export GIN_MODE=debug
   go run main.go
   ```

2. **配置验证**
   ```bash
   go run cmd/config/main.go -validate -env production
   ```

3. **配置对比**
   ```bash
   # 对比不同环境配置
   diff config.dev.yaml config.prod.yaml
   ```

## 最佳实践

1. **配置分离**: 不同环境使用不同配置文件
2. **环境变量**: 敏感信息使用环境变量
3. **配置验证**: 启动时验证配置完整性
4. **版本控制**: 配置文件纳入版本控制
5. **文档更新**: 配置变更时更新文档
6. **测试覆盖**: 配置相关代码要有测试
7. **监控告警**: 配置错误要有监控告警

## 总结

通过配置文件管理多服务配置，我们实现了：

- ✅ **消除硬编码**: 所有URL和端点使用常量
- ✅ **环境感知**: 根据环境自动选择配置
- ✅ **动态生成**: 前端URL动态生成
- ✅ **配置验证**: 启动时验证配置完整性
- ✅ **管理工具**: 提供配置管理命令行工具
- ✅ **多环境支持**: 支持开发、测试、生产环境

这种方式比数据库管理更简单、更高效，适合OAuth配置相对稳定的场景。
