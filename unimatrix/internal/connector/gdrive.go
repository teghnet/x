// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/teghnet/x/unimatrix/internal/model"
)

const gdriveBaseURL = "https://www.googleapis.com/drive/v3"

// GDrive is a connector for Google Drive.
type GDrive struct {
	name         string
	token        string
	client       *http.Client
	rootFolderID string
}

// NewGDrive creates a new Google Drive connector.
// Token should be an OAuth2 access token.
func NewGDrive(name, token string) *GDrive {
	if token == "" {
		token = os.Getenv("GDRIVE_ACCESS_TOKEN")
	}
	return &GDrive{
		name:   name,
		token:  token,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Name implements Connector.
func (g *GDrive) Name() string {
	return g.name
}

// Connect implements Connector.
func (g *GDrive) Connect(ctx context.Context) error {
	if g.token == "" {
		return fmt.Errorf("gdrive: access token not configured")
	}

	// Test connection
	req, err := g.newRequest(ctx, "GET", "/about?fields=user", nil)
	if err != nil {
		return err
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("gdrive: connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gdrive: auth failed with status %d", resp.StatusCode)
	}

	return nil
}

// Close implements Connector.
func (g *GDrive) Close() error {
	return nil
}

// List implements Connector.
func (g *GDrive) List(ctx context.Context, path string) ([]model.Node, error) {
	folderID := path
	if folderID == "" {
		folderID = "root"
	}

	query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	fields := "files(id,name,mimeType,size,modifiedTime)"

	url := fmt.Sprintf("/files?q=%s&fields=%s", query, fields)
	req, err := g.newRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result driveFilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	nodes := make([]model.Node, 0, len(result.Files))
	for _, file := range result.Files {
		nodes = append(nodes, g.fileToNode(file))
	}

	return nodes, nil
}

// Read implements Connector.
func (g *GDrive) Read(ctx context.Context, node model.Node) (io.ReadCloser, error) {
	// Check if it's a Google Doc (needs export)
	if mimeType, ok := node.Metadata["mimeType"].(string); ok {
		if strings.HasPrefix(mimeType, "application/vnd.google-apps.") {
			return g.exportFile(ctx, node.ID, mimeType)
		}
	}

	// Regular file download
	req, err := g.newRequest(ctx, "GET", "/files/"+node.ID+"?alt=media", nil)
	if err != nil {
		return nil, err
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("gdrive: download failed with status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (g *GDrive) exportFile(ctx context.Context, fileID, mimeType string) (io.ReadCloser, error) {
	// Determine export format
	exportMime := "text/plain"
	switch mimeType {
	case "application/vnd.google-apps.document":
		exportMime = "text/markdown"
	case "application/vnd.google-apps.spreadsheet":
		exportMime = "text/csv"
	case "application/vnd.google-apps.presentation":
		exportMime = "application/pdf"
	}

	url := fmt.Sprintf("/files/%s/export?mimeType=%s", fileID, exportMime)
	req, err := g.newRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("gdrive: export failed with status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// Write implements Connector.
func (g *GDrive) Write(ctx context.Context, node model.Node, r io.Reader) error {
	if r == nil {
		// Create folder
		return g.createFolder(ctx, node)
	}

	// Upload file
	return g.uploadFile(ctx, node, r)
}

func (g *GDrive) createFolder(ctx context.Context, node model.Node) error {
	metadata := map[string]any{
		"name":     node.Name,
		"mimeType": "application/vnd.google-apps.folder",
	}

	parent := filepath.Dir(node.Path)
	if parent != "" && parent != "." {
		metadata["parents"] = []string{parent}
	}

	body, _ := json.Marshal(metadata)
	req, err := g.newRequest(ctx, "POST", "/files", strings.NewReader(string(body)))
	if err != nil {
		return err
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gdrive: create folder failed with status %d", resp.StatusCode)
	}

	return nil
}

func (g *GDrive) uploadFile(ctx context.Context, node model.Node, r io.Reader) error {
	// Simplified upload - for real implementation use resumable upload
	url := "/upload/drive/v3/files?uploadType=media"
	req, err := http.NewRequestWithContext(ctx, "POST", gdriveBaseURL+url, r)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gdrive: upload failed with status %d", resp.StatusCode)
	}

	return nil
}

// Delete implements Connector.
func (g *GDrive) Delete(ctx context.Context, node model.Node) error {
	req, err := g.newRequest(ctx, "DELETE", "/files/"+node.ID, nil)
	if err != nil {
		return err
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gdrive: delete failed with status %d", resp.StatusCode)
	}

	return nil
}

// Stat implements Connector.
func (g *GDrive) Stat(ctx context.Context, path string) (*model.Node, error) {
	fields := "id,name,mimeType,size,modifiedTime"
	req, err := g.newRequest(ctx, "GET", "/files/"+path+"?fields="+fields, nil)
	if err != nil {
		return nil, err
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, os.ErrNotExist
	}

	var file driveFile
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return nil, err
	}

	node := g.fileToNode(file)
	return &node, nil
}

func (g *GDrive) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	url := gdriveBaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (g *GDrive) fileToNode(file driveFile) model.Node {
	nodeType := model.FileNode
	if file.MimeType == "application/vnd.google-apps.folder" {
		nodeType = model.FolderNode
	}

	modTime, _ := time.Parse(time.RFC3339, file.ModifiedTime)

	return model.Node{
		ID:        file.ID,
		Path:      file.ID,
		Name:      file.Name,
		Type:      nodeType,
		Size:      file.Size,
		ModTime:   modTime,
		Connector: g.name,
		Metadata: map[string]any{
			"mimeType": file.MimeType,
		},
	}
}

// Google Drive API response types

type driveFilesResponse struct {
	Files []driveFile `json:"files"`
}

type driveFile struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	MimeType     string `json:"mimeType"`
	Size         int64  `json:"size"`
	ModifiedTime string `json:"modifiedTime"`
}
