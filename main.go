package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

const count = 100000

type TXS struct {
	TX [count]map[string]interface{}
}

func main() {
	// 监视区块链
	Monitor()

	// 将交易写入到json
	// WriteTXJSON()

	// // 读取BLS key以及账户address，并把他们按规则生成DeployAccount
	// _, BLSkey, _ := ListDir("../harmony/.hmy", ".key")
	// // fmt.Println(BLSkey, err)
	// acc := ReadAcc("../go-sdk/account600.txt")
	// DeployAccount(acc, BLSkey, 600, 600)

	// // 读取BLS key，生成ReSharding.txt文件
	// BLSkey, _, _ := ListDir("../harmony/.hmy", ".key")
	// ReSharding(BLSkey, 0, 600)

}

// WriteTXJSON 辅助函数，用于将交易写入json
func WriteTXJSON() {

	var txArray [count]map[string]interface{}

	for i := 0; i < count; i++ {
		tx := make(map[string]interface{})
		tx["from"] = "one1pdv9lrdwl0rg5vglh4xtyrv3wjk3wsqket7zxy"
		tx["to"] = "one1pdv9lrdwl0rg5vglh4xtyrv3wjk3wsqket7zxy"
		tx["from-shard"] = "0"
		tx["to-shard"] = "1"
		tx["amount"] = "0.000000001"
		nonce := strconv.Itoa(i)
		tx["nonce"] = nonce

		txArray[i] = tx
		if i%1000 == 0 {
			fmt.Println(txArray[i]["nonce"])
		}
	}

	txs := TXS{
		TX: txArray,
	}

	txToJSON, err := json.MarshalIndent(txs.TX, "", "	")
	if err != nil {
		fmt.Println("error:", err)
	}
	err = ioutil.WriteFile("testtxs-shard0-100000.json", txToJSON, os.ModeAppend)
	if err != nil {
		fmt.Println("error:", err)
	}
}

//ListDir 获取指定目录下的所有文件，不进入下一级目录搜索，可以匹配后缀过滤。
func ListDir(dirPth string, suffix string) ([]string, []string, error) {
	files := make([]string, 0, 10)
	fileNameOnly := make([]string, 0, 10)
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, nil, err
	}
	// PthSep := string(os.PathSeparator)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写
	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) { //匹配文件
			files = append(files, fi.Name())
			fileTemp := fi.Name()
			fileNameOnly = append(fileNameOnly, fileTemp[0:len(fileTemp)-4]) //获取文件名
		}
	}
	return files, fileNameOnly, nil
}

// ReadAcc 读取已经注册了钱包的账户的地址
func ReadAcc(AccFile string) []string {
	file, err := os.Open(AccFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	/*
		ScanLines (默认)
		ScanWords
		ScanRunes (遍历UTF-8字符非常有用)
		ScanBytes
	*/

	var account []string
	//是否有下一行
	for scanner.Scan() {
		// fmt.Println("read string:", scanner.Text())
		if strings.Contains(scanner.Text(), "one") {
			account = append(account, scanner.Text())
		}
	}
	// fmt.Println(account)

	return account
}

func DeployAccount(account []string, BLSkey []string, nodeNum int, accNum int) {
	shardNum := nodeNum / accNum
	for i := 0; i < nodeNum; i++ {
		fmt.Println(`{Index: "`, i, `", Address: "`+account[i/shardNum]+`", BLSPublicKey: "`+BLSkey[i]+`"},`)
	}
}

func ReSharding(BLSkey []string, shardIndex int, nodePerShard int) {
	for i := shardIndex * nodePerShard; i < (shardIndex+1)*nodePerShard; i++ {
		istr := strconv.Itoa(i - shardIndex*nodePerShard)
		var port string
		if len(istr) == 1 {
			port = "00" + istr
		} else if len(istr) == 2 {
			port = "0" + istr
		} else if len(istr) == 3 {
			port = istr
		}
		fmt.Println("0.0.0.0 9" + port + " validator .hmy/" + BLSkey[i])
	}
	fmt.Println("0.0.0.0 9299 explorer null", shardIndex)
}
