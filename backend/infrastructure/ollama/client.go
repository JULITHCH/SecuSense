package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/secusense/backend/config"
	"github.com/secusense/backend/internal/domain"
)

type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

func NewClient(cfg config.OllamaConfig) *Client {
	return &Client{
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format"`
}

type generateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// rawCourseContent is used for initial parsing with flexible question handling
type rawCourseContent struct {
	Title              string                   `json:"title"`
	Description        string                   `json:"description"`
	LearningObjectives []string                 `json:"learningObjectives"`
	Outline            []domain.CourseOutlineChapter `json:"outline"`
	VideoScript        string                   `json:"videoScript"`
	Questions          []json.RawMessage        `json:"questions"`
}

// rawQuestion for flexible question parsing
type rawQuestion struct {
	QuestionType string          `json:"questionType"`
	QuestionText string          `json:"questionText"`
	QuestionData json.RawMessage `json:"questionData"`
	Points       interface{}     `json:"points"` // Can be int or string
}

func (c *Client) GenerateCourseContent(ctx context.Context, req *domain.GenerateCourseRequest) (*domain.GeneratedCourseContent, error) {
	prompt := c.buildCoursePrompt(req)

	ollamaReq := generateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
		Format: "json",
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var ollamaResp generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Log raw response for debugging
	log.Printf("Ollama raw response length: %d bytes", len(ollamaResp.Response))

	// Clean and parse the response
	content, err := c.parseAndValidateContent(ollamaResp.Response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated content: %w", err)
	}

	return content, nil
}

func (c *Client) parseAndValidateContent(rawJSON string) (*domain.GeneratedCourseContent, error) {
	// Try to extract JSON if wrapped in markdown code blocks
	rawJSON = extractJSON(rawJSON)

	// First parse into raw structure
	var raw rawCourseContent
	if err := json.Unmarshal([]byte(rawJSON), &raw); err != nil {
		return nil, fmt.Errorf("JSON unmarshal failed: %w, raw: %s", err, truncate(rawJSON, 500))
	}

	// Build the final content
	content := &domain.GeneratedCourseContent{
		Title:              raw.Title,
		Description:        raw.Description,
		LearningObjectives: raw.LearningObjectives,
		Outline:            raw.Outline,
		VideoScript:        raw.VideoScript,
		Questions:          make([]domain.GeneratedQuestion, 0),
	}

	// Parse questions with flexible handling
	for i, qRaw := range raw.Questions {
		q, err := c.parseQuestion(qRaw)
		if err != nil {
			log.Printf("Warning: Failed to parse question %d: %v", i, err)
			continue
		}
		if q != nil {
			content.Questions = append(content.Questions, *q)
		}
	}

	log.Printf("Parsed content: title=%q, questions=%d", content.Title, len(content.Questions))

	// If no questions were parsed, generate default questions
	if len(content.Questions) == 0 {
		log.Printf("No questions parsed, generating defaults for topic: %s", content.Title)
		content.Questions = c.generateDefaultQuestions(content.Title)
	}

	return content, nil
}

func (c *Client) parseQuestion(rawQ json.RawMessage) (*domain.GeneratedQuestion, error) {
	var rq rawQuestion
	if err := json.Unmarshal(rawQ, &rq); err != nil {
		return nil, fmt.Errorf("unmarshal question: %w", err)
	}

	// Validate question type
	qType := normalizeQuestionType(rq.QuestionType)
	if qType == "" {
		return nil, fmt.Errorf("invalid question type: %s", rq.QuestionType)
	}

	// Parse points (handle string or int)
	points := 10
	switch v := rq.Points.(type) {
	case float64:
		points = int(v)
	case int:
		points = v
	case string:
		fmt.Sscanf(v, "%d", &points)
	}

	// Validate and fix questionData based on type
	questionData, err := c.validateQuestionData(qType, rq.QuestionData)
	if err != nil {
		return nil, fmt.Errorf("invalid question data for type %s: %w", qType, err)
	}

	return &domain.GeneratedQuestion{
		QuestionType: qType,
		QuestionText: rq.QuestionText,
		QuestionData: questionData,
		Points:       points,
	}, nil
}

func (c *Client) validateQuestionData(qType domain.QuestionType, data json.RawMessage) (json.RawMessage, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty question data")
	}

	switch qType {
	case domain.QuestionTypeMultipleChoice:
		var mc domain.MultipleChoiceData
		if err := json.Unmarshal(data, &mc); err != nil {
			return nil, err
		}
		if len(mc.Options) < 2 {
			return nil, fmt.Errorf("multiple choice needs at least 2 options")
		}
		if len(mc.CorrectIndices) == 0 {
			mc.CorrectIndices = []int{0} // Default to first option
		}
		return json.Marshal(mc)

	case domain.QuestionTypeDragDrop:
		var dd domain.DragDropData
		if err := json.Unmarshal(data, &dd); err != nil {
			return nil, err
		}
		if len(dd.Items) == 0 || len(dd.DropZones) == 0 {
			return nil, fmt.Errorf("drag drop needs items and drop zones")
		}
		if dd.CorrectMapping == nil {
			dd.CorrectMapping = make(map[string]string)
		}
		return json.Marshal(dd)

	case domain.QuestionTypeFillBlank:
		var fb domain.FillBlankData
		if err := json.Unmarshal(data, &fb); err != nil {
			return nil, err
		}
		if fb.Template == "" || len(fb.Blanks) == 0 {
			return nil, fmt.Errorf("fill blank needs template and blanks")
		}
		return json.Marshal(fb)

	case domain.QuestionTypeMatching:
		var m domain.MatchingData
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		if len(m.LeftItems) == 0 || len(m.RightItems) == 0 {
			return nil, fmt.Errorf("matching needs left and right items")
		}
		if m.CorrectPairs == nil {
			m.CorrectPairs = make(map[string]string)
		}
		return json.Marshal(m)

	case domain.QuestionTypeOrdering:
		var o domain.OrderingData
		if err := json.Unmarshal(data, &o); err != nil {
			return nil, err
		}
		if len(o.Items) == 0 {
			return nil, fmt.Errorf("ordering needs items")
		}
		if len(o.CorrectOrder) == 0 {
			// Generate sequential order
			o.CorrectOrder = make([]int, len(o.Items))
			for i := range o.Items {
				o.CorrectOrder[i] = i
			}
		}
		return json.Marshal(o)
	}

	return data, nil
}

func normalizeQuestionType(t string) domain.QuestionType {
	t = strings.ToLower(strings.TrimSpace(t))
	t = strings.ReplaceAll(t, "-", "_")
	t = strings.ReplaceAll(t, " ", "_")

	switch t {
	case "multiple_choice", "multiplechoice", "mc", "mcq":
		return domain.QuestionTypeMultipleChoice
	case "drag_drop", "dragdrop", "drag_and_drop", "draganddrop":
		return domain.QuestionTypeDragDrop
	case "fill_blank", "fillblank", "fill_in_blank", "fill_in_the_blank", "fillintheblank":
		return domain.QuestionTypeFillBlank
	case "matching", "match":
		return domain.QuestionTypeMatching
	case "ordering", "order", "sequence":
		return domain.QuestionTypeOrdering
	}
	return ""
}

func (c *Client) generateDefaultQuestions(topic string) []domain.GeneratedQuestion {
	questions := []domain.GeneratedQuestion{
		{
			QuestionType: domain.QuestionTypeMultipleChoice,
			QuestionText: fmt.Sprintf("What is the main purpose of %s?", topic),
			QuestionData: mustMarshal(domain.MultipleChoiceData{
				Options:        []string{"To improve security", "To reduce costs", "To increase speed", "To simplify processes"},
				CorrectIndices: []int{0},
				Explanation:    "Security improvement is the primary goal.",
			}),
			Points: 10,
		},
		{
			QuestionType: domain.QuestionTypeMultipleChoice,
			QuestionText: fmt.Sprintf("Which of the following is a best practice for %s?", topic),
			QuestionData: mustMarshal(domain.MultipleChoiceData{
				Options:        []string{"Follow industry standards", "Ignore guidelines", "Use shortcuts", "Skip verification"},
				CorrectIndices: []int{0},
				Explanation:    "Following industry standards ensures proper implementation.",
			}),
			Points: 10,
		},
		{
			QuestionType: domain.QuestionTypeMultipleChoice,
			QuestionText: "What should you do first when implementing security measures?",
			QuestionData: mustMarshal(domain.MultipleChoiceData{
				Options:        []string{"Assess current risks", "Buy new software", "Change all passwords", "Disable all access"},
				CorrectIndices: []int{0},
				Explanation:    "Risk assessment helps identify what needs protection.",
			}),
			Points: 10,
		},
		{
			QuestionType: domain.QuestionTypeMultipleChoice,
			QuestionText: "How often should security practices be reviewed?",
			QuestionData: mustMarshal(domain.MultipleChoiceData{
				Options:        []string{"Regularly and after incidents", "Only once a year", "Never", "Only when problems occur"},
				CorrectIndices: []int{0},
				Explanation:    "Regular reviews and post-incident analysis ensure continued protection.",
			}),
			Points: 10,
		},
		{
			QuestionType: domain.QuestionTypeFillBlank,
			QuestionText: "Complete the security statement:",
			QuestionData: mustMarshal(domain.FillBlankData{
				Template:    "A strong {{blank}} policy is essential for {{blank}} protection.",
				Blanks:      []string{"security", "data"},
				Explanation: "Security policies protect organizational data.",
			}),
			Points: 10,
		},
		{
			QuestionType: domain.QuestionTypeFillBlank,
			QuestionText: "Fill in the security terms:",
			QuestionData: mustMarshal(domain.FillBlankData{
				Template:    "{{blank}} testing helps identify {{blank}} before attackers do.",
				Blanks:      []string{"Penetration", "vulnerabilities"},
				Explanation: "Penetration testing proactively finds security weaknesses.",
			}),
			Points: 10,
		},
		{
			QuestionType: domain.QuestionTypeOrdering,
			QuestionText: "Arrange the incident response steps in the correct order:",
			QuestionData: mustMarshal(domain.OrderingData{
				Items:        []string{"Identify the threat", "Contain the damage", "Eradicate the cause", "Recover systems", "Document lessons"},
				CorrectOrder: []int{0, 1, 2, 3, 4},
				Explanation:  "Proper incident response follows: Identify, Contain, Eradicate, Recover, Document.",
			}),
			Points: 10,
		},
		{
			QuestionType: domain.QuestionTypeMatching,
			QuestionText: "Match the security terms with their descriptions:",
			QuestionData: mustMarshal(domain.MatchingData{
				LeftItems:    []string{"Encryption", "Authentication", "Authorization"},
				RightItems:   []string{"Verifying identity", "Scrambling data", "Granting permissions"},
				CorrectPairs: map[string]string{"Encryption": "Scrambling data", "Authentication": "Verifying identity", "Authorization": "Granting permissions"},
				Explanation:  "Understanding security terminology is fundamental.",
			}),
			Points: 10,
		},
		{
			QuestionType: domain.QuestionTypeDragDrop,
			QuestionText: "Categorize these items as either 'Good Practice' or 'Bad Practice':",
			QuestionData: mustMarshal(domain.DragDropData{
				Items:          []string{"Regular updates", "Sharing passwords", "Using MFA", "Ignoring alerts"},
				DropZones:      []string{"Good Practice", "Bad Practice"},
				CorrectMapping: map[string]string{"Regular updates": "Good Practice", "Sharing passwords": "Bad Practice", "Using MFA": "Good Practice", "Ignoring alerts": "Bad Practice"},
				Explanation:    "Security awareness includes knowing good from bad practices.",
			}),
			Points: 10,
		},
		{
			QuestionType: domain.QuestionTypeDragDrop,
			QuestionText: "Sort these actions by priority level:",
			QuestionData: mustMarshal(domain.DragDropData{
				Items:          []string{"Patch critical vulnerability", "Update documentation", "Review logs", "Plan training"},
				DropZones:      []string{"High Priority", "Medium Priority"},
				CorrectMapping: map[string]string{"Patch critical vulnerability": "High Priority", "Review logs": "High Priority", "Update documentation": "Medium Priority", "Plan training": "Medium Priority"},
				Explanation:    "Critical security issues take priority over administrative tasks.",
			}),
			Points: 10,
		},
	}
	return questions
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

func extractJSON(s string) string {
	// Remove markdown code blocks
	re := regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")
	matches := re.FindStringSubmatch(s)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try to find JSON object boundaries
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}

	return strings.TrimSpace(s)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (c *Client) buildCoursePrompt(req *domain.GenerateCourseRequest) string {
	duration := req.VideoDurationMin
	if duration == 0 {
		duration = 5
	}
	wordCount := duration * 150

	difficulty := req.DifficultyLevel
	if difficulty == "" {
		difficulty = "intermediate"
	}

	audience := req.TargetAudience
	if audience == "" {
		audience = "professionals"
	}

	return fmt.Sprintf(`Generate a complete training course about: %s

Target audience: %s
Difficulty level: %s
Video script should be approximately %d words (for a %d minute video).

You must respond with a valid JSON object in this exact format:
{
  "title": "Course title",
  "description": "2-3 sentence course description",
  "learningObjectives": ["objective 1", "objective 2", "objective 3"],
  "outline": [
    {
      "title": "Chapter 1 title",
      "description": "Brief description",
      "topics": ["topic 1", "topic 2"]
    }
  ],
  "videoScript": "Full video narration script...",
  "questions": [
    {
      "questionType": "multiple_choice",
      "questionText": "Question text?",
      "questionData": {
        "options": ["Option A", "Option B", "Option C", "Option D"],
        "correctIndices": [0],
        "explanation": "Explanation of correct answer"
      },
      "points": 10
    },
    {
      "questionType": "drag_drop",
      "questionText": "Match items to categories",
      "questionData": {
        "items": ["Item 1", "Item 2"],
        "dropZones": ["Zone A", "Zone B"],
        "correctMapping": {"Item 1": "Zone A", "Item 2": "Zone B"},
        "explanation": "Explanation"
      },
      "points": 10
    },
    {
      "questionType": "fill_blank",
      "questionText": "Fill in the blanks",
      "questionData": {
        "template": "The {{blank}} is used for {{blank}}.",
        "blanks": ["answer1", "answer2"],
        "explanation": "Explanation"
      },
      "points": 10
    },
    {
      "questionType": "matching",
      "questionText": "Match the terms",
      "questionData": {
        "leftItems": ["Term 1", "Term 2"],
        "rightItems": ["Definition 1", "Definition 2"],
        "correctPairs": {"Term 1": "Definition 1", "Term 2": "Definition 2"},
        "explanation": "Explanation"
      },
      "points": 10
    },
    {
      "questionType": "ordering",
      "questionText": "Put in correct order",
      "questionData": {
        "items": ["Step 1", "Step 2", "Step 3"],
        "correctOrder": [0, 1, 2],
        "explanation": "Explanation"
      },
      "points": 10
    }
  ]
}

Generate exactly 10 questions with this distribution:
- 4 multiple_choice questions
- 2 drag_drop questions
- 2 fill_blank questions
- 1 matching question
- 1 ordering question

Each question should be worth 10 points for a total of 100 points.
Ensure all JSON is properly formatted and valid.`, req.Topic, audience, difficulty, wordCount, duration)
}

func (c *Client) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama health check failed with status %d", resp.StatusCode)
	}

	return nil
}
