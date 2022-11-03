package storage

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
)

// Zip is a convenience function that takes a map of filenames and file contents
// and wraps it into a single zipfile, returning the zipped file's byte slice
func Zip(files map[string][]byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	for filename, content := range files {
		zipFile, err := zipWriter.Create(filename)
		if err != nil {
			return nil, err
		}
		_, err = zipFile.Write(content)
		if err != nil {
			return nil, err
		}
	}
	if err := zipWriter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnZip is a convenience function that takes a zipfile, unzips it and returns
// a map of the contents' filenames and contents as a byte slice
func UnZip(data []byte) (map[string][]byte, error) {
	result := make(map[string][]byte)
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	for _, contentFile := range zipReader.File {
		f, err := contentFile.Open()
		if err != nil {
			return nil, err
		}
		contentBytes, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}
		result[contentFile.Name] = contentBytes
		f.Close()
	}
	return result, nil
}
