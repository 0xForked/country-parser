package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type (
	ByAbbreviation struct {
		Country      string `json:"country"`
		Abbreviation string `json:"abbreviation"`
	}

	ByContinent struct {
		Country   string `json:"country"`
		Continent string `json:"continent"`
	}

	ByCurrency struct {
		Country      string `json:"country"`
		CurrencyCode string `json:"currency_code"`
	}

	ByCallingCode struct {
		Country     string `json:"country"`
		CallingCode int    `json:"calling_code"`
	}

	Currency struct {
		ID   string `json:"_id,omitempty"`
		Code string `json:"code,omitempty"`
		Name string `json:"name,omitempty"`
	}

	CountryResult struct {
		ID          string    `json:"id"`
		Code        string    `json:"code"`
		Name        string    `json:"name"`
		Continent   string    `json:"continent"`
		CallingCode string    `json:"dial_code"`
		Currency    *Currency `json:"currency,omitempty"`
	}
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recover from panic")
		}
	}()
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		if regexp.MustCompile(`^\/panic[\/]*$`).MatchString(r.URL.Path) {
			time.Sleep(1 * time.Second)
			panic("Panic")
			return
		}

		if regexp.MustCompile(`^\/preview[\/]*$`).MatchString(r.URL.Path) {
			w.WriteHeader(http.StatusNotFound)
			abb, cont, cur, curD, callingCode := loadData()
			data := toCountryResult(abb, cont, cur, curD, callingCode)
			j, err := json.Marshal(&data)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Fprint(w, string(j))
			return
		}

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Can't find path requested"}`))
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Can't find method requested"}`))
	}
}

func loadData() (
	abb []ByAbbreviation,
	cont []ByContinent,
	cur []ByCurrency,
	curD []Currency,
	ccD []ByCallingCode,
) {
	var countriesByAbbreviation []ByAbbreviation
	countryByAbbreviation, err := os.Open("./ref/country-by-abbreviation.json")
	if err != nil {
		fmt.Println(err)
	}
	defer countryByAbbreviation.Close()
	countryByAbbreviationByteValue, _ := ioutil.ReadAll(countryByAbbreviation)
	json.Unmarshal(countryByAbbreviationByteValue, &countriesByAbbreviation)

	var countriesByContinent []ByContinent
	countryByContinent, err := os.Open("./ref/country-by-continent.json")
	if err != nil {
		fmt.Println(err)
	}
	defer countryByContinent.Close()
	countryByContinentByteValue, _ := ioutil.ReadAll(countryByContinent)
	json.Unmarshal(countryByContinentByteValue, &countriesByContinent)

	var countriesByCurrency []ByCurrency
	countryByCurrency, err := os.Open("./ref/country-by-currency-code.json")
	if err != nil {
		fmt.Println(err)
	}
	defer countryByCurrency.Close()
	countryByCurrencyByteValue, _ := ioutil.ReadAll(countryByCurrency)
	json.Unmarshal(countryByCurrencyByteValue, &countriesByCurrency)

	var currencies []Currency
	currenciesFile, err := os.Open("./ref/currency.json")
	if err != nil {
		fmt.Println(err)
	}
	defer currenciesFile.Close()
	currenciesByteValue, _ := ioutil.ReadAll(currenciesFile)
	json.Unmarshal(currenciesByteValue, &currencies)

	var callingCode []ByCallingCode
	callingCodeFile, err := os.Open("./ref/country-by-calling-code.json")
	if err != nil {
		fmt.Println(err)
	}
	defer callingCodeFile.Close()
	callingCodeByteValue, _ := ioutil.ReadAll(callingCodeFile)
	json.Unmarshal(callingCodeByteValue, &callingCode)

	return countriesByAbbreviation, countriesByContinent, countriesByCurrency, currencies, callingCode
}

func toCountryResult(
	abbArr []ByAbbreviation,
	contArr []ByContinent,
	curArr []ByCurrency,
	curD []Currency,
	ccD []ByCallingCode,
) []CountryResult {
	var countries []CountryResult
	for _, abbObj := range abbArr {
		var country CountryResult
		country.ID = uuid.New().String()
		country.Code = abbObj.Abbreviation
		country.Name = abbObj.Country
		for _, contObj := range contArr {
			if contObj.Country == country.Name {
				country.Continent = contObj.Continent
				break
			}
		}
		for _, curObj := range curArr {
			if curObj.Country == country.Name {
				for _, curDObj := range curD {
					if curDObj.Code == curObj.CurrencyCode {
						country.Currency = &Currency{
							Code: curDObj.Code,
							Name: curDObj.Name,
						}
						break
					}
				}
				break
			}
		}
		for _, ccDObj := range ccD {
			if strings.ToUpper(ccDObj.Country) == strings.ToUpper(country.Name) {
				country.CallingCode = fmt.Sprintf("+%d", ccDObj.CallingCode)
				break
			}
		}
		countries = append(countries, country)
	}

	var newCountries []CountryResult
	for _, cData := range countries {
		for _, curD := range curD {
			if cData.Currency != nil && curD.Code == cData.Currency.Code {
				newCountries = append(newCountries, cData)
				break
			}
		}
	}

	return newCountries
}
