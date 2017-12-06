//This handles http requests sent to steam. This relies on the httpUtil created.
//Author: Edwin Heerschap
package SteamHttp

import (
	"bitbucket.org/SneakyHideout/ItemManager/HttpUtil"
	"encoding/json"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"fmt"
	"strconv"
)

//This is set to the Tor Proxy URL
var TorProxy string

type SteamItem struct {
	NormalName string
	ImageUrl   string
	AppId int
}

//GetTotalCount gets the total number of items on the steam market for an Appid.
func GetTotalCount(appId string) int {

	//Sending request for a single search (This will also include the total_count)
	reqBody, state, _ := sendSCMSRequest(appId, "0", "1", false)

	if !state {
		log.Fatal("Failed to send SCMS request")
	}

	//TODO Handle http code != OK

	ms := marketSearch{}
	json.Unmarshal(reqBody, &ms)

	return ms.Total_count
}

//GetSteamItemsData returns a list containing SteamItem id's.
func GetSteamItemsData(Appid string, start string, count string, useTor bool) ([]SteamItem, bool, int) {
	fmt.Println("Getting steam item data")
	//Getting response from steam
	respBody, success, HttpCode := sendSCMSRequest(Appid, start, count, useTor)

	if !success {
		log.Fatal("Failed to send SCMS request")
		return nil, false, HttpCode
	}

	//Parsing the response from steam from JSON into the marketSearch object
	ms := marketSearch{}
	json.Unmarshal(respBody, &ms)

	//This will store the information to be stored in the steam DB
	steamItems := make([]SteamItem, ms.Pagesize)
	itemCounter := 0

	//Tokenizing the HTML tags
	stringReader := strings.NewReader(ms.Results_html)
	tokens := html.NewTokenizer(stringReader)

	//Looping through html tags
	for {
		token := tokens.Next()
		switch {
		case token == html.ErrorToken:
			//End of document
			return steamItems, true, HttpCode
		case token == html.StartTagToken:
			tag := tokens.Token()

			//If it is a <a> tag
			if tag.Data == "a" {

				//Looping through attributes until finding href
				for _, a := range tag.Attr {
					if a.Key == "href" {
						//Finding & setting normal item name
						href := a.Val
						slashIndex := strings.LastIndex(href, "/")
						hashName := href[slashIndex+1:]
						normalName, _ := url.QueryUnescape(hashName)
						steamItems[itemCounter].NormalName = normalName
					}
				}

			}

		//Finding the 2x image for the item
		case token == html.SelfClosingTagToken:
			tag := tokens.Token()
			if tag.Data == "img" {

				for _, imgParam := range tag.Attr {
					//Finding & setting the image URL
					if imgParam.Key == "srcset" {
						imagesStr := imgParam.Val
						imageArr := strings.Split(imagesStr, " ")
						steamItems[itemCounter].ImageUrl = imageArr[len(imageArr)-2]
						steamItems[itemCounter].AppId, _ = strconv.Atoi(Appid);
						itemCounter++
					}
				}

			}

		}
	}


	return steamItems, true, HttpCode

}

//sendSCMSRequest - Run Steam Community Market Search RequestSpeed
//This corresponds to sending a request to the following url:
//http://steamcommunity.com/market/search/render/?cc=pt&count=4&currency=3&l=english&query=appid%3A730&start=0
//This will send the request and return the text body and error if any.
func sendSCMSRequest(Appid string, start string, count string, useTor bool) ([]byte, bool, int) {

	//Building request
	reqUrl, err := url.Parse("http://steamcommunity.com/market/search/render/")

	if err != nil {
		log.Fatal("Failed to parse steam market initial URL.")
		return nil, false, 0
	}

	//Setup the url parameters.
	urlParams := url.Values{}

	urlParams.Add("query", "appid:"+Appid)
	urlParams.Add("start", start)
	urlParams.Add("count", count)
	urlParams.Add("currency", "3")
	urlParams.Add("l", "english")
	urlParams.Add("cc", "pt") //No idea what cc=pt does... :(

	//Add parameters to url
	reqUrl.RawQuery = urlParams.Encode()

	var client *http.Client

	timeout := 60

	//If we are using Tor, setup http client to use Tor proxy.
	//If we are not using Tor, setup http client with normal timeout.
	if useTor {
		client = HttpUtil.SetupProxyClient(TorProxy, timeout)
	} else {
		client = &http.Client{Timeout: time.Second * time.Duration(timeout)}
	}

	respBody, success, HttpCode := HttpUtil.SendHttpRequest(client, reqUrl.String())

	return respBody, success, HttpCode
}
