package main

//Defines a RequestWorker and the job they handle
//Author: Edwin Heerschap

import (
	"bitbucket.org/SneakyHideout/ItemManager/HashNameExtractor/SteamHttp"
	"fmt"
	"net/http"
	"log"
)

//Used to pass information on the request job for the worker
type job struct {
	appId string
	start string
	count string
}

//Worker used to send requests
type requestWorker struct {
	jobChan   chan job
	dbJobChan chan []SteamHttp.SteamItem
	err429Chan chan job
	id        int
	iteration int
}

//handleJob uses the information from the job to request information on the steam
//item and then passes the information onto a database worker.
func (rw *requestWorker) handleJob(newJob job) {
	rw.iteration++

	si, success, httpCode := SteamHttp.GetSteamItemsData(newJob.appId, newJob.start, newJob.count, true)

	if !success {
		log.Fatal("Failed to get steam item data")
		return
	}

	if httpCode != http.StatusOK {
		if httpCode == 429 {
			rw.err429Chan <- newJob
			return
		}

		fmt.Printf("Http code neither 200 or 429! Code: %d", httpCode)
	}else{

		rw.dbJobChan <- si

	}

	fmt.Printf("Id: %d, Iteration: %d, HTTPCODE: %d\n", rw.id, rw.iteration, httpCode)

}

//Starts a worker with a job then listens to a channel
func (rw *requestWorker) StartWorker(startJob job) {
	fmt.Println("Starting")
	rw.handleJob(startJob)
	rw.iteration = 0
	rw.Listen()
}

//Starts a worker listening to a channel
func (rw *requestWorker) Listen() {
	for {

		newJob, open := <-rw.jobChan
		//returning if job channel is closed.
		if !open {
			return
		}
		rw.handleJob(newJob)
	}
}
