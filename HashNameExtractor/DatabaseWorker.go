package main

import (
  "bitbucket.org/SneakyHideout/ItemManager/HashNameExtractor/SteamHttp"
  _ "github.com/go-sql-driver/mysql"
  "fmt"
  "bytes"
  "database/sql"
  "log"
)


type DatabaseWorker struct {
  databaseChan chan []SteamHttp.SteamItem
  marketID int
  dbUrl string
  dbUser string
  dbPass string
  dbName string
  dbTable string
}

func (dw *DatabaseWorker) handleJob (si []SteamHttp.SteamItem){

  //Data source name (https://github.com/go-sql-driver/mysql/#dsn-data-source-name)
  var dsn string
  dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s", dw.dbUser, dw.dbPass, dw.dbUrl, dw.dbName)
  //Creating SQL handle
  //This must use the 'mysql' as the driver.
  //This driver can be imported from github.com/go-sql-driver/mysql
  db, err := sql.Open("mysql", dsn)

  if err != nil {
    log.Fatal("Failed to create SQL handle!")
    return
  }

  defer db.Close()

  //Checking if a connection can be made to MySql
  err = db.Ping()
  if err != nil {
    log.Fatal("Cannot connect to MySQL")
    return
  }

  //Using efficient way to concatenate strings
  //https://stackoverflow.com/questions/1760757/how-to-efficiently-concatenate-strings-in-go
  //Doing this so we can have the correct amount of wild cards.
  var buffer bytes.Buffer
  //Creating Temporary table to store values in
  buffer.WriteString("CREATE temporary TABLE TempTable (ItemName VARCHAR(200), ImageUrl VARCHAR(500), MarketID INT(11));\n")
  queryformat := buffer.String()
  buffer.Reset()
  _, err = db.Exec(queryformat)

  buffer.WriteString("INSERT INTO TempTable (ItemName, ImageUrl, MarketID) Values ")
  counter := 0
  for i := 0; i < len(si) - 1; i++ {
    buffer.WriteString("(?, ?, ?),")
    counter++
  }

  //Last one should not have a comma
  buffer.WriteString("(?, ?, ?);")
  counter++
  fmt.Printf("Amount written to db: %d \n", counter)
  queryformat = buffer.String()


  //Creating array to unpack as arugments arguments.
  //https://stackoverflow.com/questions/17555857/go-unpacking-array-as-arguments
  arguments := make([]interface{}, len(si) * 3)
  for i := 0; i < len(si); i++{
    arguments[i*3] = si[i].NormalName
    arguments[(i*3) + 1] = si[i].ImageUrl
    arguments[(i*3) + 2] = dw.marketID
  }

  queryformat = buffer.String()

  _, err = db.Exec(queryformat, arguments...)

  fillTblQuery := "INSERT INTO " + dw.dbTable + " (ItemName, ImageUrl, MarketID) (SELECT DISTINCT ItemName, ImageUrl, MarketID FROM TempTable WHERE TempTable.ItemName NOT IN (SELECT " + dw.dbTable + ".ItemName FROM " + dw.dbTable + "));"

  _, err = db.Exec(fillTblQuery)

    res, err := db.Query("SELECT COUNT(CSGO_Items.ItemName) FROM homestead.CSGO_Items;")

    var count int

    if err != nil {
      fmt.Println(err)
     fmt.Println(queryformat)
    }else{

    for res.Next() {
      err = res.Scan(&count)
      if err != nil {
        fmt.Println(err)
      }
      fmt.Printf("Number of rows: %d \r\n", count)
    }

  }


}

func (dw *DatabaseWorker) StartWorker(databaseUrl string, databaseUser string, databasePass string, databaseName string, databaseTable string) {

  dw.dbUrl = databaseUrl
  dw.dbUser = databaseUser
  dw.dbPass = databasePass
  dw.dbName = databaseName
  dw.dbTable = databaseTable

  for{

    newJob, open := <- dw.databaseChan

    if !open {
      return
    }

    dw.handleJob(newJob)

  }

}
