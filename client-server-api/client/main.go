package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	serverTimeOut = 300 * time.Millisecond
	serverURL     = "http://localhost:8080/cotacao"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, serverTimeOut)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", serverURL, nil)
	if err != nil {
		log.Fatalf("error creating request: %v", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("error making request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("error reading body: %v", err)
	}

	fileCreation(body)
	fmt.Println(string(body))
}

func fileCreation(responseBody []byte) {
	var result map[string]any
	err := json.Unmarshal(responseBody, &result)
	if err != nil {
		log.Fatalf("error unmarshalling response body: %v", err)
	}
	bid := result["bid"].(string)

	// check if file exists
	_, err = os.Stat("./files/cotacao.txt")
	fileExists := !os.IsNotExist(err)

	var file *os.File
	if fileExists {
		file, err = os.OpenFile("./files/cotacao.txt", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		writer := bufio.NewWriter(file)
		fmt.Fprintf(writer, "\nDólar: {%s}", bid)
		writer.Flush()
	} else {
		file, err = os.Create("./files/cotacao.txt")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		_, err = fmt.Fprintf(file, "Dólar: {%s}", bid)
		if err != nil {
			log.Fatal(err)
		}
	}
}
