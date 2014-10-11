package main

import (
	"bytes"
	_ "encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	_ "strings"
)

func connectPOST(apiURL string, resource string, data url.Values, headerMap map[string]string) (string, http.Header, error) {

	u, _ := url.ParseRequestURI(apiURL)
	u.Path = resource
	urlStr := fmt.Sprintf("%v", u)

	var encodedData *bytes.Buffer

	client := &http.Client{}

	if data != nil {
		encodedData = bytes.NewBufferString(data.Encode())
	} else {
		encodedData = nil
	}

	req, _ := http.NewRequest("POST", urlStr, encodedData)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if headerMap != nil {
		for k, v := range headerMap {
			req.Header.Add(k, v)
		}
	}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	bodystring := string(body)
	if resp.StatusCode != 200 {
		return bodystring, resp.Header, fmt.Errorf("%s", resp.Status)
	}
	return bodystring, resp.Header, nil
}

func newfileUploadRequest(uri string, paramName, mypath string) (*http.Request, error) {
	file, err := os.Open(mypath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, path.Base(mypath))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	contentType := writer.FormDataContentType()
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	//fmt.Printf("%s\n",body)
	fmt.Printf("%s\n\n", contentType)
	req, err := http.NewRequest("POST", uri, body)
	req.Header.Add("Content-Type", contentType)

	return req, err
}

func main() {

	dummyfilePath := "./test.text"

	authAPIURL := "https://www.google.com"
	authResource := "/accounts/ClientLogin"

	authData := url.Values{}
	authData.Add("Email", "test@gmail.com")
	authData.Add("Passwd", "password")
	authData.Add("accountType", "GOOGLE")
	authData.Add("service", "writely")
	authData.Add("source", "cURL")

	v, _, err := connectPOST(authAPIURL, authResource, authData, nil)

	var authkey string
	if err == nil {
		re := regexp.MustCompile("Auth=(.*)")
		authkey = re.FindStringSubmatch(v)[1]
	} else {
		return
	}

	authToken := fmt.Sprintf("GoogleLogin auth=%s", authkey)

	mapLocation := map[string]string{
		"Content-Length": "0",
		"Authorization":  authToken,
		"GData-Version":  "3.0",
		"Slug":           dummyfilePath,
	}

	uploadURL := "https://docs.google.com"
	uploadResource := "/feeds/upload/create-session/default/private/full"

	_, mapRes, err := connectPOST(uploadURL, uploadResource, authData, mapLocation)

	if err == nil {
		uploadLocation := mapRes.Get("Location")
		fmt.Printf("%s\n", uploadLocation)

		request, err := newfileUploadRequest(uploadLocation, "file", dummyfilePath)
		if err != nil {
			log.Fatal(err)
		}
		client := &http.Client{}
		request.Header.Add("Authorization", authToken)
		request.Header.Add("GData-Version", "3.0")
		request.Header.Add("Slug", dummyfilePath)
		resp, err := client.Do(request)
		if err != nil {
			log.Fatal(err)
		} else {
			body := &bytes.Buffer{}
			_, err := body.ReadFrom(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			resp.Body.Close()
			fmt.Println(resp.StatusCode)
			fmt.Println(resp.Header)
			fmt.Println(body)
		}
	}

}
