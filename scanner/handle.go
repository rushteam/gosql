package scanner

import (
	"time"

	"github.com/pkg/errors"
)

//Marshaler ..
type Marshaler interface {
	Read()
	Marshaler()
}

//TimeMarshaler ..
type TimeMarshaler struct {
	layout string
	loc    *time.Location
}

func (t TimeMarshaler) Read(fieldAddr interface{}) (interface{}, error) {
	return new([]byte), nil
}

//Marshaler ..
func (t TimeMarshaler) Marshaler(fieldAddr interface{}, scanTarget interface{}) error {
	if t.layout == "" {
		t.layout = "2006-01-02 15:04:05"
	}
	value := string(*scanTarget.(*[]uint8))
	tv, err := time.ParseInLocation(t.layout, value, t.loc)
	if err != nil {
		return errors.Wrap(err, "TimeMarshaler")
	}
	switch fieldAddr.(type) {
	case *time.Time:
		*fieldAddr.(*time.Time) = tv
	case **time.Time:
		*fieldAddr.(**time.Time) = &tv
	default:
		return errors.Errorf("TimeMarshaler: unknown struct field type: %T", fieldAddr)
	}
	return nil
}

//CsvMarshaler ..
type CsvMarshaler struct{}

func (elt CsvMarshaler) Read(fieldAddr interface{}) (interface{}, error) {
	return new([]byte), nil
}

//Marshaler ..
func (elt CsvMarshaler) Marshaler(fieldAddr interface{}, scanTarget interface{}) error {
	return nil
}
