# 社交媒体平台项目说明

## 项目概述

这是一个多平台社交媒体授权和内容分享服务，支持YouTube、X (Twitter)、Facebook、TikTok、Instagram等主流社交媒体平台的OAuth授权和内容发布功能。

## 核心功能

### 🔐 OAuth授权管理
- **多平台支持**: YouTube、X、Facebook、TikTok、Instagram
- **OAuth 2.0流程**: 完整的授权码流程，支持PKCE
- **Token管理**: 自动token刷新和过期处理
- **多服务配置**: 支持多个项目使用不同的OAuth配置

### 📤 内容分享
- **统一接口**: 标准化的内容分享API
- **多格式支持**: 文本、图片、视频内容分享
- **平台特性**: 根据各平台特性调整内容格式和限制
- **批量操作**: 支持同时分享到多个平台

### 🛠️ 管理功能
- **配置管理**: 灵活的配置文件和环境变量支持
- **监控统计**: 内容分享统计和用户数据获取
- **健康检查**: 服务状态监控和Redis连接检查
- **API文档**: 完整的Swagger API文档

## 技术架构

### 后端技术栈
- **语言**: Go 1.21+
- **框架**: Gin Web Framework
- **存储**: Redis (PKCE验证码和token存储)
- **配置**: Viper (支持YAML配置和环境变量)
- **文档**: Swagger/OpenAPI 3.0
- **日志**: 结构化日志记录

### 前端技术栈
- **HTML5**: 现代化响应式界面
- **CSS3**: 渐变背景和动画效果
- **JavaScript**: 原生ES6+，无框架依赖
- **设计**: 移动端友好的Tab式布局

### 部署架构
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端界面      │    │   Go后端服务    │    │   Redis存储     │
│   (静态文件)    │◄──►│   (Gin框架)     │◄──►│   (Token缓存)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   用户浏览器    │    │   社交媒体API   │    │   配置管理      │
│   (OAuth回调)   │    │   (OAuth提供方) │    │   (多环境)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 项目结构

```
social/
├── cmd/                          # 命令行工具
│   └── config/                   # 配置管理工具
│       └── main.go
├── internal/                     # 内部包
│   ├── config/                   # 配置管理
│   │   ├── config.go            # 主配置结构
│   │   ├── constants.go         # OAuth端点常量
│   │   ├── env.go              # 环境变量管理
│   │   ├── utils.go            # 配置工具函数
│   │   └── validator.go        # 配置验证器
│   ├── handlers/                # HTTP处理器
│   │   ├── auth.go             # OAuth授权处理
│   │   ├── share.go            # 内容分享处理
│   │   └── health.go           # 健康检查
│   ├── oauth/                   # OAuth核心逻辑
│   │   └── oauth.go
│   ├── platforms/               # 平台特定实现
│   │   ├── youtube.go          # YouTube平台
│   │   ├── x.go                # X (Twitter)平台
│   │   ├── facebook.go         # Facebook平台
│   │   ├── tiktok.go           # TikTok平台
│   │   ├── instagram.go        # Instagram平台
│   │   └── registry.go         # 平台注册器
│   ├── storage/                 # 存储接口
│   │   ├── interface.go        # 存储接口定义
│   │   └── redis.go            # Redis实现
│   └── types/                   # 数据类型定义
│       └── types.go
├── pkg/                         # 公共包
│   ├── context/                 # 上下文管理
│   ├── errors/                  # 错误处理
│   ├── logger/                  # 日志记录
│   ├── response/                # 响应格式化
│   └── validator/               # 数据验证
├── static/                      # 静态文件
│   ├── auth.html               # 授权页面
│   ├── callback.html           # 回调处理页面
│   ├── share.html              # 内容分享页面
│   ├── test.html               # 测试页面
│   └── README.md               # 静态文件说明
├── docs/                        # API文档
│   ├── docs.go                 # Swagger文档
│   ├── swagger.json            # OpenAPI规范
│   └── swagger.yaml
├── config.yaml                  # 主配置文件
├── config.dev.yaml             # 开发环境配置
├── config.yaml.example         # 配置示例
├── main.go                     # 主程序入口
├── go.mod                      # Go模块定义
├── go.sum                      # 依赖校验和
├── Dockerfile                  # Docker镜像构建
├── docker-compose.yml          # Docker编排
├── Makefile                    # 构建脚本
└── README.md                   # 项目说明
```

## 核心组件

### 1. 配置管理 (`internal/config/`)

#### 多环境支持
- **开发环境**: `config.dev.yaml`
- **测试环境**: `config.staging.yaml`
- **生产环境**: `config.prod.yaml`
- **环境变量**: 支持环境变量覆盖配置

#### 多服务配置
```yaml
servers:
  myblog:                    # 博客应用
    youtube:
      client_id: "blog_youtube_id"
      client_secret: "blog_youtube_secret"
  marketing:                 # 营销工具
    youtube:
      client_id: "marketing_youtube_id"
      client_secret: "marketing_youtube_secret"
```

#### 配置验证
```bash
# 验证配置
go run cmd/config/main.go -validate

# 查看配置
go run cmd/config/main.go -show
```

### 2. OAuth授权 (`internal/oauth/`)

#### 支持的OAuth流程
- **授权码流程**: 标准OAuth 2.0授权码流程
- **PKCE支持**: 增强安全性的PKCE扩展
- **Token刷新**: 自动处理token过期和刷新
- **状态管理**: 安全的state参数验证

#### 平台支持
| 平台 | 授权URL | Token URL | 特殊要求 |
|------|---------|-----------|----------|
| YouTube | Google OAuth | Google OAuth | 需要Google Cloud项目 |
| X (Twitter) | X OAuth 2.0 | X OAuth 2.0 | 需要X Developer账号 |
| Facebook | Facebook OAuth | Facebook OAuth | 需要Facebook应用 |
| TikTok | TikTok OAuth | TikTok OAuth | 需要TikTok开发者账号 |
| Instagram | Facebook OAuth | Facebook OAuth | 通过Facebook应用 |

### 3. 平台处理器 (`internal/platforms/`)

#### 统一接口
```go
type Platform interface {
    GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error)
    ShareContent(ctx context.Context, token *oauth2.Token, content *ShareContent) (*ShareResult, error)
}
```

#### 平台特性
- **YouTube**: 视频上传，支持大文件
- **X**: 280字符限制，支持媒体附件
- **Facebook**: 页面管理，支持多种内容类型
- **TikTok**: 短视频分享，支持创意工具
- **Instagram**: 图片分享，支持故事和帖子

### 4. 存储层 (`internal/storage/`)

#### Redis存储
- **PKCE验证码**: 临时存储OAuth验证码
- **Token缓存**: 缓存OAuth token，减少API调用
- **会话管理**: 用户会话状态管理
- **过期处理**: 自动清理过期数据

#### 存储接口
```go
type Storage interface {
    StorePKCEVerifier(key string, verifier string, ttl time.Duration) error
    GetPKCEVerifier(key string) (string, error)
    StoreToken(key string, token *oauth2.Token) error
    GetToken(key string) (*oauth2.Token, error)
    Close() error
}
```

## API接口

### 授权接口

#### 开始授权
```http
POST /auth/start
Content-Type: application/json

{
    "provider": "youtube",
    "user_id": "user123",
    "server_name": "myblog",
    "redirect_uri": "https://myapp.com/callback"
}
```

#### 处理回调
```http
POST /auth/callback
Content-Type: application/json

{
    "provider": "youtube",
    "server_name": "myblog",
    "code": "authorization_code",
    "state": "state_parameter",
    "redirect_uri": "https://myapp.com/callback"
}
```

#### 刷新Token
```http
POST /auth/refresh
Content-Type: application/json

{
    "provider": "youtube",
    "user_id": "user123",
    "server_name": "myblog"
}
```

### 分享接口

#### 分享内容
```http
POST /api/share
Content-Type: application/json

{
    "provider": "youtube",
    "user_id": "user123",
    "server_name": "myblog",
    "content": "分享内容",
    "media_url": "https://example.com/video.mp4",
    "tags": ["tag1", "tag2"]
}
```

#### 获取统计
```http
POST /api/stats
Content-Type: application/json

{
    "provider": "youtube",
    "user_id": "user123",
    "server_name": "myblog",
    "post_id": "post123"
}
```

### RESTful接口

#### 创建帖子
```http
POST /api/{platform}/posts
```

#### 获取帖子统计
```http
GET /api/{platform}/posts/{post_id}/stats
```

#### 获取用户统计
```http
GET /api/{platform}/users/{user_id}/stats
```

## 前端界面

### 1. 授权页面 (`static/auth.html`)
- **Tab式布局**: 无图标设计，简洁专业
- **平台选择**: 支持所有配置的平台
- **配置管理**: 用户ID、服务器名称、重定向URI
- **一键授权**: 自动打开OAuth授权页面

### 2. 回调页面 (`static/callback.html`)
- **自动检测**: 自动解析URL中的OAuth参数
- **手动处理**: 支持手动输入回调参数
- **状态反馈**: 详细的处理结果和时间戳信息
- **错误处理**: 完善的错误提示和解决建议

### 3. 分享页面 (`static/share.html`)
- **Tab式布局**: 无图标设计，专注于功能
- **内容编辑**: 支持文本、媒体URL、标签等
- **扩展功能**: 预留用户信息获取等功能接口
- **平台特性**: 根据平台特性调整内容长度限制

### 4. 测试页面 (`static/test.html`)
- **流程引导**: 清晰的测试流程说明
- **快速访问**: 一键访问各个功能页面
- **配置说明**: 详细的配置和使用说明
- **状态显示**: 当前测试状态和服务器信息

## 部署指南

### 开发环境

#### 1. 环境准备
```bash
# 安装Go 1.21+
go version

# 安装Redis
redis-server --version

# 克隆项目
git clone <repository-url>
cd social
```

#### 2. 配置设置
```bash
# 复制配置示例
cp config.yaml.example config.yaml

# 编辑配置文件
vim config.yaml

# 设置环境变量
export ENVIRONMENT=development
export SERVER_BASE_URL=http://localhost:8080
```

#### 3. 启动服务
```bash
# 启动Redis
redis-server

# 启动应用
go run main.go
```

#### 4. 访问界面
- **测试页面**: http://localhost:8080/static/test.html
- **API文档**: http://localhost:8080/swagger/index.html
- **健康检查**: http://localhost:8080/health

### 生产环境

#### 1. Docker部署
```bash
# 构建镜像
docker build -t social-platform .

# 运行容器
docker run -d \
  --name social-platform \
  -p 8080:8080 \
  -e ENVIRONMENT=production \
  -e SERVER_BASE_URL=https://api.yourdomain.com \
  -e REDIS_ADDR=redis:6379 \
  social-platform
```

#### 2. Docker Compose部署
```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

#### 3. 环境变量配置
```bash
# 生产环境变量
export ENVIRONMENT=production
export SERVER_BASE_URL=https://api.yourdomain.com
export REDIS_ADDR=redis.yourdomain.com:6379
export REDIS_PASSWORD=your_secure_password

# OAuth配置
export YOUTUBE_CLIENT_ID=your_youtube_client_id
export YOUTUBE_CLIENT_SECRET=your_youtube_client_secret
export X_CLIENT_ID=your_x_client_id
export X_CLIENT_SECRET=your_x_client_secret
# ... 其他平台配置
```

## 监控和维护

### 健康检查
```http
GET /health
```

响应示例：
```json
{
    "status": "healthy",
    "timestamp": "2024-01-15T10:30:00Z",
    "services": {
        "redis": "connected",
        "oauth": "ready"
    }
}
```

### 日志监控
- **结构化日志**: JSON格式，便于解析
- **日志级别**: Debug、Info、Warn、Error
- **请求追踪**: 每个请求的唯一ID
- **性能监控**: 请求耗时和资源使用

### 配置监控
```bash
# 验证配置
go run cmd/config/main.go -validate

# 查看配置警告
go run cmd/config/main.go -show | grep "⚠️"
```

## 安全考虑

### OAuth安全
- **HTTPS强制**: 生产环境必须使用HTTPS
- **State验证**: 防止CSRF攻击
- **PKCE支持**: 增强移动端安全性
- **Token安全**: 安全的token存储和传输

### API安全
- **输入验证**: 严格的输入参数验证
- **速率限制**: API调用频率限制
- **错误处理**: 不泄露敏感信息
- **日志记录**: 记录所有API调用

### 配置安全
- **敏感信息**: 使用环境变量存储
- **访问控制**: 限制配置文件访问权限
- **版本控制**: 配置文件不包含敏感信息
- **定期轮换**: 定期更新OAuth凭据

## 扩展指南

### 添加新平台
参考 [平台开发指南](PLATFORM_DEVELOPMENT_GUIDE.md) 了解如何添加新的社交媒体平台支持。

### 添加新服务
参考 [平台开发指南](PLATFORM_DEVELOPMENT_GUIDE.md) 了解如何为现有平台添加新的服务配置。

### 自定义功能
- **新API端点**: 在 `internal/handlers/` 中添加新的处理器
- **新存储后端**: 实现 `internal/storage/interface.go` 中的接口
- **新平台支持**: 在 `internal/platforms/` 中添加新的平台实现

## 故障排除

### 常见问题

#### 1. OAuth授权失败
- 检查Client ID和Secret是否正确
- 验证重定向URI配置
- 确认scopes权限设置
- 检查网络连接

#### 2. Redis连接失败
- 检查Redis服务是否运行
- 验证连接地址和端口
- 检查防火墙设置
- 确认Redis密码配置

#### 3. 配置加载失败
- 检查配置文件格式
- 验证环境变量设置
- 确认文件权限
- 查看错误日志

### 调试工具

#### 1. 配置验证
```bash
go run cmd/config/main.go -validate
```

#### 2. 健康检查
```bash
curl http://localhost:8080/health
```

#### 3. 日志查看
```bash
# 启用调试模式
export GIN_MODE=debug
go run main.go
```

## 贡献指南

### 开发流程
1. Fork项目仓库
2. 创建功能分支
3. 编写代码和测试
4. 提交Pull Request
5. 代码审查和合并

### 代码规范
- 遵循Go官方代码规范
- 编写单元测试
- 更新相关文档
- 添加适当的注释

### 测试要求
- 单元测试覆盖率 > 80%
- 集成测试覆盖主要流程
- 性能测试验证关键路径
- 安全测试检查潜在漏洞

## 许可证

本项目采用 MIT 许可证，详见 [LICENSE](LICENSE) 文件。

## 联系方式

- **项目维护者**: [维护者姓名]
- **邮箱**: [维护者邮箱]
- **问题反馈**: [GitHub Issues链接]
- **文档**: [项目文档链接]

---

*最后更新: 2024年1月15日*
