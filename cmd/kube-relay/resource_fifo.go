package main

import (
	"container/list"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/anhk/kube-relay/pkg/cond"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const MAX_RESOURCE_FIFO_LEN = 0x10000

type Item struct {
	key   string
	ele   *list.Element
	event *metav1.WatchEvent
}

// FIFO for Resource
type ResourceFifo struct {
	version int64

	mu    sync.RWMutex     // 读写锁
	items map[string]*Item // key: version.toString()
	list  list.List        // LRU列表

	cond *cond.Cond
}

func NewResourceFifo() *ResourceFifo {
	rf := &ResourceFifo{version: 1, items: make(map[string]*Item)}
	rf.cond = cond.NewCond(&rf.mu)
	return rf
}

func (fifo *ResourceFifo) Push(event *metav1.WatchEvent) {
	fifo.mu.Lock()
	defer fifo.mu.Unlock()

	it := &Item{event: event, key: fmt.Sprintf("%d", fifo.version)}
	it.ele = fifo.list.PushBack(it)
	fifo.items[it.key] = it
	fifo.version++

	for fifo.list.Len() > MAX_RESOURCE_FIFO_LEN {
		fifo.removeOldest()
	}

	fifo.cond.Broadcast()
}

func (fifo *ResourceFifo) Get(resourceVersion string) ([]*metav1.WatchEvent, string, error) {
	fifo.mu.RLock()
	defer fifo.mu.RUnlock()
	curVersion := fmt.Sprintf("%d", fifo.version)

	resVerion, err := strconv.ParseInt(resourceVersion, 10, 64)
	if err != nil {
		return nil, curVersion, err
	}

	it, ok := fifo.items[resourceVersion]
	if !ok && resVerion < fifo.version { // 不存在则返回`410 Gone`
		return nil, curVersion, fmt.Errorf("401 Gone")
	} else if !ok {
		return nil, curVersion, nil
	}

	var result []*metav1.WatchEvent
	for ele := it.ele; ele != nil; ele = ele.Next() {
		it := ele.Value.(*Item)
		result = append(result, it.event)
	}
	return result, curVersion, nil
}

// 删除最久未使用的一个节点，无锁
func (fifo *ResourceFifo) removeOldest() {
	keyEle := fifo.list.Front()
	if keyEle == nil {
		return
	}
	key := fifo.list.Remove(keyEle)
	delete(fifo.items, key.(string))
}

// 等待，直到有消息进来
func (fifo *ResourceFifo) Wait(resourceVersion string) error {
	resVerion, err := strconv.ParseInt(resourceVersion, 10, 64)
	if err != nil {
		return err
	}

	fifo.cond.L.Lock()
	if resVerion == fifo.version {
		fifo.cond.WaitWithTimeout(time.Second)
	}
	fifo.cond.L.Unlock()
	return nil
}

func (fifo *ResourceFifo) Version() string {
	return fmt.Sprintf("%d", fifo.version)
}
