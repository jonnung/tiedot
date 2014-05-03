// Database logic.
package dbsvc

import (
	"errors"
	"fmt"
	"github.com/HouzuoGuo/tiedot/tdlog"
	"net/rpc"
	"path"
	"strconv"
	"strings"
	"time"
)

type DBSvc struct {
	srvWorkingDir   string
	dataDir         string
	totalRank       int
	data            []*rpc.Client                  // Connections to data partitions
	schema          map[string]map[string][]string // Collection => Index name => Index path ^^
	mySchemaVersion int64
}

// Create a new Client, connect to all server ranks.
func NewDBSvc(totalRank int, srvWorkingDir string, dataDir string) (db *DBSvc, err error) {
	db = &DBSvc{srvWorkingDir, dataDir, totalRank, make([]*rpc.Client, totalRank), make(map[string]map[string][]string), time.Now().UnixNano()}
	for i := 0; i < totalRank; i++ {
		if db.data[i], err = rpc.Dial("unix", path.Join(srvWorkingDir, strconv.Itoa(i))); err != nil {
			return
		}
	}
	return
}

// Shutdown all data partitions.
func (db *DBSvc) Shutdown() (err error) {
	discard := new(bool)
	errs := make([]string, 0, 1)
	for i, srv := range db.data {
		if err := srv.Call("DataSvc.Shutdown", false, discard); err == nil || !strings.Contains(fmt.Sprint(err), "unexpected EOF") {
			errs = append(errs, fmt.Sprintf("Could not shutdown server rank %d", i))
		}
	}
	if len(errs) > 0 {
		err = errors.New(strings.Join(errs, "; "))
		tdlog.Errorf("Shutdown did not fully complete, but best effort has been made: %v", err)
	}
	return
}