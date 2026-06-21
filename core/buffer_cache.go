package core

import (
	"errors"
	"syscall"
	"unsafe"
)

const InvalidPageID = -1

type Frame struct {
	PageID int32
	Dirty  uint8
	PinCnt uint32

	Page Page
}

type SharedState struct {
	PageTable [10000]int
	Frames    [1024]Frame
}

type BufferPool struct {
	mem    []byte
	shared *SharedState

	disk *DiskManager
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
