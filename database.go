package main

import (
	"database/sql"
	"time"

	"github.com/karlseguin/ccache"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sjsafranek/ligneous"
)

var DB *Database

func init() {
	db, err := NewDatabase(ligneous.AddLogger("database", "trace", "logs"))
	if nil != err {
		panic(err)
	}
	err = db.MakeTables()
	if nil != err {
		panic(err)
	}

	DB = db
}

func NewDatabase(log ligneous.Log) (*Database, error) {
	db := Database{log: log}
	db.cache = ccache.Layered(ccache.Configure().MaxSize(100).ItemsToPrune(10))
	err := db.Open()
	return &db, err
}

type Database struct {
	db    *sql.DB
	log   ligneous.Log
	cache *ccache.LayeredCache
}

func (self *Database) Open() error {
	db, err := sql.Open("sqlite3", "db.sqlite3?cache=shared&mode=rwc&_busy_timeout=50000000")
	self.db = db
	self.buildCache()
	return err
}

// buildCache
func (self *Database) buildCache() error {
	rows, err := self.db.Query("SELECT type, path, content FROM assets;")
	if err != nil {
		self.log.Error(err)
		return err
	}

	defer rows.Close()
	for rows.Next() {
		var atype string
		var path string
		var content string
		err = rows.Scan(&atype, &path, &content)
		if err != nil {
			self.log.Error(err)
			return err
		}
		self.cache.Set(atype, path, content, time.Minute*5)
	}

	return rows.Err()
}

func (self *Database) MakeTables() error {
	self.log.Debugf("create database tables")
	_, err := self.db.Exec(TABLES_SQL)
	return err
}

func (self *Database) GetAsset(atype, path string) (string, error) {
	// get from cache
	item := self.cache.Get(atype, path)
	if nil != item {
		self.log.Tracef("got asset from cache: %v %v", atype, path)
		return item.Value().(string), nil
	}

	// get from database
	self.log.Tracef("get asset from database: %v %v", atype, path)
	stmt, err := self.db.Prepare("SELECT content FROM assets WHERE path=? and type=?")
	if err != nil {
		self.log.Error(err)
		return "", err
	}
	defer stmt.Close()

	var content string
	err = stmt.QueryRow(path, atype).Scan(&content)
	if err != nil {
		self.log.Error(err)
		return "", err
	}

	// add to cache
	self.cache.Set(atype, path, content, time.Minute*5)

	return content, err
}

func (self *Database) SaveAsset(atype, path, content string) error {
	self.log.Tracef("save asset from database: %v %v", atype, path)

	// replace item in cache
	self.cache.Delete(atype, path)
	self.cache.Set(atype, path, content, time.Minute*5)

	err := self.execWriteQuery("INSERT INTO assets(type, path, content) VALUES(?, ?, ?)", atype, path, content)
	if nil != err {
		return self.UpdateAsset(atype, path, content)
	}
	return nil
}

func (self *Database) UpdateAsset(atype, path, content string) error {
	self.log.Tracef("update asset from database: %v %v", atype, path)
	return self.execWriteQuery("UPDATE assets SET content=? WHERE path=? AND type=?", content, path, atype)
}

func (self *Database) DeleteAsset(atype, path string) error {
	self.log.Tracef("delete asset from database: %v %v", atype, path)
	return self.execWriteQuery("DELETE FROM assets WHERE path=? AND type=?", path, atype)
}

func (self *Database) execWriteQuery(query string, args ...interface{}) error {
	tx, err := self.db.Begin()
	if err != nil {
		tx.Rollback()
		self.log.Error(err)
		return err
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		self.log.Error(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if nil != err {
		tx.Rollback()
		self.log.Error(err)
		return err
	}

	tx.Commit()
	return nil
}
