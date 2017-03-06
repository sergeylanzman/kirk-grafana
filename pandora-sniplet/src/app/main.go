package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	sdkbase "qiniu.com/pandora/base"
	tsdb "qiniu.com/pandora/tsdb"
)

type Proxy struct {
	Url         string `json:"tsdbHost"`
	Ak          string
	Sk          string
	Port        int  `json:"port"`
	CrossDomain bool `json:"cross_domain"`
	DebugLevel  int  `json:"debug_level"`
}

func NewClient(host, ak, sk string) (client tsdb.TsdbAPI, err error) {
	tsdbConfig := tsdb.NewConfig().
		WithAccessKeySecretKey(ak, sk).
		WithEndpoint(host).
		WithLogger(sdkbase.NewDefaultLogger()).
		WithLoggerLevel(sdkbase.LogDebug)

	client, err = tsdb.New(tsdbConfig)
	if err != nil {
		return
	}

	return
}

var client tsdb.TsdbAPI

func (proxy *Proxy) handler(w http.ResponseWriter, r *http.Request) {

	if proxy.CrossDomain {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization,Origin, X-Requested-With, Content-Type, Accept, X-Appid")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	}

	if strings.ToUpper(r.Method) == "OPTIONS" {
		return
	}

	vs, err := url.ParseRequestURI(r.RequestURI)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("recieve req url:", r.RequestURI)

	repoName := vs.Query().Get("db")
	sql := vs.Query().Get("q")
	fmt.Printf("repo: %v, sql: %v\n", repoName, sql)

	var ret *tsdb.QueryOutput
	if strings.HasPrefix(sql, "SHOW MEASUREMENTS") { //翻译show measurements语句
		results, err := client.ListSeries(&tsdb.ListSeriesInput{RepoName: repoName})
		if err != nil {
			fmt.Println(err)
			return
		}
		if results == nil {
			return
		}

		ret = &tsdb.QueryOutput{
			Results: []tsdb.Result{
				tsdb.Result{
					Series: []tsdb.Serie{
						tsdb.Serie{
							Name:    "measurements",
							Columns: []string{"name"},
						},
					},
				},
			},
		}
		ret.Results[0].Series[0].Values = make([][]interface{}, 0)
		for _, result := range *results {
			sub := []interface{}{result.Name}
			ret.Results[0].Series[0].Values = append(ret.Results[0].Series[0].Values, sub)
		}
	} else {
		ret, err = client.QueryPoints(&tsdb.QueryInput{RepoName: repoName, Sql: sql})
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	if strings.HasPrefix(sql, "SELECT") { //如果是select语句的话需要转换数据点的时间戳
		for i, result := range ret.Results {
			for j, series := range result.Series {
				for k, values := range series.Values {
					v, ok := values[0].(string)
					if !ok {
						fmt.Println("type assertion fail")
						continue
					}
					ret.Results[i].Series[j].Values[k][0] = convertTimestamp(v)
				}
			}
		}
	}

	str, err := json.Marshal(*ret)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Fprint(w, string(str))

}

func convertTimestamp(t string) (ret int64) { //string格式的timestamp转换成ms
	tm, err := time.Parse(time.RFC3339, t)
	if err != nil {
		fmt.Println(err)
		return
	}
	ret = tm.UnixNano() / 1000000
	return
}

func load_config(fileName string, proxy *Proxy) error {
	f, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(f, proxy)

	proxy.Ak = os.Getenv("DATABASE_AK")
	proxy.Sk = os.Getenv("DATABASE_SK")

	return err
}

func main() {
	flag.Parse()
	arg := flag.Arg(0)
	if arg == "" {
		fmt.Println("no config")
		return
	}
	var proxy Proxy
	err := load_config(arg, &proxy)
	if err != nil {
		fmt.Println("load config failed", err)
		return
	}
	client, err = NewClient(proxy.Url, proxy.Ak, proxy.Sk)
	if err != nil {
		fmt.Println(err)
		return
	}

	http.HandleFunc("/query", proxy.handler)
	err = http.ListenAndServe(":"+strconv.Itoa(proxy.Port), nil)
	if err != nil {
		fmt.Println(err)
	}
}
