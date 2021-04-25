package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli"
)

type ReportData struct {
	miningTime time.Time
	jpy        int
}

func main() {
	app := cli.NewApp()
	app.Name = "nhf"
	app.Usage = "nhf -i report.csv -o output.csv"
	app.Version = "1.0.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "input, i",
			Usage: "NiceHash report csv file path",
		},
		cli.StringFlag{
			Name:  "output, o",
			Usage: "output report file path",
		},
	}

	app.Action = func(c *cli.Context) error {
		input := c.String("input")
		output := c.String("output")

		// Validation
		if input == "" {
			return fmt.Errorf("Error: %s", "input/i is a required option.")
		}

		if output == "" {
			return fmt.Errorf("Error: %s", "output/o is a required option.")
		}

		file, err := os.Open(input)
		if err != nil {
			return fmt.Errorf("Error: %s", "can not open report.")
		}
		defer file.Close()

		reader := csv.NewReader(file)
		reader.ReuseRecord = true

		// レポートを売上と支出に分割
		var salsesData []ReportData
		var costData []ReportData

		var line []string
		var header = false
		for {
			line, err = reader.Read()
			if err != nil {
				break
			}

			if !header {
				header = true
				continue
			}

			t, err := time.Parse("2006-01-02 15:04:05 MST-0700", strings.Replace(line[1], "+09:00", "+0900", 1))
			if err != nil {
				return fmt.Errorf("Error: %s", err.Error())
			}

			f, err := strconv.ParseFloat(line[5], 64)
			if err != nil {
				return fmt.Errorf("Error: %s", err.Error())
			}

			if f >= 0 {
				// 売上の場合は小数点以下切り捨て
				salse := int(math.Floor(f))
				if salse > 0 {
					salsesData = append(salsesData, ReportData{t, salse})
				}
			} else {
				// 支出の場合は小数点以下切り上げ
				cost := int(math.Ceil(f))
				if cost < 0 {
					costData = append(costData, ReportData{t, cost})
				}
			}
		}

		// 日毎にまとめる
		daySales := map[string]int{}
		dayCost := map[string]int{}

		for _, v := range salsesData {
			date := v.miningTime.Format("2006-01-02")
			_, ok := daySales[date]

			if ok {
				daySales[date] = daySales[date] + v.jpy
			} else {
				daySales[date] = v.jpy
			}
		}
		for _, v := range costData {
			date := v.miningTime.Format("2006-01-02")
			_, ok := dayCost[date]

			if ok {
				dayCost[date] = dayCost[date] + v.jpy
			} else {
				dayCost[date] = v.jpy
			}
		}

		out, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return fmt.Errorf("Error: %s", "can not open result file.")
		}
		defer out.Close()

		writer := csv.NewWriter(out)
		writer.Write([]string{"発生日", "収支区分", "勘定科目", "金額", "税区分", "決済日", "決済口座", "決済金額"})

		for k, v := range daySales {
			amount := strconv.Itoa(v)
			ty := "収入"
			ac := "売上高"
			tax := "非課売上"
			b := "NiceHashMining"
			writer.Write([]string{k, ty, ac, amount, tax, k, b, amount})
		}

		for k, v := range dayCost {
			amount := strconv.Itoa(v * -1)
			ty := "支出"
			ac := "支払手数料"
			tax := "非課仕入"
			b := "NiceHashMining"
			writer.Write([]string{k, ty, ac, amount, tax, k, b, amount})
		}
		writer.Flush()

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
