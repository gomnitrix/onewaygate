package sqlhelper

import (
	"database/sql"
	"fmt"

	"controller.com/internal/OwmError"

	"controller.com/Model"

	"controller.com/config"
	_ "github.com/go-sql-driver/mysql"
)

type DbHelper struct {
	Db         *sql.DB
	DbName     string
	DbUserName string
	DbPassWd   string
	DbNetWork  string
	DbServer   string
	DbPort     int
}

type ContMap struct {
	TargetID  string `json:"target"`
	ManagerID string `json:"manager"`
}

func GetNewHelper() *DbHelper {
	helper := &DbHelper{
		DbName:     config.DbName,
		DbUserName: config.DbUserName,
		DbPassWd:   config.DbPassWd,
		DbNetWork:  config.DbNetWork,
		DbServer:   config.DbServer,
		DbPort:     config.DbPort,
	}
	helper.Open()
	helper.CreateTable()
	return helper
}

func (helper DbHelper) getConn() (conn string) {
	conn = fmt.Sprintf(
		"%s:%s@%s(%s:%d)/%s",
		helper.DbUserName,
		helper.DbPassWd,
		helper.DbNetWork,
		helper.DbServer,
		helper.DbPort,
		helper.DbName,
	)
	return
}
func (helper *DbHelper) Open() {
	conn := helper.getConn()
	db, err := sql.Open("mysql", conn)
	if err != nil {
		fmt.Println("connection to mysql failed:", err)
		return
	}
	db.SetConnMaxLifetime(config.ConnMaxLifeTime)
	db.SetMaxOpenConns(config.MaxOpenConns)
	helper.Db = db
}

func (helper *DbHelper) Close() {
	helper.Db.Close()
}

func (helper *DbHelper) CreateTable() bool {
	// create target<-->manager map table
	sql := `CREATE TABLE IF NOT EXISTS contmap(
		target VARCHAR(12) PRIMARY KEY NOT NULL,
		manager VARCHAR(12) NOT NULL
	); `

	if _, err := helper.Db.Exec(sql); err != nil {
		fmt.Println("create contmap table failed:", err)
		return false
	}

	// create user Table
	sql = `CREATE TABLE IF NOT EXISTS userlist(
		uname VARCHAR(12) PRIMARY KEY NOT NULL,
		passwd VARCHAR(12) NOT NULL
	); `
	if _, err := helper.Db.Exec(sql); err != nil {
		fmt.Println("create userlist table failed:", err)
		return false
	}

	return true
}

func (helper *DbHelper) GetChMap() map[string]chan bool {
	var chmap = make(map[string]chan bool)
	var manager string
	rows, err := helper.Db.Query("select manager from contmap")
	defer rows.Close()
	if err != nil {
		fmt.Printf("Query failed,err:%v\n", err)
		return nil
	}
	for rows.Next() {
		err = rows.Scan(&manager)
		if err != nil {
			fmt.Printf("Scan failed,err:%v\n", err)
			return nil
		}
		chmap[manager] = make(chan bool)
	}
	return chmap
}

func (helper *DbHelper) InputConts(targetID, managerID string) {
	if targetID == "" || managerID == "" {
		return
	}
	result, err := helper.Db.Exec("insert INTO contmap(target,manager) values(?,?)", targetID, managerID)
	if err != nil {
		fmt.Printf("Insert data failed,err:%v\n", err)
		return
	}
	rowsaffected, err := result.RowsAffected() //通过RowsAffected获取受影响的行数
	if err != nil {
		fmt.Printf("Get RowsAffected failed,err:%v\n", err)
		return
	}
	if rowsaffected != 1 {
		fmt.Println("some stange things happened while inserting ")
		return
	}
}

func (helper *DbHelper) DeleteConts(targetID string) {
	if targetID == "" {
		return
	}
	result, err := helper.Db.Exec("delete from contmap where target=?", targetID)
	if err != nil {
		fmt.Printf("Delete data failed,err:%v\n", err)
		return
	}
	rowsaffected, err := result.RowsAffected() //通过RowsAffected获取受影响的行数
	if err != nil {
		fmt.Printf("Get RowsAffected failed,err:%v\n", err)
		return
	}
	if rowsaffected != 1 {
		fmt.Println("some stange things happened while deleting ")
		return
	}
}

func (helper *DbHelper) queryUser(name string) {
	defer OwmError.Pack()
	result, err := helper.Db.Exec("select * from userlist where uname = ?", name)
	OwmError.Check(err, false, "Query user %s error\n", name)
	rowsaffected, err := result.RowsAffected()
	OwmError.Check(err, false, "Db RowsAffected Error\n")
	if rowsaffected >= 1 {
		OwmError.Check(err, false, "User: %s exist\n", name)
	}
}

func (helper *DbHelper) InputUser(user Model.User) {
	defer OwmError.Pack()
	helper.queryUser(user.Name)
	result, err := helper.Db.Exec("insert INTO userlist(uname,passwd) values(?,?)", user.Name, user.Passwd)
	OwmError.Check(err, false, "Insert user: %s failed\n", user.Name)
	rowsaffected, err := result.RowsAffected() //通过RowsAffected获取受影响的行数
	OwmError.Check(err, false, "Db RowsAffected Error\n")
	if rowsaffected != 1 {
		OwmError.Check(err, false, "Some thing wrong When insert user: %s\n", user.Name)
	}
}
