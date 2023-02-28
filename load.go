package demerklizator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	dbm "github.com/tendermint/tm-db"
)

const (
	latestVersionKey = "s/latest"
	commitInfoKeyFmt = "s/%d" // s/<version>
)

// openDB instantiates LevelDB database
func openDB(dbPath string) dbm.DB {
	dbName := strings.Trim(filepath.Base(dbPath), ".db")

	db, err := sdk.NewLevelDB(dbName, filepath.Dir(dbPath))
	if err != nil {
		panic(err)
	}
	return db
}

// mountKVStoresToRootStore populates rootmulti.Store with sub KV stores
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

// ApplicationDBPathFromRootDir returns default path to database
func ApplicationDBPathFromRootDir(rootDir string) string {
	return filepath.Join(rootDir, "data", "application.db")
}

// getCommitInfo fetches block's commit info from disk
func getCommitInfo(db dbm.DB, ver int64) (*storetypes.CommitInfo, error) {

	bz, err := getCommitInfoBz(db, ver)

	cInfo := &storetypes.CommitInfo{}
	if err = cInfo.Unmarshal(bz); err != nil {
		return nil, fmt.Errorf("failed unmarshal commit info: %s", err)
	}

	return cInfo, nil
}

// getCommitInfo fetches block's commit info in bytes form from the disk
func getCommitInfoBz(db dbm.DB, ver int64) ([]byte, error) {
	cInfoKey := fmt.Sprintf(commitInfoKeyFmt, ver)

	bz, err := db.Get([]byte(cInfoKey))
	if err != nil {
		return nil, fmt.Errorf("failed to get commit info: %s", err)
	} else if bz == nil {
		return nil, fmt.Errorf("no commit info found")
	}
	return bz, nil
}

// getStoreKeys gets store keys of a latest version in database
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

// loadLatestStateToRootStore loads a latest state of database to root multi store
func loadLatestStateToRootStore(applicationDBPath string, storetype storetypes.StoreType) (rootStore *rootmulti.Store, db dbm.DB, err error) {
	rootStore, db = newRootStoreAtPath(applicationDBPath)

	storeKeys := getStoreKeys(db)
	// mount all the module stores to root store
	mountKVStoresToRootStore(rootStore, storeKeys, storetype)

	err = rootStore.LoadLatestVersion()
	if err != nil {
		return nil, nil, err
	}

	return
}

// newRootStoreAtPath returns an instance of commit multistore at specified database path
func newRootStoreAtPath(dbPath string) (*rootmulti.Store, dbm.DB) {
	db := openDB(dbPath)

	rootStore := store.NewCommitMultiStore(db).(*rootmulti.Store)
	return rootStore, db
}

func fetchLatestCommitInfoFromIAVLStoreToRelationalStore(merkleDBPath, relationalDBPath string) error {
	merkleDB := openDB(merkleDBPath)
	relationalDB := openDB(relationalDBPath)

	bz, err := merkleDB.Get([]byte(latestVersionKey))
	if err != nil {
		return err
	}

	var latestVersion int64

	if err := gogotypes.StdInt64Unmarshal(&latestVersion, bz); err != nil {
		return err
	}

	bz, err = getCommitInfoBz(merkleDB, latestVersion)
	if err != nil {
		return err
	}

	commitInfoKey := []byte(fmt.Sprintf(commitInfoKeyFmt, latestVersion))

	relationalDB.Set(commitInfoKey, bz)
	return nil
}
