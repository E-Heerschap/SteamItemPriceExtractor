//This upadtes the price of steam items.
//Author: Edwin Heerschap
package main

import (
	"time"
	_ "bitbucket.org/SneakyHideout/ItemManager/HttpUtil"
	"fmt"
	"bitbucket.org/SneakyHideout/ItemManager/HttpUtil"
)

func main() {

	cfgFile := GetConfigFile()

	//Looping through configs to run
	for _, config := range cfgFile.Configurations {

		//List of items to scan information for
		itemsList := getAllItems(cfgFile.DatabaseURL, cfgFile.DatabaseUser, cfgFile.DatabasePassword, cfgFile.DatabaseName, config.DBNamesTable, cfgFile.MarketID)

		//Used to ensure we compete all items in loop
		itemsLength := itemsList.Len()
		itemsCompleted := 0

		//Will store the item information to pass off to the database
		itemInfoArray := make([]itemPriceInfo, itemsLength)

		//Normal items names are sent to this channel to get item price information
		jobChan := make(chan requestJob)

		//Channel which completed jobs are sent back on
		//This channel will contain itemPriceInfo objects from each job.
		completedChan := make(chan itemPriceInfo)

		//err429 chan is where failed jobs are sent back on
		err429Chan := make(chan requestJob)

		//This is used to maintain the correct position in the itemInfoArray
		//so that no nil values are filled into the array.
		itemsCompletedSuccessfully := 0

		//Used to counter how many requests have been done in the current Tor circuit
		torSwitchCounter := 0
		_ = torSwitchCounter

		//Creating request workers
		for i := 0; i < cfgFile.NoOfReqRoutines; i++{
			go StartNewRequestWorker(jobChan, err429Chan, completedChan, cfgFile.Currency, config.AppId, cfgFile.TorProxy)
		}

		fmt.Println(itemsList.Front().Value.(requestJob).name)
		fmt.Println(itemsList.Back().Value.(requestJob).name)
		//Looping until number of completed items is equal to the number of items
		for itemsLength != itemsCompleted - 1 {


			select {
			case failedJob := <- err429Chan:
				itemsList.PushBack(failedJob)
				continue
			case completedJob := <- completedChan:
				if completedJob.Success {
					itemInfoArray[itemsCompletedSuccessfully] = completedJob;
					fmt.Printf("ID: %d, Lowest: %s, Volume, %s\n\r", itemInfoArray[itemsCompletedSuccessfully].itemId, itemInfoArray[itemsCompletedSuccessfully].LowestPrice, itemInfoArray[itemsCompletedSuccessfully].Volume)
					itemsCompletedSuccessfully++

					if itemsCompletedSuccessfully % 5 == 0 {
						args := createItemPriceDBArguments(itemInfoArray[itemsCompletedSuccessfully-5:itemsCompletedSuccessfully-1], cfgFile.MarketID)
						fmt.Println("FAT DOG")
						go uploadItemsToDB(cfgFile.DatabaseURL, cfgFile.DatabaseUser, cfgFile.DatabasePassword, cfgFile.DatabaseName, config.DBPriceTable, args)
					}
				}
				fmt.Println("completed.")
				itemsCompleted++
				continue
			default:
			}

			if itemsList.Len() != 0 {

				newEle := itemsList.Front()
				itemsList.Remove(newEle)
				newJob := newEle.Value.(requestJob)
				fmt.Println(newJob.name)
				fmt.Println(newJob.itemId)

				jobChan <- newJob

				time.Sleep(time.Millisecond * time.Duration(60.0/cfgFile.RequestSpeed*1000))

				torSwitchCounter++
				if torSwitchCounter == cfgFile.RequestsBeforeTorSwitch {
					HttpUtil.SwitchTorEndpoint(cfgFile.TorControl, cfgFile.TorProxy)
					torSwitchCounter = 0
				}
			}

			itemsCompleted++

		}

		//Uploading elements currently in array
		args := createItemPriceDBArguments(itemInfoArray[:itemsCompletedSuccessfully - 1], cfgFile.MarketID)
		uploadItemsToDB(cfgFile.DatabaseURL, cfgFile.DatabaseUser, cfgFile.DatabasePassword, cfgFile.DatabaseName, config.DBPriceTable, args)

		close(jobChan)
		close(completedChan)
		close(err429Chan)

	}

}