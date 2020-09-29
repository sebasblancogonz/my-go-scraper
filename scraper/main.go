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

	products := MakeRequest("https://www.carrefour.es/supermercado/la-despensa/alimentacion/aceites-y-vinagres/N-1c4rm7v/c")

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
			pName := strings.Replace(s.Find(".title-product").Text(), "\n", "", -1)

			if s.Find(".price").Text() != "" {
				rawPrice := strings.Replace(s.Find(".price").Text(), ",", ".", 1)
				convPrice, err := strconv.ParseFloat(strings.ReplaceAll(rawPrice, "\u00a0€", ""), 16)
				if err != nil {
					fmt.Print(err)
				}
				p := Product{Name: pName, Price: float32(convPrice)}
				products = append(products, p)
			}
			if s.Find(".price-less").Text() != "" {
				rawPrice := strings.Replace(s.Find(".price-less").Text(), ",", ".", 1)
				delLineBreaks := strings.Replace(rawPrice, "\n", "", 1)
				convPrice, err := strconv.ParseFloat(strings.ReplaceAll(delLineBreaks, "\u00a0€", ""), 32)
				if err != nil {
					fmt.Print(err)
				}
				p := Product{Name: pName, Price: float32(convPrice)}
				products = append(products, p)
			}
		})

		return products
	}
	return nil
}
