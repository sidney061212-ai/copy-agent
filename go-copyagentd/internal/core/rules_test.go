package core

import "testing"

func TestExtractCopyText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "chinese colon", in: "复制：你好", want: "你好"},
		{name: "ascii colon", in: "复制: 你好", want: "你好"},
		{name: "whitespace", in: "复制 你好", want: "你好"},
		{name: "english whitespace", in: "copy hello", want: "hello"},
		{name: "english colon", in: "copy: hello", want: "hello"},
		{name: "mention prefix", in: "@bot 复制 你好", want: "你好"},
		{name: "plain text", in: "你好", want: "你好"},
		{name: "bare command", in: "复制", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractCopyText(tt.in)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidText(t *testing.T) {
	if ValidText("   ") {
		t.Fatal("blank text should be invalid")
	}
}
