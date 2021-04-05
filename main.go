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
	jpy        float64
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

		var data []ReportData

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

			data = append(data, ReportData{t, f})
		}

		set := [][]ReportData{}
		sliceSize := len(data)
		for i := 0; i < sliceSize; i += 2 {
			end := i + 2
			if sliceSize < end {
				end = sliceSize
			}
			set = append(set, data[i:end])
		}

		out, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return fmt.Errorf("Error: %s", "can not open result file.")
		}
		defer out.Close()

		writer := csv.NewWriter(out)
		writer.Write([]string{"発生日", "収支区分", "勘定科目", "金額", "税区分", "決済日", "決済口座", "決済金額"})

		for _, v := range set {
			date := v[0].miningTime.Format("2006-01-02")
			ty := "収入"
			ac := "売上高"
			sum := math.Trunc(v[0].jpy + v[1].jpy)
			amount := strconv.FormatFloat(sum, 'f', -1, 64)
			tax := "非課売上"
			b := "NiceHashMining"

			writer.Write([]string{date, ty, ac, amount, tax, date, b, amount})
		}
		writer.Flush()

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
