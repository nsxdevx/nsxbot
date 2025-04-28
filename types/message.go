package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"net/http"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Text struct {
	Text string `json:"text"`
}

func (t Text) Cmd(prefix string) (string, error) {
	trimmed := strings.TrimSpace(t.Text)
	if !strings.HasPrefix(trimmed, prefix) {
		return "", errors.New("not a command")
	}
	parts := strings.Fields(strings.TrimLeft(trimmed, prefix))
	if len(parts) < 2 {
		return "", errors.New("not enough parts")
	}
	return parts[0], nil
}

func (t Text) CmdIndex(prefix string, index int) (string, error) {
	if _, err := t.Cmd(prefix); err != nil {
		return "", err
	}
	trimmed := strings.TrimSpace(t.Text)
	parts := strings.Fields(trimmed)
	if len(parts)-1 < index {
		return "", errors.New("index out of range")
	}
	return parts[index+1], nil
}

func (t Text) CmdKey(key string) (string, error) {
	trimmed := strings.TrimSpace(t.Text)
	parts := strings.Fields(trimmed)
	if len(parts) < 3 {
		return "", errors.New("not enough parts")
	}
	for i := 1; i+1 < len(parts); i = i + 2 {
		if strings.EqualFold(strings.ToLower(parts[i]), strings.ToLower(key)) {
			return parts[i+1], nil
		}
	}
	return "", errors.New("key not found")
}

type Face struct {
	Id string `json:"id"`
}

type At struct {
	QQ string `json:"qq"`
}

type Reply struct {
	Id int `json:"id"`
}

var ErrNetWork = errors.New("network error")

type BaseFile struct {
	File string `json:"file"`
	Url  string `json:"url,omitzero"`
	Path string `json:"path,omitzero"`

	FileID     string `json:"file_id,omitzero"`
	FileSize   string `json:"file_size,omitzero"`
	FileUnique string `json:"file_unique,omitzero"`
}

type Image struct {
	BaseFile
	Name     string `json:"name,omitzero"`
	Summary  string `json:"summary,omitzero"`
	SubType  string `json:"subtype,omitzero"`
	realType string
}

// In Go 1.22 RSA key exchange based cipher suites were
// removed from the default list, but can be re-added with the
// GODEBUG setting tlsrsakex=1 or use noTls to get qq image Type() or Decode()
// Type returns the image real type.! For qq iamge set GODEBUG setting tlsrsakex=1 or use ToHttp
func (i *Image) RealType(noTls bool) (string, error) {
	if len(i.realType) > 0 {
		return i.realType, nil
	}
	url := i.Url
	if noTls {
		url = strings.Replace(i.Url, "https://", "http://", 1)
	}
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		if resp.Body.Close() != nil {
			slog.Error("failed to close response body", "err", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", ErrNetWork
	}

	buffer := make([]byte, 16)
	if _, err := io.ReadFull(resp.Body, buffer); err != nil {
		return "", err
	}

	var typ string
	switch {
	case bytes.HasPrefix(buffer, []byte{0xFF, 0xD8}):
		typ = "jpeg"
	case bytes.HasPrefix(buffer, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}):
		typ = "png"
	case bytes.HasPrefix(buffer, []byte{0x47, 0x49, 0x46, 0x38}):
		typ = "gif"
	case bytes.HasPrefix(buffer, []byte{0x42, 0x4D}):
		typ = "bmp"
	case len(buffer) >= 12 && bytes.HasPrefix(buffer, []byte{0x52, 0x49, 0x46, 0x46}) && bytes.Equal(buffer[8:12], []byte{0x57, 0x45, 0x42, 0x50}):
		typ = "webp"
	default:
		return "", errors.New("unknown image type")
	}
	i.realType = typ
	return typ, nil
}

// In Go 1.22 RSA key exchange based cipher suites were
// removed from the default list, but can be re-added with the
// GODEBUG setting tlsrsakex=1 or use noTls to get qq image Type() or Decode()
// Decode to image.Image ! For qq iamge set GODEBUG setting tlsrsakex=1 or use ToHttp
func (i *Image) Decode(noTls bool) (image.Image, error) {
	url := i.Url
	if noTls {
		url = strings.Replace(i.Url, "https://", "http://", 1)
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resp.Body.Close() != nil {
			slog.Error("failed to close response body", "err", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrNetWork
	}

	img, name, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	i.realType = name
	return img, nil
}
