package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

//Product struct
type Product struct {
	Name  string
	Price float32
}

//Category struct
type Category struct {
	Name string
	URL  string
}

//TODO refactor main func and try to simplify it. Also fix the "Ofertas y promociones" section, because it is saved as nil
func main() {
	host := "https://www.carrefour.es"
	url := host + "/supermercado"

	categories := GetAllCategories(url)

	if _, err := os.Stat("../data/carrefour"); os.IsNotExist(err) {
		os.MkdirAll("../data/carrefour", 0700)
	}
	var products []Product

	for _, category := range categories {
		url := host + category.URL
		subCategories := GetAllCategories(url)
		if _, err := os.Stat("../data/carrefour/" + category.Name); os.IsNotExist(err) {
			os.MkdirAll("../data/carrefour/"+category.Name, 0700)
		}
		for _, subCategory := range subCategories {
			moreCategories := GetAllCategories(host + subCategory.URL)
			if len(moreCategories) != 0 {
				if _, err := os.Stat("../data/carrefour/" + category.Name + "/" + subCategory.Name); os.IsNotExist(err) {
					os.MkdirAll("../data/carrefour/"+category.Name+"/"+subCategory.Name, 0700)
				}
				for _, subSubCategory := range moreCategories {
					pages := GetProductPages(url + subSubCategory.URL)

					for _, page := range pages {
						products = append(products, MakeRequest(host+page)...)
					}

					file, _ := json.MarshalIndent(products, "", "")

					er := ioutil.WriteFile("../data/carrefour/"+category.Name+
						"/"+subCategory.Name+"/"+subSubCategory.Name+".json", file, 0644)
					if er != nil {
						print(er.Error())
					}
					products = nil
				}
			} else {
				pages := GetProductPages(url + subCategory.URL)

				for _, page := range pages {
					products = append(products, MakeRequest(host+page)...)
				}

				file, _ := json.MarshalIndent(products, "", "")

				er := ioutil.WriteFile("../data/carrefour/"+category.Name+"/"+subCategory.Name+".json", file, 0644)
				if er != nil {
					print(er.Error())
				}

				products = nil
			}
		}
	}

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

//CheckIfPriceLess checks if the product has discount
func CheckIfPriceLess(name string, price string) bool {
	return name != "" && price != ""
}

//GetProductPages will return pagination
func GetProductPages(url string) []string {
	resp, err := http.Get(url)

	if err != nil {
		fmt.Print(err)
	}

	defer resp.Body.Close()

	var pages []string

	if resp.StatusCode == 200 {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		doc.Find(".selectPagination > option").Each(func(i int, p *goquery.Selection) {
			attr, _ := p.Attr("value")
			pages = append(pages, attr)
		})

		return pages
	}
	return nil
}

//GetAllCategories will return all categories and their urls
func GetAllCategories(url string) []Category {
	var categories []Category
	resp, err := http.Get(url)

	if err != nil {
		fmt.Print(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		doc.Find(".category").Each(func(i int, c *goquery.Selection) {
			url, _ := c.Find("a").Attr("href")
			cat := Category{Name: c.Find(".nombre-categoria").Text(), URL: url}

			categories = append(categories, cat)
		})
	}

	return categories
}
