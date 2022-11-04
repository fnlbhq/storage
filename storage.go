package storage

type Credentials struct {
	Key      string
	Secret   string
	Endpoint string
	Region   string // https://github.com/aws/aws-sdk-go/issues/2232
}

type Provider interface {
	GetBucket(name string) (Bucket, error)
	DeleteBucket(name string) error
}

type Bucket interface {
	Name() string
	Keys() ([]string, error)
	Save(object *Object) error
	Get(key string) (*Object, error)
	Delete(key string) error
	Move(originKey, destinationKey string) error
	DownloadURL(key string) (string, error)
	DefaultACL() string
}

type Object struct {
	Bucket, Key string
	Payload     []byte
	MetaData    map[string]*string
	ACL         string
}
