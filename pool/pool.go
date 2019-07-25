package pool

import (
	"database/sql"
	"fmt"
)

//Configs ...
type Configs map[string]Config

//Config ..
type Config struct {
	DbType string   `yaml:"db_type"`
	Nodes  []string `yaml:"nodes"`
}

//PoolsDB ..
type PoolsDB map[string]*sql.DB

//Pools ..
var Pools = make(PoolsDB, 0)

type Cluster struct{}

func (c Cluster) Get(name, node string) (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:123321@tcp(192.168.33.10:3306)/auth?parseTime=true")
	return db, err
}

//Init ..
func Init(dbType string, settings map[string]map[string][]string) *Cluster {
	for k, v := range settings {
		fmt.Println(k, v)
	}
	c := &Cluster{}
	return c
}

//Start ..
func Start() {
	go func() {

	}()
}
