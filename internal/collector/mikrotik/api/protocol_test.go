package api

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncodeLength(t *testing.T) {
	tests := []struct {
		name     string
		length   int
		expected []byte
	}{
		{
			name:     "single byte - 0",
			length:   0,
			expected: []byte{0x00},
		},
		{
			name:     "single byte - 127",
			length:   127,
			expected: []byte{0x7F},
		},
		{
			name:     "two bytes - 128",
			length:   128,
			expected: []byte{0x80, 0x80},
		},
		{
			name:     "two bytes - 16383",
			length:   16383,
			expected: []byte{0xBF, 0xFF},
		},
		{
			name:     "three bytes - 16384",
			length:   16384,
			expected: []byte{0xC0, 0x40, 0x00},
		},
		{
			name:     "three bytes - 2097151",
			length:   2097151,
			expected: []byte{0xDF, 0xFF, 0xFF},
		},
		{
			name:     "four bytes - 2097152",
			length:   2097152,
			expected: []byte{0xE0, 0x20, 0x00, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeLength(tt.length)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("EncodeLength(%d) = %v, want %v", tt.length, result, tt.expected)
			}
		})
	}
}

func TestDecodeLength(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected int
		wantErr  bool
	}{
		{
			name:     "single byte - 0",
			input:    []byte{0x00},
			expected: 0,
		},
		{
			name:     "single byte - 127",
			input:    []byte{0x7F},
			expected: 127,
		},
		{
			name:     "two bytes - 128",
			input:    []byte{0x80, 0x80},
			expected: 128,
		},
		{
			name:     "two bytes - 16383",
			input:    []byte{0xBF, 0xFF},
			expected: 16383,
		},
		{
			name:     "three bytes - 16384",
			input:    []byte{0xC0, 0x40, 0x00},
			expected: 16384,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.input)
			result, err := DecodeLength(reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeLength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("DecodeLength() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	lengths := []int{0, 1, 127, 128, 255, 256, 16383, 16384, 65535, 100000, 2097151}

	for _, length := range lengths {
		encoded := EncodeLength(length)
		reader := bytes.NewReader(encoded)
		decoded, err := DecodeLength(reader)
		if err != nil {
			t.Errorf("DecodeLength failed for length %d: %v", length, err)
			continue
		}
		if decoded != length {
			t.Errorf("Round-trip failed: encoded %d, decoded %d", length, decoded)
		}
	}
}

func TestEncodeWord(t *testing.T) {
	tests := []struct {
		name     string
		word     string
		expected []byte
	}{
		{
			name:     "empty word",
			word:     "",
			expected: []byte{0x00},
		},
		{
			name:     "short word",
			word:     "/login",
			expected: append([]byte{0x06}, []byte("/login")...),
		},
		{
			name:     "longer word",
			word:     "/system/resource/print",
			expected: append([]byte{0x16}, []byte("/system/resource/print")...),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeWord(tt.word)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("EncodeWord(%q) = %v, want %v", tt.word, result, tt.expected)
			}
		})
	}
}

func TestDecodeWord(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
		wantErr  bool
	}{
		{
			name:     "empty word",
			input:    []byte{0x00},
			expected: "",
		},
		{
			name:     "short word",
			input:    append([]byte{0x06}, []byte("/login")...),
			expected: "/login",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.input)
			result, err := DecodeWord(reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeWord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("DecodeWord() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNewSentence(t *testing.T) {
	s := NewSentence("/system/resource/print")
	
	if s.Word != "/system/resource/print" {
		t.Errorf("expected Word to be '/system/resource/print', got %q", s.Word)
	}
	
	if len(s.Words) != 0 {
		t.Errorf("expected Words to be empty, got %v", s.Words)
	}
}

func TestSentenceAddAttribute(t *testing.T) {
	s := NewSentence("/interface/print")
	s.AddAttribute("name", "ether1")
	
	if len(s.Words) != 1 {
		t.Errorf("expected 1 word, got %d", len(s.Words))
	}
	
	if s.Words[0] != "=name=ether1" {
		t.Errorf("expected '=name=ether1', got %q", s.Words[0])
	}
}

func TestSentenceAddQuery(t *testing.T) {
	s := NewSentence("/interface/print")
	s.AddQuery("running", "true")
	
	if len(s.Words) != 1 {
		t.Errorf("expected 1 word, got %d", len(s.Words))
	}
	
	if s.Words[0] != "?running=true" {
		t.Errorf("expected '?running=true', got %q", s.Words[0])
	}
}

func TestSentenceAddProplist(t *testing.T) {
	s := NewSentence("/interface/print")
	s.AddProplist("name", "type", "running")
	
	if len(s.Words) != 1 {
		t.Errorf("expected 1 word, got %d", len(s.Words))
	}
	
	if s.Words[0] != "=.proplist=name,type,running" {
		t.Errorf("expected '=.proplist=name,type,running', got %q", s.Words[0])
	}
}

func TestEncodeSentence(t *testing.T) {
	s := NewSentence("/login")
	s.AddAttribute("name", "admin")
	s.AddAttribute("password", "secret")
	
	encoded := EncodeSentence(s)
	
	// Verify it starts with the command word
	if !bytes.Contains(encoded, []byte("/login")) {
		t.Error("encoded sentence should contain '/login'")
	}
	
	// Verify it contains the attributes
	if !bytes.Contains(encoded, []byte("=name=admin")) {
		t.Error("encoded sentence should contain '=name=admin'")
	}
	
	// Verify it ends with an empty word (0x00)
	if encoded[len(encoded)-1] != 0x00 {
		t.Error("encoded sentence should end with 0x00")
	}
}

func TestReplyTypes(t *testing.T) {
	tests := []struct {
		replyType string
		isDone    bool
		isTrap    bool
		isFatal   bool
		isData    bool
	}{
		{"!done", true, false, false, false},
		{"!trap", false, true, false, false},
		{"!fatal", false, false, true, false},
		{"!re", false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.replyType, func(t *testing.T) {
			reply := &Reply{Type: tt.replyType}
			
			if reply.IsDone() != tt.isDone {
				t.Errorf("IsDone() = %v, want %v", reply.IsDone(), tt.isDone)
			}
			if reply.IsTrap() != tt.isTrap {
				t.Errorf("IsTrap() = %v, want %v", reply.IsTrap(), tt.isTrap)
			}
			if reply.IsFatal() != tt.isFatal {
				t.Errorf("IsFatal() = %v, want %v", reply.IsFatal(), tt.isFatal)
			}
			if reply.IsData() != tt.isData {
				t.Errorf("IsData() = %v, want %v", reply.IsData(), tt.isData)
			}
		})
	}
}

func TestReplyGetMessage(t *testing.T) {
	reply := &Reply{
		Type: "!trap",
		Data: map[string]string{
			"message": "no such item",
		},
	}
	
	if reply.GetMessage() != "no such item" {
		t.Errorf("GetMessage() = %q, want 'no such item'", reply.GetMessage())
	}
	
	emptyReply := &Reply{
		Type: "!trap",
		Data: map[string]string{},
	}
	
	if emptyReply.GetMessage() != "" {
		t.Errorf("GetMessage() = %q, want empty string", emptyReply.GetMessage())
	}
}

func TestParseAttributeWord(t *testing.T) {
	tests := []struct {
		word      string
		wantKey   string
		wantValue string
		wantOk    bool
	}{
		{"=name=ether1", "name", "ether1", true},
		{"=address=192.168.1.1", "address", "192.168.1.1", true},
		{"=disabled=", "disabled", "", true},
		{"=comment=has=equals", "comment", "has=equals", true},
		{"name=value", "", "", false},
		{"", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			key, value, ok := ParseAttributeWord(tt.word)
			if ok != tt.wantOk {
				t.Errorf("ParseAttributeWord(%q) ok = %v, want %v", tt.word, ok, tt.wantOk)
				return
			}
			if key != tt.wantKey {
				t.Errorf("ParseAttributeWord(%q) key = %q, want %q", tt.word, key, tt.wantKey)
			}
			if value != tt.wantValue {
				t.Errorf("ParseAttributeWord(%q) value = %q, want %q", tt.word, value, tt.wantValue)
			}
		})
	}
}

func TestDecodeSentence(t *testing.T) {
	// Build a mock !re response
	var buf bytes.Buffer
	buf.Write(EncodeWord("!re"))
	buf.Write(EncodeWord("=name=ether1"))
	buf.Write(EncodeWord("=type=ether"))
	buf.Write(EncodeWord("=running=true"))
	buf.WriteByte(0x00) // Empty word to terminate

	reply, err := DecodeSentence(&buf)
	if err != nil {
		t.Fatalf("DecodeSentence failed: %v", err)
	}

	if reply.Type != "!re" {
		t.Errorf("expected Type '!re', got %q", reply.Type)
	}

	if reply.Data["name"] != "ether1" {
		t.Errorf("expected name='ether1', got %q", reply.Data["name"])
	}

	if reply.Data["type"] != "ether" {
		t.Errorf("expected type='ether', got %q", reply.Data["type"])
	}

	if reply.Data["running"] != "true" {
		t.Errorf("expected running='true', got %q", reply.Data["running"])
	}
}

func TestSentenceChaining(t *testing.T) {
	s := NewSentence("/interface/print").
		AddQuery("running", "true").
		AddQuery("type", "ether").
		AddProplist("name", "comment")
	
	if s.Word != "/interface/print" {
		t.Error("chaining should preserve command")
	}
	
	if len(s.Words) != 3 {
		t.Errorf("expected 3 words, got %d", len(s.Words))
	}

	// Check that query words are present
	found := 0
	for _, w := range s.Words {
		if strings.HasPrefix(w, "?") {
			found++
		}
	}
	if found != 2 {
		t.Errorf("expected 2 query words, got %d", found)
	}
}
