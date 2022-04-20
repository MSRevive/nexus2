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

func GenerateCharFile(steamid int64, slot int, data []byte) (string, error) {
  filename := fmt.Sprintf("./runtime/temp/%s-%d.char", Steam64ToString(steamid), slot)
  err := os.WriteFile(filename, data, 0666)
  if err != nil {
    return "", err
  }
  
  return filename, nil
}