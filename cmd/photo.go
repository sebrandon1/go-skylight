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

const (
	photoExtJPG = ".jpg"
	photoExtMP4 = ".mp4"
)

var (
	photoPageToken   string
	photoLimit       int
	photoFile        string
	photoCaption     string
	photoID          []string
	photoOutputDir   string
	photoDownloadAll bool
)

var photoCmd = &cobra.Command{
	Use:   "photo",
	Short: "Photo management commands",
}

var photoListCmd = &cobra.Command{
	Use:   subList,
	Short: "List photos on a frame",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		photos, nextToken, err := client.ListPhotos(cmd.Context(), frameID, lib.PhotoListOptions{
			PageToken: photoPageToken,
		})
		if err != nil {
			return fmt.Errorf("listing photos: %w", err)
		}

		if photoLimit > 0 && len(photos) > photoLimit {
			photos = photos[:photoLimit]
			nextToken = ""
		}

		// include next_page_token in JSON so scripts can paginate without parsing stderr
		if outputFormat != outputTable {
			printJSON(map[string]any{
				"photos":          photos,
				"next_page_token": nextToken,
			})
			return nil
		}

		printOutput(photos)
		if nextToken != "" {
			fmt.Printf("next_page_token: %s\n", nextToken)
		}
		return nil
	},
}

var photoUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a photo to a frame",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		data, err := os.ReadFile(photoFile)
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}

		ext := strings.TrimPrefix(filepath.Ext(photoFile), ".")
		if ext == "" {
			ext = photoExtJPG[1:]
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		result, err := client.UploadPhoto(cmd.Context(), frameID, ext, data, photoCaption)
		if err != nil {
			return fmt.Errorf("uploading photo: %w", err)
		}

		printJSON(result)
		return nil
	},
}

var photoDeleteCmd = &cobra.Command{
	Use:   subDelete,
	Short: "Delete one or more photos by photo ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete %d photo(s): %v", len(photoID), photoID)
			return nil
		}

		if !confirmAction(fmt.Sprintf("Delete %d photo(s)?", len(photoID))) {
			return nil
		}

		ids := make([]int, 0, len(photoID))
		for _, s := range photoID {
			id, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("invalid photo ID %q: %w", s, err)
			}
			ids = append(ids, id)
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeletePhotos(cmd.Context(), frameID, ids); err != nil {
			return fmt.Errorf("deleting photos: %w", err)
		}

		printSuccess("Photos deleted successfully")
		return nil
	},
}

var photoDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download photos from a frame to local files",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if !photoDownloadAll && len(photoID) == 0 {
			return fmt.Errorf("specify --photo-id or --all")
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		wantIDs := make(map[string]bool, len(photoID))
		for _, id := range photoID {
			wantIDs[id] = true
		}

		var toDownload []lib.Photo
		ctx := cmd.Context()
		pageToken := ""
		for {
			photos, nextToken, err := client.ListPhotos(ctx, frameID, lib.PhotoListOptions{PageToken: pageToken})
			if err != nil {
				return fmt.Errorf("listing photos: %w", err)
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
			return nil
		}

		if err := os.MkdirAll(photoOutputDir, 0o755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}

		for _, p := range toDownload {
			data, err := client.DownloadPhoto(ctx, p.AssetURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: downloading %s: %v\n", p.ID, err)
				continue
			}
			ext := photoAssetExt(p.AssetURL, p.AssetType)
			filename := filepath.Join(photoOutputDir, p.ID+ext)
			if err := os.WriteFile(filename, data, 0o600); err != nil {
				fmt.Fprintf(os.Stderr, "warning: writing %s: %v\n", filename, err)
				continue
			}
			printSuccessf("Saved %s\n", filename)
		}
		return nil
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
		return photoExtMP4
	}
	return photoExtJPG
}

func init() {
	photoCmd.AddCommand(photoListCmd)
	photoCmd.AddCommand(photoUploadCmd)
	photoCmd.AddCommand(photoDeleteCmd)
	photoCmd.AddCommand(photoDownloadCmd)

	photoListCmd.Flags().StringVar(&photoPageToken, "page-token", "", "Pagination token (omit to start from beginning)")
	photoListCmd.Flags().IntVar(&photoLimit, "limit", 0, "Maximum number of photos to return (0 = server default)")

	photoUploadCmd.Flags().StringVar(&photoFile, "file", "", "Path to image file to upload")
	photoUploadCmd.Flags().StringVar(&photoCaption, "caption", "", "Optional caption for the photo")
	markFlagRequired(photoUploadCmd, "file")

	photoDeleteCmd.Flags().StringArrayVar(&photoID, "photo-id", nil, "Photo ID to delete (repeatable)")
	photoDeleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	photoDeleteCmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	markFlagRequired(photoDeleteCmd, "photo-id")

	photoDownloadCmd.Flags().StringArrayVar(&photoID, "photo-id", nil, "Photo ID to download (repeatable)")
	photoDownloadCmd.Flags().BoolVar(&photoDownloadAll, "all", false, "Download all photos")
	photoDownloadCmd.Flags().StringVar(&photoOutputDir, "output-dir", ".", "Directory to save downloaded files")
}
