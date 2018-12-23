package scanner

//Field ..
type Field interface {
	Read()
	Format()
}
type TimeField struct{}

func (elt TimeField) Read(fieldAddr interface{}) (scanTarget interface{}, err error) {
	// switch elem := fieldAddr.(type) {
	// case *time.Time:

	// }
	return fieldAddr, nil
}

// func (elt TimeField) Format(fieldAddr interface{}) (scanTarget interface{}, err error) {
// 	return nil
// }
