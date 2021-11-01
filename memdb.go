// Package memdb provides an in-memory database that supports transactions
// and MVCC.
package memdb

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/hashicorp/go-immutable-radix"
)

// MemDB is an in-memory database providing Atomicity, Consistency, and
// Isolation from ACID. MemDB doesn't provide Durability since it is an
// in-memory database.
//
// MemDB provides a table abstraction to store objects (rows) with multiple
// indexes based on inserted values. The database makes use of immutable radix
// trees to provide transactions and MVCC.
//
// Objects inserted into MemDB are not copied. It is **extremely important**
// that objects are not modified in-place after they are inserted since they
// are stored directly in MemDB. It remains unsafe to modify inserted objects
// even after they've been deleted from MemDB since there may still be older
// snapshots of the DB being read from other goroutines.
//
//
// MemDB 是一个内存数据库，提供 ACID 中的原子性、一致性和与隔离性；
// MemDB 不提供持久性，因为它是内存中的数据库。
//
// MemDB 提供了一个表抽象，用于存储具有多个索引的对象（行）。
// 数据库使用不可变基树来提供事务和 MVCC 。
//
// 插入到 MemDB 的对象不会被复制。
// 由于对象直接存储在 MemDB 中，因此在插入对象后不进行原地修改 **非常重要** 。
// 即使是已从 MemDB 中删除的对象，修改这些对象仍然是不安全的，因为可能有旧的数据库快照正在被其他 goroutine 读取。

type MemDB struct {
	schema  *DBSchema
	root    unsafe.Pointer // *iradix.Tree underneath
	primary bool

	// There can only be a single writer at once
	writer sync.Mutex
}

// NewMemDB creates a new MemDB with the given schema.
func NewMemDB(schema *DBSchema) (*MemDB, error) {
	// Validate the schema
	if err := schema.Validate(); err != nil {
		return nil, err
	}

	// Create the MemDB
	db := &MemDB{
		schema:  schema,
		root:    unsafe.Pointer(iradix.New()),
		primary: true,
	}

	// Init MemDB
	if err := db.initialize(); err != nil {
		return nil, err
	}

	return db, nil
}

// getRoot is used to do an atomic load of the root pointer
func (db *MemDB) getRoot() *iradix.Tree {
	root := (*iradix.Tree)(atomic.LoadPointer(&db.root))
	return root
}

// Txn is used to start a new transaction in either read or write mode.
// There can only be a single concurrent writer, but any number of readers.
func (db *MemDB) Txn(write bool) *Txn {
	// 写事务加锁
	if write {
		db.writer.Lock()
	}

	txn := &Txn{
		db:      db,
		write:   write,
		rootTxn: db.getRoot().Txn(),
	}
	return txn
}

// Snapshot is used to capture a point-in-time snapshot  of the database that
// will not be affected by any write operations to the existing DB.
//
// If MemDB is storing reference-based values (pointers, maps, slices, etc.),
// the Snapshot will not deep copy those values. Therefore, it is still unsafe
// to modify any inserted values in either DB.
func (db *MemDB) Snapshot() *MemDB {
	clone := &MemDB{
		schema:  db.schema,
		root:    unsafe.Pointer(db.getRoot()),
		primary: false,
	}
	return clone
}

// initialize is used to setup the DB for use after creation. This should
// be called only once after allocating a MemDB.
func (db *MemDB) initialize() error {
	root := db.getRoot()
	for tName, tableSchema := range db.schema.Tables {
		for iName := range tableSchema.Indexes {
			index := iradix.New()
			path := indexPath(tName, iName)
			root, _, _ = root.Insert(path, index)
		}
	}
	db.root = unsafe.Pointer(root)
	return nil
}

// indexPath returns the path from the root to the given table index
func indexPath(table, index string) []byte {
	return []byte(table + "." + index)
}
