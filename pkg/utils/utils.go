package utils

import (
	"strconv"
	"net"
	"strings"
	"net/http"
	"errors"
	"fmt"
	"io"

	json "github.com/sugawarayuuta/sonnet"
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
	if err := json.Unmarshal(body, v); err != nil {
		var errln error
		var syntaxErr *json.SyntaxError
		var unmarshalErr *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxErr):
			errln = fmt.Errorf("json syntax error at byte %d: %w", syntaxErr.Offset, err)
		case errors.As(err, &unmarshalErr):
			errln = fmt.Errorf("json type mismatch for field %q: %w", unmarshalErr.Field, err)
		case errors.Is(err, io.EOF):
			errln = fmt.Errorf("request body is empty")
		default:
			errln = fmt.Errorf("malformed json: %w", err)
		}

		if errln == nil {
			errln = fmt.Errorf("unknown error: %w", err)
		}

		return errln
	}

	return nil
}