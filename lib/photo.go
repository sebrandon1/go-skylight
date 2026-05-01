package lib

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ListPhotos retrieves photos (and videos) for a frame with optional pagination.
// Pass PageToken "__START__" (or empty string) to start from the beginning.
// Returns the photos, the next page token (empty when no more pages), and any error.
func (c *Client) ListPhotos(frameID string, opts PhotoListOptions) ([]Photo, string, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/messages", c.effectiveURL(), frameID))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create list photos request: %w", err)
	}

	pageToken := opts.PageToken
	if pageToken == "" {
		pageToken = "__START__"
	}
	addQueryParams(req, map[string]string{"page_token": pageToken})

	var apiResp photoAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, "", fmt.Errorf("failed to list photos: %w", err)
	}

	photos := make([]Photo, len(apiResp.Data))
	for i := range apiResp.Data {
		photos[i] = apiResp.Data[i].toPhoto()
	}

	var nextToken string
	if apiResp.Meta != nil {
		nextToken = apiResp.Meta.NextPageToken
	}

	return photos, nextToken, nil
}

// UploadPhoto uploads an image to a Skylight frame using a two-step S3 presign flow.
// ext is the file extension without the dot (e.g. "jpg", "png").
// imageData is the raw image bytes. caption is optional.
func (c *Client) UploadPhoto(frameID string, ext string, imageData []byte, caption string) (*PhotoUploadResponse, error) {
	// Step 1: request a presigned S3 upload URL
	uploadReq := photoUploadURLRequest{
		Ext:      strings.ToLower(ext),
		FrameIDs: []string{frameID},
		Caption:  caption,
	}
	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/upload_url", c.effectiveURL()), uploadReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload URL request: %w", err)
	}

	var urlResp photoUploadURLResponse
	if err := c.post(req, &urlResp); err != nil {
		return nil, fmt.Errorf("failed to get upload URL: %w", err)
	}

	presignedURL := urlResp.Data.UploadURL
	if presignedURL == "" {
		return nil, fmt.Errorf("empty upload URL in response")
	}

	// Step 2: PUT the raw bytes directly to the presigned S3 URL
	putReq, err := http.NewRequest("PUT", presignedURL, bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 PUT request: %w", err)
	}
	contentType := extToContentType(ext)
	putReq.Header.Set("Content-Type", contentType)

	resp, err := c.HTTPClient.Do(putReq)
	if err != nil {
		return nil, fmt.Errorf("failed to upload photo to S3: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) //nolint:errcheck

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("S3 upload failed with status %d", resp.StatusCode)
	}

	result := urlResp.Data
	return &result, nil
}

// DownloadPhoto fetches the raw bytes of a photo from its CDN asset URL.
// The URL is a pre-signed CloudFront URL returned by ListPhotos — no auth header is needed.
func (c *Client) DownloadPhoto(assetURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", assetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download photo: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read photo data: %w", err)
	}
	return data, nil
}

// DeletePhotos deletes one or more photos by their message IDs.
func (c *Client) DeletePhotos(frameID string, messageIDs []int) error {
	req, err := newRequestWithBody(
		"DELETE",
		fmt.Sprintf("%s/frames/%s/messages/destroy_multiple", c.effectiveURL(), frameID),
		photoDeleteRequest{MessageIDs: messageIDs},
	)
	if err != nil {
		return fmt.Errorf("failed to create delete photos request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete photos: %w", err)
	}

	return nil
}

// extToContentType maps a file extension to a MIME content type.
func extToContentType(ext string) string {
	switch strings.ToLower(ext) {
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "mp4":
		return "video/mp4"
	case "mov":
		return "video/quicktime"
	default:
		return "image/jpeg"
	}
}
