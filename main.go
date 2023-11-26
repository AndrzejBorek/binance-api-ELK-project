package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type RequestMessage struct {
	Time     int64           `json:"time"`
	URL      string          `json:"url"`
	Currency string          `json:"currency"`
	Error    string          `json:"error"`
	Response json.RawMessage `json:"response"`
}

func sendRequestWithJSON(url, currency string) ([]byte, error) {
	response, err := http.Get(url)
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf(err.Error())
	}
	responseBody := string(bodyBytes)
	err = response.Body.Close()
	if err != nil {
		log.Fatalf(err.Error(), "Error when closing response. Possible resource leak.")
	}

	message := RequestMessage{
		Time:     time.Now().UnixNano(),
		URL:      url,
		Currency: currency,
	}

	if err != nil {
		message.Error = err.Error()
		message.Response = nil
	} else {
		message.Error = ""
		message.Response = json.RawMessage(responseBody)
	}

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return jsonMessage, nil
}

func addParametersToUrl(apiUrl string, currencyParam string, windowSizeParam string) string {
	queryParameters := url.Values{}
	queryParameters.Set("symbol", currencyParam)
	queryParameters.Set("windowSize", windowSizeParam)
	urlWithParams := fmt.Sprintf("%s?%s", apiUrl, queryParameters.Encode())
	return urlWithParams
}

//Naprawić dodawanie parametru windowSize, tak żeby response nie był 400 ani 404

func saveJSONToFile(filename string, jsonData []byte) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf(err.Error())
		}
	}(file)

	// Convert the JSON data to bytes
	dataBytes := append(bytes.Replace(jsonData, []byte("\\u0026"), []byte("&"), -1), '\n')

	// Append the data to the file
	_, err = file.Write(dataBytes)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	//	apiUrl := "https://api.binance.com/api/v3/ticker/price"

	//currency := os.Getenv("CURRENCY")
	currency := "BTCUSDT"

	// Klucze API Binance
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	apiUrl := "https://api.binance.com/api/v3/ticker"

	for i := 1; i <= 100000; i++ {
		jsonData, err := sendRequestWithJSON(addParametersToUrl(apiUrl, currency, "1m"), currency)
		if err != nil {
			log.Printf("Error sending request: %s", err)
		} else {
			err := saveJSONToFile("./filebeat_ingest_data/logdata.log", jsonData)
			if err != nil {
				log.Printf("Error saving JSON to file: %s", err)
			}
		}
		time.Sleep(1 * time.Second)
	}
}
