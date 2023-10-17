package utils

import (
	"context"
	"errors"
	mediaErrors "github.com/ayo-ajayi/ecommerce/internal/errors"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
)

type MediaCloudManager struct {
	folder string
	cld    *cloudinary.Cloudinary
}

func NewMediaCloudManager(api_uri, folder string) (*MediaCloudManager, error) {
	cld, err := cloudinary.NewFromURL(api_uri)
	if err != nil {
		return nil, errors.New("failed to init cloudinary: " + err.Error())
	}
	return &MediaCloudManager{
		cld:    cld,
		folder: folder,
	}, nil
}

func (mcm *MediaCloudManager) UploadImage(ctx context.Context, files []*multipart.FileHeader, collection string) ([]string, *mediaErrors.AppError) {
	images := make([]string, len(files))
	errchan := make(chan *mediaErrors.AppError, len(files))
	var wg sync.WaitGroup
	for i, fileHeader := range files {
		wg.Add(1)
		go func(index int, fileHeader *multipart.FileHeader) {
			file, err := fileHeader.Open()
			if err != nil {
				errchan <- mediaErrors.NewError("failed to open file: "+fileHeader.Filename+" "+err.Error(), 400)
				return
			}
			res, err := mcm.cld.Upload.Upload(ctx, file, uploader.UploadParams{Folder: mcm.folder + "/" + collection, UniqueFilename: api.Bool(true), ResourceType: "image"})
			if err != nil {
				errchan <- mediaErrors.NewError("failed to upload file: "+fileHeader.Filename+" "+err.Error(), 400)
				return
			}
			images[index] = res.SecureURL
			file.Close()
			wg.Done()
		}(i, fileHeader)
	}
	go func() {
		wg.Wait()
		close(errchan)
	}()
	for err := range errchan {
		if err != nil {
			return nil, err
		}
	}
	return images, nil

}

func (mcm *MediaCloudManager) DeleteImageBySecureURL(ctx context.Context, secureUrl string) *mediaErrors.AppError {
	publicID, err := fetchPublicIdFromSecureUrl(secureUrl)
	if err != nil {
		return mediaErrors.NewError("failed to fetch public id: "+err.Error(), 400)
	}
	if publicID == "" {
		return mediaErrors.NewError("invalid public id", 400)
	}
	_, err = mcm.cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID, ResourceType: "image"})
	if err != nil {
		return mediaErrors.NewError("failed to delete image: "+err.Error(), 400)
	}
	return nil
}

func fetchPublicIdFromSecureUrl(secureUrl string) (string, error) {
	parsedURL, err := url.Parse(secureUrl)
	if err != nil {
		return "", errors.New("failed to parse url: " + secureUrl)
	}
	segments := strings.Split(parsedURL.Path, "/")
	if len(segments) < 2 {
		return "", errors.New("invalid url: " + secureUrl)
	}
	publicID := segments[len(segments)-1]
	publicID = strings.TrimSuffix(publicID, filepath.Ext(publicID))
	return publicID, nil
}
