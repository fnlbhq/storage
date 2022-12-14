package cloudstorage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"google.golang.org/api/iterator"

	"cloud.google.com/go/storage"
)

const Timeout = time.Second * 60

// https://pkg.go.dev/cloud.google.com/go/storage#section-readme
// https://cloud.google.com/appengine/docs/standard/go/using-cloud-storage
// https://cloud.google.com/appengine/docs/standard/go111/googlecloudstorageclient/read-write-to-cloud-storage

// Put puts a file to Google Cloud client and returns an error or nil
func Put(bucket, path string, payload []byte) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	wc := client.Bucket(bucket).Object(path).NewWriter(ctx)
	if _, err = io.Copy(wc, bytes.NewReader(payload)); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}
	log.Printf("%v saved to %s/%s\n", path, bucket, path)
	return nil
}

// Get fetches a file from Google Cloud client and returns a byte slice or an error
// https://cloud.google.com/storage/docs/samples/storage-download-file
func Get(bucket, path string) ([]byte, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	rc, err := client.Bucket(bucket).Object(path).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	log.Printf("%v retrieved from %s/%s\n", path, bucket, path)
	return data, nil
}

// Delete deletes a file from Google Cloud client
func Delete(bucket, path string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	o := client.Bucket(bucket).Object(path)
	return o.Delete(ctx)
}

// Exists returns true if a file exists at the bucket and path arguments. It returns false otherwise.
// If it encounters any other type of error, it panics
func Exists(bucket, path string) bool {
	ctx := context.Background()
	client, _ := storage.NewClient(ctx)
	_, err := client.Bucket(bucket).Object(path).Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false
	}
	if err != nil {
		log.Panic(err)
	}
	return true
}

// FileMetadata wraps an ObjectAttrs object
type FileMetadata struct {
	storage.ObjectAttrs
}

// FileName extracts the name of the file from the path
func (s FileMetadata) FileName() string {
	nameArray := strings.Split(s.Name, "/")
	return nameArray[len(nameArray)-1]
}

// Get fetches the referenced file from cloud storage
func (s FileMetadata) Get() ([]byte, error) {
	return Get(s.Bucket, s.Name)
}

// FilesAtPath returns a slice of StorageObjects that contains metadata about each object. An optional
// filter function can be passed in which case the results will be filtered according to the rules defined in the function
// https://cloud.google.com/storage/docs/samples/storage-list-files-with-prefix#storage_list_files_with_prefix-go
func FilesAtPath(bucket, path string, filter ...func(object FileMetadata) bool) ([]FileMetadata, error) {
	var result []FileMetadata
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	it := client.Bucket(bucket).Objects(ctx, &storage.Query{Prefix: path})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("bucket(%q).Objects: %v", bucket, err)
		}
		storageObject := FileMetadata{*attrs}
		if storageObject.FileName() != "" {
			result = append(result, storageObject)
		}
	}
	// If a filter function was passed as an argument, use it to filter the result set
	if len(filter) > 0 {
		for i, object := range result {
			// If the object does not meet the filter condition...
			if !filter[0](object) {
				// remove it: (inefficient but this approach preserves order)
				result = append(result[:i], result[i+1:]...)
			}
		}
	}
	return result, nil
}

// ProcessFile performs a non-destructive operation on a cloud storage file's data via the 'process' callback
// While it's possible to alter the file on cloud storage within the 'process' callback, there is a separate
// function, ProcessAndUpdateFile, that moves any change to the cloud storage object out of the callback
func ProcessFile(bucket, path string, process func(file []byte) error) error {
	data, err := Get(bucket, path)
	if err != nil {
		return err
	}
	return process(data)
}

// ProcessAndUpdateFile fetches a file from cloud storage, runs the 'process' callback, which must return
// a byte slice containing the updated file, or an error. This byte slice is then put to cloud storage,
// replacing the original object.
func ProcessAndUpdateFile(bucket, path string, process func(file []byte) ([]byte, error)) error {
	data, err := Get(bucket, path)
	if err != nil {
		return err
	}
	processedData, err := process(data)
	if err != nil {
		return err
	}
	return Put(bucket, path, processedData)
}

// SaveNetworkFile fetches a file at a target url and puts it to the Google Cloud client destination
// bucket and path. it returns a byte slice or nil and an error or nil.
func SaveNetworkFile(targetUrl, destinationBucket, destinationPath string, headers map[string]string) ([]byte, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", targetUrl, nil)
	if headers != nil {
		for header, value := range headers {
			req.Header.Add(header, value)
		}
	}
	if err != nil {
		return nil, err
	}
	result, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}
	return body, Put(destinationBucket, destinationPath, body)
}
