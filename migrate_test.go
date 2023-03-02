package demerklizator

import (
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/stretchr/testify/require"
)

func TestMigrateIAVLStoreToDBStore(t *testing.T) {
	dbName := t.TempDir()

	rs, db := newRootStoreAtPath(dbName)
	defer os.RemoveAll(dbName)

	mountKVStoresToRootStore(rs, []string{"store1", "store2"}, storetypes.StoreTypeIAVL)

	store1 := rs.GetStoreByName("store1").(store.KVStore)
	store2 := rs.GetStoreByName("store2").(store.KVStore)

	kvMapStore1 := setRandomDataForKVStore(store1)
	kvMapStore2 := setRandomDataForKVStore(store2)

	rs.Commit()

	err := db.Close()
	require.NoError(t, err)

	migrateDBName := t.TempDir()
	defer os.RemoveAll(migrateDBName)

	err = MigrateLatestStateDataToDBStores(dbName, migrateDBName)
	require.NoError(t, err)

	migratedRS, migratedDB, err := loadLatestStateToRootStore(migrateDBName, storetypes.StoreTypeDB)
	require.NoError(t, err)

	migratedStore1 := migratedRS.GetStoreByName("store1").(store.KVStore)
	migratedStore2 := migratedRS.GetStoreByName("store2").(store.KVStore)

	checkKVStoreData(t, migratedStore1, kvMapStore1)
	checkKVStoreData(t, migratedStore2, kvMapStore2)

	migratedDB.Close()
}
