package myreports

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tealeg/xlsx"
)

type rawTime []byte

// Convert []byte date from DB to format time.Time
func (t rawTime) Time() (time.Time, error) {
	return time.Parse("2006-01-02", string(t))
}

func connectDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:12345@/test")
	if err != nil {
		return nil, fmt.Errorf("Error func connectDB method Open:__ %s", err)
	}
	return db, nil
}

//WriteDataDB write data to DB
func WriteDataDB() (*sql.Stmt, *sql.DB, *sql.Tx, error) {
	var tx *sql.Tx

	db, err := connectDB()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error func WriteDataDB method Open:__ %s", err)
	}

	if tx, err = db.Begin(); err != nil {
		tx.Rollback()
		return nil, nil, nil, fmt.Errorf("Error func WriteDataDB method Begin:__ %s", err)
	}

	stmt, errPrepare := tx.Prepare("INSERT INTO brands2(date_parse ,brand, discription, price, price_not_discount, href, label, src, size_product) VALUES(?,?,?,?,?,?,?,?,?);")
	if errPrepare != nil {
		return nil, nil, nil, fmt.Errorf("Error func WriteDataDB method Prepare:__ %s", errPrepare)
	}

	return stmt, db, tx, nil
}

//DataChartHBarDB read data from DB
func DataChartHBarDB(date []time.Time) ([]string, error) {
	var (
		timeEmpty time.Time
		slice     []string
	)
	db, err := connectDB()
	if err != nil {
		return nil, fmt.Errorf("Error func ReadDB:__ %s", err)
	}
	defer db.Close()

	if date[1] == timeEmpty {
		date[1] = date[0]
	}

	row, err := db.Query(`select brand, count(brand) as productcount from brands where date_parse between ? and ? group by brand`, date[0], date[1])
	if err != nil {
		return nil, fmt.Errorf("Error func ReadDB method Query:__ %s", err)
	}
	defer row.Close()

	for row.Next() {
		var (
			brand        string
			productcount int
		)

		if err := row.Scan(&brand, &productcount); err != nil {
			return nil, fmt.Errorf("Error func ReadDB method Scan:__ %s", err)
		}

		slice = append(slice, brand, strconv.Itoa(productcount))
	}

	return slice, nil
}

//DataChartLineDB read data from DB
func DataChartLineDB(date []time.Time) ([][]string, error) {
	var (
		timeEmpty time.Time
		slice     [][]string
	)
	db, err := connectDB()
	if err != nil {
		return nil, fmt.Errorf("Error func ReadDB:__ %s", err)
	}
	defer db.Close()

	if date[1] == timeEmpty {
		date[1] = date[0]
	}

	row, err := db.Query(`select date_parse,brand, count(brand) from brands where date_parse between ? and ? group by brand, date_parse`, date[0], date[1])
	if err != nil {
		return nil, fmt.Errorf("Error func ReadDB method Query:__ %s", err)
	}
	defer row.Close()

	for row.Next() {
		var (
			brand        string
			productcount int
			dateparse    rawTime
		)

		if err := row.Scan(&dateparse, &brand, &productcount); err != nil {
			return nil, fmt.Errorf("Error func ReadDB method Scan:__ %s", err)
		}

		//Formate date from Base
		t, err := dateparse.Time()
		if err != nil {
			return nil, err
		}

		slice = append(slice, []string{t.Format("\"02.01\","), brand, strconv.Itoa(productcount)})
	}

	return slice, nil
}

//DataChartLineDB read data from DB
func DataChartLineCountDayDB(date []time.Time) ([][]string, error) {
	var (
		timeEmpty                  time.Time
		slice                      [][]string
		sliceDate, sliceCountBrand []string
	)
	db, err := connectDB()
	if err != nil {
		return nil, fmt.Errorf("Error func ReadDB:__ %s", err)
	}
	defer db.Close()

	if date[1] == timeEmpty {
		date[1] = date[0]
	}

	row, err := db.Query(`select date_parse, count(brand) from brands where date_parse between ? and ? group by date_parse`, date[0], date[1])
	if err != nil {
		return nil, fmt.Errorf("Error func ReadDB method Query:__ %s", err)
	}
	defer row.Close()

	for row.Next() {
		var (
			dateparse    rawTime
			productcount int
		)

		if err := row.Scan(&dateparse, &productcount); err != nil {
			return nil, fmt.Errorf("Error func ReadDB method Scan:__ %s", err)
		}

		//Formate date from Base
		t, err := dateparse.Time()
		if err != nil {
			return nil, err
		}
		sliceDate = append(sliceDate, t.Format("\"02.01\","))
		sliceCountBrand = append(sliceCountBrand, strconv.Itoa(productcount)+",")

	}
	slice = append(slice, sliceDate, sliceCountBrand)
	return slice, nil
}

//DateXLSX write data to xlsx
func DateXLSX(data [][]string) error {
	var (
		discription *xlsx.File
		sheet       *xlsx.Sheet
		row         *xlsx.Row
		style       *xlsx.Style
	)
	borderStyleThin := "thin"
	borderStyleThick := "thick"
	xlsx.SetDefaultFont(11, "Calibri")

	if _, err := os.Stat("report_2.xlsx"); os.IsNotExist(err) {
		discription = xlsx.NewFile()
		sheet, err = discription.AddSheet("Sheet1")
		if err != nil {
			return fmt.Errorf("Error func DateXLSXDB method AddSheet:__ %s", err)
		}
	} else {
		discription, err = xlsx.OpenFile("report_2.xlsx")
		if err != nil {
			return fmt.Errorf("Error func writeCSV method OpenFile:__ %s", err)
		}
	}

	for key, val := range data {
		if key == 0 {
			style = xlsx.NewStyle()
			style.ApplyBorder = true
			style.Border = *xlsx.NewBorder(borderStyleThin, borderStyleThin, borderStyleThick, borderStyleThin)
			style.Border.TopColor = "f99d9d"
			style.Fill.PatternType = "solid"
			//style.Fill.BgColor = "FF4F00"
			style.Fill.FgColor = "f99d9d"
		} else {
			style = xlsx.NewStyle()
			style.ApplyBorder = true
			style.Border = *xlsx.NewBorder(borderStyleThin, borderStyleThin, borderStyleThin, borderStyleThin)

		}

		if _, err := os.Stat("report_2.xlsx"); os.IsNotExist(err) {
			row = sheet.AddRow()
		} else {
			row = discription.Sheet[discription.Sheets[0].Name].AddRow()
		}

		row.WriteSlice(&val, len(val))
		for i := 0; i < len(val); i++ {
			row.Cells[i].SetStyle(style)
		}
	}
	discription.Save("report_2.xlsx")

	return nil
}
