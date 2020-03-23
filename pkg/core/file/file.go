package file

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
)

type Url string

var ErrUnknown = errors.New("unknown error")
var ErrResponse = errors.New("response error")

type File struct {
	url Url
}
type FileURL struct {
	Name string `json:"name"`
}

func (f *File) Save(ctx context.Context, byte []byte, token, fName string) (pic string, err error) {

	// Create multipart
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("file", fName) //give file a name
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = fw.Write(byte)
	if err != nil { //copy the file to the multipart buffer
		fmt.Println(err)
		return
	}
	w.Close()

	// print the head of the multipart data
	//bs := b.Bytes()
	//fmt.Printf("%+v\n\n", string(bs[:1000]))

		// Upload file
		req, err := http.NewRequestWithContext(ctx,"POST", string(f.url)+"/save", &b)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", w.FormDataContentType())
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
	//	fmt.Println(res.Status)
	//	fmt.Printf("%+v\n", res.Request)
		defer res.Body.Close()
		responseBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", fmt.Errorf("can't parse response: %w", err)
		}
		log.Print(string(responseBody))
	urls := []FileURL{}
	switch res.StatusCode {
		case 200:
			err = json.Unmarshal(responseBody, &urls)
			if err != nil {
				return "", fmt.Errorf("can't decode response: %w", err)
			}
			return urls[0].Name, nil
		case 400:
			return "", errors.New("error bad request")
		default:
			return "", ErrUnknown
		}
}

func (f *File) Serve(ctx context.Context, fromContext , token string) (byte []byte, err error){
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/media/%s", f.url, fromContext),
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
	all, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	switch response.StatusCode {
	case 200:
		return  all, nil
	case 400:
		return nil, errors.New("error bad request")
	default:
		return nil, ErrUnknown
	}
}

		//	request, err := http.NewRequestWithContext(
		//		ctx,
		//		http.MethodPost,
		//		fmt.Sprintf("%s/save", f.url),
		//		bytes.NewBuffer(byte),
		//	)
		//	if err != nil {
		//		return "", fmt.Errorf("can't create request: %w", err)
		//	}
		//	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		//	request.Header.Set("Content-Type", "multipart/form-data; boundary")
		//	request.Header.Set("Content-Disposition", `form-data; name="file"; filename="par.jpg"`)
		//	request.Header.Set("Content-Type", `text/plain`)
		//	response, err := http.DefaultClient.Do(request)
		//	if err != nil {
		//		return "", fmt.Errorf("can't send request: %w", err)
		//	}
		//	defer response.Body.Close()
		//	responseBody, err := ioutil.ReadAll(response.Body)
		//	if err != nil {
		//		return "", fmt.Errorf("can't parse response: %w", err)
		//	}
		//	url := FileURL{}
		//	switch response.StatusCode {
		//	case 200:
		//		err = json.Unmarshal(responseBody, &url)
		//		if err != nil {
		//			return "", fmt.Errorf("can't decode response: %w", err)
		//		}
		//		return url.Name, nil
		//	case 400:
		//		return "", errors.New("error bad request")
		//	default:
		//		return "", ErrUnknown
		//	}
		//}


//9308e26b-f2d5-4ec4-b42a-5ea80c466777
func NewFile(url Url) *File {
	return &File{url: url}
}
