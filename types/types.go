package types

const (
	USER_ROLE_ADMIN = "admin"
)
const (
	USER_WORKSPACE_ROLE_EXECUTIVE = "executive"
	USER_WORKSPACE_ROLE_HEAD      = "head"
	USER_WORKSPACE_ROLE_DHEAD     = "dhead"
	USER_WORKSPACE_ROLE_ASSISTANT = "assistant"
	USER_WORKSPACE_ROLE_STAFF     = "staff"
)

const (
	USER_MANAGEMENT_LEVEL_EXECUTIVE = 5
	USER_MANAGEMENT_LEVEL_HEAD      = 4
	USER_MANAGEMENT_LEVEL_DHEAD     = 3
	USER_MANAGEMENT_LEVEL_ASSISTANT = 2
	USER_MANAGEMENT_LEVEL_STAFF     = 2
)

var MAPPING_ROLE_TO_MANAGEMENT_LEVEL map[string]int = map[string]int{
	USER_WORKSPACE_ROLE_EXECUTIVE: USER_MANAGEMENT_LEVEL_EXECUTIVE,
	USER_WORKSPACE_ROLE_HEAD:      USER_MANAGEMENT_LEVEL_HEAD,
	USER_WORKSPACE_ROLE_DHEAD:     USER_MANAGEMENT_LEVEL_DHEAD,
	USER_WORKSPACE_ROLE_ASSISTANT: USER_MANAGEMENT_LEVEL_ASSISTANT,
	USER_WORKSPACE_ROLE_STAFF:     USER_MANAGEMENT_LEVEL_STAFF,
}

const (
	TASK_STATUS_OPEN      = "open"
	TASK_STATUS_CLOSE     = "close"
	TASK_STATUS_CANCEL    = "cancel"
	TASK_STATUS_DOING     = "doing"
	TASK_STATUS_COMPLETED = "completed"
	TASK_STATUS_REVIEW    = "review"
)

const (
	DepartmentTechnical      = "DepartmentTechnical"
	DepartmentProductionPlan = "DepartmentProductionPlan"
	DepartmentQuality        = "DepartmentQuality"
	DepartmentMaterial       = "DepartmentMaterial"
)

type Admin struct {
	ID       string `json:"id" bson:"_id,omitempty"`
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
	Role     string `json:"role" bson:"role"`
}

type User struct {
	ID              string `json:"id" bson:"_id,omitempty"`
	Username        string `json:"username" bson:"username"`
	Password        string `json:"password" bson:"password"`
	FullName        string `json:"full_name" bson:"full_name"`
	ManagementLevel int    `json:"management_level" bson:"management_level"`
	WorkspaceRole   string `json:"workspace_role" bson:"workspace_role"`
	Workspace       string `json:"workspace" bson:"workspace"`
	CreateAt        int64  `json:"created_at" bson:"created_at"`
	UpdateAt        int64  `json:"updated_at" bson:"updated_at"`
}

type Workspace struct {
	ID   string `json:"id" bson:"_id,omitempty"`
	Name string `json:"name" bson:"name"`
}

type Task struct {
	ID          string `json:"id" bson:"_id,omitempty"`
	Title       string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
	Workspace   string `json:"workspace" bson:"workspace"`
	Creator     string `json:"creator" bson:"creator"`
	Deadline    int64  `json:"deadline" bson:"deadline"`
	Assignee    string `json:"assignee" bson:"assignee"`
	Status      string `json:"status" bson:"status"`
	Progress    int    `json:"progress" bson:"progress"`
	CreateAt    int64  `json:"created_at" bson:"created_at"`
	StartAt     int64  `json:"start_at" bson:"start_at"`
	UpdateAt    int64  `json:"updated_at" bson:"updated_at"`
}

type Report struct {
	ID         string `json:"id" bson:"_id,omitempty"`
	TaskID     string `json:"task_id" bson:"task_id"`
	Creator    string `json:"creator" bson:"creator"`
	Report     string `json:"report" bson:"report"`
	ReportFile string `json:"report_file" bson:"report_file"`
	Feedback   string `json:"feedback" bson:"feedback"`
	CreatedAt  int64  `json:"created_at" bson:"created_at"`
	UpdatedAt  int64  `json:"updated_at" bson:"updated_at"`
}

type TaskFilter struct {
	Title        string `json:"title" bson:"title"`
	Workspace    string `json:"workspace" bson:"workspace"`
	Creator      string `json:"creator" bson:"creator"`
	StartFrom    int64  `json:"start_from" bson:"start_from"`
	StartTo      int64  `json:"start_to" bson:"start_to"`
	DeadlineFrom int64  `json:"deadline_from" bson:"deadline_from"`
	DeadlineTo   int64  `json:"deadline_to" bson:"deadline_to"`
	ReportFrom   int64  `json:"report_from" bson:"report_from"`
	ReportTo     int64  `json:"report_to" bson:"report_to"`
	Assignee     string `json:"assignee" bson:"assignee"`
	Status       string `json:"status" bson:"status"`
}

type ReportFilter struct {
	TaskID      string `json:"task_id" bson:"task_id"`
	Creator     string `json:"creator" bson:"creator"`
	CreatedFrom int64  `json:"created_from" bson:"created_from"`
	CreatedTo   int64  `json:"created_to" bson:"created_to"`
	Workspace   string `json:"workspace" bson:"workspace"`
}

type UserFilter struct {
	Username  string `json:"username" bson:"username"`
	FullName  string `json:"full_name" bson:"full_name"`
	Workspace string `json:"workspace" bson:"workspace"`
}
