package v1alpha1

type StorageLocation struct {
	S3    *S3Bucket     `json:"s3"`
	Local *LocalStorage `json:"local"`
}

type S3Bucket struct {
	Bucket string `json:"bucket"`
	Prefix string `json:"prefix"`
}

type LocalStorage struct{}
