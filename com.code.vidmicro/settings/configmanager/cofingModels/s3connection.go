package configModels

type S3Connection struct {
	EndPoint  string `json:"endpoint"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Bucket    string `json:"bucket"`
	Folder    string `json:"folder"`
}
