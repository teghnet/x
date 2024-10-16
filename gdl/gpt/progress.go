package main

import (
	"fmt"
	"io"
)

// ProgressWriter tracks the progress of the download or upload
type ProgressWriter struct {
	Total    int64
	Progress int64
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Progress += int64(n)
	fmt.Printf("\rDownloading: %.2f%% complete", float64(pw.Progress)/float64(pw.Total)*100)
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
	fmt.Printf("\rUploading: %.2f%% complete", float64(pr.Progress)/float64(pr.Total)*100)
	return n, err
}
