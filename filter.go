package memdb

// FilterFunc is a function that takes the results of an iterator and returns
// whether the result should be filtered out.
//
// FilterFunc 是一个函数，它获取迭代器的结果并返回是否应过滤。
type FilterFunc func(interface{}) bool

// FilterIterator is used to wrap a ResultIterator and apply a filter over it.
//
// FilterIterator 用于封装 ResultIterator 并在其上应用过滤器。
type FilterIterator struct {

	// filter is the filter function applied over the base iterator.
	// filter 是应用于基本迭代器的 filter 函数。
	filter FilterFunc

	// iter is the iterator that is being wrapped.
	// iter 是被封装的迭代器。
	iter ResultIterator
}

// NewFilterIterator wraps a ResultIterator.
// The filter function is applied to each value returned by a call to iter.Next.
//
// See the documentation for ResultIterator to understand the behaviour of the
// returned FilterIterator.
//
//
//
func NewFilterIterator(iter ResultIterator, filter FilterFunc) *FilterIterator {
	return &FilterIterator{
		filter: filter,
		iter:   iter,
	}
}

// WatchCh returns the watch channel of the wrapped iterator.
func (f *FilterIterator) WatchCh() <-chan struct{} {
	return f.iter.WatchCh()
}

// Next returns the next non-filtered result from the wrapped iterator.
func (f *FilterIterator) Next() interface{} {
	for {
		// 遍历迭代器，返回首个非空、未被过滤的 value
		if value := f.iter.Next(); value == nil || !f.filter(value) {
			return value
		}
	}
}
