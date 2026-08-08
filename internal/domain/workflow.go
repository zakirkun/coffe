package domain

type Intent string

const IntentDeploy Intent = "deploy"

func (i Intent) Valid() bool { return i == IntentDeploy }

type Workflow struct {
	composeFile ComposeFile
	stackName   StackName
	mode        DeploymentMode
	build       bool
}

func NewWorkflow(composeFile ComposeFile, build bool) Workflow {
	return Workflow{composeFile: composeFile, mode: DeploymentCompose, build: build}
}

func NewSwarmWorkflow(composeFile ComposeFile, stackName StackName) Workflow {
	return Workflow{composeFile: composeFile, stackName: stackName, mode: DeploymentSwarm}
}

func (w Workflow) ComposeFile() ComposeFile { return w.composeFile }
func (w Workflow) StackName() StackName     { return w.stackName }
func (w Workflow) Mode() DeploymentMode     { return w.mode }
func (w Workflow) Build() bool              { return w.build }

type DeploymentMode string

const (
	DeploymentCompose DeploymentMode = "compose"
	DeploymentSwarm   DeploymentMode = "swarm"
)
