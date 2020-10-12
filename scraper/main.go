package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

//Product struct
type Product struct {
	Name, Image string
	Price       float32
	IsOffer     bool
}

//Category struct
type Category struct {
	Name, URL string
}

//TODO Fix the "Ofertas y promociones" json files under every category, because it is saved as null
func main() {
	host := "https://www.carrefour.es"
	url := host + "/supermercado"

	categories := GetAllCategories(url)

	var products []Product

	var wg sync.WaitGroup

	for _, category := range categories {
		directory := "../data/carrefour/" + category.Name
		url := host + category.URL
		subCategories := GetAllCategories(url)
		CreateDirectories(directory)
		for _, subCategory := range subCategories {
			subDirectory := directory + "/" + subCategory.Name
			fileName := subDirectory + ".json"
			CreateDirectories(subDirectory)
			subSubCategories := GetAllCategories(host + subCategory.URL)
			if len(subSubCategories) != 0 {
				wg.Add(len(subSubCategories))
				go func(subDirectory string, subSubCategories []Category) {
					for _, subSubCategory := range subSubCategories {
						fileName := subDirectory + "/" + subSubCategory.Name + ".json"
						pages := GetProductPages(url + subSubCategory.URL)

						for _, page := range pages {
							products = append(products, MakeRequest(host+page)...)
						}

						WriteJSONFile(products, fileName)

						products = nil
						wg.Done()
					}
				}(subDirectory, subSubCategories)
				wg.Wait()
			} else {
				pages := GetProductPages(url + subCategory.URL)

				for _, page := range pages {
					products = append(products, MakeRequest(host+page)...)
				}

				WriteJSONFile(products, fileName)

				products = nil
			}
		}
	}

}

//MakeRequest Makes request
func MakeRequest(url string) []Product {
	emptyProduct := Product{}
	resp, err := http.Get(url)

	if err != nil {
		fmt.Print(err)
	}

	defer resp.Body.Close()

	var products []Product

	if resp.StatusCode == 200 {
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		doc.Find(".product-card-item").Each(func(i int, s *goquery.Selection) {
			p := CreateProduct(s)
			if p != emptyProduct {
				products = append(products, p)
			}

		})

		return products
	}
	return nil
}

//CreateProduct creates products
func CreateProduct(s *goquery.Selection) Product {
	var price float32
	var isOffer bool
	name := strings.Replace(s.Find(".title-product").Text(), "\n", "", -1)
	if name != "" {
		imgURL, _ := s.Find("img").Attr("src")
		if CheckIfPriceLess(name, s.Find(".price").Text()) {
			rawPrice := strings.Replace(s.Find(".price").Text(), ",", ".", 1)
			price = ConvertString(rawPrice)
		}
		if CheckIfPriceLess(name, s.Find(".price-less").Text()) {
			rawPrice := strings.Replace(s.Find(".price-less").Text(), ",", ".", 1)
			delLineBreaks := strings.Replace(rawPrice, "\n", "", 1)
			price = ConvertString(delLineBreaks)
			isOffer = true
		}
		return Product{Name: name, Price: price, IsOffer: isOffer, Image: imgURL}
	}
	return Product{}
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

//CreateDirectories will create all directories on a given path
func CreateDirectories(directory string) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.MkdirAll(directory, 0700)
	}
}

//WriteJSONFile will write the products on a JSON file
func WriteJSONFile(products []Product, fileName string) {
	file, _ := json.MarshalIndent(products, "", "")

	er := ioutil.WriteFile(fileName, file, 0644)
	if er != nil {
		print(er.Error())
	}
}
