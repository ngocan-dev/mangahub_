package api

import (
	"context"
	"net/http"
)

// ChapterReleaseNotificationRequest captures the notification payload.
type ChapterReleaseNotificationRequest struct {
	NovelID   int64 `json:"novel_id"`
	Chapter   int   `json:"chapter"`
	ChapterID int64 `json:"chapter_id,omitempty"`
}

// ChapterReleaseNotificationResponse captures the notification response.
type ChapterReleaseNotificationResponse struct {
	Message   string `json:"message"`
	NovelID   int64  `json:"novel_id"`
	NovelName string `json:"novel_name"`
	Chapter   int    `json:"chapter"`
	ChapterID int64  `json:"chapter_id"`
}

// NotifyChapterRelease triggers a chapter release notification broadcast.
func (c *Client) NotifyChapterRelease(ctx context.Context, req ChapterReleaseNotificationRequest) (*ChapterReleaseNotificationResponse, error) {
	var resp ChapterReleaseNotificationResponse
	if err := c.doRequest(ctx, http.MethodPost, "/admin/notify", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
