package xkcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"yadro.com/course/update/core"
)

type Client struct {
	log    *slog.Logger
	client http.Client
	url    string
}

type xkcdResponse struct {
	Num        int    `json:"num"`
	Transcript string `json:"transcript"`
	Alt        string `json:"alt"`
	Img        string `json:"img"`
	Title      string `json:"title"`
}

func NewClient(url string, timeout time.Duration, log *slog.Logger) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("empty base url specified")
	}
	return &Client{
		client: http.Client{Timeout: timeout},
		log:    log,
		url:    url,
	}, nil
}

func (c Client) Get(ctx context.Context, id int) (core.XKCDInfo, error) {
	url := fmt.Sprintf("%s/%d/info.0.json", c.url, id)

	info, err := c.fetch(ctx, url)
	if err != nil {
		return core.XKCDInfo{}, err
	}

	return core.XKCDInfo{
		ID:          info.Num,
		URL:         info.Img,
		Title:       info.Title,
		Description: info.Transcript + " " + info.Alt,
	}, nil
}

func (c Client) LastID(ctx context.Context) (int, error) {
	url := fmt.Sprintf("%s/info.0.json", c.url)

	info, err := c.fetch(ctx, url)
	if err != nil {
		return 0, err
	}
	return info.Num, nil
}

func (c Client) fetch(ctx context.Context, url string) (xkcdResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return xkcdResponse{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return xkcdResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return xkcdResponse{}, core.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return xkcdResponse{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var info xkcdResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return xkcdResponse{}, err
	}
	return info, nil
}
