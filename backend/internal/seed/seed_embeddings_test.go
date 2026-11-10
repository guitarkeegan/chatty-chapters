package seed

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseFileString(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		want    TextBookMetadata
		wantErr bool
	}{
		{
			name: "single author",
			file: "alice_in_wonderland-lewis_carol-3-12.txt",
			want: TextBookMetadata{Title: "Alice In Wonderland", Authors: []string{"Lewis Carol"}, Edition: 3, Chapter: 12},
		},
		{
			name: "multiple authors",
			file: "enjoyment_of_music-kristine_forney_joseph_machlis-14-01.txt",
			want: TextBookMetadata{Title: "Enjoyment Of Music", Authors: []string{"Kristine Forney", "Joseph Machlis"}, Edition: 14, Chapter: 1},
		},
		{name: "missing chapter", file: "alice_in_wonderland-lewis_carol-3.txt", wantErr: true},
		{name: "non numeric edition", file: "alice_in_wonderland-lewis_carol-x-1.txt", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFileString(tt.file)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("want error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("want %+v, got %+v", tt.want, got)
			}
		})
	}
}

// every seed book's chapters should have a matching text file in books/
func TestSeedBooksMatchTextFiles(t *testing.T) {
	for _, sb := range seedBooks() {
		dir := strings.ToLower(strings.ReplaceAll(sb.book.Title, " ", "_"))
		files, err := filepath.Glob(filepath.Join("..", "..", booksDir, dir, "*.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != len(sb.chapters) {
			t.Errorf("%s: %d chapters but %d text files", sb.book.Title, len(sb.chapters), len(files))
		}
		for _, f := range files {
			meta, err := parseFileString(filepath.Base(f))
			if err != nil {
				t.Errorf("%s: %v", f, err)
				continue
			}
			if meta.Edition != sb.book.Edition {
				t.Errorf("%s: edition %d, want %d", f, meta.Edition, sb.book.Edition)
			}
		}
	}
}

func TestSplitText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chapter.txt")
	text := "the first line\nthe second line\n\nA new paragraph."
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}

	chunks, err := splitText(path)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(chunks, " ")
	// lines must not be glued together
	if strings.Contains(joined, "linethe") {
		t.Errorf("lines were joined without whitespace: %q", joined)
	}
}
