package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
	"strconv"

	"golang.org/x/net/proxy"
	"fmt"
)

// URL to fetch
var webUrl string = "https://check.torproject.org"

// Specify Tor proxy ip and port
var torProxy string = "socks5://127.0.0.1:9050" // 9150 w/ Tor Browser

func spamTest(i int){

	// Parse Tor proxy URL string to a URL type
	torProxyUrl, err := url.Parse(torProxy)
	if err != nil {
		log.Fatal("Error parsing Tor proxy URL:", torProxy, ".", err)
	}

	// Create proxy dialer using Tor SOCKS proxy
	torDialer, err := proxy.FromURL(torProxyUrl, proxy.Direct)
	if err != nil {
		log.Fatal("Error setting Tor proxy.", err)
	}

	// Set up a custom HTTP transport to use the proxy and create the client
	torTransport := &http.Transport{Dial: torDialer.Dial}
	client := &http.Client{Transport: torTransport, Timeout: time.Second * 60}

	// Make request
	resp, err := client.Get("http://steamcommunity.com/market/priceoverview/?appid=730&currency=3&market_hash_name=StatTrak%E2%84%A2%20M4A1-S%20|%20Hyper%20Beast%20(Minimal%20Wear)")
	if err != nil {
		log.Fatal("Error making GET request.", err)
		fmt.Println("FAILING")
	}

	if (resp.StatusCode != http.StatusOK){
		fmt.Println("Status Error: ", strconv.Itoa(resp.StatusCode))
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading body of response.", err)
		fmt.Println("FAILING")
	}



	if(string(body) == ""){
		fmt.Println("FAILED")
	}else{
		fmt.Println(string(body))
	}

	fmt.Println("Ending ", strconv.Itoa(i))
}

func test() {


	for i := 0; i < 10000; i++ {
		time.Sleep(time.Millisecond * 100)
		go spamTest(i)
	}

	for(true){

	}

}
