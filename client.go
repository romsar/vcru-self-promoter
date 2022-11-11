package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const apiURL = "https://api.vc.ru/v2.31"

type Timeline struct {
	Result struct {
		Items []struct {
			Data Entry `json:"data"`
		} `json:"items"`
	} `json:"result"`
}

type Entry struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Date  int64  `json:"date"`
}

type Client struct {
	token  string
	client http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token:  token,
		client: http.Client{},
	}
}

func (c *Client) SelfPromoTimeline() (*Timeline, error) {
	resp, err := c.get("/timeline?sorting=date&hashtag=субботнийсамопиар")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := new(Timeline)
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) AddComment(id int64, text string) error {
	resp, err := c.post("/comment/add", url.Values{
		"id":   {strconv.Itoa(int(id))},
		"text": {text},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) get(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, apiURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Device-Token", c.token)

	resp, err := c.client.Do(req)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server return status `%d`", resp.StatusCode)
	}
	return resp, err
}

func (c *Client) post(endpoint string, data url.Values) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, apiURL+endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Device-Token", c.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server return status `%d`", resp.StatusCode)
	}
	return resp, err
}
