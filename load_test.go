package demerklizator

import (
	"os"
	"testing"

	"math/rand"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
)

const (
	strChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" // 62 characters
)

// Str constructs a random alphanumeric string of given length.
func randByte(length int) []byte {
	chars := []byte{}
MAIN_LOOP:
	for {
		val := rand.Int63() //nolint:gosec
		for i := 0; i < 10; i++ {
			v := int(val & 0x3f) // rightmost 6 bits
			if v >= 62 {         // only 62 characters in strChars
				val >>= 6
				continue
			} else {
				chars = append(chars, strChars[v])
				if len(chars) == length {
					break MAIN_LOOP
				}
				val >>= 6
			}
		}
	}

	return chars
}

func newTempDB(t *testing.T) (db dbm.DB, dbName string) {
	dbName = t.TempDir()
	db, err := openDB(dbName)
	require.NoError(t, err)
	return db, dbName
}

func checkKVStoreData(t *testing.T, kvStore store.KVStore, kvMap map[string]string) {
	itr := kvStore.Iterator(nil, nil)

	entries_num := 0
	for itr.Valid() {
		expectedValue := kvMap[string(itr.Key())]
		require.Equal(t, expectedValue, string(itr.Value()))
		entries_num += 1
		itr.Next()
	}
	itr.Close()

	require.Equal(t, entries_num, len(kvMap))
}

func TestLoadLatestStateToRootStore(t *testing.T) {
	db, dbName := newTempDB(t)
	defer os.RemoveAll(dbName)

	rs := store.NewCommitMultiStore(db).(*rootmulti.Store)
	rs.MountStoreWithDB(storetypes.NewKVStoreKey("s1"), storetypes.StoreTypeIAVL, nil)
	rs.MountStoreWithDB(storetypes.NewKVStoreKey("s2"), storetypes.StoreTypeIAVL, nil)

	err := rs.LoadLatestVersion()
	require.NoError(t, err)

	s1 := rs.GetStoreByName("s1").(store.KVStore)
	s2 := rs.GetStoreByName("s2").(store.KVStore)

	kvMapS1 := map[string]string{}
	kvMapS2 := map[string]string{}
	for i := 0; i < 10; i++ {
		rand1 := randByte(20)
		rand2 := randByte(20)

		kvMapS1[string(rand1)] = string(rand2)
		s1.Set(rand1, rand2)

		kvMapS2[string(rand2)] = string(rand1)
		s2.Set(rand2, rand1)
	}

	rs.Commit()

	err = db.Close()
	require.NoError(t, err)

	loadedRS, _ := loadLatestStateToRootStore(dbName)

	loadedS1 := loadedRS.GetStoreByName("s1").(store.KVStore)
	loadedS2 := loadedRS.GetStoreByName("s2").(store.KVStore)

	checkKVStoreData(t, loadedS1, kvMapS1)
	checkKVStoreData(t, loadedS2, kvMapS2)
}
