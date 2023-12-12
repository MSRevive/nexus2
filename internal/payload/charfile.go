package payload

import (
	"fmt"
	"encoding/base64"
	"strconv"
	"bytes"

	"github.com/msrevive/nexus2/pkg/utils"
)

func GenerateCharFile(steam64 string, slot int, data string) (rd *bytes.Reader, fn string, err error) {
	steamID, err := strconv.ParseInt(steam64, 10, 64)
	d, err := base64.StdEncoding.DecodeString(data)
	
	// slot+1 is done because in the game slots start at 1 while FN server's start at 0
	fn = fmt.Sprintf("%s_%d.char", utils.Steam64To32(steamID), slot+1)

	//we want to create the file in memory only to avoid unneeded io operations
	rd = bytes.NewReader(d)
	return
}