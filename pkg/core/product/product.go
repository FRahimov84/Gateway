package product

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Url string


type ResponseProduct struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	Pic         string `json:"pic"`
}


var ErrUnknown = errors.New("unknown error")
var ErrResponse = errors.New("response error")

type Product struct {
	url Url
}

func NewProduct(url Url) *Product {
	return &Product{url: url}
}

func (c *Product) ProductList(ctx context.Context, token string) (list []ResponseProduct, err error) {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/products", c.url),
		bytes.NewBuffer(nil),
	)
	if err != nil {
		return nil, fmt.Errorf("can't create request: %w", err)
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("can't send request: %w", err)
	}
	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("can't parse response: %w", err)
	}

	switch response.StatusCode {
	case 200:
		var responseData []ResponseProduct
		err = json.Unmarshal(responseBody, &responseData)
		if err != nil {
			return nil, fmt.Errorf("can't decode response: %w", err)
		}
		return responseData, nil
	case 400:
		return nil, errors.New("error bad request")
	default:
		return nil, ErrUnknown
	}

}

func (c *Product) NewProduct(ctx context.Context, prod ResponseProduct, token string) (err error) {
	byte, err := json.Marshal(prod)
	if err != nil {
		return
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/products/0", c.url),
		bytes.NewBuffer(byte),
	)
	if err != nil {
		return fmt.Errorf("can't create request: %w", err)
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("can't send request: %w", err)
	}
	defer response.Body.Close()
	log.Print(http.StatusText(response.StatusCode))
	switch response.StatusCode {
	case 204:
		return  nil
	case 400:
		return errors.New("error bad request")
	default:
		return ErrUnknown
	}

}

func (c *Product) UpdateProduct(ctx context.Context, prod ResponseProduct, token string, id int) (err error) {
	byte, err := json.Marshal(prod)
	if err != nil {
		return
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/products/%d", c.url, id),
		bytes.NewBuffer(byte),
	)
	if err != nil {
		return fmt.Errorf("can't create request: %w", err)
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("can't send request: %w", err)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case 204:
		return  nil
	case 400:
		return errors.New("error bad request")
	default:
		return ErrUnknown
	}

}

func (c *Product) RemoveByID(ctx context.Context, token string, id int) (err error) {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		fmt.Sprintf("%s/api/products/%d", c.url, id),
		bytes.NewBuffer(nil),
	)
	if err != nil {
		return fmt.Errorf("can't create request: %w", err)
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("can't send request: %w", err)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case 204:
		return  nil
	case 400:
		return errors.New("error bad request")
	default:
		return ErrUnknown
	}
}
