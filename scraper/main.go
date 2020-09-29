package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

//Product struct
type Product struct {
	Name  string
	Price float32
}

func main() {

	products := MakeRequest("https://www.carrefour.es/supermercado/la-despensa/alimentacion/pastas/N-107dg9k/c")

	file, _ := json.MarshalIndent(products, "", "")

	_ = ioutil.WriteFile("products.json", file, 0644)

}

//MakeRequest Makes request
func MakeRequest(url string) []Product {
	resp, err := http.Get(url)

	if err != nil {
		fmt.Print(err)
	}

	defer resp.Body.Close()

	var products []Product

	if resp.StatusCode == 200 {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)

		doc.Find(".product-card-item").Each(func(i int, s *goquery.Selection) {
			products = append(products, CreateProduct(s))
		})

		return products
	}
	return nil
}

//CreateProduct creates products
func CreateProduct(s *goquery.Selection) Product {
	var price float32
	name := strings.Replace(s.Find(".title-product").Text(), "\n", "", -1)
	if CheckIfPriceLess(name, s.Find(".price").Text()) {
		rawPrice := strings.Replace(s.Find(".price").Text(), ",", ".", 1)
		price = ConvertString(rawPrice)
	}
	if CheckIfPriceLess(name, s.Find(".price-less").Text()) {
		rawPrice := strings.Replace(s.Find(".price-less").Text(), ",", ".", 1)
		delLineBreaks := strings.Replace(rawPrice, "\n", "", 1)
		price = ConvertString(delLineBreaks)
	}
	return Product{Name: name, Price: price}
}

//ConvertString converts strings to float32
func ConvertString(price string) float32 {
	convPrice, err := strconv.ParseFloat(strings.ReplaceAll(price, "\u00a0€", ""), 32)
	if err != nil {
		fmt.Print(err)
	}
	return float32(convPrice)
}

func CheckIfPriceLess(name string, price string) bool {
	return name != "" && price != ""
}
