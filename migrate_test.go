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
	kvMapStore1 := setDataForKVStore(store1)
	kvMapStore2 := setDataForKVStore(store2)

	rs.Commit()

	err := db.Close()
	require.NoError(t, err)

	migrateDBName := t.TempDir()
	defer os.RemoveAll(migrateDBName)
	MigrateLatestStateDataToDBStores(dbName, migrateDBName)

	migratedRS, migratedDB := loadLatestStateToRootStore(migrateDBName, storetypes.StoreTypeDB)

	migratedStore1 := migratedRS.GetStoreByName("store1").(store.KVStore)
	migratedStore2 := migratedRS.GetStoreByName("store2").(store.KVStore)

	checkKVStoreData(t, migratedStore1, kvMapStore1)
	checkKVStoreData(t, migratedStore2, kvMapStore2)

	migratedDB.Close()
}
