package demerklizator

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

// This only copy the state data from the iavl store, which excludes the data produced by merklization
// That state data is the state data of a chain's module store
func copyStateDataFromIAVLStoreToDBStore(iavlStore storetypes.KVStore, dbStore storetypes.KVStore) {
	itr := iavlStore.Iterator(nil, nil)
	for itr.Valid() {
		dbStore.Set(itr.Key(), itr.Value())
		itr.Next()
	}
	itr.Close()
}

func MigrateLatestStateDataToDBStores(applicationDBPath string, outApplicationDBPath string) error {
	rootStore, db, err := loadLatestStateToRootStore(applicationDBPath, storetypes.StoreTypeIAVL)
	if err != nil {
		return fmt.Errorf("Failed to load latest state to root store: %s", err.Error())
	}

	latestVersion := rootmulti.GetLatestVersion(db)

	outRootStore, outDB := newRootStoreAtPath(outApplicationDBPath)

	// get all the stores from rootStore, which is all iavl stores
	iavlStores := rootStore.GetStores()

	// mount all the empty db stores to outRootStore
	// for each iavl stores mounted on rootStore,
	// we mount an empty db store on rootStore with the same key
	for storeKey := range iavlStores {
		outRootStore.MountStoreWithDB(storeKey, storetypes.StoreTypeDB, nil)
	}
	err = outRootStore.LoadLatestVersion()
	if err != nil {
		return err
	}

	err = outRootStore.SetInitialVersion(latestVersion)
	if err != nil {
		return err
	}

	// get all the stores from outRootStore, which is empty db stores
	dbStores := outRootStore.GetStores()

	// copy the state data from iavl stores of rootStore to db stores of outRootStore
	for storeKey, iavlStore := range iavlStores {
		dbStore := dbStores[storeKey]
		copyStateDataFromIAVLStoreToDBStore(iavlStore, dbStore)
	}
	outRootStore.Commit()
	outDB.Close()
	db.Close()

	return nil
}
