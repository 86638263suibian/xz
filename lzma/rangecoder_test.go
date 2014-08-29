package lzma

import (
	"bytes"
	"io"
	"testing"
)

type bitEncoder interface {
	encode(b bit) error
	flush() error
}

type bitDecoder interface {
	init() error
	decode() (bit, error)
}

type directEncoder struct {
	e *rangeEncoder
}

func (e directEncoder) encode(b bit) error {
	return e.e.encodeDirect(b)
}

func (e directEncoder) flush() error {
	return e.e.flush()
}

func newDirectEncoder(w io.ByteWriter) bitEncoder {
	return &directEncoder{e: newRangeEncoder(w)}
}

type directDecoder struct {
	d *rangeDecoder
}

func (d directDecoder) init() error {
	return d.d.init()
}

func (d directDecoder) decode() (bit, error) {
	return d.d.decodeDirect()
}

func newDirectDecoder(r io.ByteReader) bitDecoder {
	return &directDecoder{d: newRangeDecoder(r)}
}

type probEncoder struct {
	e *rangeEncoder
	p prob
}

func newProbEncoder(w io.ByteWriter) bitEncoder {
	return &probEncoder{e: newRangeEncoder(w), p: probInit}
}

func (e *probEncoder) encode(b bit) error {
	return e.e.encode(b, &e.p)
}

func (e *probEncoder) flush() error {
	return e.e.flush()
}

type probDecoder struct {
	d *rangeDecoder
	p prob
}

func newProbDecoder(r io.ByteReader) bitDecoder {
	return &probDecoder{d: newRangeDecoder(r), p: probInit}
}

func (d *probDecoder) init() error {
	d.p = probInit
	return d.d.init()
}

func (d *probDecoder) decode() (bit, error) {
	return d.d.decode(&d.p)
}

func encodeByte(e bitEncoder, b byte) error {
	for i := 7; i >= 0; i-- {
		x := bit((b >> uint(i)) & 1)
		if err := e.encode(x); err != nil {
			return err
		}
	}
	return nil
}

func encodeBytes(e bitEncoder, p []byte) error {
	for _, b := range p {
		if err := encodeByte(e, b); err != nil {
			return err
		}
	}
	return e.flush()
}

func decodeByte(d bitDecoder) (b byte, err error) {
	for i := 7; i >= 0; i-- {
		bit, err := d.decode()
		if err != nil {
			return 0, err
		}
		b |= (byte(bit) & 1) << uint(i)
	}
	return b, nil
}

func decodeBytes(t *testing.T, d bitDecoder, n int) (p []byte, err error) {
	if err = d.init(); err != nil {
		return nil, err
	}
	for ; n > 0; n-- {
		b, err := decodeByte(d)
		if err != nil {
			return nil, err
		}
		p = append(p, b)
		t.Logf("p %#v", p)
	}
	return p, nil
}

func testCodec(t *testing.T, buf *bytes.Buffer, e bitEncoder, d bitDecoder,
	w []byte,
) {
	var err error
	if err = encodeBytes(e, w); err != nil {
		t.Fatalf("encodeBytes: %s", err)
	}
	t.Logf("buf.Len() %d", buf.Len())
	t.Logf("buf %#v", buf.Bytes())
	p, err := decodeBytes(t, d, len(w))
	if err != nil {
		t.Fatalf("decodeBytes: %s", err)
	}
	if !bytes.Equal(p, w) {
		t.Logf("p=%#v; want %#v", p, w)
	}
}

func TestDirect(t *testing.T) {
	testStrings := []string{
		"HalloBallo",
	}
	for _, c := range testStrings {
		w := []byte(c)
		t.Logf("w %#v", w)
		var buf bytes.Buffer
		e := newDirectEncoder(&buf)
		d := newDirectDecoder(&buf)
		testCodec(t, &buf, e, d, w)
	}
}

func TestProb(t *testing.T) {
	testStrings := []string{
		"HalloBallo",
	}
	for _, c := range testStrings {
		w := []byte(c)
		t.Logf("w %#v", w)
		var buf bytes.Buffer
		e := newProbEncoder(&buf)
		d := newProbDecoder(&buf)
		testCodec(t, &buf, e, d, w)
	}
}
