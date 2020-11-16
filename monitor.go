package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	method  = "POST"
	client  = &http.Client{}
	TxCount = 0
)

var wg sync.WaitGroup
var rw sync.RWMutex

func Monitor() {

	interval := flag.Uint("i", 5, "Query interval in second")
	flag.Parse()
	queryStats(interval)

}

func queryStats(interval *uint) {
	titles := []string{"Timestamp\t", "TPS", "TxCount"}
	titles = append(titles, "SHARD-HEIGHT")
	fmt.Println(strings.Join(titles, "\t"))
	intv := time.Duration(*interval)
	ticker := time.NewTicker(intv * time.Second)

	// url := "https://rpc.s0.t.hmny.io"
	url := "http://172.31.9.176:9500"
	// 先query出有几个shard
	shardNum := GetShardNum(url)

	msg, blockNumOld := statsInit(shardNum, *interval)
	fmt.Println(msg)
	for {
		select {
		case <-ticker.C:
			TxCount = 0
			msg, blockNumOld = stats(shardNum, *interval, blockNumOld)
			fmt.Println(msg)
		}
	}
}

func statsInit(shardNum int, interval uint) (string, []string) {
	BlockNumber := make([]string, shardNum)
	wg.Add(shardNum)

	for i := 0; i < shardNum; i++ {
		num := strconv.Itoa(i + 9500)
		// url := "https://rpc.s" + num + ".t.hmny.io"
		url := "http://172.31.9.176:" + num

		go func(i int) {
			// 再query每个shard的当前高度
			BlockNumber[i] = GetBlockNum(url)
			rw.Lock()
			// query交易数量
			TxCount = GetTxCount(url, BlockNumber[i]) + TxCount
			rw.Unlock()

			wg.Done()
		}(i)
	}
	wg.Wait()

	t := time.Now()
	msg := t.Format("2006-01-02 15:04:05")
	msg += "\t"
	msg += fmt.Sprintf("%2.2f", float64(TxCount/int(interval)))
	msg += "\t"
	msg += fmt.Sprintf("%2.2f", float64(TxCount))
	msg += "\t"

	shards := make([]string, shardNum)
	for i := 0; i < shardNum; i++ {
		shards[i] = fmt.Sprintf("%d-%s", i, BlockNumber[i])
	}
	msg += strings.Join(shards, " ")

	return msg, BlockNumber
}

func stats(shardNum int, interval uint, blockNumOld []string) (string, []string) {
	BlockNumber := make([]string, shardNum)
	wg.Add(shardNum)

	for i := 0; i < shardNum; i++ {
		num := strconv.Itoa(i + 9500)
		// url := "https://rpc.s" + num + ".t.hmny.io"
		url := "http://172.31.9.176:" + num

		go func(i int) {
			// 再query每个shard的当前高度
			BlockNumber[i] = GetBlockNum(url)
			// 如果高度有增加，则query增加高度中的交易数量
			if BlockNumber[i] != blockNumOld[i] {
				numOld, _ := strconv.Atoi(blockNumOld[i])
				numNew, _ := strconv.Atoi(BlockNumber[i])

				for j := numOld + 1; j <= numNew; j++ {
					rw.Lock()
					TxCount = GetTxCount(url, strconv.Itoa(j)) + TxCount
					rw.Unlock()
				}

				// 更新区块高度
				blockNumOld[i] = BlockNumber[i]
			}

			wg.Done()
		}(i)
	}
	wg.Wait()

	// fmt.Println("txcount:", TxCount)

	t := time.Now()
	msg := t.Format("2006-01-02 15:04:05")
	msg += "\t"
	msg += fmt.Sprintf("%2.2f", float64(TxCount/int(interval)))
	msg += "\t"
	msg += fmt.Sprintf("%2.2f", float64(TxCount))
	msg += "\t"

	shards := make([]string, shardNum)
	for i := 0; i < shardNum; i++ {
		shards[i] = fmt.Sprintf("%d-%s", i, BlockNumber[i])
	}
	msg += strings.Join(shards, " ")

	return msg, blockNumOld
}

func GetTxCount(url string, blockNum string) int {

	payload := strings.NewReader("{\n    \"jsonrpc\": \"2.0\",\n    \"id\": 1,\n    \"method\": \"hmyv2_getBlockTransactionCountByNumber\",\n    \"params\": [\n       " + blockNum + "\n    ]\n}")

	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	bodyString := string(body)
	index := strings.LastIndex(bodyString, ":")
	txCountString := bodyString[index+1 : len(bodyString)-2]
	// fmt.Println(txCountString)
	txCount, _ := strconv.Atoi(txCountString)
	return txCount
}

func GetBlockNum(url string) string {

	payload := strings.NewReader("{\n    \"jsonrpc\": \"2.0\",\n    \"id\": 1,\n    \"method\": \"hmyv2_blockNumber\",\n    \"params\": []\n}")

	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	bodyString := string(body)
	index := strings.LastIndex(bodyString, ":")
	blockNum := bodyString[index+1 : len(bodyString)-2]
	// fmt.Println(blockNum)
	return blockNum
}

func GetShardNum(url string) int {

	payload := strings.NewReader("{\n    \"jsonrpc\": \"2.0\",\n    \"id\": 1,\n    \"method\": \"hmyv2_getShardingStructure\",\n    \"params\": []\n}")

	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	bodyString := string(body)
	count := strings.Count(bodyString, "shardID")

	return count
}
