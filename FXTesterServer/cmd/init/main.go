package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	data := url.Values{}
	data.Set("username", "admin")
	data.Set("password", "admin")
	data.Set("grant_type", "password")
	data.Set("client_id", "admin-cli")

	req, err := http.NewRequest("POST", "http://localhost:28080/realms/master/protocol/openid-connect/token", strings.NewReader(data.Encode()))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// レスポンスボディを読み取る
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// JSONデコード
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		panic(err)
	}

	for k, v := range result {
		fmt.Println(k, v)
	}
}
