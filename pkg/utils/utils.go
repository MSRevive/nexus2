package utils

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

func Steam64To32(steamID int64) (steam32 string) {
	steamID = steamID - 76561197960265728
	remainder := steamID % 2
	steamID = steamID / 2
	
	steam32 = "STEAM_0-" + strconv.FormatInt(remainder, 10) + "-" + strconv.FormatInt(steamID, 10)
	return
}