package util

import (
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

var DebugEnabled = true
var l = log.New(color.Output, "", log.Ltime)

func init() {
	if DebugEnabled {
		l.SetFlags(log.Ltime | log.Lshortfile)
	}
}

func formatMessage(args ...interface{}) string {
	var msg string

	switch v := args[0].(type) {
	case string:
		msg = v
		if strings.ContainsRune(msg, '%') {
			return fmt.Sprintf(msg, args[1:]...)
		}
	}

	if msg == "" && len(args) == 1 {
		return fmt.Sprintf("%#v", args[0])
	}

	if msg != "" {
		if len(args) == 1 {
			return msg
		}

		msg += " "
	}

	for _, a := range args[1:] {
		switch v := a.(type) {
		case string:
			msg += v
		case fmt.Stringer:
			msg += color.GreenString(v.String())
		case []byte:
			msg += fmt.Sprintf("%d bytes:\n", len(v)) + hex.Dump(v)
		case int:
			msg += color.YellowString(strconv.FormatInt(int64(v), 10))
		case int8:
			msg += color.YellowString(strconv.FormatInt(int64(v), 10))
		case int16:
			msg += color.YellowString(strconv.FormatInt(int64(v), 10))
		case int32:
			msg += color.YellowString(strconv.FormatInt(int64(v), 10))
		case int64:
			msg += color.YellowString(strconv.FormatInt(int64(v), 10))
		case uint:
			msg += color.YellowString(strconv.FormatUint(uint64(v), 10))
		case uint8:
			msg += color.YellowString(strconv.FormatUint(uint64(v), 10))
		case uint16:
			msg += color.YellowString(strconv.FormatUint(uint64(v), 10))
		case uint32:
			msg += color.YellowString(strconv.FormatUint(uint64(v), 10))
		case uint64:
			msg += color.YellowString(strconv.FormatUint(uint64(v), 10))
		default:
			msg += fmt.Sprintf("%#v", v)
		}
		msg += " "
	}

	return msg
}

// Print just prints things.
func Print(args ...interface{}) error {
	return l.Output(2, formatMessage(args...))
}

// Debug only prints if DebugEnabled is true.
func Debug(args ...interface{}) error {
	if DebugEnabled {
		return l.Output(2, color.CyanString("[DEBUG] ")+formatMessage(args...))
	}
	return nil
}
