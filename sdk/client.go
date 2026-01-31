package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client SDK客户端
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient 创建新的SDK客户端
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// post 发送POST请求
func (c *Client) post(ctx context.Context, path string, reqBody interface{}, respBody interface{}) error {
	// 从ctx中获取自定义path
	if customPath, ok := ctx.Value("path").(string); ok && customPath != "" {
		path = customPath
	}

	url := fmt.Sprintf("%s%s", c.BaseURL, path)

	// 序列化请求体
	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 从ctx中获取自定义header
	if headers, ok := ctx.Value("headers").(map[string]string); ok {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应体失败: %v", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 反序列化响应体
	if respBody != nil {
		if err := json.Unmarshal(body, respBody); err != nil {
			return fmt.Errorf("反序列化响应体失败: %v", err)
		}
	}

	return nil
}
