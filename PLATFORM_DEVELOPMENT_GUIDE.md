# 平台开发指南

## 概述

本文档详细说明如何为社交媒体平台项目添加新的服务配置和新的社交媒体平台支持。

## 目录

1. [添加新服务配置](#添加新服务配置)
2. [添加新平台支持](#添加新平台支持)
3. [配置验证和测试](#配置验证和测试)
4. [最佳实践](#最佳实践)

## 添加新服务配置

### 1. 配置文件修改

#### 在 `config.yaml` 中添加新服务
```yaml
# 现有配置...
servers:
  # 现有服务配置...

  # 新服务配置
  newservice:
    youtube:
      client_id: "newservice_youtube_client_id"
      client_secret: "newservice_youtube_client_secret"
      scopes:
        - "https://www.googleapis.com/auth/youtube.upload"
        - "openid"
        - "email"
    x:
      client_id: "newservice_x_client_id"
      client_secret: "newservice_x_client_secret"
      scopes:
        - "tweet.read"
        - "tweet.write"
        - "users.read"
        - "offline.access"
    facebook:
      client_id: "newservice_facebook_app_id"
      client_secret: "newservice_facebook_app_secret"
      scopes:
        - "pages_manage_posts"
        - "pages_read_engagement"
        - "pages_show_list"
        - "pages_read_user_content"
    tiktok:
      client_id: "newservice_tiktok_client_id"
      client_secret: "newservice_tiktok_client_secret"
      scopes:
        - "video.upload"
        - "user.info.basic"
    instagram:
      client_id: "newservice_instagram_client_id"
      client_secret: "newservice_instagram_client_secret"
      scopes:
        - "instagram_content_publish"
        - "pages_read_engagement"
```

#### 环境特定配置
```yaml
# config.dev.yaml - 开发环境
servers:
  newservice:
    youtube:
      client_id: "dev_newservice_youtube_client_id"
      client_secret: "dev_newservice_youtube_client_secret"
      # ... 其他配置

# config.prod.yaml - 生产环境
servers:
  newservice:
    youtube:
      client_id: "${NEWSERVICE_YOUTUBE_CLIENT_ID}"
      client_secret: "${NEWSERVICE_YOUTUBE_CLIENT_SECRET}"
      # ... 其他配置
```

### 2. 环境变量配置

#### 开发环境
```bash
# 可选：为特定服务设置环境变量
export NEWSERVICE_YOUTUBE_CLIENT_ID="dev_newservice_youtube_client_id"
export NEWSERVICE_YOUTUBE_CLIENT_SECRET="dev_newservice_youtube_client_secret"
```

#### 生产环境
```bash
# 生产环境必须使用环境变量
export NEWSERVICE_YOUTUBE_CLIENT_ID="prod_newservice_youtube_client_id"
export NEWSERVICE_YOUTUBE_CLIENT_SECRET="prod_newservice_youtube_client_secret"
export NEWSERVICE_X_CLIENT_ID="prod_newservice_x_client_id"
export NEWSERVICE_X_CLIENT_SECRET="prod_newservice_x_client_secret"
# ... 其他平台配置
```

### 3. 验证新服务配置

```bash
# 验证配置
/opt/homebrew/bin/go run cmd/config/main.go -validate

# 查看特定服务配置
/opt/homebrew/bin/go run cmd/config/main.go -show -format json | jq '.servers.newservice'
```

## 添加新平台支持

### 1. 后端代码修改

#### 步骤1: 添加平台常量
在 `internal/config/constants.go` 中添加新平台常量：

```go
// 新平台OAuth端点
const (
    // 现有平台...

    // 新平台 (例如: LinkedIn)
    LinkedInAuthURL  = "https://www.linkedin.com/oauth/v2/authorization"
    LinkedInTokenURL = "https://www.linkedin.com/oauth/v2/accessToken"
)

// 默认OAuth scopes
var (
    // 现有平台...

    // 新平台默认scopes
    DefaultLinkedInScopes = []string{
        "r_liteprofile",
        "r_emailaddress",
        "w_member_social",
    }
)

// 支持的平台列表
var SupportedProviders = []string{
    "youtube",
    "x",
    "facebook",
    "tiktok",
    "instagram",
    "linkedin", // 添加新平台
}
```

#### 步骤2: 更新配置结构
在 `internal/config/config.go` 中添加新平台配置：

```go
// OAuthConfig holds OAuth provider configurations
type OAuthConfig struct {
    YouTube   ProviderConfig `mapstructure:"youtube"`
    X         ProviderConfig `mapstructure:"x"`
    Facebook  ProviderConfig `mapstructure:"facebook"`
    TikTok    ProviderConfig `mapstructure:"tiktok"`
    Instagram ProviderConfig `mapstructure:"instagram"`
    LinkedIn  ProviderConfig `mapstructure:"linkedin"` // 添加新平台
}

// ServerOAuthConfig holds OAuth configuration for a specific server
type ServerOAuthConfig struct {
    YouTube   ProviderConfig `mapstructure:"youtube"`
    X         ProviderConfig `mapstructure:"x"`
    Facebook  ProviderConfig `mapstructure:"facebook"`
    TikTok    ProviderConfig `mapstructure:"tiktok"`
    Instagram ProviderConfig `mapstructure:"instagram"`
    LinkedIn  ProviderConfig `mapstructure:"linkedin"` // 添加新平台
}
```

#### 步骤3: 更新默认配置
在 `setDefaults()` 函数中添加新平台默认值：

```go
func setDefaults() {
    // 现有配置...

    // 新平台默认配置
    viper.SetDefault("oauth.linkedin.scopes", DefaultLinkedInScopes)
}
```

#### 步骤4: 更新OAuth配置方法
在 `GetOAuthConfig()` 方法中添加新平台支持：

```go
func (c *Config) GetOAuthConfig(provider string) (*oauth2.Config, error) {
    baseURL := c.Server.BaseURL

    switch provider {
    // 现有平台...

    case "linkedin":
        return &oauth2.Config{
            ClientID:     c.OAuth.LinkedIn.ClientID,
            ClientSecret: c.OAuth.LinkedIn.ClientSecret,
            Scopes:       c.OAuth.LinkedIn.Scopes,
            Endpoint: oauth2.Endpoint{
                AuthURL:  LinkedInAuthURL,
                TokenURL: LinkedInTokenURL,
            },
            RedirectURL: fmt.Sprintf("%s/auth/linkedin/callback", baseURL),
        }, nil

    default:
        return nil, fmt.Errorf("unknown provider: %s", provider)
    }
}
```

同样需要更新 `GetOAuthConfigWithRedirect()` 和 `GetServerOAuthConfig()` 方法。

#### 步骤5: 创建平台处理器
创建 `internal/platforms/linkedin.go`：

```go
package platforms

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"

    "golang.org/x/oauth2"
)

// LinkedInPlatform handles LinkedIn-specific operations
type LinkedInPlatform struct {
    config *oauth2.Config
}

// NewLinkedInPlatform creates a new LinkedIn platform handler
func NewLinkedInPlatform(config *oauth2.Config) *LinkedInPlatform {
    return &LinkedInPlatform{
        config: config,
    }
}

// GetUserInfo retrieves user information from LinkedIn
func (p *LinkedInPlatform) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
    client := p.config.Client(ctx, token)

    // LinkedIn API endpoint for user info
    resp, err := client.Get("https://api.linkedin.com/v2/people/~")
    if err != nil {
        return nil, fmt.Errorf("failed to get user info: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("LinkedIn API returned status %d", resp.StatusCode)
    }

    var userInfo UserInfo
    if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
        return nil, fmt.Errorf("failed to decode user info: %w", err)
    }

    return &userInfo, nil
}

// ShareContent shares content to LinkedIn
func (p *LinkedInPlatform) ShareContent(ctx context.Context, token *oauth2.Token, content *ShareContent) (*ShareResult, error) {
    client := p.config.Client(ctx, token)

    // LinkedIn分享API调用
    shareData := map[string]interface{}{
        "author": "urn:li:person:" + content.UserID,
        "lifecycleState": "PUBLISHED",
        "specificContent": map[string]interface{}{
            "com.linkedin.ugc.ShareContent": map[string]interface{}{
                "shareCommentary": map[string]interface{}{
                    "text": content.Text,
                },
                "shareMediaCategory": "NONE",
            },
        },
        "visibility": map[string]interface{}{
            "com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC",
        },
    }

    jsonData, err := json.Marshal(shareData)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal share data: %w", err)
    }

    resp, err := client.Post("https://api.linkedin.com/v2/ugcPosts", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("failed to share content: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("LinkedIn API returned status %d", resp.StatusCode)
    }

    var result ShareResult
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode share result: %w", err)
    }

    return &result, nil
}
```

#### 步骤6: 注册新平台
在 `internal/platforms/registry.go` 中注册新平台：

```go
// RegisterPlatforms registers all supported platforms
func (r *Registry) RegisterPlatforms() {
    r.platforms["youtube"] = NewYouTubePlatform
    r.platforms["x"] = NewXPlatform
    r.platforms["facebook"] = NewFacebookPlatform
    r.platforms["tiktok"] = NewTikTokPlatform
    r.platforms["instagram"] = NewInstagramPlatform
    r.platforms["linkedin"] = NewLinkedInPlatform // 注册新平台
}
```

### 2. 前端代码修改

#### 步骤1: 更新授权页面
在 `static/auth.html` 中添加新平台tab：

```html
<div class="tabs">
    <button class="tab active" onclick="switchTab('youtube')">YouTube</button>
    <button class="tab" onclick="switchTab('x')">X (Twitter)</button>
    <button class="tab" onclick="switchTab('facebook')">Facebook</button>
    <button class="tab" onclick="switchTab('tiktok')">TikTok</button>
    <button class="tab" onclick="switchTab('instagram')">Instagram</button>
    <button class="tab" onclick="switchTab('linkedin')">LinkedIn</button> <!-- 添加新平台 -->
</div>

<!-- LinkedIn 授权 -->
<div id="linkedin-tab" class="tab-content">
    <div class="section">
        <h3>LinkedIn 授权配置</h3>
        <div class="input-group">
            <label for="linkedin-userId">用户ID:</label>
            <input type="text" id="linkedin-userId" readonly placeholder="自动生成的UUID">
        </div>
        <div class="input-group">
            <label for="linkedin-serverName">服务器名称:</label>
            <input type="text" id="linkedin-serverName" value="testapp" placeholder="输入服务器名称">
        </div>
        <div class="input-group">
            <label for="linkedin-redirectUri">重定向URI:</label>
            <input type="text" id="linkedin-redirectUri" value="" placeholder="留空使用默认回调页面">
        </div>
        <button class="btn btn-primary" onclick="startAuth('linkedin')">开始LinkedIn授权</button>
        <button class="btn btn-secondary" onclick="clearResults('linkedin')">清空结果</button>
    </div>

    <div id="linkedin-results" class="section hidden">
        <h3>授权结果</h3>
        <div id="linkedin-output"></div>
    </div>
</div>
```

#### 步骤2: 更新分享页面
在 `static/share.html` 中添加新平台支持：

```html
<div class="tabs">
    <button class="tab active" onclick="switchTab('youtube')">YouTube</button>
    <button class="tab" onclick="switchTab('x')">X (Twitter)</button>
    <button class="tab" onclick="switchTab('facebook')">Facebook</button>
    <button class="tab" onclick="switchTab('tiktok')">TikTok</button>
    <button class="tab" onclick="switchTab('instagram')">Instagram</button>
    <button class="tab" onclick="switchTab('linkedin')">LinkedIn</button> <!-- 添加新平台 -->
</div>

<!-- LinkedIn 分享 -->
<div id="linkedin-tab" class="tab-content">
    <div class="section">
        <h3>LinkedIn 分享配置</h3>
        <div class="input-group">
            <label for="linkedin-userId">用户ID:</label>
            <input type="text" id="linkedin-userId" value="debug-user-123" placeholder="输入用户ID">
        </div>
        <div class="input-group">
            <label for="linkedin-serverName">服务器名称:</label>
            <select id="linkedin-serverName">
                <option value="default">default (使用默认配置)</option>
                <option value="testapp">testapp (使用服务器配置)</option>
            </select>
        </div>
    </div>

    <div class="section">
        <h3>📝 分享内容</h3>
        <div class="input-group">
            <label for="linkedin-content">内容:</label>
            <textarea id="linkedin-content" placeholder="输入要分享的内容" maxlength="1300">🌤️ 美丽的天空和云朵！今天天气真好，风向也很不错。分享一张美丽的风景照片给大家！ #风向 #天气 #自然 #风景</textarea>
        </div>
        <div class="input-group">
            <label for="linkedin-mediaUrl">媒体URL (可选):</label>
            <input type="url" id="linkedin-mediaUrl" value="https://d20r62ijagu47n.cloudfront.net/1755482887937-ymoolaca.png" placeholder="https://example.com/image.jpg">
        </div>
        <div class="input-group">
            <label for="linkedin-tags">标签 (可选，用逗号分隔):</label>
            <input type="text" id="linkedin-tags" value="风向,天气,自然,风景,美丽,摄影,天空,云朵" placeholder="tag1,tag2,tag3">
        </div>
    </div>

    <div class="section">
        <h3>🎯 操作</h3>
        <button class="btn btn-primary" onclick="shareContent('linkedin')">📤 分享到LinkedIn</button>
        <button class="btn btn-success" onclick="getUserInfo('linkedin')">👤 获取用户信息</button>
        <button class="btn btn-secondary" onclick="clearResults('linkedin')">🗑️ 清空结果</button>
    </div>

    <div id="linkedin-results" class="section hidden">
        <h3>📊 分享结果</h3>
        <div id="linkedin-output"></div>
    </div>
</div>
```

#### 步骤3: 更新JavaScript代码
在页面加载时为新平台生成UUID：

```javascript
// 页面加载时的初始化
window.onload = function () {
    // 为所有平台生成UUID
    const platforms = ['youtube', 'x', 'facebook', 'tiktok', 'instagram', 'linkedin']; // 添加新平台
    platforms.forEach(platform => {
        const userId = generateUUID();
        document.getElementById(`${platform}-userId`).value = userId;
    });
};
```

### 3. 配置文件更新

#### 更新默认配置
在 `config.yaml` 中添加新平台默认配置：

```yaml
# 默认OAuth配置
oauth:
  # 现有平台...

  linkedin:
    client_id: "your_linkedin_client_id"
    client_secret: "your_linkedin_client_secret"
    scopes:
      - "r_liteprofile"
      - "r_emailaddress"
      - "w_member_social"

platform:
  supported_providers:
    - "youtube"
    - "x"
    - "facebook"
    - "tiktok"
    - "instagram"
    - "linkedin" # 添加新平台
```

#### 更新多服务配置
为每个服务添加新平台配置：

```yaml
servers:
  myblog:
    # 现有平台...

    linkedin:
      client_id: "myblog_linkedin_client_id"
      client_secret: "myblog_linkedin_client_secret"
      scopes:
        - "r_liteprofile"
        - "r_emailaddress"
        - "w_member_social"

  marketing:
    # 现有平台...

    linkedin:
      client_id: "marketing_linkedin_client_id"
      client_secret: "marketing_linkedin_client_secret"
      scopes:
        - "r_liteprofile"
        - "r_emailaddress"
        - "w_member_social"
        - "w_organization_social"
```

## 配置验证和测试

### 1. 配置验证

```bash
# 验证配置完整性
/opt/homebrew/bin/go run cmd/config/main.go -validate

# 验证特定环境配置
/opt/homebrew/bin/go run cmd/config/main.go -validate -env production

# 查看新平台配置
/opt/homebrew/bin/go run cmd/config/main.go -show -format json | jq '.oauth.linkedin'
```

### 2. 编译测试

```bash
# 编译项目
/opt/homebrew/bin/go build -o tmp/main main.go

# 运行测试
/opt/homebrew/bin/go test ./...
```

### 3. 功能测试

#### 授权测试
1. 访问 `http://localhost:8080/static/auth.html`
2. 选择新平台tab
3. 填写配置信息
4. 点击"开始授权"
5. 验证OAuth流程

#### 分享测试
1. 访问 `http://localhost:8080/static/share.html`
2. 选择新平台tab
3. 填写分享内容
4. 点击"分享内容"
5. 验证分享结果

## 最佳实践

### 1. 平台开发规范

#### 命名规范
- 平台名称使用小写字母
- 常量使用大写下划线分隔
- 文件名使用小写字母和下划线

#### 代码组织
- 每个平台一个独立的文件
- 统一的接口实现
- 完善的错误处理

#### 配置管理
- 使用常量定义OAuth端点
- 提供默认scopes配置
- 支持环境变量覆盖

### 2. 测试规范

#### 单元测试
```go
func TestLinkedInPlatform_GetUserInfo(t *testing.T) {
    // 测试用户信息获取
}

func TestLinkedInPlatform_ShareContent(t *testing.T) {
    // 测试内容分享
}
```

#### 集成测试
```go
func TestLinkedInOAuthFlow(t *testing.T) {
    // 测试完整OAuth流程
}
```

### 3. 文档规范

#### API文档
- 更新Swagger文档
- 添加新平台API说明
- 提供示例代码

#### 用户文档
- 更新使用说明
- 添加新平台配置指南
- 提供故障排除指南

### 4. 安全规范

#### OAuth配置
- 使用HTTPS重定向URI
- 验证state参数
- 安全的token存储

#### API安全
- 验证用户权限
- 限制API调用频率
- 记录操作日志

## 故障排除

### 常见问题

#### 1. 配置验证失败
```bash
# 检查配置文件格式
/opt/homebrew/bin/go run cmd/config/main.go -validate

# 检查环境变量
env | grep LINKEDIN
```

#### 2. OAuth授权失败
- 检查Client ID和Secret是否正确
- 验证重定向URI配置
- 确认scopes权限设置

#### 3. API调用失败
- 检查token是否有效
- 验证API端点URL
- 确认请求格式正确

### 调试技巧

#### 1. 启用详细日志
```bash
export GIN_MODE=debug
/opt/homebrew/bin/go run main.go
```

#### 2. 配置调试
```bash
# 查看完整配置
/opt/homebrew/bin/go run cmd/config/main.go -show -format json
```

#### 3. 网络调试
```bash
# 使用curl测试API
curl -H "Authorization: Bearer YOUR_TOKEN" \
     https://api.linkedin.com/v2/people/~
```

## 总结

添加新平台支持需要：

1. **后端修改**: 常量、配置结构、OAuth方法、平台处理器
2. **前端修改**: HTML页面、JavaScript代码
3. **配置更新**: 默认配置、多服务配置
4. **测试验证**: 配置验证、编译测试、功能测试

遵循最佳实践，确保代码质量和安全性，提供完善的文档和测试覆盖。
