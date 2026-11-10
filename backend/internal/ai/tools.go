package ai

import (
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

const (
	REQUIRED_NUM_OF_RESPONSE_FEEDBACK = 3
)

type QuestionFeedback struct {
	Score    int // 0 or 1
	Feedback string
}

type QuizFeedbackRequest struct {
	Responses []ResponseFeedback `json:"responses"`
}

// this file will be where I keep all of the tools or actions that the ai can call

// availableTools simulates the tools/functions we're making available for
// the model.

var quizFeedbackTool = []llms.Tool{
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "generateQuizFeedback",
			Description: "Generates feedback for student, based on their anwsers. The 'responses' object contains an array of 3 items where, each item assesses a student's answer with a 'score' and 'feedback'. There must be an 3 objects {score: 0, feedback: 'the student received this score because...'} in the array.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"responses": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"score": map[string]any{
									"type":        "integer",
									"description": "a score of 0 or 1. 0 for no credit, and 1 for credit.",
								},
								"feedback": map[string]any{
									"type":        "string",
									"description": "feedback with an explanation for the score assigned to the student",
								},
							},
							"required": []string{"score", "feedback"},
						},
					},
				},
				"required": []string{"responses"},
			},
		},
	},
}

var questionQualityTool = []llms.Tool{
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "generateScoreAndFeedback",
			Description: "output the ai assesment in a structured format",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"score": map[string]any{
						"type":        "integer",
						"description": "a score of 0 or 1. 0 for no credit, and 1 for credit.",
					},
					"feedback": map[string]any{
						"type":        "string",
						"description": "feedback with an explanation for the score assigned to the student",
					},
				},
				"required": []string{"score", "feedback"},
			},
		},
	},
}

var quizQuestionTool = []llms.Tool{
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "generateQuizQuestions",
			Description: "ouput the generated quiz questions in the correct format",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"q1": map[string]any{
						"type":        "string",
						"description": "first quiz question",
					},
					"q2": map[string]any{
						"type":        "string",
						"description": "second quiz question",
					},
					"q3": map[string]any{
						"type":        "string",
						"description": "third quiz question",
					},
				},
				"required": []string{"q1", "q2", "q3"},
			},
		},
	},
}

func generateScoreAndFeedback(score int, feedback string) (QuestionFeedback, error) {

	if score > 1 || score < 0 {
		return QuestionFeedback{}, fmt.Errorf("invalid score, must be 0 or 1")
	}

	if feedback == "" {
		return QuestionFeedback{}, fmt.Errorf("feedback cannot be blank")
	}

	return QuestionFeedback{
		Score:    score,
		Feedback: feedback,
	}, nil
}

func generateQuizQuestions(q1, q2, q3 string) ([]QuizQuestion, error) {

	if len(q1) == 0 || len(q2) == 0 || len(q3) == 0 {
		return nil, fmt.Errorf("one of the qs was blank. doc1: %s, doc2: %s, doc3: %s", q1, q2, q3)
	}

	quizQuestions := []QuizQuestion{{strings.TrimSpace(q1)}, {strings.TrimSpace(q2)}, {strings.TrimSpace(q3)}}

	return quizQuestions, nil

}

func generateQuizFeedback(rf QuizFeedbackRequest) ([]ResponseFeedback, error) {

	if len(rf.Responses) != REQUIRED_NUM_OF_RESPONSE_FEEDBACK {
		return []ResponseFeedback{}, fmt.Errorf("generateQuizFeedback: not enough rf. Want %d, Got: %d. rf.Responses: %+v", REQUIRED_NUM_OF_RESPONSE_FEEDBACK, len(rf.Responses), rf.Responses)
	}

	for _, item := range rf.Responses {
		if item.Feedback == "" {
			return []ResponseFeedback{}, fmt.Errorf("generateQuizFeedback: missing feedback")
		}
	}

	return rf.Responses, nil

}
