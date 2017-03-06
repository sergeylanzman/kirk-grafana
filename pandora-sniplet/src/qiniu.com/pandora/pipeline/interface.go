package pipeline

import (
	"qiniu.com/pandora/base"
)

type PipelineAPI interface {
	CreateGroup(*CreateGroupInput) error

	UpdateGroup(*UpdateGroupInput) error

	StartGroupTask(*StartGroupTaskInput) error

	StopGroupTask(*StopGroupTaskInput) error

	ListGroups(*ListGroupsInput) (*ListGroupsOutput, error)

	GetGroup(*GetGroupInput) (*GetGroupOutput, error)

	DeleteGroup(*DeleteGroupInput) error

	CreateRepo(*CreateRepoInput) error

	GetRepo(*GetRepoInput) (*GetRepoOutput, error)

	ListRepos(*ListReposInput) (*ListReposOutput, error)

	DeleteRepo(*DeleteRepoInput) error

	PostData(*PostDataInput) error

	PostDataFromFile(*PostDataFromFileInput) error

	PostDataFromReader(*PostDataFromReaderInput) error

	PostDataFromBytes(*PostDataFromBytesInput) error

	UploadPlugin(*UploadPluginInput) error

	UploadPluginFromFile(*UploadPluginFromFileInput) error

	ListPlugins(*ListPluginsInput) (*ListPluginsOutput, error)

	GetPlugin(*GetPluginInput) (*GetPluginOutput, error)

	DeletePlugin(*DeletePluginInput) error

	CreateTransform(*CreateTransformInput) error

	GetTransform(*GetTransformInput) (*GetTransformOutput, error)

	ListTransforms(*ListTransformsInput) (*ListTransformsOutput, error)

	DeleteTransform(*DeleteTransformInput) error

	CreateExport(*CreateExportInput) error

	GetExport(*GetExportInput) (*GetExportOutput, error)

	ListExports(*ListExportsInput) (*ListExportsOutput, error)

	DeleteExport(*DeleteExportInput) error

	VerifyTransform(*VerifyTransformInput) (*VerifyTransformOutput, error)

	VerifyExport(*VerifyExportInput) error

	MakeToken(*base.TokenDesc) (string, error)
}
