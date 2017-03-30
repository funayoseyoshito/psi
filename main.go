package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	pagespeedonline "google.golang.org/api/pagespeedonline/v2"
)

//ResultRow 結果のrow
type ResultRow []string

//Results 結果のrowを格納する配列
type Results []ResultRow

//AnalyzeParam analyze関数へ渡す引数
type AnalyzeParam struct {
	target, strategy string
}

const (
	urlsFilePath   = "./urls.csv"   //Analyzeの対象となるURLリストファイル
	resultFilePath = "./result.csv" //結果csvのファイルパス
	strategyMOBILE = "mobile"       //Analyze対象のデバイス(mobile)
	strategyPC     = "desktop"      //Analyze対象のデバイス(PC)
	workerNum      = 5              //ワーカーの数
)

var wg sync.WaitGroup

func main() {

	fmt.Println("--- start ---")

	cxt, cancel := context.WithCancel(context.Background())
	queue := make(chan AnalyzeParam)

	//ワーカーの作成
	for i := 0; i < workerNum; i++ {
		wg.Add(1)
		go func(ctx context.Context, queue chan AnalyzeParam) {
			for {
				select {
				case <-ctx.Done():
					wg.Done()
					return
				case param := <-queue:
					writeCsv(analyze(param))
				}
			}
		}(cxt, queue)
	}

	file, _ := os.Open(urlsFilePath)
	defer file.Close()
	reader := csv.NewReader(file)
	reader.LazyQuotes = true

	//キューにジョブを積む
	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println(err)
		} else if record != nil {
			fmt.Println(record[0])
			param := AnalyzeParam{target: record[0], strategy: strategyPC}
			queue <- param
			param.strategy = strategyMOBILE
			queue <- param
		}
	}

	cancel()
	wg.Wait()
	fmt.Println("--- end ---")
}

func replaceToFormat(format string, key string, value string) string {
	return strings.Replace(format, "{{"+key+"}}", value, 1)
}

//Analyzeを行う
func analyze(param AnalyzeParam) Results {

	var results Results
	c := &http.Client{Timeout: time.Duration(60) * time.Second}
	pso, err := pagespeedonline.New(c)
	if err != nil {
		panic(err)
	}

	r, err := pso.Pagespeedapi.Runpagespeed(param.target).
		Locale("ja_jp").Strategy(param.strategy).Do()
	if err != nil {
		panic(err)
	}

	for k, v := range r.FormattedResults.RuleResults {
		if k == "OptimizeImages" {
			for _, u := range v.UrlBlocks {
				for _, urlsValue := range u.Urls {
					message := urlsValue.Result.Format
					var source string
					for _, arg := range urlsValue.Result.Args {
						if arg.Key == "URL" {
							source = arg.Value
						}
						message = replaceToFormat(message, arg.Key, arg.Value)
					}
					results = append(results,
						ResultRow{time.Now().String(),
							param.strategy,
							r.Id,
							r.Title,
							strconv.FormatInt(r.RuleGroups["SPEED"].Score, 10),
							v.LocalizedRuleName,
							v.Summary.Format,
							message,
							source,
						})
				}

			}
		}
	}
	return results
}

//CSVファイルに書き込む
func writeCsv(data Results) {
	file, err := os.OpenFile(resultFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)

	if err != nil {
		panic(err)
	}
	defer file.Close()
	w := csv.NewWriter(transform.NewWriter(file, japanese.ShiftJIS.NewEncoder()))
	for _, row := range data {
		w.Write(row)
	}
	w.Flush()
}
