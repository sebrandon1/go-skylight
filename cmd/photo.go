package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	photoPageToken string
	photoFile      string
	photoCaption   string
	photoMessageID []string
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
			fmt.Printf("Error listing photos: %v\n", err)
			os.Exit(1)
		}

		printJSON(photos)

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
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		ext := strings.TrimPrefix(filepath.Ext(photoFile), ".")
		if ext == "" {
			ext = "jpg"
		}

		client := getClient()

		result, err := client.UploadPhoto(frameID, ext, data, photoCaption)
		if err != nil {
			fmt.Printf("Error uploading photo: %v\n", err)
			os.Exit(1)
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
				fmt.Printf("Invalid message ID %q: %v\n", s, err)
				os.Exit(1)
			}
			ids = append(ids, id)
		}

		client := getClient()

		err := client.DeletePhotos(frameID, ids)
		if err != nil {
			fmt.Printf("Error deleting photos: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Photos deleted successfully")
	},
}

func init() {
	photoCmd.AddCommand(photoListCmd)
	photoCmd.AddCommand(photoUploadCmd)
	photoCmd.AddCommand(photoDeleteCmd)

	photoListCmd.Flags().StringVar(&photoPageToken, "page-token", "", "Pagination token (omit to start from beginning)")

	photoUploadCmd.Flags().StringVar(&photoFile, "file", "", "Path to image file to upload")
	photoUploadCmd.Flags().StringVar(&photoCaption, "caption", "", "Optional caption for the photo")
	photoUploadCmd.MarkFlagRequired("file") //nolint:errcheck

	photoDeleteCmd.Flags().StringArrayVar(&photoMessageID, "message-id", nil, "Message ID to delete (repeatable)")
	photoDeleteCmd.MarkFlagRequired("message-id") //nolint:errcheck
}
