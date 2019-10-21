package core

import (
	"container/heap"
	"sort"

	"github.com/simplechain-org/simplechain/core/types"
)

type timeHeap []int64

func (h timeHeap) Len() int           { return len(h) }
func (h timeHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h timeHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *timeHeap) Push(x interface{}) {
	*h = append(*h, x.(int64))
}

func (h *timeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type txQueue struct {
	items map[int64]*types.Transaction
	index *timeHeap
}

func newTxQueue() *txQueue {
	return &txQueue{
		items: make(map[int64]*types.Transaction),
		index: new(timeHeap),
	}
}

func (q *txQueue) Len() int {
	return q.index.Len()
}

func (q *txQueue) Put(tx *types.Transaction) {
	timestamp := tx.ImportTime()
	if q.items[timestamp] == nil {
		heap.Push(q.index, timestamp)
	}
}

func (q *txQueue) Remove(timestamp int64) bool {
	// Short circuit if no transaction is present
	_, ok := q.items[timestamp]
	if !ok {
		return false
	}
	// Otherwise delete the transaction and fix the heap index
	i := sort.Search(q.index.Len(), func(i int) bool {
		return (*q.index)[i] >= timestamp
	})

	if i == q.index.Len() || (*q.index)[i] != timestamp {
		return false
	}

	heap.Remove(q.index, i)
	delete(q.items, timestamp)
	return true
}

func (q *txQueue) Loop(f func(*types.Transaction) bool) {
	for _, timestamp := range *q.index {
		if f(q.items[timestamp]) {
			break
		}
	}
}
