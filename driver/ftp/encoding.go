package ftp

import (
	"fmt"
	pathpkg "path"
	"strings"
	"unicode/utf8"

	ftpclient "github.com/jlaffaye/ftp"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"github.com/no-src/gofs/core"
)

const (
	ftpEncodingAuto = "auto"
	ftpEncodingUTF8 = "utf8"
	ftpEncodingGBK  = "gbk"
)

type ftpPathCodec struct {
	mode string
}

func newFTPPathCodec(conf core.FTPConfig) (*ftpPathCodec, error) {
	mode := strings.ToLower(strings.TrimSpace(conf.Encoding))
	if mode == "" {
		mode = ftpEncodingAuto
	}
	switch mode {
	case ftpEncodingAuto, ftpEncodingUTF8, ftpEncodingGBK:
		return &ftpPathCodec{mode: mode}, nil
	default:
		return nil, fmt.Errorf("ftp: unsupported encoding %q", conf.Encoding)
	}
}

func (c *ftpPathCodec) disableUTF8Feature() bool {
	return c.mode == ftpEncodingGBK
}

func (c *ftpPathCodec) encodePath(path string) (string, error) {
	clean := cleanFTPPath(path)
	if c.mode != ftpEncodingGBK {
		return clean, nil
	}
	encoded, _, err := transform.String(simplifiedchinese.GBK.NewEncoder(), clean)
	return encoded, err
}

func (c *ftpPathCodec) decodePath(path string) string {
	if path == "" {
		return path
	}
	decoded, ok := c.decodeString(path)
	if !ok {
		return path
	}
	return pathpkg.Clean(decoded)
}

func (c *ftpPathCodec) decodeEntry(entry *ftpclient.Entry) *ftpclient.Entry {
	if entry == nil {
		return nil
	}
	decoded := *entry
	decoded.Name = c.decodeName(entry.Name)
	decoded.Target = c.decodeName(entry.Target)
	return &decoded
}

func (c *ftpPathCodec) decodeName(name string) string {
	decoded, ok := c.decodeString(name)
	if !ok {
		return name
	}
	return decoded
}

func (c *ftpPathCodec) decodeString(value string) (string, bool) {
	if value == "" {
		return value, true
	}
	switch c.mode {
	case ftpEncodingUTF8:
		return value, true
	case ftpEncodingGBK:
	decoded, _, err := transform.String(simplifiedchinese.GBK.NewDecoder(), value)
		return decoded, err == nil
	case ftpEncodingAuto:
		if utf8.ValidString(value) {
			return value, true
		}
		decoded, _, err := transform.String(simplifiedchinese.GBK.NewDecoder(), value)
		return decoded, err == nil
	default:
		return value, true
	}
}
