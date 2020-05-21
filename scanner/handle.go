package scanner

import (
	"fmt"
	"time"
)

//Marshaler ..
type Marshaler interface {
	Read()
	Marshaler()
}

//TimeMarshaler ..
type TimeMarshaler struct{}

func (elt TimeMarshaler) Read(fieldAddr interface{}) (scanTarget interface{}, err error) {
	// switch fieldAddr.(type) {
	// case *time.Time:
	// 	return fieldAddr, nil
	// case **time.Time:
	// 	return fieldAddr, nil
	// default:
	// 	return nil, fmt.Errorf("TimeMeddler.Read: unknown struct field type: %T", fieldAddr)
	// }
	// return new([]uint8), nil
	return new([]byte), nil
}

//Marshaler ..
func (elt TimeMarshaler) Marshaler(fieldAddr interface{}, scanTarget interface{}) error {
	// fmt.Println(string(*scanTarget.(*[]uint8)))
	// *fieldAddr.(*time.Time), _ = time.Parse("2006-01-02 15:04:05", string(*ptr))
	ptr := scanTarget.(*[]uint8)
	t, _ := time.Parse("2006-01-02 15:04:05", string(*ptr))
	switch fieldAddr.(type) {
	case *time.Time:
		*fieldAddr.(*time.Time) = t
	case **time.Time:
		*fieldAddr.(**time.Time) = &t
	default:
		return fmt.Errorf("unknown struct field type: %T", fieldAddr)
	}
	return nil
}

//CsvMarshaler ..
type CsvMarshaler struct{}

func (elt CsvMarshaler) Read(fieldAddr interface{}) (scanTarget interface{}, err error) {
	scanTarget = new(string)
	return scanTarget, nil
}

//Marshaler ..
func (elt CsvMarshaler) Marshaler(fieldAddr interface{}, scanTarget interface{}) error {
	return nil
}
