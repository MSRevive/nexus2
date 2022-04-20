package helper

import (
  "fmt"
  "strconv"
  "bytes"
)

func Steam64ToString(steamid int64) string {
  steamid = steamid - 76561197960265728
  remainder := steamid % 2
  steamid = steamid / 2
  
  steamStr := "STEAM_0-" + strconv.FormatInt(remainder, 10) + "-" + strconv.FormatInt(steamid, 10)
  
  return steamStr
}

func GenerateCharFile(steam64 string, slot int, data []byte) (*bytes.Reader, string, error) {
  steamid, err := strconv.ParseInt(steam64, 10, 64)
  if err != nil {
    return nil, "", err
  }
  
  filename := fmt.Sprintf("%s_%d.char", Steam64ToString(steamid), slot)
  reader := bytes.NewReader(data) //we want to create the file in memory only to avoid unneeded io operations
  
  return reader, filename, nil
}