package db

import "database/sql"

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

//Start ..
func Start() {
	go func() {

	}()
}
