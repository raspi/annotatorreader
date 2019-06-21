package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/raspi/annotatorreader"
	"log"
)

// First few bytes of Amiga 500 game Turrican II TFMX music from level 1

var testData = []byte{
	0x54, 0x46, 0x4d, 0x58, 0x2d, 0x53, 0x4f, 0x4e, 0x47, 0x20,
	0x0, 0x1, 0x0, 0x0, 0xe, 0x50,

	0x44, 0x61, 0x74, 0x65, 0x20, 0x3a, 0x20, 0x32, 0x33, 0x2e,
	0x30, 0x31, 0x2e, 0x39, 0x31, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,

	0x54, 0x69, 0x6d, 0x65, 0x20, 0x3a, 0x20, 0x31, 0x37, 0x3a,
	0x35, 0x36, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,

	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,

	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,

	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,

	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,

	0x0, 0x1, 0x0, 0x45, 0x0, 0x51, 0x0, 0x5e,
	0x0, 0x61, 0x0, 0x64, 0x0, 0x65, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x44, 0x1, 0xff,

	0x0, 0x44, 0x0, 0x50, 0x0, 0x5d, 0x0, 0x60,
	0x0, 0x63, 0x0, 0x64, 0x0, 0x65, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x4f, 0x1, 0xff,

	0x0, 0x4, 0x0, 0x3, 0x0, 0x3, 0x0, 0x3,
	0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x5,
	0x0, 0x5, 0x0, 0x5, 0x0, 0x5, 0x0, 0x5,
	0x0, 0x5, 0x0, 0x5, 0x0, 0x5, 0x0, 0x5,
	0x0, 0x5, 0x0, 0x5, 0x0, 0x5, 0x0, 0x5,
	0x0, 0x5, 0x0, 0x5, 0x0, 0x5, 0x0, 0x5,
	0x0, 0x5, 0x0, 0x5, 0x0, 0x5, 0x0, 0x5,
	0x0, 0x5, 0x0, 0x5, 0x0, 0x3, 0x0, 0x5,

	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,

	0x0, 0x0, 0x3, 0xe8,

	0x0, 0x0, 0x30, 0x78,

	0x0, 0x0, 0x31, 0xdc,

	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0,

	0x28, 0x0, 0x63, 0x0, 0x77, 0x18, 0x0, 0x2,
	0x20, 0x1, 0x2, 0x0, 0x76, 0x7f, 0x0, 0x0,
}

type SongRaw struct {
	Header             [10]byte
	Pad                [6]byte
	Text               [6][40]byte
	StartingPositions  [32]uint16
	EndingPositions    [32]uint16
	TempoInformation   [32]uint16
	Mute               [8]int16
	TrackStepPointer   uint32
	PatternDataPointer uint32
	MacroDataPointer   uint32
	Pad2               [36]byte
}

func main() {
	bu := bytes.NewReader(testData)

	var song SongRaw

	a := annotatorreader.New(binary.LittleEndian, bu)

	err := a.Marshal(&song, "song")
	if err != nil {
		log.Fatal(err)
	}

	var track [8]uint16

	err = a.Marshal(&track, "track data")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(`%v`, a.Dump())

}
