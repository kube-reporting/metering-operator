package v1alpha1

type StorageLocation struct {
	S3    *S3Bucket     `json:"s3"`
	Local *LocalStorage `json:"local"`
}

type S3Bucket struct {
	Bucket string `json:"bucket"`
	Pre***REMOVED***x string `json:"pre***REMOVED***x"`
}

type LocalStorage struct{}
