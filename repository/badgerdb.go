package repository

import (
	"bytes"
	"fmt"

	"github.com/Inflowenger/inflow-inspector-api/env"
	"github.com/bytedance/sonic"
	"github.com/dgraph-io/badger/v4"
)

var db *badger.DB

type BadgerHolder[T any] struct {
	db    *badger.DB
	model T
}

func connect() {

	var err error
	db, err = badger.Open(badger.DefaultOptions(env.GetDbStoreBasePath()))
	if err != nil {
		panic(err)
	}
	firstLoad := true
	if err := GetRaw("idx_seq", &[]byte{}); err == nil {
		firstLoad = false
	}

	if firstLoad {
		idx, err := Seq() // initialize sequence
		if err != nil {
			panic(err)
		}

		if idx < 10 {
			for {
				idx, err := Seq()
				if err != nil {
					panic(err)
				}
				if idx > 9 {
					break
				}
			}
		}
	}
}
func GetBadgerDb[T any](model T) *BadgerHolder[T] {

	if db == nil || db.IsClosed() {
		connect()
	}
	badgerHolder := &BadgerHolder[T]{db: db, model: model}

	return badgerHolder
}

func (b *BadgerHolder[T]) ListValues(seek string, last string, limit int, desc bool) ([]T, string, error) {

	var result = make([]T, 0)
	var lastKey string

	err := b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Reverse = desc
		it := txn.NewIterator(opts)
		defer it.Close()
		seekbyte := []byte(seek)
		if last != "" {
			seekbyte = []byte(last)
		}

		if opts.Reverse {
			seekbyte = append(seekbyte, 0xFF)
		}
		for it.Seek(seekbyte); len(result) < limit && it.Valid(); it.Next() {
			key := it.Item().Key()
			if !bytes.HasPrefix(key, []byte(seek)) {
				continue
			}
			if bytes.Equal(key, []byte(last)) {
				continue
			}
			item, err := txn.Get(key)
			if err != nil {
				continue
			}

			_ = item.Value(func(v []byte) error {
				var val T
				sonic.Unmarshal(v, &val)
				result = append(result, val)
				return nil
			})
			lastKey = string(key)

		}

		return nil
	})

	return result, lastKey, err
}

func Seq() (uint64, error) {
	if db == nil || db.IsClosed() {
		connect()
	}
	seq, err := db.GetSequence([]byte("idx_seq"), 100)
	if err != nil {
		return 0, err
	}
	defer seq.Release()
	return seq.Next()

}
func Upsert(key string, value any) error {
	if db == nil || db.IsClosed() {
		connect()
	}
	return db.Update(func(txn *badger.Txn) error {
		pk := []byte(key)

		valueBytes, err := sonic.Marshal(value)
		if err != nil {
			return err
		}
		// store new value
		if err := txn.Set(pk, valueBytes); err != nil {
			return err
		}
		// fmt.Printf("key: %s, value: %s\n", key, string(valueBytes))
		return nil
	})
}
func UpsertEntry(value *badger.Entry) error {
	if db == nil || db.IsClosed() {
		connect()
	}
	return db.Update(func(txn *badger.Txn) error {

		// store new value
		if err := txn.SetEntry(value); err != nil {
			return err
		}
		return nil
	})
}
func UpsertWithKeys(keys []string, value any) error {
	if db == nil || db.IsClosed() {
		connect()
	}
	return db.Update(func(txn *badger.Txn) error {

		valueBytes, err := sonic.Marshal(value)
		if err != nil {
			return err
		}
		for _, key := range keys {
			pk := []byte(key)
			// store new value
			if err := txn.Set(pk, valueBytes); err != nil {
				return err
			}
			// fmt.Printf("key: %s, Added\n", key)

		}
		return nil
	})
}
func (b *BadgerHolder[T]) Get(key string) (*T, error) {
	var out T
	err := b.db.View(func(txn *badger.Txn) error {
		pk := []byte(key)
		item, err := txn.Get(pk)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			// decode stored wrapper
			if err := sonic.Unmarshal(val, &out); err != nil {
				return err
			}

			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}
func GetRaw(key string, out *[]byte) error {
	if db == nil || db.IsClosed() {
		connect()
	}
	return db.View(func(txn *badger.Txn) error {
		// pk := makePrimaryKey(entity, id)

		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			*out = append(*out, val...)
			// decode stored wrapper

			return nil
		})
	})
}
func SetRaw(
	key string,
	data []byte,
) error {
	if db == nil || db.IsClosed() {
		connect()
	}
	return db.Update(func(txn *badger.Txn) error {

		// save primary record
		if err := txn.Set([]byte(key), data); err != nil {
			return err
		}

		// fmt.Printf("key: %s, value: %s\n", key, string(data))

		return nil
	})
}
func GetAllKeys() ([]string, error) {
	if db == nil || db.IsClosed() {
		connect()
	}
	var keys []string

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		// faster if you only need keys
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			key := item.KeyCopy(nil)

			keys = append(keys, string(key))
		}

		return nil
	})

	return keys, err
}

func Delete(keys ...string) error {
	if db == nil || db.IsClosed() {
		connect()
	}
	return db.Update(func(txn *badger.Txn) error {
		for _, key := range keys {
			pk := []byte(key)

			if err := txn.Delete(pk); err != nil {
				return err
			}

		}
		return nil
	})
}
func InsertWithSeq(
	key string,
	data any,
) error {

	return db.Update(func(txn *badger.Txn) error {
		// // generate sequence
		seq, err := Seq()
		if err != nil {
			return err
		}

		pk := append([]byte(key), []byte(fmt.Sprintf(":%d", seq))...)

		databytes, err := sonic.Marshal(data)
		if err != nil {
			return err
		}
		// save primary record
		if err := txn.Set(pk, databytes); err != nil {
			return err
		}

		// fmt.Println(string(pk))

		return nil
	})
}

// Best option is using DropWithPrefix
func DeleteByPrefix(prefix string) error {
	if db == nil || db.IsClosed() {
		connect()
	}
	return db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		p := []byte(prefix)

		for it.Seek(p); it.ValidForPrefix(p); it.Next() {
			key := it.Item().KeyCopy(nil)

			if err := txn.Delete(key); err != nil {
				return err
			}
		}

		return nil
	})
}

func DropWithPrefix(prefix string) error {
	if db == nil || db.IsClosed() {
		connect()
	}
	return db.DropPrefix([]byte(prefix))
}
