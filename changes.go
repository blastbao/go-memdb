package memdb

// Changes describes a set of mutations to memDB tables performed during a
// transaction.
//
// Changes 描述了在一个事务期间对 memDB 表执行的一组变更。
type Changes []Change

// Change describes a mutation to an object in a table.
//
// Change 描述对表中对象的更改。
type Change struct {
	Table  string			// 表
	Before interface{}		// 修改前的值
	After  interface{}		// 修改后的值

	// primaryKey stores the raw key value from the primary index so that we can
	// de-duplicate multiple updates of the same object in the same transaction
	// but we don't expose this implementation detail to the consumer.
	//
	// primaryKey 存储主键索引中的原始键值，以便我们可以在同一事务中对同一对象进行多次更新，
	// 但不向使用者公开这个实现细节。
	primaryKey []byte
}

// Created returns true if the mutation describes a new object being inserted.
//
// 如果 mutation 描述插入的新对象，则 Created 返回 true 。
func (m *Change) Created() bool {
	return m.Before == nil && m.After != nil
}

// Updated returns true if the mutation describes an existing object being
// updated.
//
// 如果 mutation 描述了正在更新的现有对象，则 Updated 返回 true 。
func (m *Change) Updated() bool {
	return m.Before != nil && m.After != nil
}

// Deleted returns true if the mutation describes an existing object being
// deleted.
//
// 如果 mutation 描述正在删除的现有对象，则 Deleted 返回 true 。
func (m *Change) Deleted() bool {
	return m.Before != nil && m.After == nil
}
