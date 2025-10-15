# 多项目配置快速入门

## 🚀 快速开始

### 1. 查看当前配置

```bash
# 列出所有已配置的服务器
make config-list

# 验证配置文件
make config-validate
```

### 2. 添加新项目

```bash
# 添加一个新项目（例如：myblog）
make config-add SERVER=myblog

# 查看项目配置
make config-show SERVER=myblog
```

### 3. 编辑配置文件

编辑 `config.yaml` 文件，为你的项目填入真实的OAuth凭据：

```yaml
servers:
  myblog:
    youtube:
      client_id: "你的YouTube客户端ID"
      client_secret: "你的YouTube客户端密钥"
      scopes:
        - "https://www.googleapis.com/auth/youtube.upload"
        - "openid"
        - "email"
    x:
      client_id: "你的X客户端ID"
      client_secret: "你的X客户端密钥"
      scopes:
        - "tweet.read"
        - "tweet.write"
        - "users.read"
        - "offline.access"
```

### 4. 测试配置

```bash
# 启动服务
make run

# 访问测试页面
open https://test-pubproject.wondera.io/static/test.html
```

## 📋 项目配置示例

### 个人博客项目
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

### 企业营销工具
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
    facebook:
      client_id: "marketing_facebook_app_id"
      client_secret: "marketing_facebook_app_secret"
      scopes:
        - "pages_manage_posts"
        - "pages_read_engagement"
        - "pages_show_list"
        - "pages_read_user_content"
        - "ads_management"
```

## 🔧 常用命令

```bash
# 配置管理
make config-list                    # 列出所有服务器
make config-add SERVER=project1     # 添加新项目
make config-remove SERVER=project1  # 删除项目
make config-show SERVER=project1    # 显示项目配置
make config-validate                # 验证配置文件

# 开发命令
make run                            # 运行服务
make build                          # 构建应用
make swagger                        # 生成API文档
make test                           # 运行测试
```

## 📱 前端调用示例

### JavaScript调用
```javascript
// 开始OAuth授权
const startAuth = async (provider, serverName, userId, redirectUri) => {
  const response = await fetch('/auth/start', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      provider: provider,
      user_id: userId,
      server_name: serverName,
      redirect_uri: redirectUri
    })
  });

  const data = await response.json();
  if (response.ok) {
    window.open(data.auth_url, '_blank');
  }
};

// 使用示例
startAuth('x', 'myblog', 'user123', 'https://myblog.com/callback');
```

### cURL调用
```bash
# 开始授权
curl -X POST https://test-pubproject.wondera.io/auth/start \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "x",
    "user_id": "user123",
    "server_name": "myblog",
    "redirect_uri": "https://myblog.com/callback"
  }'

# 处理回调
curl -X POST https://test-pubproject.wondera.io/auth/callback \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "x",
    "server_name": "myblog",
    "code": "authorization_code",
    "state": "encoded_state"
  }'
```

## 🎯 最佳实践

### 1. 项目命名
- 使用有意义的名称：`myblog`, `marketing`, `cms`
- 避免特殊字符和空格
- 保持名称简短但描述性

### 2. 权限配置
- 只配置项目需要的权限
- 定期审查和更新权限范围
- 为不同环境使用不同的OAuth应用

### 3. 安全考虑
- 不要在代码中硬编码凭据
- 使用环境变量或配置文件
- 定期轮换客户端密钥

### 4. 测试流程
1. 使用测试页面验证配置
2. 检查OAuth应用的回调URL设置
3. 验证权限范围是否正确
4. 测试完整的OAuth流程

## 🐛 故障排除

### 常见问题

1. **服务器名称不存在**
   ```
   错误: server_name "unknown" not found
   解决: 在配置文件中添加对应的服务器配置
   ```

2. **权限不足**
   ```
   错误: insufficient scopes
   解决: 在配置中添加所需的权限范围
   ```

3. **回调URL不匹配**
   ```
   错误: redirect_uri_mismatch
   解决: 确保OAuth应用配置中的回调URL与请求中的一致
   ```

### 调试步骤

1. 检查配置文件格式：`make config-validate`
2. 查看服务器配置：`make config-show SERVER=your_server`
3. 使用测试页面验证：访问测试页面进行OAuth流程测试
4. 查看服务日志：检查详细的错误信息

## 📚 更多信息

- [多项目配置详细说明](./multi-project-config.md)
- [API文档](../docs/swagger.json)
- [测试页面使用说明](../static/README.md)
