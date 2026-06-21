package client

// ---- Projects ----

type Project struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Status            string   `json:"status"`
	Organization      string   `json:"organization"`
	BillingAccountID  string   `json:"billingAccountId"`
	ClassificationIDs []string `json:"classificationIds"`
	Tags              []Tag    `json:"tags"`
}

type ProjectDetail struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	Status            string          `json:"status"`
	Organization      string          `json:"organization"`
	BillingAccountID  string          `json:"billingAccountId"`
	ClassificationIDs []string        `json:"classificationIds"`
	Tags              []Tag           `json:"tags"`
	Contacts          []Contact       `json:"contacts"`
	Settings          ProjectSettings `json:"settings"`
	Account           *CloudAccount   `json:"account,omitempty"`
	StatusMessage     string          `json:"statusMessage"`
	CreatedBy         string          `json:"createdBy"`
	CreatedAt         string          `json:"createdAt"`
	UpdatedAt         string          `json:"updatedAt"`
	DeployedAt        string          `json:"deployedAt"`
}

type ProjectInput struct {
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	BillingAccountID  string          `json:"billingAccountId"`
	Settings          ProjectSettings `json:"settings"`
	Contacts          []Contact       `json:"contacts"`
	Account           *CloudAccount   `json:"account,omitempty"`
	ClassificationIDs []string        `json:"classificationIds,omitempty"`
	Tags              []Tag           `json:"tags,omitempty"`
}

type ProjectSettings struct {
	BudgetAmount                 int      `json:"budgetAmount"`
	BudgetPeriod                 string   `json:"budgetPeriod"`
	EnableBackup                 bool     `json:"enableBackup"`
	EnableSftp                   bool     `json:"enableSftp"`
	RetentionPolicyDays          int      `json:"retentionPolicyDays"`
	TemporaryStorageLifetimeDays int      `json:"temporaryStorageLifetimeDays"`
	ServiceConnections           []string `json:"serviceConnections,omitempty"`
	KmsArn                       string   `json:"kmsArn,omitempty"`
	VpcID                        string   `json:"vpcId,omitempty"`
	BatchSubnets                 []string `json:"batchSubnets,omitempty"`
	WorkspaceSubnets             []string `json:"workspaceSubnets,omitempty"`
	MaxSpotVCPU                  int      `json:"maxSpotVCPU,omitempty"`
	MaxFPGAVCPU                  int      `json:"maxFPGAVCPU,omitempty"`
	MaxGPUVCPU                   int      `json:"maxGPUVCPU,omitempty"`
	EnableDragen                 bool     `json:"enableDragen,omitempty"`
	DragenAmi                    string   `json:"dragenAmi,omitempty"`
	MaxWorkspacesVCPU            int      `json:"maxWorkspacesVCPU,omitempty"`
	MaxWorkspacesGPUVCPU         int      `json:"maxWorkspacesGPUVCPU,omitempty"`
	MaxWorkspacesPerUser         int      `json:"maxWorkspacesPerUser,omitempty"`
	EnableAdvancedGpuConfig      bool     `json:"enableAdvancedGpuConfig,omitempty"`
	EnableCustomWorkspaceRoles   bool     `json:"enableCustomWorkspaceRoles,omitempty"`
	MaxSharedFilesystems         int      `json:"maxSharedFilesystems,omitempty"`
	IsDiscoverable               bool     `json:"isDiscoverable,omitempty"`
	IsShareable                  bool     `json:"isShareable,omitempty"`
	IsAiEnabled                  bool     `json:"isAiEnabled,omitempty"`

	// Computed — set by Cirro, not sent in create/update requests.
	HasPipelinesEnabled         bool `json:"hasPipelinesEnabled,omitempty"`
	HasWorkspacesEnabled        bool `json:"hasWorkspacesEnabled,omitempty"`
	HasSharedFilesystemsEnabled bool `json:"hasSharedFilesystemsEnabled,omitempty"`
}

type Contact struct {
	Name         string `json:"name"`
	Organization string `json:"organization"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
}

type CloudAccount struct {
	AccountID   string `json:"accountId,omitempty"`
	AccountName string `json:"accountName,omitempty"`
	RegionName  string `json:"regionName,omitempty"`
	AccountType string `json:"accountType"`
}

type Tag struct {
	Value string `json:"value"`
}

type CreateResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// ---- Project Members ----

type ProjectUser struct {
	Name         string `json:"name"`
	Username     string `json:"username"`
	Organization string `json:"organization"`
	Department   string `json:"department"`
	Email        string `json:"email"`
	JobTitle     string `json:"jobTitle"`
	Role         string `json:"role"`
}

type SetUserProjectRoleRequest struct {
	Username             string `json:"username"`
	Role                 string `json:"role"`
	SuppressNotification bool   `json:"suppressNotification"`
}

// ---- Users ----

type UserDetail struct {
	Username           string                  `json:"username"`
	Name               string                  `json:"name"`
	Email              string                  `json:"email"`
	Organization       string                  `json:"organization"`
	Phone              string                  `json:"phone"`
	OrcidID            string                  `json:"orcidId"`
	JobTitle           string                  `json:"jobTitle"`
	Department         string                  `json:"department"`
	InvitedBy          string                  `json:"invitedBy"`
	GlobalRoles        []string                `json:"globalRoles"`
	ProjectAssignments []UserProjectAssignment `json:"projectAssignments"`
}

type UserProjectAssignment struct {
	ProjectID string `json:"projectId"`
	Role      string `json:"role"`
	CreatedBy string `json:"createdBy"`
	CreatedAt string `json:"createdAt"`
}

type UserDto struct {
	Username     string `json:"username"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Organization string `json:"organization"`
}

type PaginatedUsersResponse struct {
	Items     []UserDto `json:"items"`
	NextToken string    `json:"nextToken"`
}

type InviteUserRequest struct {
	Name         string `json:"name"`
	Organization string `json:"organization"`
	Email        string `json:"email"`
}

type InviteUserResponse struct {
	Message string `json:"message"`
}

type UpdateUserRequest struct {
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	Phone        string   `json:"phone,omitempty"`
	Department   string   `json:"department,omitempty"`
	JobTitle     string   `json:"jobTitle,omitempty"`
	Organization string   `json:"organization,omitempty"`
	GlobalRoles  []string `json:"globalRoles,omitempty"`
}

// ---- Billing Accounts ----

type BillingAccount struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	Organization        string    `json:"organization"`
	Contacts            []Contact `json:"contacts"`
	CustomerType        string    `json:"customerType"`
	BillingMethod       string    `json:"billingMethod"`
	PrimaryBudgetNumber string    `json:"primaryBudgetNumber"`
	Owner               string    `json:"owner"`
	SharedWith          []string  `json:"sharedWith"`
	IsArchived          bool      `json:"isArchived"`
}

type BillingAccountRequest struct {
	Name                string    `json:"name"`
	Contacts            []Contact `json:"contacts"`
	CustomerType        string    `json:"customerType"`
	BillingMethod       string    `json:"billingMethod"`
	PrimaryBudgetNumber string    `json:"primaryBudgetNumber"`
	Owner               string    `json:"owner"`
	SharedWith          []string  `json:"sharedWith"`
}

// ---- Agents ----

type AgentDetail struct {
	ID                       string             `json:"id"`
	Name                     string             `json:"name"`
	AgentRoleArn             string             `json:"agentRoleArn"`
	Status                   string             `json:"status"`
	Registration             *AgentRegistration `json:"registration,omitempty"`
	Tags                     map[string]string  `json:"tags,omitempty"`
	EnvironmentConfiguration map[string]string  `json:"environmentConfiguration,omitempty"`
	CreatedBy                string             `json:"createdBy"`
	CreatedAt                string             `json:"createdAt"`
	UpdatedAt                string             `json:"updatedAt"`
}

type AgentInput struct {
	Name                     string            `json:"name"`
	AgentRoleArn             string            `json:"agentRoleArn"`
	Tags                     map[string]string `json:"tags,omitempty"`
	EnvironmentConfiguration map[string]string `json:"environmentConfiguration,omitempty"`
}

type AgentRegistration struct {
	LocalIP      string `json:"localIp"`
	RemoteIP     string `json:"remoteIp"`
	AgentVersion string `json:"agentVersion"`
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
}

// ---- Classifications ----

type GovernanceClassification struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	RequirementIDs []string `json:"requirementIds"`
	CreatedBy      string   `json:"createdBy"`
	CreatedAt      string   `json:"createdAt"`
	UpdatedAt      string   `json:"updatedAt"`
}

type ClassificationInput struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	RequirementIDs []string `json:"requirementIds"`
}

// ---- Pipelines (Processes) ----

type PipelineCode struct {
	RepositoryPath  string `json:"repositoryPath"`
	Version         string `json:"version"`
	RepositoryType  string `json:"repositoryType"`
	EntryPoint      string `json:"entryPoint"`
	ExecutorVersion string `json:"executorVersion,omitempty"`
}

type CustomPipelineSettings struct {
	Repository     string `json:"repository"`
	Branch         string `json:"branch,omitempty"`
	Folder         string `json:"folder,omitempty"`
	RepositoryType string `json:"repositoryType,omitempty"`
	LastSync       string `json:"lastSync,omitempty"`
	SyncStatus     string `json:"syncStatus,omitempty"`
	CommitHash     string `json:"commitHash,omitempty"`
	IsAuthorized   bool   `json:"isAuthorized,omitempty"`
}

type ProcessInput struct {
	ID                   string                  `json:"id"`
	Name                 string                  `json:"name"`
	Description          string                  `json:"description"`
	Executor             string                  `json:"executor"`
	ChildProcessIDs      []string                `json:"childProcessIds"`
	ParentProcessIDs     []string                `json:"parentProcessIds"`
	LinkedProjectIDs     []string                `json:"linkedProjectIds"`
	DataType             string                  `json:"dataType,omitempty"`
	Category             string                  `json:"category,omitempty"`
	DocumentationURL     string                  `json:"documentationUrl,omitempty"`
	FileRequirementsMsg  string                  `json:"fileRequirementsMessage,omitempty"`
	IsTenantWide         bool                    `json:"isTenantWide,omitempty"`
	AllowMultipleSources bool                    `json:"allowMultipleSources,omitempty"`
	UsesSampleSheet      bool                    `json:"usesSampleSheet,omitempty"`
	PipelineCode         *PipelineCode           `json:"pipelineCode,omitempty"`
	CustomSettings       *CustomPipelineSettings `json:"customSettings,omitempty"`
}

type ProcessDetail struct {
	ID                   string                  `json:"id"`
	Name                 string                  `json:"name"`
	Description          string                  `json:"description"`
	Executor             string                  `json:"executor"`
	DataType             string                  `json:"dataType"`
	Category             string                  `json:"category"`
	PipelineType         string                  `json:"pipelineType"`
	DocumentationURL     string                  `json:"documentationUrl"`
	FileRequirementsMsg  string                  `json:"fileRequirementsMessage"`
	ChildProcessIDs      []string                `json:"childProcessIds"`
	ParentProcessIDs     []string                `json:"parentProcessIds"`
	LinkedProjectIDs     []string                `json:"linkedProjectIds"`
	Owner                string                  `json:"owner"`
	IsTenantWide         bool                    `json:"isTenantWide"`
	AllowMultipleSources bool                    `json:"allowMultipleSources"`
	UsesSampleSheet      bool                    `json:"usesSampleSheet"`
	IsArchived           bool                    `json:"isArchived"`
	PipelineCode         *PipelineCode           `json:"pipelineCode"`
	CustomSettings       *CustomPipelineSettings `json:"customSettings"`
	CreatedAt            string                  `json:"createdAt"`
	UpdatedAt            string                  `json:"updatedAt"`
}
