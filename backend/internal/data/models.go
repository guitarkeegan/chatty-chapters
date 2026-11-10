package data

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	User               UserModel
	Book               BookModel
	Chapter            ChapterModel
	UserBook           UserBookModel
	UserActivity       UserActivityModel
	UserMessage        UserMessageModel
	UserProgress       UserProgressModel
	UserCourse         UserCourseModel
	AIMessage          AIMessageModel
	AIUserConversation AIUserConversationModel
	Transaction        TrasactionModel
	ChatEval           ChatEvalModel
	Quiz               QuizModel
	QuizQuestion       QuizQuestionModel
	QuizResponse       QuizResponseModel
	QuizResponseEval   QuizResponseEvalModel
	Course             CourseModel
	CourseBookChapter  CourseBookChapterModel
}

func NewModels(db *pgxpool.Pool) Models {
	return Models{
		User:               UserModel{DB: db},
		Book:               BookModel{DB: db},
		Chapter:            ChapterModel{DB: db},
		UserBook:           UserBookModel{DB: db},
		UserActivity:       UserActivityModel{DB: db},
		UserMessage:        UserMessageModel{DB: db},
		UserProgress:       UserProgressModel{DB: db},
		UserCourse:         UserCourseModel{DB: db},
		AIMessage:          AIMessageModel{DB: db},
		AIUserConversation: AIUserConversationModel{DB: db},
		Transaction:        TrasactionModel{DB: db},
		ChatEval:           ChatEvalModel{DB: db},
		Quiz:               QuizModel{DB: db},
		QuizQuestion:       QuizQuestionModel{DB: db},
		QuizResponse:       QuizResponseModel{DB: db},
		QuizResponseEval:   QuizResponseEvalModel{DB: db},
		Course:             CourseModel{DB: db},
		CourseBookChapter:  CourseBookChapterModel{DB: db},
	}
}
