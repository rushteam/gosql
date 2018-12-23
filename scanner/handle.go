package scanner

//Marshaler ..
type Marshaler interface {
	Read()
	Marshaler()
}

//TimeMarshaler ..
type TimeMarshaler struct{}

func (elt TimeMarshaler) Read(fieldAddr interface{}) (scanTarget interface{}, err error) {
	// switch elem := fieldAddr.(type) {
	// case *time.Time:

	// }
	return fieldAddr, nil
}

//Marshaler ..
func (elt TimeMarshaler) Marshaler(fieldAddr interface{}, scanTarget interface{}) error {
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
