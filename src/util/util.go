package util

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"strings"
	"time"
)

type color string
type debugLevel int

// global var set in main for debug level
var Debug debugLevel

const (
	Lock  debugLevel = 1
	Trace debugLevel = 2
	Info  debugLevel = 3
	None  debugLevel = 4
)

const (
	Default   color = ""
	Underline color = "\033[4m"
	Red       color = "\033[31m"
	Green     color = "\033[32m"
	Yellow    color = "\033[33m"
	Blue      color = "\033[34m"
	Purple    color = "\033[35m"
	Cyan      color = "\033[36m"
	Reset     color = "\033[0m"
)

func StartTest(desc string) {
	if Debug != None {
		desc = desc + "\n"
	}
	ColorPrintf(Underline, desc)
}

func EndTest() {
	ColorPrintf(Green, "  pass\n")
}

// default printing
func Printf(format string, a ...interface{}) (n int, err error) {
	ColorPrintf(Default, format, a...)
	return
}

func Printfln(format string, a ...interface{}) (n int, err error) {
	ColorPrintf(Default, format+"\n", a...)
	return
}

// error logging
func EPrintf(format string, a ...interface{}) (n int, err error) {
	ColorPrintf(Red, "[ERROR] "+format, a...)
	return
}

// info logging
func IPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug <= Info {
		ColorPrintf(Blue, format, a...)
	}
	return
}

// trace logging
func TPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug <= Trace {
		format = strings.Replace(format, "\n", "\n    ", strings.Count(format, "\n")-1)
		ColorPrintf(Default, "    "+format, a...)
	}
	return
}

//lock logging
func LPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug <= Lock {
		ColorPrintf(Default, format, a...)
	}
	return
}

// generic printing with color
func ColorPrintf(c color, format string, a ...interface{}) (n int, err error) {
	str := string(c) + format + string(Reset)
	fmt.Printf(str, a...)
	return
}

func Wait(milliseconds int) {
	<-time.After(time.Millisecond * time.Duration(milliseconds))
}

func ByteArrayEquals(first []byte, second []byte) bool {
	if first == nil && second == nil {
		return true
	}
	if first == nil || second == nil {
		return false
	}
	if len(first) != len(second) {
		return false
	}
	for i := range first {
		if first[i] != second[i] {
			return false
		}
	}
	return true
}

func BoolArrayEquals(first []bool, second []bool) bool {
	if first == nil && second == nil {
		return true
	}
	if first == nil || second == nil {
		return false
	}
	if len(first) != len(second) {
		return false
	}
	for i := range first {
		if first[i] != second[i] {
			return false
		}
	}
	return true
}

// from http://stackoverflow.com/questions/25686109/split-string-by-length-in-golang
func SplitEveryN(s string, n int) []string {
	sub := ""
	subs := []string{}

	runes := bytes.Runes([]byte(s))
	l := len(runes)
	for i, r := range runes {
		sub = sub + string(r)
		if (i+1)%n == 0 {
			subs = append(subs, sub)
			sub = ""
		} else if (i + 1) == l {
			subs = append(subs, sub)
		}
	}

	return subs
}

func GenerateRandStr(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return string(b[:length])
}

func BoolsToBytes(data []bool) []byte {
	if len(data)%8 != 0 {
		// we need to pad the message with zeros
		padBuf := make([]bool, 8-(len(data)%8))
		// the extra pad needs to be false (0-padded)
		data = append(data, padBuf...)
	}

	if len(data)%8 != 0 {
		EPrintf("boolsToBytes: wtf")
	}

	output := make([]byte, len(data)/8)
	for i := 0; i < len(data); i++ {
		val := byte(0)
		if data[i] {
			val = byte(0x80)
		}
		output[i/8] = output[i/8] | (val >> uint(i%8))
	}
	return output
}

func BytesToBools(data []byte) []bool {
	output := make([]bool, len(data)*8)
	for i := 0; i < len(data)*8; i++ {
		mask := (byte(0x80) >> uint(i%8))
		if (data[i/8] & mask) > 0 {
			output[i] = true
		}
		// We dont have to set the false bits because thats done for us by the make operation
	}
	return output
}
