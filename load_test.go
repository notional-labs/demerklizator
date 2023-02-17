package demerklizator

import (
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/stretchr/testify/require"
)

func TestLoadLatestStateToRootStore(t *testing.T) {
	dbName := t.TempDir()
	defer os.RemoveAll(dbName)

	rs, db := newRootStoreAtPath(dbName)

	mountKVStoresToRootStore(rs, []string{"s1", "s2"}, storetypes.StoreTypeIAVL)

	s1 := rs.GetStoreByName("s1").(store.KVStore)
	s2 := rs.GetStoreByName("s2").(store.KVStore)

	kvMapS1 := setRandomDataForKVStore(s1)
	kvMapS2 := setRandomDataForKVStore(s2)

	rs.Commit()

	err := db.Close()
	require.NoError(t, err)

	loadedRS, db := loadLatestStateToRootStore(dbName, storetypes.StoreTypeIAVL)

	loadedS1 := loadedRS.GetStoreByName("s1").(store.KVStore)
	loadedS2 := loadedRS.GetStoreByName("s2").(store.KVStore)

	checkKVStoreData(t, loadedS1, kvMapS1)
	checkKVStoreData(t, loadedS2, kvMapS2)

	db.Close()
}
