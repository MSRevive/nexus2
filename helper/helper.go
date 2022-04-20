package helper

import (
  "os"
  "fmt"
  "strconv"
)

func Steam64ToString(steamid int64) string {
  steamid = steamid - 76561197960265728
  remainder := steamid % 2
  steamid = steamid / 2
  
  steamStr := "STEAM_0-" + strconv.FormatInt(remainder, 10) + "-" + strconv.FormatInt(steamid, 10)
  
  return steamStr
}

func GenerateCharFile(steam64 string, slot int, data []byte) (*os.File, string, error) {
  steamid, err := strconv.ParseInt(steam64, 10, 64)
  if err != nil {
    return nil, "", err
  }
  
  filename := fmt.Sprintf("./runtime/temp/%s_%d.char", Steam64ToString(steamid), slot)
  err = os.WriteFile(filename, data, 0666)
  if err != nil {
    return nil, "", err
  }
  
  file, err := os.Open(filename)
  if err != nil {
    return nil, "", err
  }
  
  return file, filename, nil
}