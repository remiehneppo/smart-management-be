package types

import "errors"

var (
	ErrInvalidUser              = errors.New("invalid user")
	ErrInvalidTask              = errors.New("invalid task")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrTaskNotCreatorOrAssignee = errors.New("task not creator or assignee")
)

var (
	ErrQuestAssignNotWorkspaceMember = errors.New("quest assign not workspace member")
	ErrQuestAssignNoPermission       = errors.New("quest assign no permission")
	ErrTaskNotInWorkspace            = errors.New("task not in workspace")
	ErrTaskNotCreator                = errors.New("task not creator")
	ErrInvalidReport                 = errors.New("invalid report")
	ErrInvalidCount                  = errors.New("invalid count")
	ErrTaskNotAssignee               = errors.New("task not assignee")
	ErrInvalidProgress               = errors.New("invalid progress")
	ErrReportNotCreator              = errors.New("report not creator")
	ErrUserNotFound                  = errors.New("user not found")
)

var (
	ErrUnsupportedFileType = errors.New("unsupported file type")
	ErrFileTooLarge        = errors.New("file too large")
)

var (
	ErrFailedExtractTextFromPDF = errors.New("failed to extract text from PDF")
)
