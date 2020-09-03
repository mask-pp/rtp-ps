package buffer

import (
	"encoding/binary"
	"io"
)

type RawBuffer struct {
	raw    []byte
	buffer int
}

func (r *RawBuffer) LoadBuffer(data []byte) {
	r.raw = data
	r.buffer = len(r.raw)
}

func (r *RawBuffer) Skip(length int) error {
	if err := r.checkLen(length); err != nil {
		return err
	}
	r.skip(length)
	return nil
}

func (r *RawBuffer) Uint8() (uint8, error) {
	if err := r.checkLen(1); err != nil {
		return 0, err
	}
	res := r.raw[0]
	r.skip(1)

	return res, nil
}

func (r *RawBuffer) Uint16() (uint16, error) {
	if err := r.checkLen(2); err != nil {
		return 0, err
	}
	res := binary.BigEndian.Uint16(r.raw[:2])
	r.skip(2)

	return res, nil
}

func (r *RawBuffer) Uint32() (uint32, error) {
	if err := r.checkLen(4); err != nil {
		return 0, err
	}
	res := binary.BigEndian.Uint32(r.raw[:4])
	r.skip(4)

	return res, nil
}

func (r *RawBuffer) Bytes(length int) ([]byte, error) {
	if err := r.checkLen(length); err != nil {
		return nil, err
	}
	res := r.raw[:length]
	r.skip(length)

	return res, nil
}

func (r *RawBuffer) checkLen(length int) error {
	if length > r.buffer {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (r *RawBuffer) skip(length int) {
	r.buffer -= length
	r.raw = r.raw[length:]
}
