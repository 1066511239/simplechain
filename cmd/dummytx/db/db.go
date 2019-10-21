package db

import (
	"git.dev.tencent.com/baoquan2017/ccc-resonance/log"
	"github.com/simplechain-org/simplechain/common"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type IdHash struct {
	Id   common.Hash `json:"id"`
	Hash common.Hash `json:"hash"`
}

type LDBDatabase struct {
	fn string      // filename for reporting
	db *leveldb.DB // LevelDB instance

	quitLock sync.Mutex      // Mutex protecting the quit channel access
	quitChan chan chan error // Quit channel to stop the metrics collection before closing the database

	log log.Logger // Contextual logger tracking the database path
}

// NewLDBDatabase returns a LevelDB wrapped object.
func NewLDBDatabase(file string, cache int, handles int) (*LDBDatabase, error) {
	logger := log.New("database")
	logger.SetHandler(log.DiscardHandler())

	// Ensure we have some minimal caching and file guarantees
	if cache < 16 {
		cache = 16
	}
	if handles < 16 {
		handles = 16
	}
	logger.Info("Allocated cache and file handles cache %d, handles %d", cache, handles)

	// Open the db and recover any potential corruptions
	ldb, err := leveldb.OpenFile(file, &opt.Options{
		OpenFilesCacheCapacity: handles,
		BlockCacheCapacity:     cache / 2 * opt.MiB,
		WriteBuffer:            cache / 4 * opt.MiB, // Two of these are used internally
		Filter:                 filter.NewBloomFilter(10),
	})
	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		ldb, err = leveldb.RecoverFile(file, nil)
	}
	// (Re)check for errors and abort if opening of the db failed
	if err != nil {
		return nil, err
	}
	return &LDBDatabase{
		fn:  file,
		db:  ldb,
		log: logger,
	}, nil
}

// Path returns the path to the database directory.
func (ldb *LDBDatabase) Path() string {
	return ldb.fn
}

// Put puts the given key / value to the queue
func (ldb *LDBDatabase) Put(key []byte, value []byte) error {
	return ldb.db.Put(key, value, nil)
}

func (ldb *LDBDatabase) Has(key []byte) (bool, error) {
	return ldb.db.Has(key, nil)
}

// Get returns the given key if it's present.
func (ldb *LDBDatabase) Get(key []byte) ([]byte, error) {
	dat, err := ldb.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return dat, nil
}

// Delete deletes the key from the queue and database
func (ldb *LDBDatabase) Delete(key []byte) error {
	return ldb.db.Delete(key, nil)
}

func (ldb *LDBDatabase) NewIterator() iterator.Iterator {
	return ldb.db.NewIterator(nil, nil)
}

// NewIteratorWithPrefix returns a iterator to iterate over subset of database content with a particular prefix.
func (ldb *LDBDatabase) NewIteratorWithPrefix(prefix []byte) iterator.Iterator {
	return ldb.db.NewIterator(util.BytesPrefix(prefix), nil)
}

func (ldb *LDBDatabase) Close() {
	// Stop the metrics collection to avoid internal database races
	ldb.quitLock.Lock()
	defer ldb.quitLock.Unlock()

	if ldb.quitChan != nil {
		errc := make(chan error)
		ldb.quitChan <- errc
		if err := <-errc; err != nil {
			ldb.log.Error("Metrics collection failed", "err", err)
		}
		ldb.quitChan = nil
	}
	err := ldb.db.Close()
	if err == nil {
		ldb.log.Info("Database closed")
	} else {
		ldb.log.Error("Failed to close database", "err", err)
	}
}

func (ldb *LDBDatabase) LDB() *leveldb.DB {
	return ldb.db
}

func (ldb *LDBDatabase) InsertHash(hash, id common.Hash) error {
	return ldb.Put(hash.Bytes(), id.Bytes())
}

func (ldb *LDBDatabase) GetHashId(hash common.Hash) (common.Hash, error) {
	data, err := ldb.Get(hash.Bytes())
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(data), nil
}

//func (ldb *LDBDatabase) GetAllRawTrades() (trades protocol.BtcTrades, err error) {
//	itr := ldb.NewIterator()
//	itr.First()
//
//	for len(itr.Value()) != 0 {
//		var trade protocol.BtcTrade
//		if err := rlp.DecodeBytes(itr.Value(), &trade); err != nil {
//			return nil, err
//		}
//		trades = append(trades, trade)
//		itr.Next()
//	}
//
//	sort.Sort(trades)
//	return trades, nil
//}
