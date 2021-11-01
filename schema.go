package memdb

import "fmt"

// DBSchema is the schema to use for the full database with a MemDB instance.
//
// MemDB will require a valid schema. Schema validation can be tested using
// the Validate function. Calling this function is recommended in unit tests.
//
// DBSchema 包含数据库中所有的表模式。
// MemDB 需要一个有效的 DBSchema ，可以使用 Validate 函数来验证。
type DBSchema struct {
	// Tables is the set of tables within this database.
	// The key is the table name and must match the Name in TableSchema.
	//
	// Tables 是此数据库中的一组表，key 是表名，必须与 TableSchema 中的名称匹配。
	Tables map[string]*TableSchema
}

// Validate validates the schema.
func (s *DBSchema) Validate() error {
	if s == nil {
		return fmt.Errorf("schema is nil")
	}

	if len(s.Tables) == 0 {
		return fmt.Errorf("schema has no tables defined")
	}

	for name, table := range s.Tables {
		if name != table.Name {
			return fmt.Errorf("table name mis-match for '%s'", name)
		}

		if err := table.Validate(); err != nil {
			return fmt.Errorf("table %q: %s", name, err)
		}
	}

	return nil
}

// TableSchema is the schema for a single table.
type TableSchema struct {
	// Name of the table. This must match the key in the Tables map in DBSchema.
	Name string

	// Indexes is the set of indexes for querying this table.
	// The key is a unique name for the index and must match the Name in the IndexSchema.
	//
	// Indexes 是表的索引集合。
	// key 是索引的唯一名称，必须与 IndexSchema 中的名称匹配。
	Indexes map[string]*IndexSchema
}

// Validate is used to validate the table schema
func (s *TableSchema) Validate() error {

	// 表名非空
	if s.Name == "" {
		return fmt.Errorf("missing table name")
	}

	// 索引非空
	if len(s.Indexes) == 0 {
		return fmt.Errorf("missing table indexes for '%s'", s.Name)
	}

	// 至少要包含 ID 索引
	if _, ok := s.Indexes["id"]; !ok {
		return fmt.Errorf("must have id index")
	}

	// ID 索引必须是唯一索引
	if !s.Indexes["id"].Unique {
		return fmt.Errorf("id index must be unique")
	}

	// ID 索引必须是单值索引
	if _, ok := s.Indexes["id"].Indexer.(SingleIndexer); !ok {
		return fmt.Errorf("id index must be a SingleIndexer")
	}

	// 校验各个索引合法性
	for name, index := range s.Indexes {
		if name != index.Name {
			return fmt.Errorf("index name mis-match for '%s'", name)
		}
		if err := index.Validate(); err != nil {
			return fmt.Errorf("index %q: %s", name, err)
		}
	}

	return nil
}

// IndexSchema is the schema for an index. An index defines how a table is queried.
//
// IndexSchema 是索引的模式，定义如何查询表。
type IndexSchema struct {
	// Name of the index.
	// This must be unique among a tables set of indexes.
	// This must match the key in the map of Indexes for a TableSchema.
	//
	// 索引名
	Name string

	// AllowMissing if true ignores this index if it doesn't produce a value.
	// For example, an index that extracts a field that doesn't exist from a structure.
	//
	// 是否允许空值
	AllowMissing bool

	// 唯一索引
	Unique  bool

	// 索引对象
	Indexer Indexer
}

func (s *IndexSchema) Validate() error {
	// 索引名非空
	if s.Name == "" {
		return fmt.Errorf("missing index name")
	}
	// 索引非空
	if s.Indexer == nil {
		return fmt.Errorf("missing index function for '%s'", s.Name)
	}
	// 索引类型
	switch s.Indexer.(type) {
	case SingleIndexer:	// 单值索引
	case MultiIndexer:	// 多值索引
	default:
		return fmt.Errorf("indexer for '%s' must be a SingleIndexer or MultiIndexer", s.Name)
	}
	return nil
}
