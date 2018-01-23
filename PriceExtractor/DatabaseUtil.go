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


//getItemsToUpdate returns the list of items which have the longest time since their price has been updated.
func getItemsToUpdate(dbUrl string, dbUser string, dbPass string, dbName string, dbItemTable, dbPriceTable string, marketID int, limit int) *list.List{

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


	//TODO look at incorporating the SQL formatter for security.
	//Query explanation:
	//This query gets the rows with the maximum time value from the price table for each
	//item id. A left join is used with these values so every item has its latest time displayed
	//and null is shown if no time exists. It then orders this in ascending list so the first values
	//are the items that have not been updated for the longest period of time.
	queryString := fmt.Sprintf("SELECT  %[1]s.ItemID, %[1]s.ItemName FROM %[1]s " +
"LEFT JOIN (SELECT %[2]s.ItemID as 'ItemID', MAX(%[2]s.created_at) as 'created_at' " +
"FROM %[2]s GROUP BY %[2]s.ItemID) As maxTbl ON maxTbl.ItemID = %[1]s.ItemID " +
"WHERE %[1]s.MarketID = %[3]d " +
"ORDER BY maxTbl.created_at ASC " +
"LIMIT %[4]d", dbItemTable, dbPriceTable, marketID, limit)

	fmt.Println(queryString)

	//Getting results & error checking
	results, err := db.Query(queryString)

	defer results.Close()

	if err != nil {
		log.Println(err)
		return returnList
	}


	//Filling returnList with items
	for results.Next() {
		var tempJob requestJob
		results.Scan(&tempJob.itemId, &tempJob.name)
		fmt.Printf("ID: %d, Name: %s", tempJob.itemId, tempJob.name)
		returnList.PushFront(tempJob)
	}

	return returnList
}

//stripLeftZeros removes all of the '0' characters on the left hand
//side of the string.
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

//StripNonNumericCharacters removes all non-numeric characters from the passed string
func StripNonNumericCharacters(s string) string{
	var newString string
	for _,ele := range s {
		if _, err := strconv.Atoi(string(ele)); err == nil {
			newString = newString + string(ele)
		}
	}

	return newString
}

//cleanStringNumber removes all 0 characters on the left and removes all non-numeric characters.
//-1 Is returned if the string is empty.
func cleanStringNumber(s *string){
	*s = StripNonNumericCharacters(*s)
	*s = stripLeftZeros(*s)
	if *s == "" {
		*s = "-1"
	}
}

//createItemPriceDBArguments creates the array of arguments to pass into the upload SQL
//query that is executed in the uploadItemsToDB function. The parameters from the
//itemPriceInfo objects are unpacked into an array.
func createItemPriceDBArguments(data []itemPriceInfo, marketID int) []interface{} {

	//Using an interface array so it stores generic types.
	//len(data) * 5 because each itemPriceInfo has 5 fields to unpack
	//into the array.
	returnArray := make([]interface{}, len(data) * 5)


	//Unpacking each itemPriceInfo into the array.
	for i := 0; i < len(data); i++ {

		//Cleaning strings
		//e.g $0.52 -> 52 or $1.52 -> 152
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

//uploadsItemsToDB uploads all the items passed in the itemsToUpload array to the database
//with the information supplied. This assumes the database matches the schematics to work with
//this program.
//TODO Autocheck DB schematics.
func uploadItemsToDB(dbUrl string, dbUser string, dbPass string, dbName string, dbPriceTable string, itemsToUpload []itemPriceInfo){

	//Creating Data Source Name (DSN) (Used to connect to database)
	var dsn string
	dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPass, dbUrl, dbName)

	//Getting database handle
	db, err := sql.Open("mysql", dsn)

	defer db.Close()

	if err != nil {
		log.Println("Failed to open database handle. ", err)
	}

	//Creating buffer to create SQL query
	var buffer bytes.Buffer

	//Building SQL query
	buffer.WriteString("INSERT INTO " + dbPriceTable + " (ItemID, MarketID, LowestPrice, Volume, MedianPrice) VALUES " )

	//Creating parameters to fill for each itemPriceInfo
	for i := 1; i < len(itemsToUpload); i++ {
		buffer.WriteString("(?, ?, ?, ?, ?),")
	}

	//Writing last value with semicolon
	buffer.WriteString("(?, ?, ?, ?, ?);")

	fmt.Println("Executing upload")

	//Creating query string with place holders
	queryformat := buffer.String()
	fmt.Println(queryformat)

	itemPriceArgs := createItemPriceDBArguments(itemsToUpload, 0)

	_, err = db.Exec(queryformat, itemPriceArgs...)

	if err != nil {
		log.Println("Upload to database failed. ", err)
	}

}
