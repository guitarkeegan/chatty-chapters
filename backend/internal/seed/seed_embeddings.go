package seed

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const booksDir = "books"

type TextBookMetadata struct {
	Title   string
	Authors []string
	Edition int
	Chapter int
}

// SeedBookEmbeddings splits every chapter file in books/ into chunks, embeds
// them with OpenAI, and stores them in pgvector with the book title, edition,
// authors, and chapter as metadata.
func SeedBookEmbeddings(db *pgxpool.Pool, openaiKey string) error {
	ctx := context.Background()

	store, err := newVectorStore(ctx, db, openaiKey)
	if err != nil {
		return err
	}

	err = filepath.WalkDir(booksDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".txt" {
			return nil
		}

		meta, err := parseFileString(d.Name())
		if err != nil {
			return fmt.Errorf("parsing file name %q: %w", d.Name(), err)
		}

		chunks, err := splitText(path)
		if err != nil {
			return err
		}
		if len(chunks) == 0 {
			return fmt.Errorf("%s: no text found", path)
		}

		if err = insertChapterEmbeddings(ctx, store, chunks, meta); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}

		log.Printf("embedded %s chapter %d (%d chunks)", meta.Title, meta.Chapter, len(chunks))
		return nil
	})
	if err != nil {
		return seedError(fmt.Sprintf("error walking the path %q", booksDir), err.Error())
	}

	log.Println("🌱 embeddings seeded successfully 🌱")
	return nil
}

func newVectorStore(ctx context.Context, db *pgxpool.Pool, openaiKey string) (pgvector.Store, error) {
	llm, err := openai.New(
		openai.WithToken(openaiKey),
		openai.WithModel("gpt-4o-mini"),
		openai.WithEmbeddingModel("text-embedding-3-large"),
	)
	if err != nil {
		return pgvector.Store{}, seedError("fail to connect with openai", err.Error())
	}

	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return pgvector.Store{}, seedError("failed to make new embedder", err.Error())
	}

	store, err := pgvector.New(
		ctx,
		pgvector.WithEmbedder(e),
		pgvector.WithConn(db),
		pgvector.WithCollectionTableName("documents"),
		pgvector.WithEmbeddingTableName("embedding_table"),
		pgvector.WithCollectionName("my_collection"),
	)
	if err != nil {
		return pgvector.Store{}, seedError("failed to create new store", err.Error())
	}

	return store, nil
}

// parseFileString parses a file name like
// "alice_in_wonderland-lewis_carol-3-1.txt" into its title, authors, edition,
// and chapter.
func parseFileString(f string) (TextBookMetadata, error) {
	titleCaser := cases.Title(language.English)

	parts := strings.Split(strings.TrimSuffix(f, ".txt"), "-")
	if len(parts) != 4 {
		return TextBookMetadata{}, fmt.Errorf("expected title-authors-edition-chapter.txt, got %q", f)
	}

	tbm := TextBookMetadata{
		Title: titleCaser.String(strings.ReplaceAll(parts[0], "_", " ")),
	}

	// authors are first_last pairs joined by underscores
	names := strings.Split(parts[1], "_")
	for i := 0; i < len(names)-1; i += 2 {
		tbm.Authors = append(tbm.Authors, titleCaser.String(names[i]+" "+names[i+1]))
	}

	edition, err := strconv.Atoi(parts[2])
	if err != nil {
		return TextBookMetadata{}, fmt.Errorf("edition: %w", err)
	}
	tbm.Edition = edition

	chapter, err := strconv.Atoi(parts[3])
	if err != nil {
		return TextBookMetadata{}, fmt.Errorf("chapter: %w", err)
	}
	tbm.Chapter = chapter

	return tbm, nil
}

func insertChapterEmbeddings(ctx context.Context, store pgvector.Store, chunks []string, tb TextBookMetadata) error {
	docs := make([]schema.Document, 0, len(chunks))
	for _, chunk := range chunks {
		docs = append(docs, schema.Document{
			PageContent: chunk,
			// files will seperate the chapters
			Metadata: map[string]any{
				"title":   tb.Title,
				"edition": tb.Edition,
				"author":  tb.Authors,
				"chapter": tb.Chapter,
			},
		})
	}

	if _, err := store.AddDocuments(ctx, docs); err != nil {
		return seedError("failed to add docs to store", err.Error())
	}

	return nil
}

// splitText reads a chapter file and splits it into llm sized chunks
func splitText(path string) ([]string, error) {
	text, err := os.ReadFile(path)
	if err != nil {
		return nil, seedError("unable to read text file", err.Error())
	}

	chunks, err := textsplitter.NewRecursiveCharacter().SplitText(string(text))
	if err != nil {
		return nil, seedError("could not split text", err.Error())
	}

	return chunks, nil
}
