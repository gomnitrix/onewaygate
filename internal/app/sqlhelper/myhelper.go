package sqlhelper

import (
	"database/sql"
	"errors"
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
		manager VARCHAR(12) NOT NULL,
		usrname VARCHAR(15) NOT NULL Default 'Admin'
	); `

	if _, err := helper.Db.Exec(sql); err != nil {
		fmt.Println("create contmap table failed:", err)
		return false
	}

	// create user Table
	sql = `CREATE TABLE IF NOT EXISTS userlist(
		uname VARCHAR(12) PRIMARY KEY NOT NULL,
		passwd VARCHAR(32) NOT NULL
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

func (helper *DbHelper) GetTargetsByMgr(mgrID string) []string {
	defer OwmError.Pack()
	rows, err := helper.Db.Query("select target from contmap where manager=?", mgrID)
	defer rows.Close()
	OwmError.Check(err, false, "Query targets by manager: %s failed\n", mgrID)
	var targets []string
	for rows.Next() {
		var target string
		err = rows.Scan(&target)
		OwmError.Check(err, false, "Scan targets by manager: %s failed\n", mgrID)
		targets = append(targets, target)
	}
	return targets
}

func (helper *DbHelper) GetManagerByTgt(tgtID string) string {
	defer OwmError.Pack()
	row := helper.Db.QueryRow("select manager from contmap where target=?", tgtID)
	var manager string
	err := row.Scan(&manager)
	if err == sql.ErrNoRows {
		OwmError.Check(err, false, "Scan manager by target ID: %s failed\n", tgtID)
	}
	return manager
}

func (helper *DbHelper) GetMgrsByUser(userName string) []string {
	defer OwmError.Pack()
	rows, err := helper.Db.Query("select distinct manager from contmap where usrname=?", userName)
	defer rows.Close()
	OwmError.Check(err, false, "Query containers by user: %s failed\n", userName)
	var managers []string
	for rows.Next() {
		var manager string
		err = rows.Scan(&manager)
		OwmError.Check(err, false, "Scan targets by manager: %s failed\n", userName)
		managers = append(managers, manager)
	}
	return managers
}

func (helper *DbHelper) InputConts(usrName, targetID, managerID string) {
	defer OwmError.Pack()
	if targetID == "" || managerID == "" {
		OwmError.Check(errors.New("Target or Manager ID could not be empty"), false, "Input Containers Failed\n")
	}
	result, err := helper.Db.Exec("insert INTO contmap(target,manager,usrname) values(?,?,?)", targetID, managerID, usrName)
	OwmError.Check(err, false, "Insert Containers failed\n")
	rowsaffected, err := result.RowsAffected()
	OwmError.Check(err, false, "Get RowsAffected failed\n")
	if rowsaffected != 1 {
		OwmError.Check(err, false, "some stange things happened while inserting\n")
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
	rows, err := helper.Db.Query("select * from userlist where uname=?", name)
	OwmError.Check(err, false, "Query user %s error\n", name)
	defer rows.Close()
	if rows.Next() {
		OwmError.Check(errors.New("UserExistError"), false, "User %s Already Exist\n", name)
	}
}

func (helper *DbHelper) QueryPasswd(name string) string {
	defer OwmError.Pack()
	var passwd string
	row := helper.Db.QueryRow("select passwd from userlist where uname=?", name)
	err := row.Scan(&passwd)
	if err == sql.ErrNoRows {
		OwmError.Check(OwmError.GetUserNotExistError(name), false, "Query Password failed")
	}
	return passwd
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
