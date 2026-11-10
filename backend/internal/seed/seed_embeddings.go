package seed

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type TextBookMetadata struct {
	Title   string
	Authors []string
	Edition int
	Chapter int
}

func SeedBookEmbeddings(db *pgxpool.Pool) error {
	err := collectRawTextFiles(db)
	if err != nil {
		return err
	}
	log.Println("🌱 embeddings seeded successfully 🌱")
	return nil
}

func collectRawTextFiles(db *pgxpool.Pool) error {
	fmt.Println("collectRawTextFiles")
	err := godotenv.Load()
	if err != nil {
		return seedError("error on env load", err.Error())
	}

	dirPath := "books"

	fmt.Println("WalkDir")
	err = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".txt" {
			filename := d.Name()
			textBookMeta, err := parseFileString(filename)
			if err != nil {
				return err
			}
			chunks, err := splitText(path)
			if err != nil {
				return err
			}
			if len(chunks) == 0 {
				return errors.New("chunks are empty, expected some text")
			}

			if err = insertChapterEmbeddings(chunks, textBookMeta, db); err != nil {
				return err
			}
		}
		fmt.Println("  at return")
		return nil
	})

	if err != nil {
		log.Fatalf("error walking the path %q: %v", dirPath, err)
		return seedError(fmt.Sprintf("error walking the path %q:", dirPath), err.Error())
	}

	fmt.Println("at outer return")
	return nil
}

func parseFileString(f string) (TextBookMetadata, error) {

	fmt.Println("parseFileString")

	titleCaser := cases.Title(language.English)

	// title, authors, edition, chapter
	parsedString := strings.Split(f, "-")
	tbm := TextBookMetadata{}

	tbm.Title = titleCaser.String(strings.ReplaceAll(parsedString[0], "_", " "))

	authorsSlice := strings.Split(parsedString[1], "_")
	var authorsFullName []string
	for i := 0; i < len(authorsSlice)-1; i += 2 {
		authorsFullName = append(authorsFullName, titleCaser.String(authorsSlice[i]+" "+authorsSlice[i+1]))
	}
	tbm.Authors = authorsFullName

	fmt.Println("Parse the editions")
	// Parse the edition
	edition, err := strconv.Atoi(parsedString[2])
	if err != nil {
		return TextBookMetadata{}, err
	}
	tbm.Edition = edition
	// Parse the chapter
	chapter, err := strconv.Atoi(strings.ReplaceAll(parsedString[3], ".txt", ""))
	if err != nil {
		return TextBookMetadata{}, seedError("chapter conversion broke", err.Error())
	}
	tbm.Chapter = chapter

	fmt.Println("at return")
	return tbm, nil
}

func insertChapterEmbeddings(chunks []string, tb TextBookMetadata, pool *pgxpool.Pool) error {

	fmt.Println("insertChapterEmbeddings")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("get model for embeddings")
	opts := []openai.Option{
		openai.WithToken(os.Getenv("OPENAI")),
		openai.WithModel("gpt-4o-mini"),
		openai.WithEmbeddingModel("text-embedding-3-large"),
	}

	llm, err := openai.New(opts...)
	if err != nil {
		return seedError("fail to connect with openai", err.Error())
	}

	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return seedError("failed to make new embedder", err.Error())
	}

	fmt.Println("create the vector store")
	// Create a new pgvector store.
	ctx := context.Background()
	store, err := pgvector.New(
		ctx,
		pgvector.WithEmbedder(e),
		pgvector.WithConn(pool),
		pgvector.WithCollectionTableName("documents"),
		pgvector.WithEmbeddingTableName("embedding_table"),
		pgvector.WithCollectionName("my_collection"),
	)
	if err != nil {
		return seedError("failed to create new store", err.Error())
	}

	var docs []schema.Document
	fmt.Println("adding metadata")

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

	// Add documents to the pgvector store.
	// returns a list of ids that were added
	_, err = store.AddDocuments(context.Background(), docs)

	if err != nil {
		return seedError("failed to add docs to store", err.Error())
	}

	fmt.Println("at return")
	return nil
}

// Get text from file and split into llm optimized format
func splitText(filepath string) ([]string, error) {

	fmt.Println("splitText")
	file, err := os.Open(filepath)

	if err != nil {
		return []string{}, seedError("unable to read text file", err.Error())
	}
	defer file.Close()

	var buffer bytes.Buffer

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		buffer.WriteString(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return []string{}, seedError("Error reading file", err.Error())
	}

	fmt.Println("NewRecursiveCharacter()")
	ts := textsplitter.NewRecursiveCharacter()
	chunks, err := ts.SplitText(buffer.String())
	if err != nil {
		return []string{}, seedError("could not split text", err.Error())
	}

	fmt.Println("at return")
	return chunks, nil
}
