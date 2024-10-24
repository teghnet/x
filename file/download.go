package file

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func Download(url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer CloseFile(out)

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer CloseIO(resp.Body)

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func CloseIO(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Fatal(err)
	}
}
func CloseFile(f *os.File) {
	if err := f.Close(); err != nil {
		log.Fatal(f.Name(), err)
	}
}

// ProgressWriter tracks the progress of the download or upload
type ProgressWriter struct {
	Total    int64
	Progress int64
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Progress += int64(n)
	log.Printf("\rDownloading: %.2f%% complete", float64(pw.Progress)/float64(pw.Total)*100)
	return n, nil
}

// ProgressReader tracks the progress of the upload
type ProgressReader struct {
	Reader   io.Reader
	Total    int64
	Progress int64
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Progress += int64(n)
	log.Printf("\rUploading: %.2f%% complete", float64(pr.Progress)/float64(pr.Total)*100)
	return n, err
}
