package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func piss() {
	password := "test"

	hash := sha256.Sum256([]byte(password))

	fmt.Println(hex.EncodeToString(hash[:]))

}

type Common struct {
	Collection []int
}

func lie() {
	os.Setenv("SOUNDCLOUD_TOKEN_PATH", "./soundcloud_token.json")

	tokenFile, err := os.Open(os.Getenv("SOUNDCLOUD_TOKEN_PATH"))
	if err != nil {
		panic(err)
	}

	tokenBytes, err := io.ReadAll(tokenFile)
	if err != nil {
		panic(err)
	}
	oauth := string(tokenBytes)

	fmt.Println(oauth)

	client := &http.Client{}

	request, err := http.NewRequest("GET", "https://api-v2.soundcloud.com/search?q=no&limit=5", nil)
	if err != nil {
		panic(err)
	}
	request.Header.Set("Authorization", oauth)
	// request.Header.Set("Content-type", "application/json")

	fmt.Println(request.Header)

	res, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	all, _ := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var obj Common

	json.Unmarshal(all, &obj)

	//Will need to trash the ones that don't belong

	fmt.Println(len(obj.Collection))

}
