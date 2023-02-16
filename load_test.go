package demerklizator

import (
	"os"
	"testing"

	"math/rand"

	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/stretchr/testify/require"
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

func setDataForKVStore(kvStore store.KVStore) (kvMap map[string]string) {
	kvMap = map[string]string{}

	for i := 0; i < 10; i++ {
		key := randByte(20)
		value := randByte(20)

		kvMap[string(key)] = string(value)
		kvStore.Set(key, value)
	}

	return kvMap
}

func TestLoadLatestStateToRootStore(t *testing.T) {
	dbName := t.TempDir()
	defer os.RemoveAll(dbName)

	rs, db := newRootStoreAtPath(dbName)

	mountKVStoresToRootStore(rs, []string{"s1", "s2"}, storetypes.StoreTypeIAVL)

	s1 := rs.GetStoreByName("s1").(store.KVStore)
	s2 := rs.GetStoreByName("s2").(store.KVStore)

	kvMapS1 := setDataForKVStore(s1)
	kvMapS2 := setDataForKVStore(s2)

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
