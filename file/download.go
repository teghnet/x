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
