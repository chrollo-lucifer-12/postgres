package core

import (
	"errors"
	"syscall"
	"unsafe"
)

const InvalidPageID = -1

const (
	MaxKeyLen     = 64
	IndexCapacity = 100003
)

type Frame struct {
	PageID int32
	Dirty  uint8
	PinCnt uint32

	Page Page
}

type IndexEntry struct {
	Key    [MaxKeyLen]byte
	KeyLen uint8
	Used   uint8
	PageID int32
	SlotID int32
}

type SharedState struct {
	PageTable [10000]int
	Frames    [2]Frame

	NextPageID int32

	Index [IndexCapacity]IndexEntry
}

type BufferPool struct {
	mem    []byte
	shared *SharedState

	disk *DiskManager
}

func hashKey(key string) uint64 {
	var h uint64 = 14695981039346656037
	for i := 0; i < len(key); i++ {
		h ^= uint64(key[i])
		h *= 1099511628211
	}
	return h
}

func (bp *BufferPool) IndexGet(key string) (RID, bool) {
	idx := hashKey(key) % IndexCapacity
	for i := uint64(0); i < IndexCapacity; i++ {
		pos := (idx + i) % IndexCapacity
		e := &bp.shared.Index[pos]
		if e.Used == 0 {
			return RID{}, false
		}
		if e.Used == 1 && string(e.Key[:e.KeyLen]) == key {
			return RID{PageID: int(e.PageID), SlotID: int(e.SlotID)}, true
		}
	}
	return RID{}, false
}

func (bp *BufferPool) IndexSet(key string, rid RID) {
	idx := hashKey(key) % IndexCapacity
	for i := uint64(0); i < IndexCapacity; i++ {
		pos := (idx + i) % IndexCapacity
		e := &bp.shared.Index[pos]
		if e.Used != 1 || string(e.Key[:e.KeyLen]) == key {
			copy(e.Key[:], key)
			e.KeyLen = uint8(len(key))
			e.Used = 1
			e.PageID = int32(rid.PageID)
			e.SlotID = int32(rid.SlotID)
			return
		}
	}
}

func (bp *BufferPool) IndexDelete(key string) {
	idx := hashKey(key) % IndexCapacity
	for i := uint64(0); i < IndexCapacity; i++ {
		pos := (idx + i) % IndexCapacity
		e := &bp.shared.Index[pos]
		if e.Used == 0 {
			return
		}
		if e.Used == 1 && string(e.Key[:e.KeyLen]) == key {
			e.Used = 2
			return
		}
	}
}

func NewBufferPool(
	frameCount int,
	disk *DiskManager,
) (*BufferPool, error) {

	sizeBytes := int(unsafe.Sizeof(SharedState{}))

	mem, err := syscall.Mmap(-1, 0, sizeBytes, syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED|syscall.MAP_ANON)

	if err != nil {
		return nil, err
	}

	shared := (*SharedState)(
		unsafe.Pointer(&mem[0]),
	)

	for i := range shared.PageTable {
		shared.PageTable[i] = InvalidPageID
	}

	for i := range shared.Frames {
		shared.Frames[i].PageID = InvalidPageID
	}

	return &BufferPool{
		mem:    mem,
		shared: shared,
		disk:   disk,
	}, nil
}

func (bp *BufferPool) NewPageID() int {
	id := int(bp.shared.NextPageID)
	bp.shared.NextPageID++
	return id
}

func (bp *BufferPool) Frame(i int) *Frame {
	return &bp.shared.Frames[i]
}

func (bp *BufferPool) findFreeFrame() int {
	for i := range bp.shared.Frames {
		if bp.shared.Frames[i].PageID == InvalidPageID {
			return i
		}
	}

	return -1
}

func (bp *BufferPool) FetchPage(pageId int) (*Frame, error) {

	frameID := bp.shared.PageTable[pageId]

	if frameID != -1 {
		frame := bp.Frame(frameID)

		frame.PinCnt++

		return frame, nil
	}

	frameID = bp.findFreeFrame()

	if frameID == -1 {
		return nil, errors.New("buffer pool full")
	}

	frame := &bp.shared.Frames[frameID]
	frame.Page = *NewPage()

	err := bp.disk.ReadPage(pageId, frame.Page.Data[:])

	if err != nil {
		return nil, err
	}

	frame.PageID = int32(pageId)
	frame.PinCnt = 1
	frame.Dirty = 0

	bp.shared.PageTable[pageId] = frameID

	return frame, nil
}

func (bp *BufferPool) UnpinPage(
	pageID int,
	dirty bool,
) error {

	frameID := bp.shared.PageTable[pageID]

	if frameID == -1 {
		return errors.New("page not found")
	}

	frame := &bp.shared.Frames[frameID]

	if frame.PinCnt > 0 {
		frame.PinCnt--
	}

	if dirty {
		frame.Dirty = 1
	}

	return nil
}

func (bp *BufferPool) FlushPage(pageId int) error {

	frameID := bp.shared.PageTable[pageId]

	if frameID == -1 {
		return errors.New("page not found")
	}

	frame := &bp.shared.Frames[frameID]

	if frame.Dirty == 0 {
		return nil
	}

	err := bp.disk.WritePage(pageId, frame.Page.Data[:])

	if err != nil {
		return err
	}

	frame.Dirty = 0

	return nil
}
