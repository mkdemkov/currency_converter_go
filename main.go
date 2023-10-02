package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ResponseData struct {
	Rates map[string]float64 `json:"rates"`
}

func main() {
	http.HandleFunc("/api/list", getListOfCurrency)
	http.HandleFunc("/api/convert", convertValue)
	http.ListenAndServe(":8080", nil)
}

func convertValue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Выбранный HTTP vетод недоступен", http.StatusMethodNotAllowed)
		return
	}

	var requestBody map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		fmt.Println("Ошибка при попытке прочитать тело запроса")
		http.Error(w, "Плохой запрос", http.StatusBadRequest)
		return
	}
	fromCurrency, fromOK := requestBody["from"].(string)
	toCurrency, toOK := requestBody["to"].(string)
	amount, amountOK := requestBody["amount"].(float64)

	if !fromOK || !toOK || !amountOK {
		http.Error(w, "Неккоректное тело запроса. Убедитесь, что указанная сумма - валидное вещественное число", http.StatusBadRequest)
		return
	}

	apiUrl := fmt.Sprintf("https://open.er-api.com/v6/latest/%s", fromCurrency)
	response, err := http.Get(apiUrl)
	if err != nil {
		fmt.Println("Ошибка при выполнении запроса на внешнее API")
		http.Error(w, "Ошибка на стороне внешнего АПИ", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var responseData ResponseData
	if err := json.Unmarshal(responseBody, &responseData); err != nil {
		fmt.Println("Error decoding response body")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	rate, rateExists := responseData.Rates[toCurrency]
	if !rateExists {
		http.Error(w, "Invalid currency pair", http.StatusBadRequest)
		return
	}

	convertedAmount := amount * rate

	// Prepare the response
	responseJSON := map[string]float64{"converted_amount": convertedAmount}

	if err := json.NewEncoder(w).Encode(responseJSON); err != nil {
		fmt.Println("Error encoding JSON response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func getListOfCurrency(w http.ResponseWriter, r *http.Request) {
	apiUrl := "https://open.er-api.com/v6/latest/AED"
	response, err := http.Get(apiUrl)
	if err != nil {
		fmt.Println("Что-то пошло не так при попытке получить список доступных валют с внешнего API")
		return
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Что-то пошло не так при обработке ответа от API конвертатора валют")
		return
	}

	var responseData ResponseData
	if err := json.Unmarshal(responseBody, &responseData); err != nil {
		fmt.Println("Ошибка при чтении JSON")
		http.Error(w, "Ошибка на стороне сервера", http.StatusInternalServerError)
		return
	}

	var currencies []string
	for currency := range responseData.Rates {
		currencies = append(currencies, currency)
	}

	if err := json.NewEncoder(w).Encode(currencies); err != nil {
		fmt.Println("Error encoding JSON response")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}
