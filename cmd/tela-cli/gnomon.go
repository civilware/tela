package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/civilware/Gnomon/indexer"
	"github.com/civilware/Gnomon/storage"
	"github.com/civilware/Gnomon/structures"
	"github.com/civilware/tela"
	"github.com/civilware/tela/logger"
	"github.com/civilware/tela/shards"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/walletapi"
)

type gnomes struct {
	Indexer        *indexer.Indexer
	fastsync       bool
	parallelBlocks int
}

var gnomon gnomes

const maxParallelBlocks = 10

const gnomonSearchFilterDOC = `Function init() Uint64
10 IF EXISTS("owner") == 0 THEN GOTO 30
20 RETURN 1
30 STORE("owner", address())
50 STORE("docVersion", "1`

const gnomonSearchFilterINDEX = `Function init() Uint64
10 IF EXISTS("owner") == 0 THEN GOTO 30
20 RETURN 1
30 STORE("owner", address())
50 STORE("telaVersion", "1`

// Stop all indexers and close Gnomon
func stopGnomon() (stopped bool) {
	if gnomon.Indexer != nil {
		gnomon.Indexer.Close()
		gnomon.Indexer = nil
		logger.Printf("[Gnomon] Closed all indexers\n")
		stopped = true
	}

	return
}

// Start the Gnomon indexer
func startGnomon(endpoint string) {
	if walletapi.Connected {
		if gnomon.Indexer == nil {
			dir, _ := os.Getwd()
			path := filepath.Join(dir, "datashards", "gnomon")

			switch getNetworkInfo() {
			case shards.Value.Network.Testnet():
				path = filepath.Join(dir, "datashards", "gnomon_testnet")
			case shards.Value.Network.Simulator():
				path = filepath.Join(dir, "datashards", "gnomon_simulator")
			}

			boltDB, boltErr := storage.NewBBoltDB(path, "gnomon")

			gravDB, gravErr := storage.NewGravDB(path, "25ms")

			dbType := shards.GetDBType()

			var err error
			var height int64
			switch dbType {
			case "boltdb":
				if boltErr != nil {
					if !strings.HasPrefix(boltErr.Error(), "[") {
						boltErr = fmt.Errorf("[NewBBoltDB] %s", boltErr)
					}
					logger.Errorf("%s\n", boltErr)
					return
				}

				height, err = boltDB.GetLastIndexHeight()
				if err != nil {
					height = 0
				}
			default:
				if gravErr != nil {
					logger.Errorf("%s\n", gravErr)
					return
				}

				height, err = gravDB.GetLastIndexHeight()
				if err != nil {
					height = 0
				}
			}

			exclusions := []string{"bb43c3eb626ee767c9f305772a6666f7c7300441a0ad8538a0799eb4f12ebcd2"}
			filter := []string{gnomonSearchFilterDOC, gnomonSearchFilterINDEX}

			// Fastsync Config
			config := &structures.FastSyncConfig{
				Enabled:           gnomon.fastsync,
				SkipFSRecheck:     false,
				ForceFastSync:     true,
				ForceFastSyncDiff: 100,
				NoCode:            false,
			}

			gnomon.Indexer = indexer.NewIndexer(gravDB, boltDB, dbType, filter, height, endpoint, "daemon", false, false, config, exclusions)

			indexer.InitLog(globals.Arguments, os.Stdout)

			go gnomon.Indexer.StartDaemonMode(gnomon.parallelBlocks)

			logger.Printf("[Gnomon] Scan Status: [%d / %d]\n", gnomon.Indexer.LastIndexedHeight, walletapi.Get_Daemon_Height())
		}
	}
}

// Method of Gnomon GetAllOwnersAndSCIDs() where DB type is defined by Indexer.DBType
func (g *gnomes) GetAllOwnersAndSCIDs() (scids map[string]string) {
	switch g.Indexer.DBType {
	case "gravdb":
		return g.Indexer.GravDBBackend.GetAllOwnersAndSCIDs()
	case "boltdb":
		return g.Indexer.BBSBackend.GetAllOwnersAndSCIDs()
	default:
		return
	}
}

// Method of Gnomon GetAllSCIDVariableDetails() where DB type is defined by Indexer.DBType
func (g *gnomes) GetAllSCIDVariableDetails(scid string) (vars []*structures.SCIDVariable) {
	switch g.Indexer.DBType {
	case "gravdb":
		return g.Indexer.GravDBBackend.GetAllSCIDVariableDetails(scid)
	case "boltdb":
		return g.Indexer.BBSBackend.GetAllSCIDVariableDetails(scid)
	default:
		return
	}
}

// Method of Gnomon GetSCIDValuesByKey() where DB type is defined by Indexer.DBType
func (g *gnomes) GetSCIDValuesByKey(scid string, key interface{}) (valuesstring []string, valuesuint64 []uint64) {
	switch g.Indexer.DBType {
	case "gravdb":
		return g.Indexer.GravDBBackend.GetSCIDValuesByKey(scid, key, g.Indexer.ChainHeight, true)
	case "boltdb":
		return g.Indexer.BBSBackend.GetSCIDValuesByKey(scid, key, g.Indexer.ChainHeight, true)
	default:
		return
	}
}

// Method of Gnomon GetSCIDKeysByValue() where DB type is defined by Indexer.DBType
func (g *gnomes) GetSCIDKeysByValue(scid string, key interface{}) (valuesstring []string, valuesuint64 []uint64) {
	switch g.Indexer.DBType {
	case "gravdb":
		return g.Indexer.GravDBBackend.GetSCIDKeysByValue(scid, key, g.Indexer.ChainHeight, true)
	case "boltdb":
		return g.Indexer.BBSBackend.GetSCIDKeysByValue(scid, key, g.Indexer.ChainHeight, true)
	default:
		return
	}
}

// Get the "owner" store from a SCID
func (g *gnomes) GetOwnerAddress(scid string) (owner string) {
	if g.Indexer == nil {
		return
	}

	address, _ := g.GetSCIDValuesByKey(scid, "owner")
	if address != nil && address[0] != "anon" {
		owner = address[0]
	}

	return
}

// Method of Gnomon Indexer.AddSCIDToIndex() configured for TELA-CLI
func (g *gnomes) AddSCIDToIndex(scids map[string]*structures.FastSyncImport) error {
	return g.Indexer.AddSCIDToIndex(scids, false, false)
}

// Manually add SCID(s) to Gnomon index if they can be validated as TELA contracts
func (t *tela_cli) addToIndex(scids []string) (err error) {
	for _, scid := range scids {
		if len(scid) == 64 {
			_, err := tela.GetDOCInfo(scid, t.endpoint)
			if err != nil {
				_, errr := tela.GetINDEXInfo(scid, t.endpoint)
				if errr != nil {
					logger.Errorf("[%s] GetDOCInfo: %s\n", appName, err)
					logger.Errorf("[%s] GetINDEXInfo: %s\n", appName, errr)
					return fmt.Errorf("could not validate %s as TELA INDEX or DOC", scid)
				}
			}
		}
	}

	filters := gnomon.Indexer.SearchFilter
	gnomon.Indexer.SearchFilter = []string{}
	scidsToAdd := map[string]*structures.FastSyncImport{}

	for _, sc := range scids {
		scidsToAdd[sc] = &structures.FastSyncImport{}
	}

	err = gnomon.AddSCIDToIndex(scidsToAdd)
	gnomon.Indexer.SearchFilter = filters

	return
}
