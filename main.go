package main

import (
	"net/http"

	pagespeedonline "google.golang.org/api/pagespeedonline/v2"

	"fmt"
	"time"
	"os"
	"encoding/csv"
	"io"
)

func main() {
	//https://developers.google.com/speed/docs/insights/v2/reference/pagespeedapi/runpagespeed
	//読み込みファイル準備
	urlFile, _:= os.Open("./urls.csv")
	defer urlFile.Close()
	reader := csv.NewReader(urlFile)
	for {
		record, err := reader.Read() // 1行読み出す

		if err == io.EOF || len(record) == 0 {
			break
		} else if err != nil {
			fmt.Println(err)
		}

		fmt.Println(record[0])

		//for i, v := range record {
		//	fmt.Println()
		//}
		//var new_record []string
		//for i, v := range record {
		//	if i > 0 {
		//		new_record = append(new_record, fmt.Sprint(i) + ":" + v)
		//	}
		//}
		//writer.Write(new_record) // 1行書き出す
		//      log.Printf("%#v", record[0] + "," + record[1])
	}
//panic("------------------")

	target := "http://techblog-sokuhou.com"

	c := &http.Client{Timeout: time.Duration(60) * time.Second}
	pso, _ := pagespeedonline.New(c)
	r, _ := pso.Pagespeedapi.Runpagespeed(target).Locale("ja_jp").Strategy("desktop").Do()

	fmt.Println("---->Tartget:", r.Id)
	fmt.Println("---->Title:", r.Title)
	fmt.Println("---->RuleGroup-SPEED:", r.RuleGroups["SPEED"])

	fmt.Println("--------------------------------------------------------")
	for k, v := range r.FormattedResults.RuleResults {
		if k == "OptimizeImages" {
			fmt.Println("---->", v.LocalizedRuleName)
			fmt.Println("---->", v.Summary.Format)
			for _, uv := range v.UrlBlocks {
				fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
				for _, urlsValue := range uv.Urls {
					fmt.Println(urlsValue.Result.Format)
					fmt.Println("+++++++++++++++++++")
					for _, arg := range urlsValue.Result.Args {
						fmt.Println("**********")
						//fmt.Println(arg.Key)
						//fmt.Println(arg.Rect)
						//fmt.Println(arg.SecondaryRects)
						fmt.Println(arg.Type)
						fmt.Println(arg.Value)
						fmt.Println("**********")
					}
				}
			}
		}
	}
	fmt.Println("--------------------------------------------------------")
}
