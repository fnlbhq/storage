package storage

import (
	"fmt"
	"os"
	"testing"
)

const (
	_bucket = "test-1125197802232007"
	key     = "primary_doc.xml"
	url     = "https://www.sec.gov/Archives/edgar/data/924628/000175272420220545/primary_doc.xml"
)

//func TestClient_List(t *testing.T) {
//	credentials := Credentials{
//		Key:      os.Getenv("SPACES_KEY"),
//		Secret:   os.Getenv("SPACES_SECRET"),
//		Endpoint: "https://" + "nyc3.digitaloceanspaces.com",
//		Region:   "nyc3",
//	}
//	client := New(DigitalOceanSpaces, credentials)
//	client.List("fnlb")
//}

func TestClient_SaveURL(t *testing.T) {
	//headers := map[string]string{
	//	"User-Agent":      fmt.Sprintf("%s %s", "test", "test@test.com"),
	//	"Accept-Encoding": "gzip, deflate",
	//	"Host":            "www.sec.gov",
	//}
	credentials := Credentials{
		Key:      os.Getenv("SPACES_KEY"),
		Secret:   os.Getenv("SPACES_SECRET"),
		Endpoint: "https://" + "nyc3.digitaloceanspaces.com",
		Region:   "us-east-1", // https://github.com/aws/aws-sdk-go/issues/2232
	}
	digitalOceanSpaces := AmazonS3(credentials)

	testBucket, err := digitalOceanSpaces.GetBucket("foobarbaz69696")
	if err != nil {
		panic(err)
	}

	err = testBucket.Save(&Object{
		Bucket:  "",
		Key:     "wakawaka2",
		Payload: []byte("Hello World!"),

		ACL: "private",
	})
	if err != nil {
		panic(err)
	}

	keys, _ := testBucket.Keys()
	for _, key := range keys {
		fmt.Println(key)
		if err = testBucket.Delete(key); err != nil {
			panic(err)
		}
	}

	err = digitalOceanSpaces.DeleteBucket("foobarbaz69696")
	if err != nil {
		panic(err)
	}

	newBucket, _ := digitalOceanSpaces.GetBucket(_bucket)
	//url, _ := newBucket.DownloadURL(key)

	if err = newBucket.Move(key, "primary_key2.xml"); err != nil {
		panic(err)
	}

	if err = newBucket.Move(key, "primary_key2.xml"); err != nil {
		panic(err)
	}
}
