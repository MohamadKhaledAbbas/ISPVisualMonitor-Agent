// Package api implements the MikroTik RouterOS API protocol.
package api

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// Sentence represents a RouterOS API sentence (command + attributes).
type Sentence struct {
	Word  string
	Words []string
}

// NewSentence creates a new sentence with the given command word.
func NewSentence(command string) *Sentence {
	return &Sentence{
		Word:  command,
		Words: make([]string, 0),
	}
}

// AddAttribute adds an attribute word (=name=value format).
func (s *Sentence) AddAttribute(name, value string) *Sentence {
	s.Words = append(s.Words, fmt.Sprintf("=%s=%s", name, value))
	return s
}

// AddQuery adds a query word (?name=value format).
func (s *Sentence) AddQuery(name, value string) *Sentence {
	s.Words = append(s.Words, fmt.Sprintf("?%s=%s", name, value))
	return s
}

// AddProplist adds a .proplist attribute to specify which fields to return.
func (s *Sentence) AddProplist(fields ...string) *Sentence {
	if len(fields) > 0 {
		s.Words = append(s.Words, fmt.Sprintf("=.proplist=%s", strings.Join(fields, ",")))
	}
	return s
}

// Reply represents a response from the RouterOS API.
type Reply struct {
	Type string            // !re, !done, !trap, !fatal
	Data map[string]string // Attributes from =key=value pairs
	Tag  string            // Optional tag for async operations
}

// IsDone returns true if this is a !done reply.
func (r *Reply) IsDone() bool {
	return r.Type == "!done"
}

// IsTrap returns true if this is a !trap (error) reply.
func (r *Reply) IsTrap() bool {
	return r.Type == "!trap"
}

// IsFatal returns true if this is a !fatal reply.
func (r *Reply) IsFatal() bool {
	return r.Type == "!fatal"
}

// IsData returns true if this is a !re (data) reply.
func (r *Reply) IsData() bool {
	return r.Type == "!re"
}

// GetMessage returns the message from a trap or fatal reply.
func (r *Reply) GetMessage() string {
	if msg, ok := r.Data["message"]; ok {
		return msg
	}
	return ""
}

// EncodeLength encodes a word length according to RouterOS API protocol.
// Length encoding:
//   - 0x00-0x7F: 1 byte
//   - 0x80-0x3FFF: 2 bytes (0x8000 + length)
//   - 0x4000-0x1FFFFF: 3 bytes (0xC00000 + length)
//   - 0x200000-0xFFFFFFF: 4 bytes (0xE0000000 + length)
//   - 0x10000000-0xFFFFFFFF: 5 bytes (0xF0 + 4 bytes)
func EncodeLength(length int) []byte {
	if length < 0x80 {
		return []byte{byte(length)}
	}
	if length < 0x4000 {
		return []byte{
			byte((length >> 8) | 0x80),
			byte(length & 0xFF),
		}
	}
	if length < 0x200000 {
		return []byte{
			byte((length >> 16) | 0xC0),
			byte((length >> 8) & 0xFF),
			byte(length & 0xFF),
		}
	}
	if length < 0x10000000 {
		return []byte{
			byte((length >> 24) | 0xE0),
			byte((length >> 16) & 0xFF),
			byte((length >> 8) & 0xFF),
			byte(length & 0xFF),
		}
	}
	return []byte{
		0xF0,
		byte((length >> 24) & 0xFF),
		byte((length >> 16) & 0xFF),
		byte((length >> 8) & 0xFF),
		byte(length & 0xFF),
	}
}

// DecodeLength decodes a word length from the reader.
func DecodeLength(r io.Reader) (int, error) {
	var firstByte [1]byte
	if _, err := io.ReadFull(r, firstByte[:]); err != nil {
		return 0, err
	}

	b := firstByte[0]

	// Single byte (0x00-0x7F)
	if b < 0x80 {
		return int(b), nil
	}

	// Two bytes (0x80-0xBF)
	if b < 0xC0 {
		var secondByte [1]byte
		if _, err := io.ReadFull(r, secondByte[:]); err != nil {
			return 0, err
		}
		return int(b&0x3F)<<8 | int(secondByte[0]), nil
	}

	// Three bytes (0xC0-0xDF)
	if b < 0xE0 {
		var extra [2]byte
		if _, err := io.ReadFull(r, extra[:]); err != nil {
			return 0, err
		}
		return int(b&0x1F)<<16 | int(extra[0])<<8 | int(extra[1]), nil
	}

	// Four bytes (0xE0-0xEF)
	if b < 0xF0 {
		var extra [3]byte
		if _, err := io.ReadFull(r, extra[:]); err != nil {
			return 0, err
		}
		return int(b&0x0F)<<24 | int(extra[0])<<16 | int(extra[1])<<8 | int(extra[2]), nil
	}

	// Five bytes (0xF0-0xFF)
	var extra [4]byte
	if _, err := io.ReadFull(r, extra[:]); err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(extra[:])), nil
}

// EncodeWord encodes a single word (length-prefixed string).
func EncodeWord(word string) []byte {
	length := EncodeLength(len(word))
	return append(length, []byte(word)...)
}

// DecodeWord decodes a single word from the reader.
func DecodeWord(r io.Reader) (string, error) {
	length, err := DecodeLength(r)
	if err != nil {
		return "", err
	}

	if length == 0 {
		return "", nil
	}

	word := make([]byte, length)
	if _, err := io.ReadFull(r, word); err != nil {
		return "", err
	}

	return string(word), nil
}

// EncodeSentence encodes a complete sentence (command + attributes + empty word).
func EncodeSentence(s *Sentence) []byte {
	var buf bytes.Buffer

	// Write command word
	buf.Write(EncodeWord(s.Word))

	// Write attribute words
	for _, word := range s.Words {
		buf.Write(EncodeWord(word))
	}

	// Write empty word to terminate sentence
	buf.WriteByte(0)

	return buf.Bytes()
}

// DecodeSentence decodes a complete sentence from the reader.
func DecodeSentence(r io.Reader) (*Reply, error) {
	reply := &Reply{
		Data: make(map[string]string),
	}

	// Read first word (type: !re, !done, !trap, !fatal)
	firstWord, err := DecodeWord(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read reply type: %w", err)
	}
	reply.Type = firstWord

	// Read subsequent words until empty word
	for {
		word, err := DecodeWord(r)
		if err != nil {
			return nil, fmt.Errorf("failed to read word: %w", err)
		}

		// Empty word terminates the sentence
		if word == "" {
			break
		}

		// Parse attribute word (=key=value)
		if strings.HasPrefix(word, "=") {
			parts := strings.SplitN(word[1:], "=", 2)
			if len(parts) == 2 {
				reply.Data[parts[0]] = parts[1]
			} else if len(parts) == 1 {
				reply.Data[parts[0]] = ""
			}
		} else if strings.HasPrefix(word, ".tag=") {
			reply.Tag = strings.TrimPrefix(word, ".tag=")
		}
	}

	return reply, nil
}

// ParseAttributeWord parses a =key=value word and returns key, value.
func ParseAttributeWord(word string) (string, string, bool) {
	if !strings.HasPrefix(word, "=") {
		return "", "", false
	}
	parts := strings.SplitN(word[1:], "=", 2)
	if len(parts) < 2 {
		return parts[0], "", true
	}
	return parts[0], parts[1], true
}
