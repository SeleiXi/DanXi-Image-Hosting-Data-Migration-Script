package main

import (
	"fmt"
	"github.com/opentreehole/backend/model"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
	"log/slog"
)

func main() {
	model.Init()

	batchSize := 10000

	errFile, err := os.OpenFile("error.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open error file: %v", err)
	}
	defer func() {
		if err := errFile.Close(); err != nil {
			slog.Error("Error closing error file", "err", err)
		}
	}()
	log.SetOutput(errFile)

	var images []model.OriginalImageTable
	err = model.OriginalDB.FindInBatches(&images, batchSize, func(tx *gorm.DB, batch int) error {
		slog.Info("Processing batch", "batch", batch)

		for _, image := range images {
			imageIdentifier := strings.TrimSuffix(image.Name, filepath.Ext(image.Name))

			imageURL := "https://pic.jingyijun.xyz:8443/i/" + image.Path + "/" + image.Name
			slog.Info("Downloading image", "identifier", imageIdentifier, "url", imageURL)

			imageData, err := downloadImage(imageURL)
			if err != nil {
				slog.Error("Error downloading image", "identifier", imageIdentifier, "err", err)
				continue
			}
			slog.Info("Image downloaded successfully", "identifier", imageIdentifier)

			originalFileName := image.OriginName
			fileExtension := strings.TrimPrefix(filepath.Ext(image.Name), ".")
			createdAt := image.CreatedAt
			updatedAt := image.UpdatedAt

			err = storeImageInDatabase(originalFileName, fileExtension, imageData, imageIdentifier, createdAt, updatedAt)
			if err != nil {
				slog.Error("Error storing image in database", "identifier", imageIdentifier, "err", err)
			} else {
				slog.Info("Image stored in database successfully", "identifier", imageIdentifier)
			}
		}

		images = nil

		return nil
	}).Error

	if err != nil {
		slog.Error("Error processing batches", "err", err)
	} else {
		slog.Info("All batches processed successfully")
	}
}

func downloadImage(imageURL string) ([]byte, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		slog.Error("Failed to fetch image", "err", err)
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("Error closing response body", "err", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Bad status", "status", resp.Status)
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read image data", "err", err)
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	return imageData, nil
}

func storeImageInDatabase(originalFileName, fileExtension string, imageData []byte, imageIdentifier string, createdAt time.Time, updatedAt time.Time) error {
	uploadedImage := &model.NewImageTable{
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		ImageIdentifier:  imageIdentifier,
		OriginalFileName: originalFileName,
		ImageType:        fileExtension,
		ImageFileData:    imageData,
	}

	err := model.NewDB.Create(&uploadedImage).Error
	if err != nil {
		slog.Error("Database cannot store the image", "identifier", imageIdentifier, "err", err)
		return err
	}

	return nil
}

// ------------------------------------------------------- original -------------------------------------------------------
// package main
//
// import (
// 	"context"
// 	"fmt"
// 	"github.com/opentreehole/backend/model"
// 	"io"
// 	"log"
// 	"log/slog"
// 	"net/http"
// 	"path/filepath"
// 	"strings"
// 	"time"
// )
//
// func main() {
// 	model.Init()
// 	var images []model.OriginalImageTable
// 	result := model.OriginalDB.Find(&images)
// 	slog.Info("find all original images successfully")
//
// 	if result.Error != nil {
// 		log.Fatal(result.Error)
// 	}
//
// 	for _, image := range images {
// 		imageURL := "https://pic.jingyijun.xyz:8443/i/" + image.Path + "/" + image.Name
// 		fmt.Println("Downloading image from:", imageURL)
//
// 		imageData, err := downloadImage(imageURL)
// 		if err != nil {
// 			log.Printf("Error downloading image: %v\n", err)
// 			continue
// 		}
// 		fmt.Println("Image downloaded successfully.")
//
// 		originalFileName := image.OriginName // 用户上传的文件名
// 		imageFullName := image.Name          // 66f2cbaf9c143.png
// 		imageIdentifier := strings.TrimSuffix(imageFullName, filepath.Ext(imageFullName))
// 		fileExtension := strings.TrimPrefix(filepath.Ext(imageFullName), ".")
//
// 		createdAt := image.CreatedAt
// 		updatedAt := image.UpdatedAt
//
// 		// -------------------------------------------------------
//
// 		err = storeImageInDatabase(originalFileName, fileExtension, imageData, imageIdentifier, createdAt, updatedAt)
// 		if err != nil {
// 			log.Printf("Error storing image in database: %v\n", err)
// 		} else {
// 			fmt.Println("Image stored in database successfully.")
// 		}
//
// 	}
// }
//
// func downloadImage(imageURL string) ([]byte, error) {
//
// 	resp, err := http.Get(imageURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch image: %w", err)
// 	}
// 	defer func() {
// 		if err := resp.Body.Close(); err != nil {
// 			log.Println("Error closing response body:", err)
// 		}
// 	}()
//
// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("bad status: %s", resp.Status)
// 	}
//
// 	// 读取图片数据
// 	imageData, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read image data: %w", err)
// 	}
//
// 	return imageData, nil
// }
//
// func storeImageInDatabase(originalFileName, fileExtension string, imageData []byte, imageIdentifier string, createdAt time.Time, updatedAt time.Time) error {
//
// 	uploadedImage := &model.NewImageTable{
// 		CreatedAt:       createdAt,
// 		UpdatedAt:       updatedAt,
// 		ImageIdentifier: imageIdentifier,
// 		// 待替换
// 		OriginalFileName: originalFileName,
// 		ImageType:        fileExtension,
// 		ImageFileData:    imageData,
// 	}
//
// 	err := model.NewDB.Create(&uploadedImage).Error
// 	if err != nil {
// 		slog.LogAttrs(context.Background(), slog.LevelError, "Database cannot store the image",
// 			slog.String("err", err.Error()), slog.String("fileName", originalFileName))
// 		return err
// 	}
//
// 	return nil
// }
