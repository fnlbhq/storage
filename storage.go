package storage

type Credentials struct {
	Key      string
	Secret   string
	Endpoint string
	Region   string // https://github.com/aws/aws-sdk-go/issues/2232
}

type Bucket interface {
	Name() string
	Keys() ([]string, error)
	Save(object *Object) error
	Get(key string) (*Object, error)
	Delete(key string) error
	Move(originKey, destinationKey string) error
	DownloadURL(key string) (string, error)
}

type Object struct {
	Bucket, Key string
	Payload     []byte
	MetaData    map[string]*string
	ACL         string
}
