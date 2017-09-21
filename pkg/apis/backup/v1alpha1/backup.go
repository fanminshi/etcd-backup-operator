package v1alpha1

type StorageSource struct {
	S3 *S3Source `json:"s3,omitempty"`
}

type S3Source struct {
	// The name of the AWS S3 bucket to store backups in.
	//
	// S3Bucket overwrites the default etcd operator wide bucket.
	S3Bucket string `json:"s3Bucket,omitempty"`

	// Prefix is the S3 prefix used to prefix the bucket path.
	// It's the prefix at the beginning.
	// After that, it will have version and cluster specific paths.
	Prefix string `json:"prefix,omitempty"`

	// The name of the secret object that stores the AWS credential and config files.
	// The file name of the credential MUST be 'credentials'.
	// The file name of the config MUST be 'config'.
	// The profile to use in both files will be 'default'.
	//
	// AWSSecret overwrites the default etcd operator wide AWS credential and config.
	AWSSecret string `json:"awsSecret,omitempty"`
}
