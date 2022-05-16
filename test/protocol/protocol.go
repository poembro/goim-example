package protocol

import (
    "errors"
)

const (
    // MaxBodySize max proto body size
    MaxBodySize = int32(1 << 12)
)

const (
    // size
    _packSize      = 4
    _headerSize    = 2
    _verSize       = 2
    _opSize        = 4
    _seqSize       = 4
    _heartSize     = 4
    _rawHeaderSize = _packSize + _headerSize + _verSize + _opSize + _seqSize
    _maxPackSize   = MaxBodySize + int32(_rawHeaderSize)
    // offset
    _packOffset   = 0
    _headerOffset = _packOffset + _packSize
    _verOffset    = _headerOffset + _headerSize
    _opOffset     = _verOffset + _verSize
    _seqOffset    = _opOffset + _opSize
    _heartOffset  = _seqOffset + _seqSize
)

var (
    // ErrProtoPackLen proto packet len error
    ErrProtoPackLen = errors.New("default server codec pack length error")
    // ErrProtoHeaderLen proto header len error
    ErrProtoHeaderLen = errors.New("default server codec header length error")
)

type Proto struct {
    Ver  int32
    Op   int32
    Seq  int32
    Body []byte
}

func (p *Proto) Decode(buf []byte) (err error) {
    var (
        bodyLen   int
        headerLen int16
        packLen   int32
    )

    if len(buf) < _rawHeaderSize {
        return ErrProtoPackLen
    }
    packLen = BigEndian.Int32(buf[_packOffset:_headerOffset])
    headerLen = BigEndian.Int16(buf[_headerOffset:_verOffset])
    p.Ver = int32(BigEndian.Int16(buf[_verOffset:_opOffset]))
    p.Op = BigEndian.Int32(buf[_opOffset:_seqOffset])
    p.Seq = BigEndian.Int32(buf[_seqOffset:])
    if packLen > _maxPackSize {
        return ErrProtoPackLen
    }
    if headerLen != _rawHeaderSize {
        return ErrProtoHeaderLen
    }
    if bodyLen = int(packLen - int32(headerLen)); bodyLen > 0 {
        p.Body = buf[headerLen:packLen]
    } else {
        p.Body = nil
    }
    return
}

func (p *Proto) Encode() ([]byte, error) {
    var (
        buf     []byte
        packLen int
    )
    packLen = _rawHeaderSize + len(p.Body)
    buf = make([]byte, packLen)
    BigEndian.PutInt32(buf[_packOffset:_headerOffset], int32(packLen))
    BigEndian.PutInt16(buf[_headerOffset:_verOffset], int16(_rawHeaderSize))
    BigEndian.PutInt16(buf[_verOffset:_opOffset], int16(2)) // 协议版本固定2
    BigEndian.PutInt32(buf[_opOffset:_seqOffset], p.Op)
    BigEndian.PutInt32(buf[_seqOffset:_heartOffset], p.Seq)
    if p.Body != nil {
        // 将消息内容内容写入buffer
        copy(buf[_heartOffset:], p.Body)
    }

    return buf, nil
}
