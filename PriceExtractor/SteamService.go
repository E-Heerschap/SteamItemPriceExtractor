package main

import (
	"bitbucket.org/SneakyHideout/ItemManager/HttpUtil"
	"net/http"
	"net/url"
	"log"
	"encoding/json"
	"strconv"
)

type requestWorker struct {
	reqJobChan chan requestJob
	err429Chan chan requestJob
	completedChan chan itemPriceInfo
	currency string
	proxy string
	appId string
}

type itemPriceInfo struct {
	Success bool `json:"success"`
	LowestPrice string `json:"lowest_price"`
	Volume string `json:"volume"`
	MedianPrice string `json:"median_price"`
	itemId int
}

type requestJob struct {
	name string
	itemId int
}

func (rw *requestWorker) getItemPriceInfo(job requestJob, proxy string) {

	//Getting *http.Client using proxy
	client := HttpUtil.SetupProxyClient(proxy, 60)

	//Steam uses path escape instead of query escape for some reason...
	urlStr := "http://steamcommunity.com/market/priceoverview/?"
	query := "appid=" + rw.appId + "&currency=" + rw.currency + "&market_hash_name=" + job.name

	urlStr = urlStr + url.PathEscape(query)

	bytes, success, httpCode := HttpUtil.SendHttpRequest(client, urlStr)

	if !success {
		log.Println("Failed to send http request")
		rw.completedChan <- itemPriceInfo{Success: false}
	}

	if httpCode != http.StatusOK {
		if httpCode == 429 {
			rw.err429Chan <- job
		}else{
			log.Println("Http status code that is neither 200 or 429: " + strconv.Itoa(httpCode))
			rw.completedChan <- itemPriceInfo{Success: false}
		}
	}

	var returnItemPriceInfo itemPriceInfo

	if string(bytes) == "null"{
		log.Println("NULL returned from steam! URL: " + urlStr)
	}

	err := json.Unmarshal(bytes, &returnItemPriceInfo)

	if err != nil {
		log.Println("Failed to parse json for item price information: ", err)
		rw.completedChan <- itemPriceInfo{Success: false}
	}

	if !returnItemPriceInfo.Success {

		log.Println(string(bytes))

	}

	returnItemPriceInfo.itemId = job.itemId

	rw.completedChan <- returnItemPriceInfo

}


//handleJob is a wrapper function to handle
//jobs that arrive on the reqJobChan channel.
func (rw *requestWorker) handleJob(newJob requestJob) {

	//Handles the situation if a panic() occurs.
	defer func() {
		if (x := recover(); x != nil){
			log.Printf("Failed to handle job & catch error. Error: %v", x)
			rw.completedChan <- itemPriceInfo{Success: false}
		}
	}

	rw.getItemPriceInfo(newJob, rw.proxy)

}

//listen makes the requestWorker listen to the channel and handle
//jobs which arrive on the channel.
func (rw *requestWorker) listen() {
	for {
		newJob, ok := <- rw.reqJobChan

		if !ok {
			return
		}

		rw.handleJob(newJob)

	}
}

//startNewRequestWorker initializes a new requestWorker and starts it listening for jobs.
func StartNewRequestWorker(reqJobChan chan requestJob, err429Chan chan requestJob, completedChan chan itemPriceInfo, currency string, appId string, proxy string){
	rw := requestWorker{reqJobChan: reqJobChan, err429Chan: err429Chan, completedChan: completedChan, currency: currency, appId: appId, proxy: proxy}
	rw.listen()
}
