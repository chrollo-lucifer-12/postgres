package core

import (
	"encoding/binary"
	"errors"
)

const (
	PageSize   = 8192
	HeaderSize = 8
	SlotSize   = 4
)

type PageHeader struct {
	Lower uint16
	Upper uint16
	Count uint16
}

type Slot struct {
	Offset uint16
	Length uint16
}

type Tuple struct {
	Data []byte
}

type Page struct {
	Data [PageSize]byte
}

func NewPage() *Page {
	p := &Page{}

	p.SetLower(HeaderSize)
	p.SetUpper(PageSize)
	p.Count()

	return p
}

func (p *Page) Lower() uint16 {
	return binary.LittleEndian.Uint16(p.Data[0:2])
}

func (p *Page) Upper() uint16 {
	return binary.BigEndian.Uint16(p.Data[2:4])
}

func (p *Page) Count() uint16 {
	return binary.BigEndian.Uint16(p.Data[4:6])
}

func (p *Page) SetLower(v uint16) {
	binary.LittleEndian.PutUint16(p.Data[0:2], v)
}

func (p *Page) SetUpper(v uint16) {
	binary.LittleEndian.PutUint16(p.Data[2:4], v)
}

func (p *Page) SetCount(v uint16) {
	binary.LittleEndian.PutUint16(p.Data[4:6], v)
}

func (p *Page) FreeSpace() uint16 {
	return p.Upper() - p.Lower()
}

func (p *Page) Slot(slotID int) Slot {
	pos := HeaderSize + slotID*SlotSize

	offset := binary.LittleEndian.Uint16(p.Data[pos : pos+2])

	length := binary.LittleEndian.Uint16(p.Data[pos+2 : pos+4])

	return Slot{
		Offset: offset,
		Length: length,
	}
}

func (p *Page) Insert(data []byte) (int, error) {
	needed := SlotSize + len(data)

	if int(p.FreeSpace()) < needed {
		return -1, errors.New("page full")
	}

	upper := p.Upper()
	lower := p.Lower()

	upper -= uint16(len(data))

	copy(p.Data[upper:], data)

	binary.LittleEndian.PutUint16(p.Data[lower:lower+2], upper)
	binary.LittleEndian.PutUint16(p.Data[lower+2:lower+4], uint16(len(data)))

	p.SetUpper(upper)
	p.SetLower(lower)
	p.SetCount(p.Count() + 1)

	return int(p.Count() - 1), nil
}

func (p *Page) Get(slotID int) ([]byte, error) {
	if slotID > int(p.Count()) {
		return nil, errors.New("invalid slot")
	}

	slot := p.Slot(slotID)

	if slot.Offset == 0 {
		return nil, errors.New("deleted")
	}

	return p.Data[slot.Offset : slot.Offset+slot.Length], nil
}

func (p *Page) Delete(slotID int) error {
	if slotID >= int(p.Count()) {
		return errors.New("invalid slot")
	}

	pos := HeaderSize + slotID*SlotSize

	binary.LittleEndian.PutUint16(p.Data[pos:pos+2], 0)
	binary.LittleEndian.PutUint16(p.Data[pos+2:pos+4], 0)

	return nil
}

func (p *Page) IsDeleted(slotID int) bool {
	slot := p.Slot(slotID)
	return slot.Offset == 0
}
