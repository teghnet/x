// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/teghnet/x/unimatrix/internal/model"
)

const notionAPIVersion = "2022-06-28"
const notionBaseURL = "https://api.notion.com/v1"

// Notion is a connector for Notion workspaces.
type Notion struct {
	name   string
	token  string
	client *http.Client
}

// NewNotion creates a new Notion connector.
// Token can be provided directly or via NOTION_API_TOKEN env var.
func NewNotion(name, token string) *Notion {
	if token == "" {
		token = os.Getenv("NOTION_API_TOKEN")
	}
	return &Notion{
		name:   name,
		token:  token,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Name implements Connector.
func (n *Notion) Name() string {
	return n.name
}

// Connect implements Connector.
func (n *Notion) Connect(ctx context.Context) error {
	if n.token == "" {
		return fmt.Errorf("notion: API token not configured")
	}

	// Test connection by getting user info
	req, err := n.newRequest(ctx, "GET", "/users/me", nil)
	if err != nil {
		return err
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("notion: connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notion: auth failed with status %d", resp.StatusCode)
	}

	return nil
}

// Close implements Connector.
func (n *Notion) Close() error {
	return nil
}

// List implements Connector.
// For Notion, path is a database ID or empty for root pages.
func (n *Notion) List(ctx context.Context, path string) ([]model.Node, error) {
	if path == "" {
		return n.listRootPages(ctx)
	}
	return n.queryDatabase(ctx, path)
}

func (n *Notion) listRootPages(ctx context.Context) ([]model.Node, error) {
	body := map[string]any{
		"filter": map[string]any{
			"property": "object",
			"value":    "page",
		},
		"page_size": 100,
	}

	jsonBody, _ := json.Marshal(body)
	req, err := n.newRequest(ctx, "POST", "/search", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	nodes := make([]model.Node, 0, len(result.Results))
	for _, item := range result.Results {
		node := n.itemToNode(item)
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (n *Notion) queryDatabase(ctx context.Context, dbID string) ([]model.Node, error) {
	req, err := n.newRequest(ctx, "POST", "/databases/"+dbID+"/query", nil)
	if err != nil {
		return nil, err
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result queryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	nodes := make([]model.Node, 0, len(result.Results))
	for _, item := range result.Results {
		node := n.itemToNode(item)
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// Read implements Connector.
// Returns the page content as markdown.
func (n *Notion) Read(ctx context.Context, node model.Node) (io.ReadCloser, error) {
	// Get page blocks
	req, err := n.newRequest(ctx, "GET", "/blocks/"+node.ID+"/children", nil)
	if err != nil {
		return nil, err
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result blocksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert blocks to markdown
	var md strings.Builder
	md.WriteString("# " + node.Name + "\n\n")

	for _, block := range result.Results {
		md.WriteString(n.blockToMarkdown(block))
	}

	return io.NopCloser(strings.NewReader(md.String())), nil
}

// Write implements Connector.
func (n *Notion) Write(ctx context.Context, node model.Node, r io.Reader) error {
	// TODO: Implement page creation/update
	return fmt.Errorf("notion: write not yet implemented")
}

// Delete implements Connector.
func (n *Notion) Delete(ctx context.Context, node model.Node) error {
	// Archive the page (Notion doesn't truly delete)
	body := map[string]any{"archived": true}
	jsonBody, _ := json.Marshal(body)

	req, err := n.newRequest(ctx, "PATCH", "/pages/"+node.ID, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notion: delete failed with status %d", resp.StatusCode)
	}

	return nil
}

// Stat implements Connector.
func (n *Notion) Stat(ctx context.Context, path string) (*model.Node, error) {
	req, err := n.newRequest(ctx, "GET", "/pages/"+path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, os.ErrNotExist
	}

	var page notionPage
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, err
	}

	node := n.pageToNode(page)
	return &node, nil
}

func (n *Notion) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, notionBaseURL+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+n.token)
	req.Header.Set("Notion-Version", notionAPIVersion)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (n *Notion) itemToNode(item searchItem) model.Node {
	nodeType := model.FileNode
	if item.Object == "database" {
		nodeType = model.FolderNode
	}

	name := "Untitled"
	if len(item.Properties.Title.Title) > 0 {
		name = item.Properties.Title.Title[0].PlainText
	}

	modTime, _ := time.Parse(time.RFC3339, item.LastEditedTime)

	return model.Node{
		ID:        item.ID,
		Path:      item.ID,
		Name:      name,
		Type:      nodeType,
		ModTime:   modTime,
		Connector: n.name,
		Metadata: map[string]any{
			"object": item.Object,
			"url":    item.URL,
		},
	}
}

func (n *Notion) pageToNode(page notionPage) model.Node {
	name := "Untitled"
	if len(page.Properties.Title.Title) > 0 {
		name = page.Properties.Title.Title[0].PlainText
	}

	modTime, _ := time.Parse(time.RFC3339, page.LastEditedTime)

	return model.Node{
		ID:        page.ID,
		Path:      page.ID,
		Name:      name,
		Type:      model.FileNode,
		ModTime:   modTime,
		Connector: n.name,
	}
}

func (n *Notion) blockToMarkdown(block notionBlock) string {
	switch block.Type {
	case "paragraph":
		return n.richTextToMarkdown(block.Paragraph.RichText) + "\n\n"
	case "heading_1":
		return "# " + n.richTextToMarkdown(block.Heading1.RichText) + "\n\n"
	case "heading_2":
		return "## " + n.richTextToMarkdown(block.Heading2.RichText) + "\n\n"
	case "heading_3":
		return "### " + n.richTextToMarkdown(block.Heading3.RichText) + "\n\n"
	case "bulleted_list_item":
		return "- " + n.richTextToMarkdown(block.BulletedListItem.RichText) + "\n"
	case "numbered_list_item":
		return "1. " + n.richTextToMarkdown(block.NumberedListItem.RichText) + "\n"
	case "code":
		return "```" + block.Code.Language + "\n" + n.richTextToMarkdown(block.Code.RichText) + "\n```\n\n"
	case "quote":
		return "> " + n.richTextToMarkdown(block.Quote.RichText) + "\n\n"
	case "divider":
		return "---\n\n"
	default:
		return ""
	}
}

func (n *Notion) richTextToMarkdown(texts []richText) string {
	var sb strings.Builder
	for _, t := range texts {
		text := t.PlainText
		if t.Annotations.Bold {
			text = "**" + text + "**"
		}
		if t.Annotations.Italic {
			text = "*" + text + "*"
		}
		if t.Annotations.Code {
			text = "`" + text + "`"
		}
		sb.WriteString(text)
	}
	return sb.String()
}

// Notion API response types

type searchResponse struct {
	Results []searchItem `json:"results"`
}

type searchItem struct {
	ID             string         `json:"id"`
	Object         string         `json:"object"`
	LastEditedTime string         `json:"last_edited_time"`
	URL            string         `json:"url"`
	Properties     itemProperties `json:"properties"`
}

type itemProperties struct {
	Title titleProperty `json:"title"`
}

type titleProperty struct {
	Title []richText `json:"title"`
}

type richText struct {
	PlainText   string      `json:"plain_text"`
	Annotations annotations `json:"annotations"`
}

type annotations struct {
	Bold   bool `json:"bold"`
	Italic bool `json:"italic"`
	Code   bool `json:"code"`
}

type queryResponse struct {
	Results []searchItem `json:"results"`
}

type notionPage struct {
	ID             string         `json:"id"`
	LastEditedTime string         `json:"last_edited_time"`
	Properties     itemProperties `json:"properties"`
}

type blocksResponse struct {
	Results []notionBlock `json:"results"`
}

type notionBlock struct {
	Type             string           `json:"type"`
	Paragraph        blockContent     `json:"paragraph,omitempty"`
	Heading1         blockContent     `json:"heading_1,omitempty"`
	Heading2         blockContent     `json:"heading_2,omitempty"`
	Heading3         blockContent     `json:"heading_3,omitempty"`
	BulletedListItem blockContent     `json:"bulleted_list_item,omitempty"`
	NumberedListItem blockContent     `json:"numbered_list_item,omitempty"`
	Code             codeBlockContent `json:"code,omitempty"`
	Quote            blockContent     `json:"quote,omitempty"`
}

type blockContent struct {
	RichText []richText `json:"rich_text"`
}

type codeBlockContent struct {
	RichText []richText `json:"rich_text"`
	Language string     `json:"language"`
}
