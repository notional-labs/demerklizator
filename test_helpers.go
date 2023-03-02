package demerklizator

import (
	"testing"

	"math/rand"

	"github.com/cosmos/cosmos-sdk/store"
	"github.com/stretchr/testify/require"
)

const (
	strChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" // 62 characters
)

// Str constructs a random alphanumeric string of given length.
func randByte(length int) []byte {
	chars := []byte{}
	for i := 0; i < length; i++ {
		v := rand.Int()
		a := v % 62
		chars = append(chars, strChars[a])
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

func setRandomDataForKVStore(kvStore store.KVStore) (kvMap map[string]string) {
	kvMap = map[string]string{}

	for i := 0; i < 10; i++ {
		key := randByte(20)
		value := randByte(20)

		kvMap[string(key)] = string(value)
		kvStore.Set(key, value)
	}

	return kvMap
}
