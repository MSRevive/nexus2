package utils

import (
	"bytes"
	"strconv"
	"net"
	"strings"
	"net/http"
	"errors"
	"fmt"

	"encoding/json/jsontext"
	"encoding/json/v2"
)

// JSONOptions pins the encoding/json/v2 behaviors that would otherwise change the wire
// format relative to the v1-compatible library this replaced: nil slices and maps encode
// as null rather than []/{}, map keys stay sorted, HTML characters stay escaped, and
// object member names match case-insensitively.
var JSONOptions = json.JoinOptions(
	json.FormatNilSliceAsNull(true),
	json.FormatNilMapAsNull(true),
	json.Deterministic(true),
	json.MatchCaseInsensitiveNames(true),
	jsontext.EscapeForHTML(true),
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

// Credit to https://github.com/tidwall/jsonc - MIT License https://github.com/tidwall/jsonc/blob/master/LICENSE
func StandardJSON(src, dst []byte) []byte {
	dst = dst[:0]

	for i := 0; i < len(src); i++ {
		if src[i] == '/' {
			if i < len(src)-1 {
				if src[i+1] == '/' {
					dst = append(dst, ' ', ' ')
					i += 2
					for ; i < len(src); i++ {
						if src[i] == '\n' {
							dst = append(dst, '\n')
							break
						} else if src[i] == '\t' || src[i] == '\r' {
							dst = append(dst, src[i])
						} else {
							dst = append(dst, ' ')
						}
					}
					continue
				}
				if src[i+1] == '*' {
					dst = append(dst, ' ', ' ')
					i += 2
					for ; i < len(src)-1; i++ {
						if src[i] == '*' && src[i+1] == '/' {
							dst = append(dst, ' ', ' ')
							i++
							break
						} else if src[i] == '\n' || src[i] == '\t' ||
							src[i] == '\r' {
							dst = append(dst, src[i])
						} else {
							dst = append(dst, ' ')
						}
					}
					continue
				}
			}
		}
		
		dst = append(dst, src[i])
		if src[i] == '"' {
			for i = i + 1; i < len(src); i++ {
				dst = append(dst, src[i])
				if src[i] == '"' {
					j := i - 1
					for ; ; j-- {
						if src[j] != '\\' {
							break
						}
					}
					if (j-i)%2 != 0 {
						break
					}
				}
			}
		} else if src[i] == '}' || src[i] == ']' {
			for j := len(dst) - 2; j >= 0; j-- {
				if dst[j] <= ' ' {
					continue
				}
				if dst[j] == ',' {
					dst[j] = ' '
				}
				break
			}
		}
	}

	return dst
}

func ProcessJSON(body []byte, v any) error {
	if len(bytes.TrimSpace(body)) == 0 {
		return fmt.Errorf("request body is empty")
	}

	if err := json.Unmarshal(body, v, JSONOptions); err != nil {
		var syntaxErr *jsontext.SyntacticError
		var semanticErr *json.SemanticError

		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("json syntax error at byte %d: %w", syntaxErr.ByteOffset, err)
		case errors.As(err, &semanticErr):
			return fmt.Errorf("json type mismatch for field %q: %w", semanticErr.JSONPointer.LastToken(), err)
		default:
			return fmt.Errorf("malformed json: %w", err)
		}
	}

	return nil
}