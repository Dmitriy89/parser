package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"myreports"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v2"
)

type сountBrand struct {
	Brand string
	Count int
}

type yamlConf struct {
	URL  string      `yaml:"url"`
	Date []time.Time `yaml:"date"`
}

var (
	hashBrand           = make(map[string]int)
	hashSize            = make(map[string]string)
	sumProduct          int
	jsonSRC, sliceSpeed []string
	allProduct,
	flagBool,
	chartHBar,
	chartLine bool
	todayDate,
	title string
	wg sync.WaitGroup
)

const (
	noBreakSpace     rune   = 160
	host             string = "https://www.tsum.ru"
	jsonURLSelection string = "http://api.int.tsum.com/catalog/search?selection="
	jsonURLSection   string = "http://api.int.tsum.com/catalog/search?section="
)

func init() {
	todayDate = time.Now().Format("02.01.2006_15:04:05")
	os.Mkdir("report", 0644)
	flag.BoolVar(&allProduct, "all", false, "Parsing products on all pages")
	flag.BoolVar(&chartHBar, "chartHBar", false, "Create chart brands")
	flag.BoolVar(&chartLine, "chartLine", false, "Create chart line brands")
	flag.Parse()
}

// ReadYAML Parsing URL from *.yaml file
func ReadYAML() (yamlConf, error) {

	var yf yamlConf

	discription, err := os.Open("conf.yaml")
	if err != nil {
		return yf, fmt.Errorf("Error func ReadYAML method Open:__ %s", err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(discription)
	body := buf.Bytes()
	err = yaml.Unmarshal(body, &yf)
	if err != nil {
		return yf, fmt.Errorf("Error func ReadYAML method Unmarshal:__ %s", err)
	}

	return yf, nil
}

// NewURL if not the start page, then remove the page settings
func NewURL(src string) (string, error) {
	urlOld, err := url.Parse(src)
	if err != nil {
		return "", fmt.Errorf("Error func NewURL method Parse:__ %s", err)
	}
	val := urlOld.Query()
	val.Del("page")
	getParamNew := val.Encode()
	urlNew := urlOld.Scheme + "://" + urlOld.Host + urlOld.EscapedPath() + "?" + getParamNew

	return urlNew, nil
}

// TrimPrice trim price
func TrimPrice(price string) string {
	sliceRune := make([]rune, 0, 15)
	for _, r := range price {
		if r != noBreakSpace {
			sliceRune = append(sliceRune, r)
		}
	}
	price = strings.Trim(string(sliceRune), "₽ ")
	return price
}

// ParseJSON parsing json SRC
func ParseJSON(jsonTsum string) error {
	var d interface{}
	readJSON := strings.NewReader(jsonTsum)
	decode := json.NewDecoder(readJSON)
	for {
		tok, errTok := decode.Token()
		if errTok != nil && errTok != io.EOF {
			return fmt.Errorf("Error func ParseJSON method Token:__ %s", errTok)
		} else if errTok == io.EOF {
			break
		}
		switch tok := tok.(type) {
		case string:
			if strings.Contains(tok, jsonURLSelection) || strings.Contains(tok, jsonURLSection) {
				err := decode.Decode(&d)
				if err != nil {
					return fmt.Errorf("Error func ParseJSON method Decode:__ %s", errTok)
				}
				switch d := d.(map[string]interface{})["body"].(type) {
				case []interface{}:
					for _, val := range d {
						switch d := val.(map[string]interface{})["photos"].(type) {
						case []interface{}:
							for key, val := range d {
								if key == 0 {
									jsonSRC = append(jsonSRC, val.(map[string]interface{})["small"].(string))
								}
							}
						}
					}
				}
			}
		}
	}
	//fmt.Printf("%T\n", d.(map[string]interface{})["body"])
	/*
		switch reflect.TypeOf(d.(map[string]interface{})["body"]).Kind() {
		case reflect.Slice:
			s := reflect.ValueOf(d.(map[string]interface{})["body"])

			for i := 0; i < s.Len(); i++ {
				fmt.Printf("%v\n", s.Index(i).Interface().(map[string]interface{})["photos"])
			}
		}
	*/
	return nil
}

// ParsingProduct parsing product
func ParsingProduct(src string) error {

	resp, err := http.Get(src)
	if err != nil {
		return fmt.Errorf("Error func ParsingProduct method Get:__ %s", err)
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("Error func ParsingProduct method NewDocumentFromReader:__ %s", err)
	}

	if doc.Find(".product__data").Text() == "" {
		flagBool = false
		return nil
	}

	title = doc.Find("h1").Text()
	json := strings.Replace(doc.Find("#frontend-state").Text(), "&q;", `"`, -1)

	//Parsing JSON to file !!!need
	/*
		readerString := strings.NewReader(json)
		discriptor, err := os.Create("Обувь.json")
		io.Copy(discriptor, readerString)
	*/

	err = ParseJSON(json)
	if err != nil {
		return fmt.Errorf("Error func ParsingProduct to func ParseJSON:__ %s", err)
	}

	sliceWripper := make([][]string, 0, 60)
	slice := make([]string, 0, 7)

	stmt, db, tx, err := myreports.WriteDataDB()
	defer db.Close()
	defer stmt.Close()

	doc.Find(".product_type_catalog").Each(func(i int, s *goquery.Selection) {
		brand := s.Find(".product__brand").Text()
		discription := s.Find(".product__description").Text()
		priceSaleNew := s.Find(".price_type_new").Text()
		priceSaleOld := s.Find(".price_type_old").Text()
		label := s.Find(".label").Text()
		imgHref, _ := s.Find("a").Attr("href")
		price := s.Find(".price").Text()

		if priceSaleNew != "" {
			price = priceSaleNew
		}
		if priceSaleOld == "" {
			priceSaleOld = "0"
		}

		priceInt, err := strconv.Atoi(TrimPrice(price))
		if err != nil {
			log.Fatal("Error func ParsingProduct method Atoi1:__", err)
		}
		priceOldInt, err := strconv.Atoi(TrimPrice(priceSaleOld))
		if err != nil {
			log.Fatal("Error func ParsingProduct method Atoi2:__", err)
		}

		_, err = stmt.Exec(time.Now(), brand, discription, priceInt, priceOldInt, host+imgHref, label, jsonSRC[i], hashSize[imgHref])
		if err != nil {
			log.Fatal("Error func ParsingProduct method Exec:__", err)
		}
		slice = []string{brand, discription, TrimPrice(price), TrimPrice(priceSaleOld), host + imgHref, label, jsonSRC[i], hashSize[imgHref]}
		sliceWripper = append(sliceWripper, slice)
	})

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Error func ParsingProduct method Commit:__ %s", err)
	}

	for _, val := range sliceWripper {
		if r, ok := hashBrand[val[0]]; !ok {
			hashBrand[val[0]] = 1
		} else {
			hashBrand[val[0]] = r + 1
		}
	}

	nameCSV := time.Now().Format("report/report_02.01.2006.csv")

	if _, err := os.Stat(nameCSV); os.IsNotExist(err) {
		sliceTitle := make([][]string, len(sliceWripper)+1)
		sliceTitle[0] = []string{"Brand", "Description", "Price", "Price not discount", "Link", "Label", "Src img", "Size product"}
		copy(sliceTitle[1:], sliceWripper)
		sliceWripper = sliceTitle
	}

	err = WriteCSV(nameCSV, sliceWripper)
	if err != nil {
		return fmt.Errorf("Error func ParsingProduct func WriteCSV:__ %s", err)
	}

	return nil
}

func newCsvWriter(w io.Writer, bom bool) *csv.Writer {
	bw := bufio.NewWriter(w)
	if bom {
		bw.Write([]byte{0xEF, 0xBB, 0xBF})
	}
	return csv.NewWriter(bw)
}

// WriteCSV write slice to file *.csv
func WriteCSV(nameCSV string, slice [][]string) error {

	fileCSV, err := os.OpenFile(nameCSV, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("Error func WriteCSV method OpenFile:__ %s", err)
	}
	defer fileCSV.Close()

	writerFile := newCsvWriter(fileCSV, true)
	defer writerFile.Flush()
	writerFile.Comma = ';'

	for _, value := range slice {
		if err = writerFile.Write(value); err != nil {
			return fmt.Errorf("Error func WriteCSV method Write:__ %s", err)
		}
	}
	return nil
}

func errLogger(err error) {
	if err != nil {
		l, errOpen := os.OpenFile("log_file.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if errOpen != nil {
			log.Fatal("Error opening utm log file:__", errOpen)
		}
		defer l.Close()
		multi := io.MultiWriter(l, os.Stdout)
		logger := log.New(multi, "", log.Ldate|log.Ltime)

		logger.Println(err)

	}
}

func parseSizeProduct(src string) {
	imgPath := make([]string, 0)
	resp, err := http.Get(src)
	if err != nil {
		log.Fatal("Error func parseSizeProduct method Get:__ ", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal("Error func parseSizeProduct method NewDocumentFromReader:__ ", err)
	}

	doc.Find(".product_type_catalog").Each(func(i int, s *goquery.Selection) {
		imgHref, _ := s.Find("a").Attr("href")
		imgPath = append(imgPath, imgHref)
	})

	for _, val := range imgPath {
		wg.Add(1)
		go func(src string) {
			defer wg.Done()
			var sliceSpeed = []string{}
			var sizeStr string
			resp, err := http.Get(host + src)
			if err != nil {
				log.Fatal("Error gorutine method Get:__ ", err)
			}
			defer resp.Body.Close()

			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				log.Fatal("Error gorutine method NewDocumentFromReader:__ ", err)
			}

			doc.Find(".select_size_full ul .select__item").Each(func(i int, s *goquery.Selection) {
				if s := s.Find(".select__text").Text(); strings.Contains(s, "|") {
					sliceSpeed = append(sliceSpeed, s)
					sizeStr = strings.Join(sliceSpeed, " , ")
				}
			})
			if _, ok := hashSize[src]; !ok {
				hashSize[src] = sizeStr
			}

		}(val)
	}
}

func main() {

	yc, err := ReadYAML()
	errLogger(err)

	switch {
	case chartHBar:
		err := myreports.ChartHBar(yc.Date)
		errLogger(err)
	case chartLine:
		/*
			errLG := myreports.ChartLineGeneral(yc.Date)
			errLogger(errLG)
			errLI := myreports.ChartLineIndividual(yc.Date)
			errLogger(errLI)
		*/
		errLCD := myreports.ChartLineCountDay(yc.Date)
		errLogger(errLCD)
	default:
		var brand []сountBrand

		time1 := time.Now()

		if allProduct {
			newURL, errNU := NewURL(yc.URL)
			errLogger(errNU)

			parseSizeProduct(newURL)
			wg.Wait()

			errPP := ParsingProduct(newURL)
			errLogger(errPP)

			flagBool = true
			p, i := "&page=", 2

			for flagBool {

				iString := strconv.Itoa(i)
				param := p + iString
				parseSizeProduct(newURL + param)
				wg.Wait()
				err := ParsingProduct(newURL + param)
				errLogger(err)
				i++
			}

		} else {
			parseSizeProduct(yc.URL)
			wg.Wait()
			err := ParsingProduct(yc.URL)
			errLogger(err)
		}

		for _, c := range hashBrand {
			sumProduct = sumProduct + c
		}

		sliceXLSX := make([][]string, 0, 3)
		sliceXLSXTitle := make([]string, 0, 60)
		sliceXLSXValue := make([]string, 0, 60)
		sliceXLSXTitle = append(sliceXLSXTitle, "Data Parcing", "Type page", "URL", "Count Product")
		sliceXLSXValue = append(sliceXLSXValue, todayDate, title, yc.URL, strconv.Itoa(sumProduct))

		for key, val := range hashBrand {
			brand = append(brand, сountBrand{key, val})
		}

		sort.SliceStable(brand, func(i, j int) bool { return strings.ToLower(brand[i].Brand) < strings.ToLower(brand[j].Brand) })

		for _, val := range brand {
			sliceXLSXTitle = append(sliceXLSXTitle, val.Brand)
			sliceXLSXValue = append(sliceXLSXValue, strconv.Itoa(val.Count))
		}

		sliceXLSX = append(sliceXLSX, sliceXLSXTitle, sliceXLSXValue)

		errDXDB := myreports.DateXLSX(sliceXLSX)
		errLogger(errDXDB)

		time2 := time.Now()
		fmt.Println(time2.Sub(time1).Seconds())
	}
}
