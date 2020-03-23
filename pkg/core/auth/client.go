package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Url string

type TokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

// errors are part API
var ErrUnknown = errors.New("unknown error")
var ErrResponse = errors.New("response error")

type ErrorResponse struct {
	Errors []string `json:"errors"`
}

func (e *ErrorResponse) Error() string {
	return strings.Join(e.Errors, ", ")
}

// for errors.Is
func (e *ErrorResponse) Unwrap() error {
	return ErrResponse
}

type Client struct {
	url Url
}

func (c *Client) Profile(ctx context.Context, token string) (user UserResponseDTO, err error) {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/users", c.url),
		bytes.NewBuffer(nil),
	)
	if err != nil {
		err = fmt.Errorf("can't create request: %w", err)
		return
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		err = fmt.Errorf("can't send request: %w", err)
		return
	}
	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("can't parse response: %w", err)
		return
	}
	switch response.StatusCode {
	case 200:
		err = json.Unmarshal(responseBody, &user)
		if err != nil {
			err = fmt.Errorf("can't decode response: %w", err)
			return
		}
		return
	case 400:
		err = errors.New("error bad request")
		return
	default:
		err = ErrUnknown
		return
	}
}

type UserResponseDTO struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

func NewClient(url Url) *Client {
	return &Client{url: url}
}

func (c *Client) Login(ctx context.Context, login string, password string) (token string, err error) {
	requestData := TokenRequest{
		Username: login,
		Password: password,
	}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("can't encode requestBody %v: %w", requestData, err)
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/tokens", c.url),
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return "", fmt.Errorf("can't create request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	// in other request
	// request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("can't send request: %w", err)
	}
	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("can't parse response: %w", err)
	}

	switch response.StatusCode {
	case 200:
		var responseData *TokenResponse
		err = json.Unmarshal(responseBody, &responseData)
		if err != nil {
			return "", fmt.Errorf("can't decode response: %w", err)
		}
		return responseData.Token, nil
	case 400:
		var responseData *ErrorResponse
		err = json.Unmarshal(responseBody, &responseData)
		if err != nil {
			return "", fmt.Errorf("can't decode response: %w", err)
		}
		return "", responseData
	default:
		return "", ErrUnknown
	}

}
