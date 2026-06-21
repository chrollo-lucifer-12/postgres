package core

const InvalidPageID = -1

type Frame struct {
	PageID int32
	Dirty  bool
	PinCnt uint32

	Data [PageSize]byte
}

type BufferPool struct {
	Frames []Frame
}
