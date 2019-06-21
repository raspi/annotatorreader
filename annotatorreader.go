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

type ReaderDumper struct {
	Endian               binary.ByteOrder
	r                    io.ReadSeeker
	currentOffset        Offset
	DebugInfo            DebugInfoMap
	valuePerimeterColors []clr.Color
	lastVPC              int // last used valuePerimeterColor
	characterColors      map[uint8]clr.Color
	endianDumps          []binary.ByteOrder
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
		endianDumps:          []binary.ByteOrder{endiannessType},
	}
}

func (bd *ReaderDumper) Seek(offset Offset, whence int) (Offset, error) {
	off, err := bd.r.Seek(int64(offset), whence)
	return Offset(off), err
}

// Offset returns file's current offset
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
		iface := v.Interface()
		val := &iface

		s := uint64(binary.Size(iface))

		bd.DebugInfo[bd.currentOffset] = DebugInformation{
			Size:     s,
			Name:     name,
			Kind:     v.Kind(),
			Type:     v.Type(),
			Value:    val,
			LastKind: bd.getLastKind(reflect.ValueOf(val)),
		}

		bd.currentOffset += Offset(s)
	} else {
		panic(`no interfacing!!?!?!`)
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

		var rtree reflTree

		for i := 0; i < rdv.NumField(); i++ {
			f := rdv.Field(i)
			t := rdt.Field(i)

			rtree.getTree(f)

			if len(rtree.Tree) > 2 && rtree.Tree[0] == reflect.Array && rtree.Tree[1] == reflect.Array {
				iface := f.Interface()
				tmp := reflect.ValueOf(iface)

				for i := 0; i < tmp.Len(); i++ {
					bd.readInterface(reflect.ValueOf(tmp.Index(i).Interface()), strings.TrimSpace(fmt.Sprintf(`%v.%v[%v]`, nameprefix, t.Name, i)))
				}
			} else {

				bd.readInterface(f, strings.TrimSpace(fmt.Sprintf(`%v.%v`, nameprefix, t.Name)))
			}

			rtree.Tree = []reflect.Kind{}
		}

	case reflect.Array:
		bd.readInterface(rdv, nameprefix)

	default:
		return fmt.Errorf(`not supported: %T`, data)

	}

	return err
}

func (bd *ReaderDumper) getLastKind(v reflect.Value) reflect.Kind {
	v = bd.unpackValue(v)
	tree := reflTree{}
	tree.getTree(v)
	return tree.Tree[len(tree.Tree)-1]
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

func (bd *ReaderDumper) getVariableColor() clr.Color {
	return bd.valuePerimeterColors[bd.lastVPC]
}

// Highlight every mod X first integer
func (bd *ReaderDumper) getHighlightEffect(k reflect.Kind, c clr.Color, i uint64) clr.Color {
	eff := clr.UnderlineFm

	switch {
	case i%2 == 0 && k == reflect.Uint16:
		c = c | eff
	case i%4 == 0 && k == reflect.Uint32:
		c = c | eff
	case i%8 == 0 && k == reflect.Uint64:
		c = c | eff
	}

	return c
}

// Dump dumps readable hex dump where:
// -each variable boundary line has it's own color
// -ASCII, space, null etc characters have their own colors
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
			sb.WriteString(clr.Sprintf(`%08[1]x  `, clr.Colorize(startingOffset+Offset(idx), bd.getVariableColor())))

			paddingRequired := 16

			// ---- Print character's hex presentation
			for i := uint64(0); i < 16; i++ {
				if idx+i >= item.Size {
					break
				}

				paddingRequired--

				b := byts[idx+i]

				color, ok := bd.characterColors[b]

				if !ok {
					color = clr.WhiteFg
				}

				color = bd.getHighlightEffect(item.LastKind, color, i)

				sb.WriteString(clr.Sprintf(`%02x`, clr.Colorize(b, color)))

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

			sb.WriteString(` `)
			sb.WriteString(clr.Sprintf(`%v`, clr.Bold(`|`)))

			// ---- Show ASCII characters
			for i := uint64(0); i < 16; i++ {
				if idx+i >= item.Size {
					break
				}

				// Print character's ASCII presentation
				b := byts[idx+i]

				color, ok := bd.characterColors[b]

				if !ok {
					color = clr.WhiteFg
				}

				color = bd.getHighlightEffect(item.LastKind, color, i)

				sb.WriteString(fmt.Sprintf(`%c`, clr.Colorize(bd.toChar(b), color)))

			}

			sb.WriteString(clr.Sprintf(`%v`, clr.Bold(`|`)))
			sb.WriteString(` `)

			for i := 0; i < paddingRequired; i++ {
				sb.WriteString(` `)
			}

			// --- Variable information
			// Type
			sb.WriteString(clr.Sprintf(`%v `, clr.Colorize(item.Type, bd.getVariableColor())))
			// Name
			sb.WriteString(clr.Sprintf(`%v: `, clr.Colorize(item.Name, bd.getVariableColor())))
			// Length of data
			sb.WriteString(clr.Sprintf(`L: 0x%02[1]X %[1]v `, item.Size))

			if item.Value != nil {
				// Value
				switch (*item.Value).(type) {
				case uint8, int8:
					sb.WriteString(clr.Sprintf(`V: 0x%02[1]x %[1]v`, *item.Value))
				case uint16, int16:
					sb.WriteString(clr.Sprintf(`V: 0x%04[1]x %[1]v`, *item.Value))
				case uint32, int32:
					sb.WriteString(clr.Sprintf(`V: 0x%04[1]x %[1]v`, *item.Value))
				case uint64, int64:
					sb.WriteString(clr.Sprintf(`V: 0x%04[1]x %[1]v`, *item.Value))
				case uint, int:
					sb.WriteString(clr.Sprintf(`V: 0x%04[1]x %[1]v`, *item.Value))
				}
			}

			sb.WriteString("\n")
		}

		bd.nextColor()

	}

	return sb.String()
}
