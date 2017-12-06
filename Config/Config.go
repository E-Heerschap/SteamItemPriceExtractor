//This is a container to hold handle the configurations for the PriceExtractor.
//Author: Edwin Heerschap

package Config

import (
  "io/ioutil"
  "log"
  "encoding/json"
)

//Holds specific configurations to run.
//For example an extract region for csgo items from 0 -> 2000.
type Config struct{

  Name string `json:"Name"`
  AppId string `json:"Appid"`//Steam app id to scan
  RangeStart int `json:"ExtractRangeStart"`
  RangeEnd int `json:"ExtractRangeEnd"`//Ranges of items scan.
}



type ConfigFile struct{

  RequestSpeed float64 `json:"RequestSpeed"`
  UseTorPivots bool `json:"UseTorPivots"` //When true this will use the tor proxy service.
  TorProxy string `json:"TorProxy"` //URL to local Tor proxy service.
  ConfigArray []Config `json:"configurations"`//Array of different configurations to run.
  DatabaseURL string `json:"DatabaseURL"`
  DatabaseUser string `json:"DatabaseUser"`
  DatabasePassword string `json:"DatabasePassword"`
  DatabaseName string `json:"DatabaseName"`
  MaxGoHeRoutines int `json:"MaxGoHeRoutines"`
  MaxGoDBRoutines int `json:"MaxGoDBRoutines"`
  TorControl string `json:"TorControl"`
  TorControlPass string `json:"TorControlPass"`
  RequestsBeforeTorSwitch int `json:"RequestsBeforeTorSwitch"`

}

//NewConfigFile creates a new config file loaded from the 'config.json'
//which should be loacted in the same directory as Config.go
func NewConfigFile() (ConfigFile, bool) {

  cfgf := ConfigFile{}

  //Reading in JSONFile
  jsonData, err := ioutil.ReadFile("Config.json")
  if err != nil {
    log.Fatal(err)
    return ConfigFile{}, false
  }

  //Injecting cfgf with json data.
  json.Unmarshal(jsonData, &cfgf)

  return cfgf, true

}
