package packet

import (
	"github.com/mask-pp/rtp-ps/buffer"
)

/*
https://github.com/videolan/vlc/blob/master/modules/demux/mpeg
*/
type DecPSPackage struct {
	systemClockReferenceBase      uint64
	systemClockReferenceExtension uint64
	programMuxRate                uint32

	StreamType uint32

	*buffer.RawBuffer
	payloads [][]byte
}

func (dec *DecPSPackage) clean() {
	dec.systemClockReferenceBase = 0
	dec.systemClockReferenceExtension = 0
	dec.programMuxRate = 0

	dec.payloads = dec.payloads[:0]
}

func (dec *DecPSPackage) decPackHeader(data []byte) ([][]byte, error) {
	dec.clean()

	// 加载数据
	dec.LoadBuffer(data)

	if startcode, err := dec.Uint32(); err != nil {
		return nil, err
	} else if startcode != StartCodePS {
		return nil, ErrNotFoundStartCode
	}

	if err := dec.Skip(4 + 9); err != nil {
		return nil, err
	}

	psl, err := dec.Uint8()
	if err != nil {
		return nil, err
	}
	psl &= 0x07
	dec.Skip(int(psl))

	for {
		nextStartCode, err := dec.Uint32()
		if err != nil {
			return nil, err
		}

		switch nextStartCode {
		case StartCodeSYS:
			if err := dec.decSystemHeader(); err != nil {
				return nil, err
			}
		case StartCodeMAP:
			if err := dec.decProgramStreamMap(); err != nil {
				return nil, err
			}
		case StartCodeVideo:
			fallthrough
		case StartCodeAudio:
			if err := dec.decPESPacket(); err != nil {
				return nil, err
			}
		case HaiKangCode, MEPGProgramEndCode:
			return dec.payloads, nil
		}
	}
}

func (dec *DecPSPackage) decSystemHeader() error {
	syslens, err := dec.Uint16()
	if err != nil {
		return err
	}
	// drop rate video audio bound and lock flag
	syslens -= 6
	if err = dec.Skip(6); err != nil {
		return err
	}

	// ONE WAY: do not to parse the stream  and skip the buffer
	//br.Skip(syslen * 8)

	// TWO WAY: parse every stream info
	for syslens > 0 {
		if nextbits, err := dec.Uint8(); err != nil {
			return err
		} else if (nextbits&0x80)>>7 != 1 {
			break
		}
		if err = dec.Skip(2); err != nil {
			return err
		}
		syslens -= 3
	}
	return nil
}

func (dec *DecPSPackage) decProgramStreamMap() error {
	psm, err := dec.Uint16()
	if err != nil {
		return err
	}
	//drop psm version infor
	if err = dec.Skip(2); err != nil {
		return err
	}
	psm -= 2

	programStreamInfoLen, err := dec.Uint16()
	if err != nil {
		return err
	}
	if err = dec.Skip(int(programStreamInfoLen)); err != nil {
		return err
	}
	psm -= programStreamInfoLen + 2

	programStreamMapLen, err := dec.Uint16()
	if err != nil {
		return err
	}
	psm -= 2 + programStreamMapLen

	for programStreamMapLen > 0 {
		streamType, err := dec.Uint8()
		if err != nil {
			return err
		}

		elementaryStreamID, err := dec.Uint8()
		if err != nil {
			return err
		}

		if elementaryStreamID >= 0xe0 && elementaryStreamID <= 0xef {
			dec.StreamType = uint32(streamType)
		} else if elementaryStreamID >= 0xc0 && elementaryStreamID <= 0xdf {
			dec.StreamType = uint32(streamType)
		}

		elementaryStreamInfoLength, err := dec.Uint16()
		if err != nil {
			return err
		}
		if err = dec.Skip(int(elementaryStreamInfoLength)); err != nil {
			return err
		}
		programStreamMapLen -= 4 + elementaryStreamInfoLength
	}

	// crc 32
	if psm != 4 {
		return ErrFormatPack
	}
	if err = dec.Skip(4); err != nil {
		return err
	}
	return nil
}

func (dec *DecPSPackage) decPESPacket() error {
	payloadlen, err := dec.Uint16()
	if err != nil {
		return err
	}
	if err = dec.Skip(2); err != nil {
		return err
	}

	payloadlen -= 2
	pesHeaderDataLen, err := dec.Uint8()
	if err != nil {
		return err
	}
	payloadlen -= uint16(pesHeaderDataLen) + 1

	if err = dec.Skip(int(pesHeaderDataLen)); err != nil {
		return err
	}

	if payload, err := dec.Bytes(int(payloadlen)); err != nil {
		return err
	} else {
		dec.payloads = append(dec.payloads, payload[4:])
	}

	return nil
}
