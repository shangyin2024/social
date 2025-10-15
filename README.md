# 社交媒体平台项目

多平台社交媒体授权分享API服务，支持YouTube、X (Twitter)、Facebook、TikTok、Instagram等主流社交媒体平台。

## ✨ 新功能

### YouTube 音频/视频自动分类
- **智能文件检测**：根据文件扩展名自动识别音频和视频文件
- **自动分类上传**：音频文件自动使用音乐分类，视频文件使用默认分类
- **优化标签**：自动添加相应的标签，提高在YouTube Music和YouTube中的发现性
- **支持格式**：音频（MP3、WAV、FLAC、AAC、OGG、M4A、WMA），视频（MP4、AVI、MOV、WMV、FLV、WebM、MKV、M4V）

## 🚀 快速开始

### 环境要求
- Go 1.21+
- Redis 6.0+
- 各平台的OAuth应用凭据

### 安装运行
```bash
# 1. 克隆项目
git clone <repository-url>
cd social

# 2. 安装依赖
go mod download

# 3. 配置设置
cp config.yaml.example config.yaml
# 编辑配置文件，填入OAuth应用凭据

# 4. 启动服务
redis-server &  # 启动Redis
go run main.go  # 启动应用

# 5. 访问界面
# 测试页面: http://localhost:8080/static/test.html
# API文档: http://localhost:8080/swagger/index.html
```

## 📚 文档

### 文档索引
- **[📖 完整文档索引](docs/README.md)** - 所有文档的完整索引和使用指南

### 主要文档
- **[项目说明](docs/PROJECT_OVERVIEW.md)** - 完整的项目架构和功能说明
- **[平台开发指南](docs/PLATFORM_DEVELOPMENT_GUIDE.md)** - 添加新平台和服务配置的详细指南
- **[配置管理说明](docs/CONFIG_MANAGEMENT.md)** - 配置管理和多环境部署指南
- **[YouTube音频/视频自动分类指南](docs/YOUTUBE_AUDIO_VIDEO_GUIDE.md)** - YouTube根据文件类型自动分类上传功能

### 静态文件说明
- **[静态文件说明](static/README.md)** - 前端界面使用说明

## 🔧 配置管理

### 验证配置
```bash
go run cmd/config/main.go -validate
```

### 查看配置
```bash
go run cmd/config/main.go -show
```

### 环境变量支持
```bash
export ENVIRONMENT=development
export SERVER_BASE_URL=http://localhost:8080
export REDIS_ADDR=localhost:6379
```

## 🛠️ 开发工具

### 配置管理工具
```bash
# 验证配置
go run cmd/config/main.go -validate

# 查看配置
go run cmd/config/main.go -show

# 环境特定验证
go run cmd/config/main.go -validate -env production
```

### 健康检查
```bash
curl http://localhost:8080/health
```

## 🐳 部署

### Docker部署
```bash
docker build -t social-platform .
docker run -d --name social-platform -p 8080:8080 social-platform
```

### Docker Compose部署
```bash
docker-compose up -d
```

## 📋 支持的平台

| 平台 | 授权 | 分享 | 特殊要求 |
|------|------|------|----------|
| YouTube | ✅ | ✅ | 需要Google Cloud项目 |
| X (Twitter) | ✅ | ✅ | 需要X Developer账号 |
| Facebook | ✅ | ✅ | 需要Facebook应用 |
| TikTok | ✅ | ✅ | 需要TikTok开发者账号 |
| Instagram | ✅ | ✅ | 通过Facebook应用 |

## 🔗 主要功能

- 🔐 **OAuth授权管理**: 完整的OAuth 2.0流程，支持PKCE
- 📤 **内容分享**: 统一的内容分享接口，支持多格式内容
- 🛠️ **多服务配置**: 支持多个项目使用不同的OAuth配置
- 📊 **统计监控**: 内容分享统计和用户数据获取
- 🔄 **Token管理**: 自动token刷新和过期处理
- 📚 **API文档**: 完整的Swagger API文档
- 🎨 **现代化界面**: 响应式设计，Tab式布局

## 🚨 故障排除

### 常见问题
1. **OAuth授权失败** - 检查Client ID/Secret和重定向URI配置
2. **Redis连接失败** - 检查Redis服务状态和连接配置
3. **配置加载失败** - 验证配置文件格式和环境变量

### 调试工具
```bash
# 启用调试模式
export GIN_MODE=debug
go run main.go

# 查看详细配置
go run cmd/config/main.go -show -format json
```

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交Issue和Pull Request！请参考 [平台开发指南](docs/PLATFORM_DEVELOPMENT_GUIDE.md) 了解如何添加新功能。

## 📞 联系方式

- 邮箱: [your-email@example.com]
- GitHub: [your-github-username]
