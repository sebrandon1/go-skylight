package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	photoPageToken   string
	photoFile        string
	photoCaption     string
	photoMessageID   []string
	photoOutputDir   string
	photoDownloadAll bool
)

var photoCmd = &cobra.Command{
	Use:   "photo",
	Short: "Photo management commands",
}

var photoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List photos on a frame",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		photos, nextToken, err := client.ListPhotos(frameID, lib.PhotoListOptions{
			PageToken: photoPageToken,
		})
		if err != nil {
			fatal("listing photos", err)
		}

		printOutput(photos)

		if nextToken != "" {
			fmt.Fprintf(os.Stderr, "Next page token: %s\n", nextToken)
		}
	},
}

var photoUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a photo to a frame",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		data, err := os.ReadFile(photoFile)
		if err != nil {
			fatal("reading file", err)
		}

		ext := strings.TrimPrefix(filepath.Ext(photoFile), ".")
		if ext == "" {
			ext = "jpg"
		}

		client := getClient()

		result, err := client.UploadPhoto(frameID, ext, data, photoCaption)
		if err != nil {
			fatal("uploading photo", err)
		}

		printJSON(result)
	},
}

var photoDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete one or more photos by message ID",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		ids := make([]int, 0, len(photoMessageID))
		for _, s := range photoMessageID {
			id, err := strconv.Atoi(s)
			if err != nil {
				fatal(fmt.Sprintf("invalid message ID %q", s), err)
			}
			ids = append(ids, id)
		}

		client := getClient()

		err := client.DeletePhotos(frameID, ids)
		if err != nil {
			fatal("deleting photos", err)
		}

		fmt.Println("Photos deleted successfully")
	},
}

var photoDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download photos from a frame to local files",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		if !photoDownloadAll && len(photoMessageID) == 0 {
			fmt.Fprintln(os.Stderr, "Error: specify --message-id or --all")
			os.Exit(1)
		}

		client := getClient()

		wantIDs := make(map[string]bool, len(photoMessageID))
		for _, id := range photoMessageID {
			wantIDs[id] = true
		}

		var toDownload []lib.Photo
		pageToken := ""
		for {
			photos, nextToken, err := client.ListPhotos(frameID, lib.PhotoListOptions{PageToken: pageToken})
			if err != nil {
				fatal("listing photos", err)
			}
			for _, p := range photos {
				if photoDownloadAll || wantIDs[p.ID] {
					toDownload = append(toDownload, p)
				}
			}
			if nextToken == "" {
				break
			}
			pageToken = nextToken
		}

		if len(toDownload) == 0 {
			fmt.Println("No matching photos found")
			return
		}

		if err := os.MkdirAll(photoOutputDir, 0o755); err != nil {
			fatal("creating output directory", err)
		}

		for _, p := range toDownload {
			data, err := client.DownloadPhoto(p.AssetURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: downloading %s: %v\n", p.ID, err)
				continue
			}
			ext := photoAssetExt(p.AssetURL, p.AssetType)
			filename := filepath.Join(photoOutputDir, p.ID+ext)
			if err := os.WriteFile(filename, data, 0o600); err != nil {
				fmt.Fprintf(os.Stderr, "Error: writing %s: %v\n", filename, err)
				continue
			}
			fmt.Printf("Saved %s\n", filename)
		}
	},
}

// photoAssetExt infers the file extension from the asset URL path, falling back to asset type.
func photoAssetExt(assetURL, assetType string) string {
	if u, err := url.Parse(assetURL); err == nil {
		if ext := filepath.Ext(u.Path); ext != "" {
			return ext
		}
	}
	if strings.HasPrefix(assetType, "video") {
		return ".mp4"
	}
	return ".jpg"
}

func init() {
	photoCmd.AddCommand(photoListCmd)
	photoCmd.AddCommand(photoUploadCmd)
	photoCmd.AddCommand(photoDeleteCmd)
	photoCmd.AddCommand(photoDownloadCmd)

	photoListCmd.Flags().StringVar(&photoPageToken, "page-token", "", "Pagination token (omit to start from beginning)")

	photoUploadCmd.Flags().StringVar(&photoFile, "file", "", "Path to image file to upload")
	photoUploadCmd.Flags().StringVar(&photoCaption, "caption", "", "Optional caption for the photo")
	photoUploadCmd.MarkFlagRequired("file") //nolint:errcheck

	photoDeleteCmd.Flags().StringArrayVar(&photoMessageID, "message-id", nil, "Message ID to delete (repeatable)")
	photoDeleteCmd.MarkFlagRequired("message-id") //nolint:errcheck

	photoDownloadCmd.Flags().StringArrayVar(&photoMessageID, "message-id", nil, "Message ID to download (repeatable)")
	photoDownloadCmd.Flags().BoolVar(&photoDownloadAll, "all", false, "Download all photos")
	photoDownloadCmd.Flags().StringVar(&photoOutputDir, "output-dir", ".", "Directory to save downloaded files")
}
