package annotatorreader

import (
	"encoding/binary"
	"fmt"
	clr "github.com/logrusorgru/aurora"
	"io"
	"reflect"
	"sort"
	"strings"
)

type Offset int64
type DebugInfoMap map[Offset]DebugInformation

type ReaderDumper struct {
	Endian               binary.ByteOrder
	r                    io.ReadSeeker
	currentOffset        Offset
	DebugInfo            DebugInfoMap
	valuePerimeterColors []clr.Color
	lastVPC              int // last used valuePerimeterColor
	characterColors      map[uint8]clr.Color
}

type DebugInformation struct {
	Name string
	Size uint64
	Kind reflect.Kind
	Type reflect.Type
}

func New(endiannessType binary.ByteOrder, rdr io.ReadSeeker) ReaderDumper {
	return ReaderDumper{
		Endian:               endiannessType,
		r:                    rdr,
		currentOffset:        0,
		DebugInfo:            DebugInfoMap{},
		valuePerimeterColors: []clr.Color{clr.CyanFg, clr.MagentaFg, clr.GreenFg, clr.BlueFg, clr.YellowFg},
		lastVPC:              0,
		characterColors:      defaultCharacterColors,
	}
}

func (bd *ReaderDumper) Seek(offset Offset, whence int) (Offset, error) {
	off, err := bd.r.Seek(int64(offset), whence)
	return Offset(off), err
}

func (bd *ReaderDumper) Offset() (Offset, error) {
	off, err := bd.r.Seek(0, io.SeekCurrent)
	return Offset(off), err
}

func (bd *ReaderDumper) unpackValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}

	if v.Kind() == reflect.Interface && !v.IsNil() {
		v = v.Elem()
	}

	return v
}

func (bd *ReaderDumper) unpackType(v reflect.Type) reflect.Type {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	return v
}

func (bd *ReaderDumper) toChar(b byte) byte {
	if b < 32 || b > 126 {
		return '.'
	}
	return b
}

func (bd *ReaderDumper) readInterface(v reflect.Value, name string) {
	if v.CanInterface() {
		s := uint64(binary.Size(v.Interface()))

		bd.DebugInfo[bd.currentOffset] = DebugInformation{
			Size: s,
			Name: name,
			Kind: v.Kind(),
			Type: v.Type(),
		}

		bd.currentOffset += Offset(s)
	}
}

func (bd *ReaderDumper) Marshal(data interface{}, nameprefix string) (err error) {
	rdv := bd.unpackValue(reflect.ValueOf(data))
	rdt := bd.unpackType(reflect.TypeOf(data))

	bd.currentOffset, err = bd.Offset()
	if err != nil {
		return err
	}

	switch rdt.Kind() {
	case reflect.Struct:
		err = binary.Read(bd.r, bd.Endian, data)
		if err != nil {
			return err
		}

		for i := 0; i < rdv.NumField(); i++ {
			f := rdv.Field(i)
			t := rdt.Field(i)
			bd.readInterface(f, strings.TrimSpace(fmt.Sprintf(`%v.%v`, nameprefix, t.Name)))
		}

	case reflect.Array:
		bd.readInterface(rdv, nameprefix)

	default:
		return fmt.Errorf(`not supported: %T`, data)

	}

	return err
}

func (bd *ReaderDumper) readBytes(data []byte) (n int, err error) {
	return bd.r.Read(data)
}

func (bd *ReaderDumper) nextColor() {
	bd.lastVPC++

	if bd.lastVPC >= len(bd.valuePerimeterColors) {
		// Loop back to start
		bd.lastVPC = 0
	}
}

func (bd *ReaderDumper) getColor() clr.Color {
	return bd.valuePerimeterColors[bd.lastVPC]
}

func (bd *ReaderDumper) Dump() string {
	var sb strings.Builder

	startingOffsetsSorted := make([]Offset, 0)
	for k := range bd.DebugInfo {
		startingOffsetsSorted = append(startingOffsetsSorted, k)
	}

	sort.Slice(startingOffsetsSorted, func(i, j int) bool {
		return startingOffsetsSorted[i] < startingOffsetsSorted[j]
	})

	for _, startingOffset := range startingOffsetsSorted {
		_, _ = bd.Seek(startingOffset, io.SeekStart)
		item := bd.DebugInfo[startingOffset]

		byts := make([]byte, item.Size)
		_, _ = bd.r.Read(byts)

		for idx := uint64(0); idx < item.Size; idx += 16 {
			// Print starting offset
			sb.WriteString(clr.Sprintf(`%08[1]x  `, clr.Colorize(startingOffset+Offset(idx), bd.getColor())))

			paddingRequired := 16

			for i := uint64(0); i < 16; i++ {
				if idx+i >= item.Size {
					break
				}

				paddingRequired--

				// Print character's hex presentation
				b := byts[idx+i]

				color, ok := bd.characterColors[b]

				if !ok {
					sb.WriteString(clr.Sprintf(`%02x`, b))
				} else {
					sb.WriteString(clr.Sprintf(`%02x`, clr.Colorize(b, color)))
				}

				sb.WriteString(` `)

				if i == 7 {
					// Add extra space for better readability
					sb.WriteString(` `)
				}
			}

			// Add possible spacing before ASCII
			if paddingRequired > 8 {
				sb.WriteString(` `)
			}

			for i := 0; i < paddingRequired; i++ {
				sb.WriteString(`   `)
			}

			// Show ASCII characters
			sb.WriteString(` `)
			sb.WriteString(clr.Sprintf(`%v`, clr.Bold(`|`)))

			for i := uint64(0); i < 16; i++ {
				if idx+i >= item.Size {
					break
				}

				// Print character's hex presentation
				b := byts[idx+i]

				color, ok := bd.characterColors[b]

				if !ok {
					sb.WriteString(fmt.Sprintf(`%c`, bd.toChar(b)))
				} else {
					sb.WriteString(fmt.Sprintf(`%c`, clr.Colorize(bd.toChar(b), color)))
				}

			}

			sb.WriteString(clr.Sprintf(`%v`, clr.Bold(`|`)))
			sb.WriteString(` `)

			// Variable names
			for i := 0; i < paddingRequired; i++ {
				sb.WriteString(` `)
			}

			sb.WriteString(clr.Sprintf(`%[1]v: L: 0x%02[2]X %[2]v (%[3]v %[4]v)`, clr.Colorize(item.Name, bd.getColor()), item.Size, clr.Bold(item.Kind), item.Type))
			sb.WriteString("\n")
		}

		bd.nextColor()

	}

	return sb.String()
}
