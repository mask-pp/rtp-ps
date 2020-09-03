package packet

import "github.com/mask-pp/rtp-ps/buffer"

type RtpParsePacket struct {
	*DecPSPackage
}

func NewRtpParsePacket() *RtpParsePacket {
	return &RtpParsePacket{
		DecPSPackage: &DecPSPackage{
			payloads:  make([][]byte, 0, 8),
			RawBuffer: &buffer.RawBuffer{},
		},
	}
}

// data包含 接受到完整一帧数据后，所有的payload, 解析出去后是一阵完整的raw数据
func (rtp *RtpParsePacket) Read(data []byte) ([][]byte, error) {

	// add the MPEG Program end code
	data = append(data, 0x00, 0x00, 0x01, 0xb9)

	if rtp.DecPSPackage != nil {
		return rtp.decPackHeader(data)
	}

	return nil, nil
}
