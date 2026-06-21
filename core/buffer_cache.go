package core

import "errors"

const InvalidPageID = -1

type Frame struct {
	PageID int32
	Dirty  bool
	PinCnt uint32

	Page Page
}

type BufferPool struct {
	Frames    []Frame
	PageTable map[int]int

	disk *DiskManager
}

func NewBufferPool(
	size int,
	disk *DiskManager,
) *BufferPool {

	frames := make([]Frame, size)

	for i := range frames {
		frames[i].PageID = InvalidPageID
	}

	return &BufferPool{
		Frames: frames,

		PageTable: make(map[int]int),

		disk: disk,
	}
}

func (bp *BufferPool) findFreeFrame() int {
	for i := range bp.Frames {
		if bp.Frames[i].PageID == InvalidPageID {
			return i
		}
	}

	return -1
}

func (bp *BufferPool) fetchPage(pageId int) (*Frame, error) {

	if frameID, ok := bp.PageTable[pageId]; ok {
		frame := &bp.Frames[frameID]
		frame.PinCnt++

		return frame, nil
	}

	frameID := bp.findFreeFrame()

	if frameID == -1 {
		return nil, errors.New("buffer pool full")
	}

	frame := &bp.Frames[frameID]

	err := bp.disk.ReadPage(pageId, frame.Page.Data[:])

	if err != nil {
		return nil, err
	}

	frame.PageID = int32(pageId)
	frame.PinCnt = 1
	frame.Dirty = false

	bp.PageTable[pageId] = frameID

	return frame, nil
}

func (b *BufferPool) UnpinPage(
	pageID int,
	dirty bool,
) error {

	frameID, ok := b.PageTable[pageID]

	if !ok {
		return errors.New("page not found")
	}

	frame := &b.Frames[frameID]

	if frame.PinCnt > 0 {
		frame.PinCnt--
	}

	if dirty {
		frame.Dirty = true
	}

	return nil
}

func (bp *BufferPool) FlushPage(pageId int) error {

	frameID, ok := bp.PageTable[pageId]

	if !ok {
		return errors.New("page not found")
	}

	frame := &bp.Frames[frameID]

	if !frame.Dirty {
		return nil
	}

	err := bp.disk.WritePage(pageId, frame.Page.Data[:])

	if err != nil {
		return err
	}

	frame.Dirty = false

	return nil
}
