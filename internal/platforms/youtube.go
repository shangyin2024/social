package platforms

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"social/internal/types"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// Context key type for storing video ID

// Media types
const (
	MediaTypeAudio = "audio"
	MediaTypeVideo = "video"
)

// VideoDetails represents detailed video information
type VideoDetails struct {
	Tags []string
}

// Audio file extensions
var audioExtensions = map[string]bool{
	".mp3":  true,
	".wav":  true,
	".flac": true,
	".aac":  true,
	".ogg":  true,
	".m4a":  true,
	".wma":  true,
}

// Video file extensions
var videoExtensions = map[string]bool{
	".mp4":  true,
	".avi":  true,
	".mov":  true,
	".wmv":  true,
	".flv":  true,
	".webm": true,
	".mkv":  true,
	".m4v":  true,
}

// YouTubePlatform implements the YouTube platform
type YouTubePlatform struct{}

// NewYouTubePlatform creates a new YouTube platform instance
func NewYouTubePlatform() *YouTubePlatform {
	return &YouTubePlatform{}
}

// GetName returns the platform name
func (y *YouTubePlatform) GetName() string {
	return "youtube"
}

// detectMediaType detects if the file is audio or video based on URL extension
func (y *YouTubePlatform) detectMediaType(mediaURL string) string {
	// Extract file extension from URL
	ext := strings.ToLower(filepath.Ext(mediaURL))

	// Check if it's an audio file
	if audioExtensions[ext] {
		return MediaTypeAudio
	}

	// Check if it's a video file
	if videoExtensions[ext] {
		return MediaTypeVideo
	}

	// Default to video for unknown extensions
	return MediaTypeVideo
}

// Share shares content to YouTube
func (y *YouTubePlatform) Share(ctx context.Context, client *http.Client, req *types.ShareRequest) (string, error) {
	// Debug logging to help diagnose metadata issues
	fmt.Printf("YouTube Share request - Title: '%s', Description: '%s', Content: '%s', Tags: %v\n",
		req.Title, req.Desc, req.Content, req.Tags)

	// Check if we have a media URL to upload
	if req.MediaURL == "" {
		return "", fmt.Errorf("media_url is required for YouTube upload")
	}

	// Detect media type (audio or video)
	mediaType := y.detectMediaType(req.MediaURL)
	fmt.Printf("Detected media type: %s for URL: %s\n", mediaType, req.MediaURL)

	// Download the media file from the URL
	mediaData, err := y.downloadMedia(ctx, client, req.MediaURL)
	if err != nil {
		return "", fmt.Errorf("failed to download media: %w", err)
	}

	// Create metadata based on media type
	metadata := y.createMetadata(req, mediaType)

	// Upload based on media type
	var mediaID string
	if mediaType == MediaTypeAudio {
		// For audio files, upload to YouTube with music-specific metadata
		mediaID, err = y.uploadAudio(ctx, client, mediaData, metadata)
		if err != nil {
			return "", fmt.Errorf("failed to upload audio: %w", err)
		}
	} else {
		// For video files, upload to YouTube normally
		mediaID, err = y.uploadVideo(ctx, client, mediaData, metadata)
		if err != nil {
			return "", fmt.Errorf("failed to upload video: %w", err)
		}
	}

	return mediaID, nil
}

// GetStats retrieves statistics from YouTube using the official SDK
func (y *YouTubePlatform) GetStats(ctx context.Context, client *http.Client, mediaID string) (types.StatsData, error) {
	if mediaID == "" {
		return types.StatsData{}, fmt.Errorf("media_id required")
	}

	// Create YouTube service using the authenticated client
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return types.StatsData{}, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// Call the videos.list method to get video statistics
	call := service.Videos.List([]string{"statistics"}).Id(mediaID)
	response, err := call.Context(ctx).Do()
	if err != nil {
		return types.StatsData{}, fmt.Errorf("failed to get video statistics: %w", err)
	}

	if len(response.Items) == 0 {
		return types.StatsData{}, fmt.Errorf("video not found")
	}

	stats := response.Items[0].Statistics

	// Parse counts - YouTube SDK returns uint64 values directly
	views := int(stats.ViewCount)
	likes := int(stats.LikeCount)
	comments := int(stats.CommentCount)

	return types.StatsData{
		Views:    views,
		Likes:    likes,
		Replies:  comments,
		Shares:   0, // YouTube doesn't provide share count in basic stats
		Retweets: 0, // YouTube doesn't have retweets
	}, nil
}

// GetUserInfo retrieves user information from YouTube platform using the official SDK
func (y *YouTubePlatform) GetUserInfo(ctx context.Context, client *http.Client) (types.UserInfo, error) {
	// Create YouTube service using the authenticated client
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return types.UserInfo{}, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// Call the channels.list method to get channel info
	call := service.Channels.List([]string{"snippet", "statistics"}).Mine(true)
	response, err := call.Context(ctx).Do()
	if err != nil {
		return types.UserInfo{}, fmt.Errorf("failed to get channel info: %w", err)
	}

	if len(response.Items) == 0 {
		return types.UserInfo{}, fmt.Errorf("no channel found for user")
	}

	channel := response.Items[0]

	// Build profile URL
	profileURL := fmt.Sprintf("https://www.youtube.com/channel/%s", channel.Id)

	// Parse subscriber count - YouTube SDK returns uint64 values directly
	subscriberCount := int(channel.Statistics.SubscriberCount)

	return types.UserInfo{
		ID:          channel.Id,
		Username:    channel.Id, // YouTube uses channel ID as username
		DisplayName: channel.Snippet.Title,
		Email:       "", // YouTube doesn't provide email in channel info
		AvatarURL:   channel.Snippet.Thumbnails.Default.Url,
		ProfileURL:  profileURL,
		Verified:    false, // YouTube verification status is not available in basic channel info
		Followers:   subscriberCount,
		Following:   0, // YouTube doesn't provide following count in channel info
	}, nil
}

// GetRecentPosts retrieves recent posts from YouTube
func (y *YouTubePlatform) GetRecentPosts(ctx context.Context, client *http.Client, limit int, startTime, endTime int64) ([]types.Post, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Create YouTube service using the authenticated client
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// First, get the user's channel ID
	channelsCall := service.Channels.List([]string{"id"}).Mine(true)
	channelsResponse, err := channelsCall.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get user channel: %w", err)
	}

	if len(channelsResponse.Items) == 0 {
		return nil, fmt.Errorf("no channel found for user")
	}

	channelID := channelsResponse.Items[0].Id
	fmt.Printf("DEBUG: Found channel ID: %s\n", channelID)

	// Get the channel's uploads playlist ID
	channelsCall2 := service.Channels.List([]string{"contentDetails"}).Id(channelID)
	channelsResponse2, err := channelsCall2.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get channel details: %w", err)
	}

	if len(channelsResponse2.Items) == 0 {
		return nil, fmt.Errorf("no channel details found")
	}

	uploadsPlaylistID := channelsResponse2.Items[0].ContentDetails.RelatedPlaylists.Uploads
	fmt.Printf("DEBUG: Found uploads playlist ID: %s\n", uploadsPlaylistID)

	// Validate uploads playlist ID format
	if uploadsPlaylistID == "" {
		return nil, fmt.Errorf("uploads playlist ID is empty")
	}

	fmt.Printf("DEBUG: Using uploads playlist ID: %s\n", uploadsPlaylistID)

	// Get videos from the uploads playlist with more detailed information
	playlistItemsCall := service.PlaylistItems.List([]string{"snippet", "contentDetails"}).PlaylistId(uploadsPlaylistID).MaxResults(int64(limit))

	// Note: YouTube PlaylistItems API doesn't support time filtering directly
	// We'll need to filter the results after fetching them
	fmt.Printf("DEBUG: YouTube PlaylistItems API doesn't support time filtering, will filter results after fetching\n")

	// Execute the playlist items request
	fmt.Printf("DEBUG: Executing playlist items request for playlist ID: %s\n", uploadsPlaylistID)
	playlistResponse, err := playlistItemsCall.Context(ctx).Do()
	if err != nil {
		fmt.Printf("DEBUG: Playlist items request failed with error: %v\n", err)
		return nil, fmt.Errorf("failed to get playlist items: %w", err)
	}

	fmt.Printf("DEBUG: Playlist items request successful, found %d items\n", len(playlistResponse.Items))

	// Convert to Post structs and apply time filtering
	var posts []types.Post
	for _, item := range playlistResponse.Items {
		// Safety check for required fields
		if item.Snippet == nil {
			fmt.Printf("DEBUG: Skipping item with nil snippet\n")
			continue
		}

		if item.Snippet.ResourceId == nil || item.Snippet.ResourceId.VideoId == "" {
			fmt.Printf("DEBUG: Skipping item with missing video ID\n")
			continue
		}

		// Parse published time
		publishedTime, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil {
			publishedTime = time.Now()
		}

		publishedUnix := publishedTime.Unix()

		// Apply time filtering if specified
		if startTime > 0 {
			// Handle both second and millisecond timestamps
			var startTimeUnix int64
			if startTime > 1e12 { // If timestamp is larger than 1e12, it's likely in milliseconds
				startTimeUnix = startTime / 1000
			} else {
				startTimeUnix = startTime
			}
			if publishedUnix < startTimeUnix {
				fmt.Printf("DEBUG: Skipping video %s (published: %d, start_time: %d)\n", item.Snippet.ResourceId.VideoId, publishedUnix, startTimeUnix)
				continue
			}
		}

		if endTime > 0 {
			// Handle both second and millisecond timestamps
			var endTimeUnix int64
			if endTime > 1e12 { // If timestamp is larger than 1e12, it's likely in milliseconds
				endTimeUnix = endTime / 1000
			} else {
				endTimeUnix = endTime
			}
			if publishedUnix > endTimeUnix {
				fmt.Printf("DEBUG: Skipping video %s (published: %d, end_time: %d)\n", item.Snippet.ResourceId.VideoId, publishedUnix, endTimeUnix)
				continue
			}
		}

		// Get video statistics and tags
		stats, err := y.getVideoStats(ctx, service, item.Snippet.ResourceId.VideoId)
		if err != nil {
			// If stats fail, continue with zero stats
			stats = types.StatsData{}
		}

		// Get video details including tags
		videoDetails, err := y.getVideoDetails(ctx, service, item.Snippet.ResourceId.VideoId)
		if err != nil {
			// If video details fail, continue with empty tags
			fmt.Printf("DEBUG: Failed to get video details for %s: %v\n", item.Snippet.ResourceId.VideoId, err)
			videoDetails = &VideoDetails{}
		} else {
			fmt.Printf("DEBUG: Video %s has %d tags: %v\n", item.Snippet.ResourceId.VideoId, len(videoDetails.Tags), videoDetails.Tags)
		}

		// Build video URL
		videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.Snippet.ResourceId.VideoId)

		// Safely get thumbnail URL
		thumbnailURL := ""
		if item.Snippet.Thumbnails != nil && item.Snippet.Thumbnails.Default != nil {
			thumbnailURL = item.Snippet.Thumbnails.Default.Url
		}

		// Ensure tags is never nil
		tags := videoDetails.Tags
		if tags == nil {
			tags = []string{}
		}

		// Use title as content if description is empty
		content := item.Snippet.Description
		if content == "" {
			content = item.Snippet.Title
		}

		post := types.Post{
			ID:          item.Snippet.ResourceId.VideoId,
			Content:     content,
			Title:       item.Snippet.Title,
			Description: item.Snippet.Description,
			CreatedAt:   publishedUnix,
			UpdatedAt:   publishedUnix, // YouTube doesn't provide separate updated time
			Stats:       stats,
			URL:         videoURL,
			MediaType:   "video",
			MediaURL:    thumbnailURL,
			Tags:        tags,
		}

		fmt.Printf("DEBUG: Final post data - ID: %s, Title: %s, Tags: %v\n", post.ID, post.Title, post.Tags)

		posts = append(posts, post)
	}

	return posts, nil
}

// getVideoStats gets statistics for a specific video
func (y *YouTubePlatform) getVideoStats(ctx context.Context, service *youtube.Service, videoID string) (types.StatsData, error) {
	call := service.Videos.List([]string{"statistics"}).Id(videoID)
	response, err := call.Context(ctx).Do()
	if err != nil {
		return types.StatsData{}, err
	}

	if len(response.Items) == 0 {
		return types.StatsData{}, fmt.Errorf("video not found")
	}

	stats := response.Items[0].Statistics

	return types.StatsData{
		Views:   int(stats.ViewCount),
		Likes:   int(stats.LikeCount),
		Replies: int(stats.CommentCount),
		Shares:  0, // YouTube doesn't provide share count in basic stats
	}, nil
}

// getVideoDetails gets detailed information for a specific video including tags
func (y *YouTubePlatform) getVideoDetails(ctx context.Context, service *youtube.Service, videoID string) (*VideoDetails, error) {
	call := service.Videos.List([]string{"snippet"}).Id(videoID)
	response, err := call.Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("video not found")
	}

	video := response.Items[0]
	tags := make([]string, 0)

	fmt.Printf("DEBUG: Video snippet: %+v\n", video.Snippet)

	if video.Snippet != nil {
		if video.Snippet.Tags != nil {
			tags = video.Snippet.Tags
			fmt.Printf("DEBUG: Found %d tags in video snippet: %v\n", len(tags), tags)
		} else {
			fmt.Printf("DEBUG: Video snippet exists but no tags found\n")
		}
	} else {
		fmt.Printf("DEBUG: Video snippet is nil\n")
	}

	return &VideoDetails{
		Tags: tags,
	}, nil
}

// HandleOAuthCallback handles OAuth callback for YouTube platform
func (y *YouTubePlatform) HandleOAuthCallback(ctx context.Context, code, state string) error {
	// YouTube平台特定的OAuth回调处理逻辑
	// 这里可以添加YouTube平台特有的处理逻辑
	return nil
}

// downloadMedia downloads media file from the given URL
func (y *YouTubePlatform) downloadMedia(ctx context.Context, client *http.Client, mediaURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", mediaURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download media: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to download media: status=%d", resp.StatusCode)
	}

	mediaData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read media data: %w", err)
	}

	return mediaData, nil
}

// createMetadata creates metadata for YouTube upload based on media type
func (y *YouTubePlatform) createMetadata(req *types.ShareRequest, mediaType string) map[string]any {
	title := y.getTitle(req, mediaType)
	description := y.getDescription(req, mediaType)

	// Debug logging to verify metadata creation
	fmt.Printf("YouTube metadata creation - Type: %s, Title: '%s', Description: '%s', Tags: %v\n",
		mediaType, title, description, req.Tags)

	metadata := map[string]any{
		"snippet": map[string]any{
			"title":       title,
			"description": description,
			"tags":        y.getTags(req, mediaType),
		},
		"status": map[string]any{
			"privacyStatus": y.getPrivacyStatus(req),
		},
	}

	// Add category ID based on media type
	if mediaType == MediaTypeAudio {
		// Category 10 is "Music" on YouTube
		metadata["snippet"].(map[string]any)["categoryId"] = "10"
	} else {
		// Category 22 is "People & Blogs" for general videos
		metadata["snippet"].(map[string]any)["categoryId"] = "22"
	}

	return metadata
}

// getTitle returns the title based on media type
func (y *YouTubePlatform) getTitle(req *types.ShareRequest, mediaType string) string {
	if req.Title != "" {
		return req.Title
	}
	if req.Content != "" {
		// Use content as title if no title provided, truncate if too long
		if len(req.Content) > 100 {
			return req.Content[:97] + "..."
		}
		return req.Content
	}

	// Default title based on media type
	if mediaType == MediaTypeAudio {
		return "Untitled Audio"
	}
	return "Untitled Video"
}

// getDescription returns the description based on media type
func (y *YouTubePlatform) getDescription(req *types.ShareRequest, mediaType string) string {
	if req.Desc != "" {
		return req.Desc
	}
	if req.Content != "" {
		return req.Content
	}

	// Default description based on media type
	if mediaType == MediaTypeAudio {
		return "Music content uploaded via API"
	}
	return "Video content uploaded via API"
}

// getTags returns tags with media type specific additions
func (y *YouTubePlatform) getTags(req *types.ShareRequest, mediaType string) []string {
	tags := make([]string, 0)

	// Add existing tags
	if req.Tags != nil {
		tags = append(tags, req.Tags...)
	}

	// Add media type specific tags
	if mediaType == MediaTypeAudio {
		tags = append(tags, "music", "audio", "youtube-music")
	} else {
		tags = append(tags, "video", "youtube")
	}

	return tags
}

// getPrivacyStatus returns the privacy status for the video
func (y *YouTubePlatform) getPrivacyStatus(req *types.ShareRequest) string {
	switch req.Privacy {
	case "private":
		return "private"
	case "unlisted":
		return "unlisted"
	case "public":
		return "public"
	default:
		return "public" // Default to public
	}
}

// uploadAudio uploads audio to YouTube with music-specific settings
func (y *YouTubePlatform) uploadAudio(ctx context.Context, client *http.Client, audioData []byte, metadata map[string]any) (string, error) {
	// For audio files, we upload to YouTube but with music-specific metadata
	// This will make the content more discoverable in YouTube Music
	fmt.Printf("Uploading audio file to YouTube with music metadata\n")

	// Use the same upload logic as video, but with music-specific metadata
	return y.uploadVideo(ctx, client, audioData, metadata)
}

// uploadVideo uploads video to YouTube using the official YouTube Go client library
func (y *YouTubePlatform) uploadVideo(ctx context.Context, client *http.Client, videoData []byte, metadata map[string]any) (string, error) {
	// Create YouTube service using the authenticated client
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// Extract metadata from the map
	snippetData, ok := metadata["snippet"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("invalid snippet data in metadata")
	}

	statusData, ok := metadata["status"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("invalid status data in metadata")
	}

	// Create video object using official YouTube types
	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       getStringFromInterface(snippetData["title"]),
			Description: getStringFromInterface(snippetData["description"]),
			// CategoryId:  "22", // Default category for "People & Blogs"
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: getStringFromInterface(statusData["privacyStatus"]),
		},
	}

	// Add tags if they exist
	if tags, ok := snippetData["tags"].([]string); ok && len(tags) > 0 {
		upload.Snippet.Tags = tags
	}

	// Debug logging
	fmt.Printf("YouTube upload - Title: '%s', Description: '%s', Tags: %v, Privacy: '%s'\n",
		upload.Snippet.Title, upload.Snippet.Description, upload.Snippet.Tags, upload.Status.PrivacyStatus)

	// Create the insert call
	call := service.Videos.Insert([]string{"snippet", "status"}, upload)

	// Create a reader from the video data
	videoReader := bytes.NewReader(videoData)

	// Execute the upload
	response, err := call.Media(videoReader).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failed to upload video: %w", err)
	}

	// Debug logging
	fmt.Printf("YouTube upload successful! Video ID: %s\n", response.Id)
	fmt.Printf("YouTube response - Title: '%s', Description: '%s'\n",
		response.Snippet.Title, response.Snippet.Description)

	return response.Id, nil
}

// Helper function to safely extract string from any
func getStringFromInterface(val any) string {
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}
