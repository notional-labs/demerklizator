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

func openDB(dbPath string) (dbm.DB, error) {
	dbName := strings.Trim(filepath.Base(dbPath), ".db")

	return sdk.NewLevelDB(dbName, filepath.Dir(dbPath))
}

func applicationDBPathFromRootDir(rootDir string) string {
	return filepath.Join(rootDir, "data", "applica")
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

func loadLatestStateToRootStore(applicationDBPath string) (rootStore *rootmulti.Store, latestVersion int64) {
	db, err := openDB(applicationDBPath)
	if err != nil {
		panic(err)
	}

	rs := store.NewCommitMultiStore(db)
	storeKeys := getStoreKeys(db)
	// mount all the module stores to root store
	for _, storeKey := range storeKeys {
		rs.MountStoreWithDB(storetypes.NewKVStoreKey(storeKey), storetypes.StoreTypeIAVL, nil)
	}
	latestVersion = rootmulti.GetLatestVersion(db)
	rs.LoadVersion(latestVersion)
	return rs.(*rootmulti.Store), latestVersion
}

func NewEmptyRootStore(applicationDBPath string) *rootmulti.Store {
	outDB, err := openDB(applicationDBPath)
	if err != nil {
		panic(err)
	}

	return store.NewCommitMultiStore(outDB).(*rootmulti.Store)
}
