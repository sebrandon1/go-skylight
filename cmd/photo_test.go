package cmd

import "testing"

func TestPhotoAssetExt(t *testing.T) {
	cases := []struct {
		url       string
		assetType string
		want      string
	}{
		{"https://cdn.example.com/photos/abc.jpg", "image/jpeg", ".jpg"},
		{"https://cdn.example.com/photos/abc.png", "image/png", ".png"},
		{"https://cdn.example.com/photos/abc.mp4", "video/mp4", ".mp4"},
		{"https://cdn.example.com/photos/abc.gif", "image/gif", ".gif"},
		// no extension in URL: fall back to asset type
		{"https://cdn.example.com/photos/abc", "video/mp4", ".mp4"},
		{"https://cdn.example.com/photos/abc", "video/quicktime", ".mp4"},
		{"https://cdn.example.com/photos/abc", "image/jpeg", ".jpg"},
		// empty URL, video type maps to .mp4; non-video type defaults to .jpg regardless of subtype
		{"", "video/quicktime", ".mp4"},
		{"", "image/png", ".jpg"},
	}
	for _, c := range cases {
		got := photoAssetExt(c.url, c.assetType)
		if got != c.want {
			t.Errorf("photoAssetExt(%q, %q) = %q, want %q", c.url, c.assetType, got, c.want)
		}
	}
}
