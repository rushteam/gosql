package scanner

//Handle ..
type Handle interface {
	Read()
	Format()
}
type Time struct{}

func (elt Time) Read(fieldAddr, scanTarget interface{}) error {
	return nil
}
