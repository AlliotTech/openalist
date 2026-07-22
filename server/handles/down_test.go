package handles

import "testing"

func TestShouldConvertMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		filter   bool
		accept   string
		want     bool
	}{
		{name: "browser preview", filename: "README.md", filter: true, accept: "text/html,application/xhtml+xml", want: true},
		{name: "text editor fetch", filename: "README.md", filter: true, accept: "*/*", want: false},
		{name: "markdown disabled", filename: "README.md", filter: false, accept: "text/html", want: false},
		{name: "non markdown", filename: "index.txt", filter: true, accept: "text/html", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldConvertMarkdown(tt.filename, tt.filter, tt.accept); got != tt.want {
				t.Fatalf("shouldConvertMarkdown() = %v, want %v", got, tt.want)
			}
		})
	}
}
