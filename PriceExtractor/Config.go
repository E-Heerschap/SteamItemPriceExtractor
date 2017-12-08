package main

import (
	"io/ioutil"
	"log"
	"encoding/json"
)

type config struct {

	DBNamesTable string `json:"DBNamesTable"`
	DBPriceTable string `json:"DBPriceTable`
	AppId string `json:"AppId"`

}

type configFile struct {
	RequestSpeed int `json:"RequestSpeed"`
	MarketID int `json:"MarketID"`
	NoOfReqRoutines int `json:"NoOfReqRoutines"`
	TorProxy string `json:"TorProxy"`
	TorControl string `json:"TorControl`
	TorControlPass string `json:"TorControlPass"`
	RequestsBeforeTorSwitch int `json:"RequestsBeforeTorSwitch"`
	DatabaseURL string `json:"DatabaseURL"`
	DatabaseUser string `json:"DatabaseUser"`
	DatabasePassword string `json:"DatabasePassword"`
	DatabaseName string `json:"DatabaseName"`
	Currency string `json:"Currency"`
	Configurations []config `json:"configurations"`
}

//GetConfigFile returns the the configFile object loaded with settings from
//the Config.json which should be located in the same folder as the app.
func GetConfigFile() configFile{

	jsonFile, err := ioutil.ReadFile("Config.Json")

	if err != nil {
		log.Fatal("Failed to read config file. ", err)
	}

	var cfgFile configFile

	err = json.Unmarshal(jsonFile, &cfgFile)

	if err != nil {
		log.Fatal("Failed to unmarshal json from config file. Ensure that it is valid. ", err)
	}

	return cfgFile
}