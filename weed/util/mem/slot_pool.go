package mem

import (
	"github.com/seaweedfs/seaweedfs/weed/glog"
	"sync"
	"sync/atomic"
)

var pools []*sync.Pool

const (
	min_size = 1024
)

func bitCount(size int) (count int) {
	for ; size > min_size; count++ {
		size = (size + 1) >> 1
	}
	return
}

func init() {
	// 1KB ~ 256MB
	pools = make([]*sync.Pool, bitCount(1024*1024*256))
	for i := 0; i < len(pools); i++ {
		slotSize := 1024 << i
		pools[i] = &sync.Pool{
			New: func() interface{} {
				buffer := make([]byte, slotSize)
				return &buffer
			},
		}
	}
}

func getSlotPool(size int) (*sync.Pool, bool) {
	index := bitCount(size)
	if index >= len(pools) {
		return nil, false
	}
	return pools[index], true
}

var total int64

func Allocate(size int) []byte {
	if pool, found := getSlotPool(size); found {
		newVal := atomic.AddInt64(&total, 1)
		glog.V(4).Infof("++> %d", newVal)

		slab := *pool.Get().(*[]byte)
		return slab[:size]
	}
	return make([]byte, size)
}

func Free(buf []byte) {
	if pool, found := getSlotPool(cap(buf)); found {
		newVal := atomic.AddInt64(&total, -1)
		glog.V(4).Infof("--> %d", newVal)
		pool.Put(&buf)
	}
}
