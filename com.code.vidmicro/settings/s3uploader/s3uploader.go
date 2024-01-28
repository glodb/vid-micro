package s3uploader

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sync"

	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/xid"
)

var (
	instance *S3Uploader
	once     sync.Once
)

type S3Uploader struct {
}

func GetInstance() *S3Uploader {
	once.Do(func() {
		instance = &S3Uploader{}
	})
	return instance
}

func (u *S3Uploader) UploadToSCW(fileHeader *multipart.FileHeader) (string, int, error) {

	ext := filepath.Ext(fileHeader.Filename)

	if !configmanager.GetInstance().AllowedExtensions[ext] {
		return "", http.StatusUnsupportedMediaType, errors.New("the image extension is not allowed")
	}

	if fileHeader.Size > (int64(configmanager.GetInstance().AllowedSizeInMbs) * 1 << 20) {
		return "", http.StatusRequestEntityTooLarge, errors.New("file too big")
	}

	ctx := context.Background()

	minioClient, err := minio.New(configmanager.GetInstance().S3Settings.EndPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(configmanager.GetInstance().S3Settings.AccessKey, configmanager.GetInstance().S3Settings.SecretKey, ""),
		Secure: true,
	})
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	defer file.Close()

	xidString := xid.New().String() + ext
	objectName := filepath.Join(configmanager.GetInstance().S3Settings.Folder, xidString)

	_, err = minioClient.PutObject(ctx, configmanager.GetInstance().S3Settings.Bucket, objectName, file, fileHeader.Size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	link := fmt.Sprintf("https://%s/%s/%s", configmanager.GetInstance().S3Settings.EndPoint, configmanager.GetInstance().S3Settings.Bucket, objectName)
	return link, http.StatusOK, nil
}
