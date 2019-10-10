package types

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/simplechain-org/simplechain/common"
	"io"
	"math/big"
	"strings"
)

var (
	// EOL is returned when the end of the current list
	// has been reached during streaming.
	EOL = errors.New("rlp: end of list")

	// Actual Errors
	ErrExpectedString   = errors.New("rlp: expected String or Byte")
	ErrExpectedList     = errors.New("rlp: expected List")
	ErrCanonInt         = errors.New("rlp: non-canonical integer format")
	ErrCanonSize        = errors.New("rlp: non-canonical size information")
	ErrElemTooLarge     = errors.New("rlp: element is larger than containing list")
	ErrValueTooLarge    = errors.New("rlp: value size exceeds available input length")
	ErrMoreThanOneValue = errors.New("rlp: input contains more than one value")

	// internal errors
	errNotInList     = errors.New("rlp: call of ListEnd outside of any list")
	errNotAtEOL      = errors.New("rlp: call of ListEnd not positioned at EOL")
	errUintOverflow  = errors.New("rlp: uint overflow")
	errNoPointer     = errors.New("rlp: interface given to Decode must be a pointer")
	errDecodeIntoNil = errors.New("rlp: pointer given to Decode must not be nil")
)

// DecodeRLP implements rlp.Decoder
func (tx *Transaction) DecodeHashTx(data []byte) error {
	//fmt.Println("pool: ", data, len(data))

	r := bytes.NewReader(data)
	stream := NewStream(r, uint64(len(data)))
	//a, length, err := stream.readKind()
	//fmt.Println(a, length, err,stream.remaining)
	_, _, err := stream.readKind()

	tx.data.AccountNonce, err = stream.uint(64)
	if err != nil {
		fmt.Println("decode AccountNonce failed: ", err)
	}

	tx.data.Price, err = stream.bigInt()
	tx.data.GasLimit, err = stream.uint(64)
	tx.data.Recipient, err = stream.decodeByteArray()
	tx.data.Amount, err = stream.bigInt()
	tx.data.Payload, err = stream.decodeByteSlice()
	tx.data.V, err = stream.bigInt()
	tx.data.R, err = stream.bigInt()
	tx.data.S, err = stream.bigInt()

	//fmt.Println("tx:   =>>", tx.data)
	//a,b,err :=stream.readKind()
	//fmt.Println(a,b,err,stream.remaining)
	return err
}

// ByteReader must be implemented by any input reader for a Stream. It
// is implemented by e.g. bufio.Reader and bytes.Reader.
type ByteReader interface {
	io.Reader
	io.ByteReader
}
type listpos struct{ pos, size uint64 }

// Kind represents the kind of value contained in an RLP stream.
type Kind int

const (
	Byte Kind = iota
	String
	List
)

// Stream is not safe for concurrent use.
type Stream struct {
	r ByteReader

	// number of bytes remaining to be read from r.
	remaining uint64
	limited   bool

	// auxiliary buffer for integer decoding
	uintbuf []byte

	kind    Kind   // kind of value ahead
	size    uint64 // size of value ahead
	byteval byte   // value of single byte in type tag
	kinderr error  // error from last readKind
	stack   []listpos
}

func NewStream(r io.Reader, inputLimit uint64) *Stream {
	s := new(Stream)
	s.Reset(r, inputLimit)
	return s
}

// Reset discards any information about the current decoding context
// and starts reading from r. This method is meant to facilitate reuse
// of a preallocated Stream across many decoding operations.
//
// If r does not also implement ByteReader, Stream will do its own
// buffering.
func (s *Stream) Reset(r io.Reader, inputLimit uint64) {
	if inputLimit > 0 {
		s.remaining = inputLimit
		s.limited = true
	} else {
		// Attempt to automatically discover
		// the limit when reading from a byte slice.
		switch br := r.(type) {
		case *bytes.Reader:
			s.remaining = uint64(br.Len())
			s.limited = true
		case *strings.Reader:
			s.remaining = uint64(br.Len())
			s.limited = true
		default:
			s.limited = false
		}
	}
	// Wrap r with a buffer if it doesn't have one.
	bufr, ok := r.(ByteReader)
	if !ok {
		bufr = bufio.NewReader(r)
	}
	s.r = bufr
	// Reset the decoding context.
	s.stack = s.stack[:0]
	s.size = 0
	s.kind = -1
	s.kinderr = nil
	if s.uintbuf == nil {
		s.uintbuf = make([]byte, 8)
	}
}

func (s *Stream) readKind() (kind Kind, size uint64, err error) {
	b, err := s.readByte()
	if err != nil {
		if len(s.stack) == 0 {
			// At toplevel, Adjust the error to actual EOF. io.EOF is
			// used by callers to determine when to stop decoding.
			switch err {
			case io.ErrUnexpectedEOF:
				err = io.EOF
			case ErrValueTooLarge:
				err = io.EOF
			}
		}
		return 0, 0, err
	}
	s.byteval = 0
	switch {
	case b < 0x80:
		// For a single byte whose value is in the [0x00, 0x7F] range, that byte
		// is its own RLP encoding.
		s.byteval = b
		return Byte, 0, nil
	case b < 0xB8:
		// Otherwise, if a string is 0-55 bytes long,
		// the RLP encoding consists of a single byte with value 0x80 plus the
		// length of the string followed by the string. The range of the first
		// byte is thus [0x80, 0xB7].
		return String, uint64(b - 0x80), nil
	case b < 0xC0:
		// If a string is more than 55 bytes long, the
		// RLP encoding consists of a single byte with value 0xB7 plus the length
		// of the length of the string in binary form, followed by the length of
		// the string, followed by the string. For example, a length-1024 string
		// would be encoded as 0xB90400 followed by the string. The range of
		// the first byte is thus [0xB8, 0xBF].
		size, err = s.readUint(b - 0xB7)
		if err == nil && size < 56 {
			err = ErrCanonSize
		}
		return String, size, err
	case b < 0xF8:
		// If the total payload of a list
		// (i.e. the combined length of all its items) is 0-55 bytes long, the
		// RLP encoding consists of a single byte with value 0xC0 plus the length
		// of the list followed by the concatenation of the RLP encodings of the
		// items. The range of the first byte is thus [0xC0, 0xF7].
		return List, uint64(b - 0xC0), nil
	default:
		// If the total payload of a list is more than 55 bytes long,
		// the RLP encoding consists of a single byte with value 0xF7
		// plus the length of the length of the payload in binary
		// form, followed by the length of the payload, followed by
		// the concatenation of the RLP encodings of the items. The
		// range of the first byte is thus [0xF8, 0xFF].
		size, err = s.readUint(b - 0xF7)
		if err == nil && size < 56 {
			err = ErrCanonSize
		}
		return List, size, err
	}
}

func (s *Stream) readUint(size byte) (uint64, error) {
	switch size {
	case 0:
		s.kind = -1 // rearm Kind
		return 0, nil
	case 1:
		b, err := s.readByte()
		return uint64(b), err
	default:
		start := int(8 - size)
		for i := 0; i < start; i++ {
			s.uintbuf[i] = 0
		}
		if err := s.readFull(s.uintbuf[start:]); err != nil {
			return 0, err
		}
		if s.uintbuf[start] == 0 {
			// Note: readUint is also used to decode integer
			// values. The error needs to be adjusted to become
			// ErrCanonInt in this case.
			return 0, ErrCanonSize
		}
		return binary.BigEndian.Uint64(s.uintbuf), nil
	}
}

func (s *Stream) uint(maxbits int) (uint64, error) {
	kind, size, err := s.readKind()
	if err != nil {
		return 0, err
	}
	switch kind {
	case Byte:
		if s.byteval == 0 {
			return 0, ErrCanonInt
		}
		s.kind = -1 // rearm Kind
		return uint64(s.byteval), nil
	case String:
		if size > uint64(maxbits/8) {
			return 0, errUintOverflow
		}
		v, err := s.readUint(byte(size))
		switch {
		case err == ErrCanonSize:
			// Adjust error because we're not reading a size right now.
			return 0, ErrCanonInt
		case err != nil:
			return 0, err
		case size > 0 && v < 128:
			return 0, ErrCanonSize
		default:
			return v, nil
		}
	default:
		return 0, ErrExpectedString
	}
}

func (s *Stream) bigInt() (*big.Int, error) {
	b, err := s.Bytes()
	if err != nil {
		return nil, err
	}

	// Reject leading zero bytes
	if len(b) > 0 && b[0] == 0 {
		return nil, ErrCanonInt
	}
	re := new(big.Int)
	return re.SetBytes(b), nil
}

func (s *Stream) decodeByteArray() (*common.Address, error) {
	kind, size, err := s.readKind()
	if err != nil {
		return nil, err
	}
	switch kind {
	case Byte:
		//if vlen == 0 {
		//	return &decodeError{msg: "input string too long", typ: val.Type()}
		//}
		//if vlen > 1 {
		//	return &decodeError{msg: "input string too short", typ: val.Type()}
		//}
		//bv, _ := s.Uint()
		//val.Index(0).SetUint(bv)
	case String:
		//if uint64(vlen) < size {
		//	return &decodeError{msg: "input string too long", typ: val.Type()}
		//}
		//if uint64(vlen) > size {
		//	return &decodeError{msg: "input string too short", typ: val.Type()}
		//}
		slice := make([]byte, common.AddressLength)
		if err := s.readFull(slice); err != nil {
			return nil, err
		}
		// Reject cases where single byte encoding should have been used.
		if size == 1 && slice[0] < 128 {
			return nil, ErrCanonSize
		}
		addr := common.BytesToAddress(slice[:])
		return &addr, nil
	case List:
		return nil, ErrExpectedString
	}

	return nil, nil
}

func (s *Stream) decodeByteSlice() ([]byte, error) {
	b, err := s.Bytes()
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Bytes reads an RLP string and returns its contents as a byte slice.
// If the input does not contain an RLP string, the returned
// error will be ErrExpectedString.
func (s *Stream) Bytes() ([]byte, error) {
	kind, size, err := s.readKind()
	if err != nil {
		return nil, err
	}
	switch kind {
	case Byte:
		s.kind = -1 // rearm Kind
		return []byte{s.byteval}, nil
	case String:
		b := make([]byte, size)
		if err = s.readFull(b); err != nil {
			return nil, err
		}
		if size == 1 && b[0] < 128 {
			return nil, ErrCanonSize
		}
		return b, nil
	default:
		return nil, ErrExpectedString
	}
}

func (s *Stream) readFull(buf []byte) (err error) {
	if err := s.willRead(uint64(len(buf))); err != nil {
		return err
	}
	var nn, n int
	for n < len(buf) && err == nil {
		nn, err = s.r.Read(buf[n:])
		n += nn
	}
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return err
}

func (s *Stream) readByte() (byte, error) {
	if err := s.willRead(1); err != nil {
		return 0, err
	}
	b, err := s.r.ReadByte()
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return b, err
}

func (s *Stream) willRead(n uint64) error {
	s.kind = -1 // rearm Kind

	if len(s.stack) > 0 {
		// check list overflow
		tos := s.stack[len(s.stack)-1]
		if n > tos.size-tos.pos {
			return ErrElemTooLarge
		}
		s.stack[len(s.stack)-1].pos += n
	}
	if s.limited {
		if n > s.remaining {
			return ErrValueTooLarge
		}
		s.remaining -= n
	}
	return nil
}
