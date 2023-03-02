package demerklizator

import (
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	gogotypes "github.com/cosmos/gogoproto/types"
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

	loadedRS, db, err := loadLatestStateToRootStore(dbName, storetypes.StoreTypeIAVL)
	require.NoError(t, err)

	loadedS1 := loadedRS.GetStoreByName("s1").(store.KVStore)
	loadedS2 := loadedRS.GetStoreByName("s2").(store.KVStore)

	checkKVStoreData(t, loadedS1, kvMapS1)
	checkKVStoreData(t, loadedS2, kvMapS2)

	db.Close()
}

func TestFetchLatestCommitInfoFromIAVLStoreToRelationalStore(t *testing.T) {
	// Setup dbs
	merkleDBPath := t.TempDir()

	relationalDBPath := t.TempDir()
	relationalDB := openDB(relationalDBPath)

	// Cleanup
	defer func() {
		os.RemoveAll(merkleDBPath)
		os.RemoveAll(relationalDBPath)
	}()

	merkleRS, merkleDB := newRootStoreAtPath(merkleDBPath)

	mountKVStoresToRootStore(merkleRS, []string{"s1", "s2"}, storetypes.StoreTypeIAVL)

	s1 := merkleRS.GetStoreByName("s1").(store.KVStore)
	s2 := merkleRS.GetStoreByName("s2").(store.KVStore)

	setRandomDataForKVStore(s1)
	setRandomDataForKVStore(s2)

	merkleRS.Commit()

	err := merkleDB.Close()
	require.NoError(t, err)

	merkleRS, merkleDB, err = loadLatestStateToRootStore(merkleDBPath, storetypes.StoreTypeIAVL)
	require.NoError(t, err)

	fetchLatestCommitInfoFromIAVLStoreToRelationalStore(merkleDB, relationalDB)

	bz, err := merkleDB.Get([]byte(latestVersionKey))
	require.NoError(t, err)

	var latestVersion int64

	err = gogotypes.StdInt64Unmarshal(&latestVersion, bz)
	require.NoError(t, err)

	expectedCommitInfo, err := getCommitInfo(merkleDB, latestVersion)
	require.NoError(t, err)

	actualCommitInfo, err := getCommitInfo(relationalDB, latestVersion)
	require.NoError(t, err)

	require.Equal(t, expectedCommitInfo, actualCommitInfo)
}
