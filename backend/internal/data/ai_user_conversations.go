package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatConversation struct {
	UserCreatedAt time.Time `json:"user_created_at"`
	AICreatedAt   time.Time `json:"ai_created_at"`
	UserMessage   string    `json:"user_message"`
	AIMessage     string    `json:"ai_message"`
}

type AIUserConversation struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	BookID        int64     `json:"book_id"`
	ChapterID     int64     `json:"chapter_id"`
	UserMessageID int64     `json:"user_message_id"`
	AiMessageID   int64     `json:"ai_message_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type AIUserConversationModel struct {
	DB *pgxpool.Pool
}

func (aiuc *AIUserConversationModel) Insert(ctx context.Context, userID, bookID, chapterID, userMessageID, aiMessageID int64, tx pgx.Tx) (*AIUserConversation, error) {

	stmt := `
	INSERT INTO ai_user_conversations (user_id, book_id, chapter_id, user_message_id, ai_message_id)
	VALUES ( $1, $2, $3, $4, $5 )
	RETURNING id, user_id, book_id, chapter_id, user_message_id, ai_message_id, created_at
	`

	aiUserConversation := &AIUserConversation{}

	err := tx.QueryRow(ctx, stmt, userID, bookID, chapterID, userMessageID, aiMessageID).Scan(&aiUserConversation.ID, &aiUserConversation.UserID, &aiUserConversation.BookID, &aiUserConversation.ChapterID, &aiUserConversation.UserMessageID, &aiUserConversation.AiMessageID, &aiUserConversation.CreatedAt)
	if err != nil {
		return nil, err
	}

	return aiUserConversation, nil
}

func (aiuc *AIUserConversationModel) GetChapterConversation(ctx context.Context, userID, bookID, chapterID int64) ([]ChatConversation, error) {

	// TODO: really think about this one...
	stmt := `
	SELECT
        um.message AS user_message,
        um.created_at AS user_message_time,
        am.message AS ai_message,
        am.created_at AS ai_message_time
	FROM
	ai_user_conversations auc
	INNER JOIN
	user_messages um ON auc.user_message_id = um.id
	INNER JOIN
	ai_messages am ON auc.ai_message_id = am.id
	INNER JOIN
	chapters ch ON auc.chapter_id = ch.id
	INNER JOIN
	books b ON ch.book_id = b.id
	WHERE
	auc.user_id = $1 AND
	ch.book_id = $2 AND
	auc.chapter_id = $3
	ORDER BY
	um.created_at, am.created_at;

	`

	rows, err := aiuc.DB.Query(context.Background(), stmt, userID, bookID, chapterID)
	if err != nil {
		return nil, fmt.Errorf("GetChapterConversion: %w", err)
	}
	defer rows.Close()

	var chatConversations []ChatConversation
	for rows.Next() {

		var chatConversation ChatConversation
		err := rows.Scan(&chatConversation.UserMessage, &chatConversation.UserCreatedAt, &chatConversation.AIMessage, &chatConversation.AICreatedAt)
		if err != nil {
			return nil, err
		}

		chatConversations = append(chatConversations, chatConversation)

	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(chatConversations) == 0 {
		return nil, nil
	}

	return chatConversations, nil

}

func (aiuc *AIUserConversationModel) GetConversationHistory(ctx context.Context, userID, bookID, chapterID int64) (*[]AIUserConversation, error) {

	stmt := `
	SELECT id, user_id, book_id, chapter_id, user_message_id, ai_message_id, created_at
	FROM ai_user_conversations
	WHERE user_id = $1 AND book_id = $2 AND chapter_id = $3
	ORDER BY created_at DESC
	`
	rows, err := aiuc.DB.Query(ctx, stmt, userID, bookID, chapterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []AIUserConversation
	for rows.Next() {
		var conversation AIUserConversation
		err := rows.Scan(&conversation.ID, &conversation.UserID, &conversation.BookID, &conversation.ChapterID, &conversation.UserMessageID, &conversation.AiMessageID, &conversation.CreatedAt)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, conversation)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(conversations) == 0 {
		return &[]AIUserConversation{}, nil
	}

	return &conversations, nil
}

func InsertAll(um *UserMessageModel, am *AIMessageModel, aiuc *AIUserConversationModel, userID, bookID, chapterID int64, userMessage, aiMessageContent string) (*AIUserConversation, error) {

	dbg("InsertAll")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbg(" start tx")
	// Start a transaction
	tx, err := um.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	dbg(" userMsg")
	dbg(fmt.Sprintf("userID: %d, bookID: %d, chapterID: %d, userMessage: %s", userID, bookID, chapterID, userMessage))
	// Insert into user_messages
	userMsg, err := um.Insert(ctx, userID, bookID, chapterID, userMessage, tx)
	if err != nil {
		return nil, err
	}

	dbg(" aiMsG")
	// Insert into ai_messages
	aiMsg, err := am.Insert(ctx, aiMessageContent, tx)
	if err != nil {
		return nil, err
	}

	dbg(" aiUserConv")
	// Insert into ai_user_conversations
	aiUserConv, err := aiuc.Insert(ctx, userID, bookID, chapterID, userMsg.ID, aiMsg.ID, tx)
	if err != nil {
		return nil, err
	}

	dbg(" commit")
	// Commit the transaction
	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}
	dbg(" at return")

	return aiUserConv, nil
}
