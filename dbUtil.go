package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"

	_ "github.com/go-sql-driver/mysql"
)

var (
	ruokDB          *sql.DB
	defaultUser     string
	defaultPassword string
)

const (
	dbConfig = "db.json"
)

type dbStruct struct {
	Host        string `json:"host"`
	User        string `json:"user"`
	Port        int    `json:"port"`
	Password    string `json:"password"`
	Database    string `json:"database"`
	DefUser     string `json:"defaultUser"`
	DefPassword string `json:"defaultPassword"`
}

func connDB() {
	dbContent, err := ioutil.ReadFile(dbConfig)
	checkError(err)
	conf := dbStruct{}
	err = json.Unmarshal(dbContent, &conf)
	checkError(err)
	defaultUser = conf.DefUser
	defaultPassword = conf.DefPassword
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%d)/", conf.User, conf.Password, conf.Host, conf.Port)
	ruokDB, err = sql.Open("mysql", dataSource)
	checkError(err)

	dbExec("create database if not exists " + conf.Database)
	dbExec("use " + conf.Database)
}

func closeDB() {
	ruokDB.Close()
}

func createTables() {
	var sqlStr string
	for _, v := range pageConfigs {
		menuItem := v.(map[string]interface{})

		if forceDb {
			sqlStr = fmt.Sprintf("drop table if exists `%v`", menuItem["name"])
			logF("[drop table]:%s\n", sqlStr)
			dbExec(sqlStr)
		}

		mib, err := json.Marshal(menuItem)
		checkError(err)

		type menu struct {
			Data []struct {
				Name  string `json:"name"`
				Title string `json:"title"`
			} `json:"data"`
		}

		var pageMenu menu
		err = json.Unmarshal(mib, &pageMenu)
		checkError(err)

		sqlStr = fmt.Sprintf("create table if not exists `%v` (`id` int not null auto_increment primary key", menuItem["name"])
		for _, val := range pageMenu.Data {
			sqlStr = fmt.Sprintf("%s, `%s` varchar(200)", sqlStr, val.Name)
		}
		sqlStr += ")"
		logF("[create table]:%s\n", sqlStr)
		dbExec(sqlStr)
	}
	userCount := sqlQuery("select count(*) as count from `user` where `username` = ? and `password` = ?", defaultUser, defaultPassword)
	if userCount[0]["count"] == int64(0) {
		dbExec("insert into `user` (`username`, `password`) values (?,?)", defaultUser, defaultPassword)
	}
}

func dbExec(query string, args ...interface{}) {
	_, err := ruokDB.Exec(query, args...)
	checkError(err)
}

func sqlQuery(query string, args ...interface{}) []map[string]interface{} {
	rows, err := ruokDB.Query(query, args...)
	checkError(err)
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}

	values := make([]interface{}, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var queryResult []map[string]interface{}
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		checkError(err)

		record := make(map[string]interface{})
		for i, value := range values {
			switch value.(type) {
			case nil:
				record[columns[i]] = "NULL"
			case []byte:
				record[columns[i]] = string(value.([]byte))
			default:
				record[columns[i]] = value
			}
		}
		queryResult = append(queryResult, record)
	}

	return queryResult
}
