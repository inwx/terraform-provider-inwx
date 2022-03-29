package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

const (
	COMMAND_SUCCESSFUL         float64 = 1000
	COMMAND_SUCCESSFUL_PENDING float64 = 1001
)

type Response map[string]interface{}

func (r Response) Code() float64 {
	return r["code"].(float64)
}

func (r Response) ApiError() string {
	jsonStr, err := json.Marshal(r)
	if err != nil {
		return fmt.Sprintf("could not parse error: %w", err)
	}
	return string(jsonStr)
}

type Client struct {
	httpClient *http.Client
	logger     *logr.Logger
	BaseURL    *url.URL
	Username   string
	Password   string
	Debug      bool
}

func NewClient(username string, password string, baseURL *url.URL, logger *logr.Logger, debug bool) (*Client, error) {
	logger.V(10).Info("initializing new http client")

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("could not create http client cookie jar: %w", err))
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
		Jar: jar,
	}

	return &Client{
		httpClient,
		logger,
		baseURL,
		username,
		password,
		debug,
	}, nil
}

func (c *Client) Call(ctx context.Context, method string, parameters map[string]interface{}) (Response, error) {
	requestBody := map[string]interface{}{}
	requestBody["method"] = method
	requestBody["params"] = parameters
	requestJsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("could not marshal rpc request parameters to json: %w", err))
	}

	if c.Debug {
		fmt.Printf("Request (%s): %s", method, requestJsonBody)
		c.logger.Info(fmt.Sprintf("Request (%s): %s", method, requestJsonBody))
	}

	request, err := http.NewRequest("POST", c.BaseURL.String(), bytes.NewReader(requestJsonBody))
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("could not create rpc request: %w", err))
	}
	request = request.WithContext(ctx)
	request.Header.Set("content-type", "application/json; charset=UTF-8")

	post, err := c.httpClient.Do(request)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("could not execute rpc request: %w", err))
	}

	responseBody, err := ioutil.ReadAll(post.Body)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("could not read rpc response: %w", err))
	}
	fmt.Printf("Response (%s): %s", method, string(responseBody))

	var response map[string]interface{}
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("could not unmarshal rpc response to json: %w, %s, %s, %s", err, requestJsonBody, c.BaseURL.String(), post.Status))
	}

	// Make sure body is valid json before debug message
	if c.Debug {
		c.logger.Info(fmt.Sprintf("Request (%s): %s", method, responseBody))
	}

	return response, nil
}

func (c *Client) CallNoParams(ctx context.Context, method string) (Response, error) {
	return c.Call(ctx, method, map[string]interface{}{})
}

func (c *Client) Logout(ctx context.Context) (Response, error) {
	return c.CallNoParams(ctx, "account.logout")
}
