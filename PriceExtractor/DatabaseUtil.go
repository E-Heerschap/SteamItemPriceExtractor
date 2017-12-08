//This defines the database worker objects and how to interact with them.
//Author: Edwin Heerschap
package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"fmt"
	"container/list"
	"strconv"
	"bytes"
)


//GetAllItems gets all the records f
func getAllItems(dbUrl string, dbUser string, dbPass string, dbName string, dbTable string, marketID int) *list.List{

	returnList := list.New()

	//Creating Data Source Name (DSN)
	var dsn string
	dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPass, dbUrl, dbName)

	//Getting database handle
	db, err := sql.Open("mysql", dsn)

	defer db.Close()

	if err != nil {
		log.Println(err)
		return returnList
	}


	//Getting records (I think wild card only works after the WHERE statement)
	queryString := "SELECT ItemID, ItemName FROM " + dbTable + " WHERE MarketID = ?;"

	results, err := db.Query(queryString, marketID)

	defer results.Close()

	if err != nil {
		log.Println(err)
		return returnList
	}


	//Filling the list with items
	for results.Next() {
		var tempJob requestJob
		results.Scan(&tempJob.itemId, &tempJob.name)
		fmt.Printf("ID: %d, Name: %s", tempJob.itemId, tempJob.name)
		returnList.PushFront(tempJob)
	}

	return returnList
}


func stripLeftZeros(s string) string {
	counter := 0
	for _, ele := range s {
		if string(ele) == "0" {
			counter++
		}else{
			return s[counter:]
		}
	}
	return ""
}

func StripNonNumericCharacters(s string) string{
	var newString string
	for _,ele := range s {
		if _, err := strconv.Atoi(string(ele)); err == nil {
			newString = newString + string(ele)
		}
	}

	return newString
}

func cleanStringNumber(s *string){
	*s = StripNonNumericCharacters(*s)
	*s = stripLeftZeros(*s)
	if *s == "" {
		*s = "-1"
	}
}

func createItemPriceDBArguments(data []itemPriceInfo, marketID int) []interface{} {


	returnArray := make([]interface{}, len(data) * 5)

	for i := 0; i < len(data); i++ {
		//Clean strings first
		cleanStringNumber(&data[i].LowestPrice)
		cleanStringNumber(&data[i].MedianPrice)
		cleanStringNumber(&data[i].Volume)

		//Place into arguments
		fmt.Println(data[i].itemId)
		returnArray[(i*5)] = data[i].itemId
		returnArray[(i*5)+1] = marketID
		returnArray[(i*5)+2], _ = strconv.Atoi(data[i].LowestPrice)
		returnArray[(i*5)+3], _ = strconv.Atoi(data[i].Volume)
		returnArray[(i*5)+4], _ = strconv.Atoi(data[i].MedianPrice)
	}

	return returnArray
}

func uploadItemsToDB(dbUrl string, dbUser string, dbPass string, dbName string, dbTable string, itemPriceArgs []interface{}){

	//Creating Data Source Name (DSN)
	var dsn string
	dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPass, dbUrl, dbName)

	//Getting database handle
	db, err := sql.Open("mysql", dsn)

	defer db.Close()

	if err != nil {
		log.Println("Failed to open database handle. ", err)
	}

	var buffer bytes.Buffer

	buffer.WriteString("INSERT INTO " + dbTable + " (ItemID, MarketID, LowestPrice, Volume, MedianPrice) VALUES " )

	for i := 1; i <= len(itemPriceArgs)/5 - 1; i++ {
		buffer.WriteString("(?, ?, ?, ?, ?),")
	}

	//Writing last value with semicolon
	buffer.WriteString("(?, ?, ?, ?, ?);")


	fmt.Println("Executing upload")

	queryformat := buffer.String()
	fmt.Println(queryformat)

	for _, ele := range itemPriceArgs {
		fmt.Println(ele)
	}

	_, err = db.Exec(queryformat, itemPriceArgs...)

	if err != nil {
		log.Println("Upload to database failed. ", err)
	}


}