package types

type Response struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PaginatedData holds pagination metadata and the actual items
type PaginatedData struct {
	Total int64       `json:"total"`
	Limit int64       `json:"limit"`
	Page  int64       `json:"page"`
	Items interface{} `json:"items"`
}

// PaginatedResponse for paginated API responses
type PaginatedResponse struct {
	Status  bool          `json:"status"`
	Message string        `json:"message"`
	Data    PaginatedData `json:"data"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TaskResponse struct {
	ID          string            `json:"id" bson:"_id,omitempty"`
	Title       string            `json:"title" bson:"title"`
	Description string            `json:"description" bson:"description"`
	Workspace   string            `json:"workspace" bson:"workspace"`
	Creator     string            `json:"creator" bson:"creator"`
	Deadline    int64             `json:"deadline" bson:"deadline"`
	Assignee    string            `json:"assignee" bson:"assignee"`
	Status      string            `json:"status" bson:"status"`
	CreateAt    int64             `json:"created_at" bson:"created_at"`
	UpdateAt    int64             `json:"updated_at" bson:"updated_at"`
	Reports     []*ReportResponse `json:"reports" bson:"reports"`
}

type ReportResponse struct {
	ID       string `json:"id" bson:"_id,omitempty"`
	Creator  string `json:"creator" bson:"creator"`
	Report   string `json:"report" bson:"report"`
	Feedback string `json:"feedback" bson:"feedback"`
}
