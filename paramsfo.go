/* rpcs3-gameupdater param.sfo related */
/* cf https://github.com/xXxTheDarkprogramerxXx/PS3Tools */

package main

import (
	"encoding/binary"
	"os"
	"strconv"
)

const (
	GameData int = 0x4744
	SaveData int = 0x5344
	HDDGame  int = 0x4847
	DiscGame int = 0x4447
)

type Header struct {
	Magic             [4]byte
	Version           [4]byte
	KeyTableStart     uint32
	DataTableStart    uint32
	IndexTableEntries uint32
}

type FMT uint16

const (
	UTF_8  FMT = 0x0400
	ASCII      = 0x0402
	UINT32     = 0x0404
)

type IndexTable struct {
	ParamKeyOffset  uint16
	ParamDataFmt    FMT
	ParamDataLen    uint32
	ParamDataMaxLen uint32
	ParamDataOffset uint32
}

type Table struct {
	IndexTable IndexTable
	Name       string
	Value      string
	Index      int32
}

/* reads from a file descriptor into a header struct */

func readHeader(file *os.File) Header {
	var header Header

	binary.Read(file, binary.LittleEndian, &header.Magic)
	if header.Magic != [4]byte{0, 0x50, 0x53, 0x46} {
		printDebug("The magic is not the one expected")
		return header
	}
	binary.Read(file, binary.LittleEndian, &header.Version)
	binary.Read(file, binary.LittleEndian, &header.KeyTableStart)
	binary.Read(file, binary.LittleEndian, &header.DataTableStart)
	binary.Read(file, binary.LittleEndian, &header.IndexTableEntries)

	return header
}

/* reads from a file descriptor into an indextable struct */

func readIndexTable(file *os.File) IndexTable {
	var indexTable IndexTable

	binary.Read(file, binary.LittleEndian, &indexTable.ParamKeyOffset)
	binary.Read(file, binary.BigEndian, &indexTable.ParamDataFmt)
	binary.Read(file, binary.LittleEndian, &indexTable.ParamDataLen)
	binary.Read(file, binary.LittleEndian, &indexTable.ParamDataMaxLen)
	binary.Read(file, binary.LittleEndian, &indexTable.ParamDataOffset)

	return indexTable
}

func readName(file *os.File, header Header, table IndexTable) string {
	file.Seek(int64(header.KeyTableStart)+int64(table.ParamKeyOffset), 0)
	name := ""
	var c byte
	for binary.Read(file, binary.LittleEndian, &c); c != 0; binary.Read(file, binary.LittleEndian, &c) {
		name = name + string(c)
	}
	file.Seek(-1, 1)
	return name
}

func readValue(file *os.File, header Header, table IndexTable) string {
	file.Seek(int64(header.DataTableStart+table.ParamDataOffset), 0)
	switch table.ParamDataFmt {
	case UINT32:
		var buf uint32
		binary.Read(file, binary.LittleEndian, &buf)
		return strconv.FormatUint(uint64(buf), 10)
	case ASCII:
		buf := make([]byte, table.ParamDataMaxLen)
		binary.Read(file, binary.LittleEndian, &buf)
		return string(buf)
	case UTF_8:
		buf := make([]byte, table.ParamDataMaxLen)
		binary.Read(file, binary.LittleEndian, &buf)
		return string(buf)
	default:
		printDebug("unknown ParamDataFmt")
		return "NAN"
	}
}

/* parses the binary PARAM.sfo file into structs and returns a map of Name/Values */

func readParamSFO(file *os.File) map[string]string {
	header := readHeader(file)
	var indextables []IndexTable
	var i uint32
	for i = 0; i < header.IndexTableEntries; i++ {
		indextables = append(indextables, readIndexTable(file))
	}
	var tables []Table
	kvp := make(map[string]string)
	for count, it := range indextables {
		var t Table
		t.Index = int32(count)
		t.IndexTable = it
		t.Name = readName(file, header, it)
		t.Value = readValue(file, header, it)
		kvp[t.Name] = t.Value
		tables = append(tables, t)
	}
	return kvp
}

func getVersionFromSFO(kvp map[string]string) string {
	return kvp["APP_VER"]
}
