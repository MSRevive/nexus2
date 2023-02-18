package helper

import (
	"fmt"
	"strconv"
	"bytes"
	"net"
	"strings"
	"net/http"
	"encoding/base64"
)

func GetIP(r *http.Request) string {
	ip,_,_ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func GetRealIP(r *http.Request) string {
	ip := r.Header.Get("X_Real_IP")
	if ip == "" {
		ips := strings.Split(r.Header.Get("X_Forwarded_For"), ", ")
		if ips[0] != "" {
			return ips[0]
		}

		ip,_,_ = net.SplitHostPort(ip)
		return ip
	}

	return ip
}

func Steam64ToString(steamid int64) string {
	steamid = steamid - 76561197960265728
	remainder := steamid % 2
	steamid = steamid / 2
	
	steamStr := "STEAM_0-" + strconv.FormatInt(remainder, 10) + "-" + strconv.FormatInt(steamid, 10)
	
	return steamStr
}

func GenerateCharFile(steam64 string, slot int, data string) (*bytes.Reader, string, error) {
	steamid, err := strconv.ParseInt(steam64, 10, 64)
	if err != nil {
		return nil, "", err
	}
	
	d, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, "", err
	}
	
	filename := fmt.Sprintf("%s_%d.char", Steam64ToString(steamid), slot)
	reader := bytes.NewReader(d) //we want to create the file in memory only to avoid unneeded io operations
	
	return reader, filename, nil
}