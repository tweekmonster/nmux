//go:generate sh -c "go-bindata $BIN_DATA_ARGS -pkg=nmux -prefix data/ data/..."
//go:generate stringer -type=Op,Attr,Mode -output screen/const_string.go screen/

package nmux
