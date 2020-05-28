package gosql

import (
	"sync"
)

var _collect = make(map[string]Cluster, 0)
var _collectMutex sync.RWMutex

const defaultCollect = "default"

//NewCollect ..
func NewCollect(clst Cluster, name ...string) {
	_collectMutex.Lock()
	defer _collectMutex.Unlock()
	if len(name) == 0 {
		_collect[defaultCollect] = clst
		return
	}
	_collect[name[0]] = clst
	return
}

//Collect get a cluster by name
func Collect(name ...string) Cluster {
	_collectMutex.RLock()
	defer _collectMutex.RUnlock()
	if len(name) == 0 {
		return _collect[defaultCollect]
	}
	return _collect[name[0]]
}
