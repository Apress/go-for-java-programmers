package main

import "fmt"

type FileSize uint64
const (
	B FileSize = 1 << (10 * iota)
	KiB
	MiB
	GiB
	TiB
	PiB
	EiB
)

var fsNames = [...]string{"EiB","PiB","TiB","GiB","MiB","KiB",""}
func (fs FileSize) scaleFs(scale FileSize, index int) string {
	return fmt.Sprintf("%d%v", FileSize(fs + scale / 2) / scale, fsNames[index])
}
func (fs FileSize) String() (r string) {
	switch {
	case fs >= EiB:
		r = fs.scaleFs(EiB, 0)
	case fs >= PiB:
		r = fs.scaleFs(PiB, 1)
	case fs >= TiB:
		r = fs.scaleFs(TiB, 2)
	case fs >= GiB:
		r = fs.scaleFs(GiB, 3)
	case fs >= MiB:
		r = fs.scaleFs(MiB, 4)
	case fs >= KiB:
		r = fs.scaleFs(KiB, 5)
	default:
		r = fs.scaleFs(1, 6)
	}
	return
}


func main() {
	var v1, v2 FileSize = 1_000_000, 2 * 1e9
	fmt.Printf("FS1: %v; FS2: %v\n", v1, v2)
}

