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
func randStr(length int) string {
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

	return string(chars)
}

func newTempDB(t *testing.T) (db dbm.DB, dbName string) {
	dbName = randStr(12) + ".db"
	db, err := openDB(dbName)
	require.NoError(t, err)
	return db, dbName
}

func checkStore

func TestLoadLatestStateToRootStore(t *testing.T) {
	db, dbName := newTempDB(t)
	defer os.RemoveAll(dbName)

	rs := store.NewCommitMultiStore(db).(*rootmulti.Store)
	rs.MountStoreWithDB(storetypes.NewKVStoreKey("s1"), storetypes.StoreTypeIAVL, nil)
	rs.MountStoreWithDB(storetypes.NewKVStoreKey("s2"), storetypes.StoreTypeIAVL, nil)

	s1 := rs.GetStoreByName("s1").(store.KVStore)
	s2 := rs.GetStoreByName("s2").(store.KVStore)

	s1.Set([]byte("key1"), []byte("value1"))
	s2.Set([]byte("key2"), []byte("value2"))
	
	rs.Commit()

	err := db.Close()
	require.NoError(t, err)

	rs, latestVer := loadLatestStateToRootStore(dbName)
	
	





}
