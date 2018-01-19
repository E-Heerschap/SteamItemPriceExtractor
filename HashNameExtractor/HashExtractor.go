//This program extracts steam hash names from the steam market.
//Author: Edwin Heerschap

package main

import (
	"bitbucket.org/SneakyHideout/ItemManager/HashNameExtractor/SteamHttp"
	"bitbucket.org/SneakyHideout/ItemManager/HttpUtil"
	"container/list"
	"fmt"
	"strconv"
	"time"
	"log"
)

//createJobsList creates a new list of jobs from a Config.
func createJobsList(scanConfig Config) *list.List {

	//Finding range max value to request items for.
	var rangeEnd int
	if scanConfig.RangeEnd == -1 {
		//Set end of range to total number of steam items for Appid
		rangeEnd = SteamHttp.GetTotalCount(scanConfig.AppId)
		fmt.Println(rangeEnd)
	} else {
		rangeEnd = scanConfig.RangeEnd
	}

	//List that wills store all the jobs
	jobsList := list.New()

	//Populating jobs list
	for i := scanConfig.RangeStart; i < rangeEnd; i += 100 {

		newJob := job{appId: scanConfig.AppId, start: strconv.Itoa(i), count: "100"}
		jobsList.PushBack(newJob)

	}

	return jobsList
}

func main() {

	log.Print("Starting HashExtractor")

	//Load Config
	cfgFile, success := NewConfigFile()
	if !success {
		log.Print("Stopping HashExtractor")
		return
	}

	SteamHttp.TorProxy = cfgFile.TorProxy

	//Run configs
	for _, config := range cfgFile.ConfigArray {

		idCount := 0

		//Channel on which database jobs will be sent
		dbJobChan := make(chan []SteamHttp.SteamItem)

		//Jobs will be sent along this channel
		jobChan := make(chan job, 1)

		//Failed jobs will be sent back on this channel to re-enter the job list.
		err429Chan := make(chan job)

		//A completed job will send true to this channel
		jobCompletedChan := make(chan bool)

		//Creating all the jobs
		jobsList := createJobsList(config)

		totalNumberOfJobs := jobsList.Len()

		initialReqWorker := requestWorker{id: idCount, jobChan: jobChan, err429Chan: err429Chan, jobCompletedChan: jobCompletedChan, dbJobChan: dbJobChan}
		go initialReqWorker.Listen()
		idCount++

		initialDBWorker := DatabaseWorker{databaseChan: dbJobChan}
		go initialDBWorker.StartWorker(cfgFile.DatabaseURL, cfgFile.DatabaseUser, cfgFile.DatabasePassword, cfgFile.DatabaseName, config.DBTable)

		requestsBeforeTorSwitch := cfgFile.RequestsBeforeTorSwitch
		torSwitchCounter := 0
		//Creating workers to send requests
		fmt.Println("Starting loop")
		for jobsCompleted := 0; jobsCompleted < totalNumberOfJobs; {

			//Ensuring that job completed channel takes responsibility.
			select {
			//Add to jobs completed if compeleted flag in channel
			case completed := <-jobCompletedChan:
				//TODO find a better way to do this.
				if completed {
					jobsCompleted++
				}
				continue
			case failedJob := <-err429Chan:
				jobsList.PushBack(failedJob)
				continue
			default:
			}

			//Only send jobs or create new workers if there are jobs in list.
			if jobsList.Len() != 0 {

				newJob := jobsList.Remove(jobsList.Front()).(job)

				select {

				case jobChan <- newJob:
					if requestsBeforeTorSwitch == torSwitchCounter {
						HttpUtil.SwitchTorEndpoint(cfgFile.TorControl, cfgFile.TorControlPass)
						torSwitchCounter = 0
					}
					torSwitchCounter++

				default:
					rw := requestWorker{jobChan: jobChan, id: idCount, err429Chan: err429Chan, dbJobChan: dbJobChan, jobCompletedChan: jobCompletedChan}
					go rw.StartWorker(newJob)
					idCount++
					time.Sleep(time.Millisecond * 1)
				}

				time.Sleep(time.Millisecond * time.Duration(60.0/cfgFile.RequestSpeed*1000))
			}

			//Creating new database worker if the channel has more than 1 job in it.
			if len(dbJobChan) > 1 {
				dw := DatabaseWorker{databaseChan: dbJobChan, marketID: cfgFile.DatabaseMarketID}
				go dw.StartWorker(cfgFile.DatabaseURL, cfgFile.DatabaseUser, cfgFile.DatabasePassword, cfgFile.DatabaseName, config.DBTable)
			}

		}

		var waitStr string
		fmt.Scanln(&waitStr)
		close(jobChan)
		close(dbJobChan)
		close(err429Chan)

	}

	//Start database service



}
