package purshase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Url string

var ErrUnknown = errors.New("unknown error")
var ErrResponse = errors.New("response error")

type Purchase struct {
	url Url
}

func (p Purchase) PurchaseList(ctx context.Context, token string, id int64) (list []PurchaseDto, err error) {
	request, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/api/purshases/%d", p.url, id),
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
	fmt.Print(response.StatusCode )
	switch response.StatusCode {
	case 200:

		err = json.Unmarshal(responseBody, &list)
		if err != nil {
			return nil, fmt.Errorf("can't decode response: %w", err)
		}
		return list, nil
	case 400:
		return nil, errors.New("error bad request")
	default:
		return nil, ErrUnknown
	}

}

func NewPurchase(url Url) *Purchase {
	return &Purchase{url: url}
}

type PurchaseDto struct {
	ID         int64     `json:"id"`
	User_id    int64     `json:"user_id"`
	Product_id int64     `json:"product_id"`
	Price      int       `json:"price"`
	Quantity   int       `json:"quantity"`
	Date       time.Time `json:"date"`
}
