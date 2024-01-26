//UUID encoder/decoder from https://gist.github.com/SupaHam/3afe982dc75039356723600ccc91ff77
package bsoncoder

import (
	"fmt"
	"reflect"
	"bytes"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

var (
	tUUID = reflect.TypeOf(uuid.UUID{})
	uuidSubtype = byte(0x04)
)

func GetRegistry() *bsoncodec.Registry {
	registry := bson.NewRegistry()
	registry.RegisterTypeEncoder(tUUID, bsoncodec.ValueEncoderFunc(uuidEncodeValue))
	registry.RegisterTypeDecoder(tUUID, bsoncodec.ValueDecoderFunc(uuidDecodeValue))

	return registry
}

func Encode(val interface{}) ([]byte, error) {
	registry := bson.NewRegistry()
	registry.RegisterTypeEncoder(tUUID, bsoncodec.ValueEncoderFunc(uuidEncodeValue))
	registry.RegisterTypeDecoder(tUUID, bsoncodec.ValueDecoderFunc(uuidDecodeValue))

	dst := make([]byte, 0)
	buf := bytes.NewBuffer(dst)

	vw, err := bsonrw.NewBSONValueWriter(buf)
	if err != nil {
		return dst, err
	}

	enc, err := bson.NewEncoder(vw)
	if err != nil {
		return dst, err
	}

	if err := enc.Encode(val); err != nil {
		return dst, err
	}
	enc.SetRegistry(registry)

	return buf.Bytes(), nil
}

func Decode(data []byte, val interface{}) error {
	registry := bson.NewRegistry()
	registry.RegisterTypeEncoder(tUUID, bsoncodec.ValueEncoderFunc(uuidEncodeValue))
	registry.RegisterTypeDecoder(tUUID, bsoncodec.ValueDecoderFunc(uuidDecodeValue))

	dec, err := bson.NewDecoder(bsonrw.NewBSONDocumentReader(data))
	if err != nil {
		return err
	}

	if err := dec.Decode(val); err != nil {
		return err
	}
	dec.SetRegistry(registry)

	return nil
}

func uuidEncodeValue(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tUUID {
		return bsoncodec.ValueEncoderError{Name: "uuidEncodeValue", Types: []reflect.Type{tUUID}, Received: val}
	}
	b := val.Interface().(uuid.UUID)
	return vw.WriteBinaryWithSubtype(b[:], uuidSubtype)
}

func uuidDecodeValue(dc bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if !val.CanSet() || val.Type() != tUUID {
		return bsoncodec.ValueDecoderError{Name: "uuidDecodeValue", Types: []reflect.Type{tUUID}, Received: val}
	}

	var data []byte
	var subtype byte
	var err error
	switch vrType := vr.Type(); vrType {
	case bsontype.Binary:
		data, subtype, err = vr.ReadBinary()
		if subtype != uuidSubtype {
			return fmt.Errorf("unsupported binary subtype %v for UUID", subtype)
		}
	case bsontype.Null:
		err = vr.ReadNull()
	case bsontype.Undefined:
		err = vr.ReadUndefined()
	default:
		return fmt.Errorf("cannot decode %v into a UUID", vrType)
	}

	if err != nil {
		return err
	}
	uuid2, err := uuid.FromBytes(data)
	if err != nil {
		return err
	}
	val.Set(reflect.ValueOf(uuid2))
	return nil
}