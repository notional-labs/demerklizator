package demerklizator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tm-db"
)

const (
	commitInfoKeyFmt = "s/%d" // s/<version>
)

func openDB(dbPath string) dbm.DB {
	dbName := strings.Trim(filepath.Base(dbPath), ".db")

	db, err := sdk.NewLevelDB(dbName, filepath.Dir(dbPath))
	if err != nil {
		panic(err)
	}
	return db
}

func mountKVStoresToRootStore(rs *rootmulti.Store, keys []string, storetyp storetypes.StoreType) {
	for _, key := range keys {
		rs.MountStoreWithDB(storetypes.NewKVStoreKey(key), storetyp, nil)
	}

	// load lastest version so that store is added to rs.stores as per LoadVersion() logic
	err := rs.LoadLatestVersion()
	if err != nil {
		panic(err)
	}
}

func ApplicationDBPathFromRootDir(rootDir string) string {
	return filepath.Join(rootDir, "data", "application.db")
}

func getCommitInfo(db dbm.DB, ver int64) (*storetypes.CommitInfo, error) {
	cInfoKey := fmt.Sprintf(commitInfoKeyFmt, ver)

	bz, err := db.Get([]byte(cInfoKey))
	if err != nil {
		return nil, fmt.Errorf("failed to get commit info: %s", err)
	} else if bz == nil {
		return nil, fmt.Errorf("no commit info found")
	}

	cInfo := &storetypes.CommitInfo{}
	if err = cInfo.Unmarshal(bz); err != nil {
		return nil, fmt.Errorf("failed unmarshal commit info: %s", err)
	}

	return cInfo, nil
}

func getStoreKeys(db dbm.DB) (storeKeys []string) {
	latestVer := rootmulti.GetLatestVersion(db)
	latestCommitInfo, err := getCommitInfo(db, latestVer)
	if err != nil {
		panic(err)
	}

	for _, storeInfo := range latestCommitInfo.StoreInfos {
		storeKeys = append(storeKeys, storeInfo.Name)
	}
	return
}

func loadLatestStateToRootStore(applicationDBPath string, storetype storetypes.StoreType) (rootStore *rootmulti.Store, db dbm.DB) {
	rootStore, db = newRootStoreAtPath(applicationDBPath)

	storeKeys := getStoreKeys(db)
	// mount all the module stores to root store
	mountKVStoresToRootStore(rootStore, storeKeys, storetype)

	rootStore.LoadLatestVersion()
	return rootStore, db
}

func newRootStoreAtPath(dbPath string) (*rootmulti.Store, dbm.DB) {
	db := openDB(dbPath)

	rootStore := store.NewCommitMultiStore(db).(*rootmulti.Store)
	return rootStore, db
}
