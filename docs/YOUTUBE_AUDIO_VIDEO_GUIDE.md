# YouTube 音频/视频自动分类指南

## 概述

本系统现在支持根据文件类型自动将内容分类上传到YouTube：

- **音频文件** → 上传到YouTube，使用音乐分类和标签，便于在YouTube Music中发现
- **视频文件** → 正常上传到YouTube

## 功能特性

### 1. 自动文件类型检测
系统会根据文件扩展名自动检测文件类型：

**支持的音频格式：**
- `.mp3` - MP3音频
- `.wav` - WAV音频
- `.flac` - FLAC无损音频
- `.aac` - AAC音频
- `.ogg` - OGG音频
- `.m4a` - M4A音频
- `.wma` - WMA音频

**支持的视频格式：**
- `.mp4` - MP4视频
- `.avi` - AVI视频
- `.mov` - MOV视频
- `.wmv` - WMV视频
- `.flv` - FLV视频
- `.webm` - WebM视频
- `.mkv` - MKV视频
- `.m4v` - M4V视频

### 2. 智能分类和标签
- **音频文件**：自动添加音乐分类（Category ID: 10）和音乐相关标签
- **视频文件**：使用默认分类（Category ID: 22）和视频相关标签

### 3. 元数据优化
根据文件类型自动优化标题、描述和标签，提高内容在相应平台的发现性。

## API 使用方法

### 上传音频文件

**请求示例：**
```bash
curl -X POST http://localhost:8080/api/share \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "youtube",
    "user_id": "user123",
    "server_name": "myapp",
    "title": "我的新歌",
    "description": "这是一首原创歌曲",
    "media_url": "https://example.com/song.mp3",
    "tags": ["原创", "流行音乐"],
    "privacy": "public"
  }'
```

**系统处理：**
1. 检测到 `.mp3` 扩展名，识别为音频文件
2. 自动添加音乐分类（Category ID: 10）
3. 自动添加音乐相关标签：`["原创", "流行音乐", "music", "audio", "youtube-music"]`
4. 上传到YouTube，内容将在YouTube Music中更容易被发现

### 上传视频文件

**请求示例：**
```bash
curl -X POST http://localhost:8080/api/share \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "youtube",
    "user_id": "user123",
    "server_name": "myapp",
    "title": "我的视频教程",
    "description": "这是一个编程教程视频",
    "media_url": "https://example.com/tutorial.mp4",
    "tags": ["编程", "教程"],
    "privacy": "public"
  }'
```

**系统处理：**
1. 检测到 `.mp4` 扩展名，识别为视频文件
2. 使用默认分类（Category ID: 22）
3. 自动添加视频相关标签：`["编程", "教程", "video", "youtube"]`
4. 正常上传到YouTube

## 响应示例

**成功响应：**
```json
{
  "status": "success",
  "message": "Content shared successfully",
  "data": {
    "provider": "youtube",
    "user_id": "user123",
    "server_name": "myapp",
    "content": "我的新歌",
    "media_url": "https://example.com/song.mp3",
    "tags": ["原创", "流行音乐", "music", "audio", "youtube-music"],
    "media_id": "dQw4w9WgXcQ"
  }
}
```

## 配置要求

### YouTube OAuth 权限

确保您的YouTube OAuth配置包含以下权限：

```yaml
youtube:
  client_id: "your_youtube_client_id"
  client_secret: "your_youtube_client_secret"
  scopes:
    - "https://www.googleapis.com/auth/youtube.upload"
    - "https://www.googleapis.com/auth/youtube.readonly"
    - "openid"
    - "email"
```

### Google Cloud Console 设置

1. 在 Google Cloud Console 中启用 YouTube Data API v3
2. 确保您的应用有上传视频的权限
3. 配置正确的重定向URI

## 使用场景

### 场景1：音乐内容发布

```javascript
// 发布音乐文件
const musicResponse = await fetch('/api/share', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    provider: 'youtube',
    user_id: 'user123',
    server_name: 'myapp',
    title: '新专辑 - 歌曲名称',
    description: '来自新专辑的歌曲，希望大家喜欢',
    media_url: 'https://example.com/album-song.mp3',
    tags: ['新专辑', '流行音乐', '原创'],
    privacy: 'public'
  })
});

// 系统会自动：
// 1. 检测为音频文件
// 2. 添加音乐分类
// 3. 添加音乐标签
// 4. 优化在YouTube Music中的发现性
```

### 场景2：视频内容发布

```javascript
// 发布视频文件
const videoResponse = await fetch('/api/share', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    provider: 'youtube',
    user_id: 'user123',
    server_name: 'myapp',
    title: '编程教程 - 第1集',
    description: '从零开始学习编程',
    media_url: 'https://example.com/programming-tutorial.mp4',
    tags: ['编程', '教程', '学习'],
    privacy: 'public'
  })
});

// 系统会自动：
// 1. 检测为视频文件
// 2. 使用默认分类
// 3. 添加视频标签
// 4. 正常上传到YouTube
```

### 场景3：批量上传不同类型内容

```javascript
const uploads = [
  {
    title: '歌曲1',
    media_url: 'https://example.com/song1.mp3',
    tags: ['音乐', '原创']
  },
  {
    title: '歌曲2',
    media_url: 'https://example.com/song2.wav',
    tags: ['音乐', '翻唱']
  },
  {
    title: '视频教程',
    media_url: 'https://example.com/tutorial.mp4',
    tags: ['教程', '编程']
  }
];

for (const upload of uploads) {
  const response = await fetch('/api/share', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      provider: 'youtube',
      user_id: 'user123',
      server_name: 'myapp',
      ...upload,
      privacy: 'public'
    })
  });

  const result = await response.json();
  console.log(`上传完成: ${result.data.media_id}`);
}
```

## 技术实现细节

### 文件类型检测逻辑

```go
// 检测媒体类型
func (y *YouTubePlatform) detectMediaType(mediaURL string) string {
    ext := strings.ToLower(filepath.Ext(mediaURL))

    if audioExtensions[ext] {
        return MediaTypeAudio
    }

    if videoExtensions[ext] {
        return MediaTypeVideo
    }

    return MediaTypeVideo // 默认为视频
}
```

### 分类和标签策略

**音频文件：**
- 分类ID：10（音乐）
- 自动标签：`["music", "audio", "youtube-music"]`
- 默认描述：`"Music content uploaded via API"`

**视频文件：**
- 分类ID：22（人物与博客）
- 自动标签：`["video", "youtube"]`
- 默认描述：`"Video content uploaded via API"`

## 注意事项

1. **文件格式支持**：确保上传的文件格式在支持列表中
2. **文件大小限制**：YouTube对上传文件有大小限制（通常最大128GB）
3. **处理时间**：音频文件可能需要额外处理时间
4. **版权问题**：确保您有上传内容的版权或授权

## 错误处理

### 常见错误及解决方案

1. **不支持的文件格式**
   ```json
   {
     "error": "Unsupported file format",
     "code": "UNSUPPORTED_FORMAT"
   }
   ```
   解决方案：使用支持的文件格式

2. **文件下载失败**
   ```json
   {
     "error": "Failed to download media",
     "code": "DOWNLOAD_FAILED"
   }
   ```
   解决方案：检查媒体URL是否可访问

3. **上传失败**
   ```json
   {
     "error": "Failed to upload to YouTube",
     "code": "UPLOAD_FAILED"
   }
   ```
   解决方案：检查OAuth权限和网络连接

## 更新日志

- **v1.0.0**: 初始版本，支持基本的音频/视频检测
- **v1.1.0**: 添加自动分类和标签功能
- **v1.2.0**: 优化YouTube Music发现性
- **v1.3.0**: 完善错误处理和文档
