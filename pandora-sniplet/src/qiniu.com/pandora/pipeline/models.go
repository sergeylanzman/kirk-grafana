package pipeline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"

	"qiniu.com/pandora/base"
	"qiniu.com/pandora/base/reqerr"
)

type PipelineToken struct {
	Token string `json:"-"`
}

const (
	schemaKeyPattern = "^[a-zA-Z_][a-zA-Z0-9_]{0,99}$"
	groupNamePattern = "^[a-zA-Z_][a-zA-Z0-9_]{2,127}$"
	repoNamePattern  = "^[a-zA-Z_][a-zA-Z0-9_]{2,127}$"
)

var schemaTypes = map[string]bool{
	"float":  true,
	"string": true,
	"long":   true,
	"date":   true,
}

func validateGroupName(g string) error {
	matched, err := regexp.MatchString(groupNamePattern, g)
	if err != nil {
		return reqerr.NewInvalidArgs("GroupName", err.Error())
	}
	if !matched {
		return reqerr.NewInvalidArgs("GroupName", fmt.Sprintf("invalid group name: %s", g))
	}
	return nil
}

func validateRepoName(r string) error {
	matched, err := regexp.MatchString(repoNamePattern, r)
	if err != nil {
		return reqerr.NewInvalidArgs("RepoName", err.Error())
	}
	if !matched {
		return reqerr.NewInvalidArgs("RepoName", fmt.Sprintf("invalid repo name: %s", r))
	}
	return nil
}

type Container struct {
	Type   string `json:"type"`
	Count  int    `json:"count"`
	Status string `json:"status,omitempty"`
}

func (c *Container) Validate() (err error) {
	if c.Type != "M16C4" && c.Type != "M32C8" {
		err = reqerr.NewInvalidArgs("ContainerType", fmt.Sprintf("invalid container type: %s, should be one of \"M16C4\" and \"M32C8\"", c.Type))
		return
	}
	if c.Count < 1 || c.Count > 128 {
		err = reqerr.NewInvalidArgs("ContainerCount", fmt.Sprintf("invalid container count: %d", c.Count))
		return
	}
	return
}

type CreateGroupInput struct {
	PipelineToken
	GroupName       string     `json:"-"`
	Region          string     `json:"region"`
	Container       *Container `json:"container"`
	AllocateOnStart bool       `json:"allocateOnStart,omitempty"`
}

func (g *CreateGroupInput) Validate() (err error) {
	if err = validateGroupName(g.GroupName); err != nil {
		return
	}
	if g.Region == "" {
		err = reqerr.NewInvalidArgs("Region", "region should not be empty")
		return
	}
	if g.Container == nil {
		err = reqerr.NewInvalidArgs("Container", "container should not be empty")
		return
	}
	if err = g.Container.Validate(); err != nil {
		return
	}
	return
}

type UpdateGroupInput struct {
	PipelineToken
	GroupName string     `json:"-"`
	Container *Container `json:"container"`
}

func (g *UpdateGroupInput) Validate() (err error) {
	if err = validateGroupName(g.GroupName); err != nil {
		return
	}
	if g.Container == nil {
		err = reqerr.NewInvalidArgs("Container", "container should not be empty")
		return
	}
	if err = g.Container.Validate(); err != nil {
		return
	}
	return
}

type StartGroupTaskInput struct {
	PipelineToken
	GroupName string
}

type StopGroupTaskInput struct {
	PipelineToken
	GroupName string
}

type GetGroupInput struct {
	PipelineToken
	GroupName string
}

type GetGroupOutput struct {
	Region     string     `json:"region"`
	Container  *Container `json:"container"`
	CreateTime string     `json:"createTime"`
	UpdateTime string     `json:"updateTime"`
}

type DeleteGroupInput struct {
	PipelineToken
	GroupName string
}

type GroupDesc struct {
	GroupName string     `json:"name"`
	Region    string     `json:"region"`
	Container *Container `json:"container"`
}

type ListGroupsInput struct {
	PipelineToken
}

type ListGroupsOutput struct {
	Groups []GroupDesc `json:"groups"`
}

type RepoSchemaEntry struct {
	Key       string `json:"key"`
	ValueType string `json:"valtype"`
	Required  bool   `json:"required"`
}

func (e RepoSchemaEntry) String() string {
	bytes, _ := json.Marshal(e)
	return string(bytes)
}

func (e *RepoSchemaEntry) Validate() (err error) {
	matched, err := regexp.MatchString(schemaKeyPattern, e.Key)
	if err != nil {
		err = reqerr.NewInvalidArgs("Schema", err.Error())
		return
	}
	if !matched {
		err = reqerr.NewInvalidArgs("Schema", fmt.Sprintf("invalid field key: %s", e.Key))
		return

	}
	if !schemaTypes[e.ValueType] {
		err = reqerr.NewInvalidArgs("Schema", fmt.Sprintf("invalid field type: %s, invalid field type should be one of \"float\", \"string\", \"date\" and \"long\"", e.ValueType))
		return
	}

	return
}

type CreateRepoInput struct {
	PipelineToken
	RepoName  string
	Region    string            `json:"region"`
	Schema    []RepoSchemaEntry `json:"schema"`
	GroupName string            `json:"group"`
}

func (r *CreateRepoInput) Validate() (err error) {
	if err = validateRepoName(r.RepoName); err != nil {
		return
	}

	if r.Schema == nil || len(r.Schema) == 0 {
		err = reqerr.NewInvalidArgs("Schema", "schema should not be empty")
		return
	}
	for _, item := range r.Schema {
		if err = item.Validate(); err != nil {
			return
		}
	}

	if r.GroupName != "" {
		if err = validateGroupName(r.GroupName); err != nil {
			return
		}
	}

	if r.Region == "" {
		err = reqerr.NewInvalidArgs("Region", "region should not be empty")
		return
	}
	return
}

type GetRepoInput struct {
	PipelineToken
	RepoName string
}

type GetRepoOutput struct {
	Region      string            `json:"region"`
	Schema      []RepoSchemaEntry `json:"schema"`
	GroupName   string            `json:"group"`
	DerivedFrom string            `json:"derivedFrom"`
}

type RepoDesc struct {
	RepoName    string `json:"name"`
	Region      string `json:"region"`
	GroupName   string `json:"group"`
	DerivedFrom string `json:"derivedFrom"`
}

type ListReposInput struct {
	PipelineToken
}

type ListReposOutput struct {
	Repos []RepoDesc `json:"repos"`
}

type DeleteRepoInput struct {
	PipelineToken
	RepoName string
}

type PointField struct {
	Key   string
	Value interface{}
}

func (p *PointField) String() string {
	return fmt.Sprintf("%s=%s\t", p.Key, escapeStringField(fmt.Sprintf("%v", p.Value)))
}

type Point struct {
	Fields []PointField
}

type Points []Point

func (ps Points) Buffer() []byte {
	var buf bytes.Buffer
	for _, p := range ps {
		for _, field := range p.Fields {
			buf.WriteString(field.String())
		}
		if len(p.Fields) > 0 {
			buf.Truncate(buf.Len() - 1)
		}
		buf.WriteByte('\n')
	}
	if len(ps) > 0 {
		buf.Truncate(buf.Len() - 1)
	}
	return buf.Bytes()
}

func escapeStringField(in string) string {
	var out []byte
	for i := 0; i < len(in); i++ {
		switch in[i] {
		case '\t': // escape tab
			out = append(out, '\\')
			out = append(out, 't')
		case '\n': // escape new line
			out = append(out, '\\')
			out = append(out, 'n')
		default:
			out = append(out, in[i])
		}
	}
	return string(out)
}

type PostDataInput struct {
	PipelineToken
	RepoName string
	Points   Points
}

type PostDataFromFileInput struct {
	PipelineToken
	RepoName string
	FilePath string
}

type PostDataFromReaderInput struct {
	PipelineToken
	RepoName string
	Reader   io.ReadSeeker
}

type PostDataFromBytesInput struct {
	PipelineToken
	RepoName string
	Buffer   []byte
}

type UploadPluginInput struct {
	PipelineToken
	PluginName string
	Buffer     *bytes.Buffer
}

type UploadPluginFromFileInput struct {
	PipelineToken
	PluginName string
	FilePath   string
}

type GetPluginInput struct {
	PipelineToken
	PluginName string
}

type PluginDesc struct {
	PluginName string `json:"name"`
	CreateTime string `json:"createTime"`
}

type GetPluginOutput struct {
	PluginDesc
}

type ListPluginsInput struct {
	PipelineToken
}

type ListPluginsOutput struct {
	Plugins []PluginDesc `json:"plugins"`
}

type DeletePluginInput struct {
	PipelineToken
	PluginName string
}

type TransformPluginOutputEntry struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

type TransformPlugin struct {
	Name   string                       `json:"name"`
	Output []TransformPluginOutputEntry `json:"output"`
}

type TransformSpec struct {
	Plugin    *TransformPlugin `json:"plugin,omitempty"`
	Mode      string           `json:"mode,omitempty"`
	Code      string           `json:"code,omitempty"`
	Interval  string           `json:"interval,omitempty"`
	Container *Container       `json:"container,omitempty"`
}

func (t *TransformSpec) Validate() (err error) {
	if t.Mode == "" && t.Code == "" && t.Plugin == nil {
		err = reqerr.NewInvalidArgs("TransformSpec", "all mode, code and plugin can not be empty")
		return
	}
	if t.Container != nil {
		if err = t.Container.Validate(); err != nil {
			return
		}
	}
	return
}

type CreateTransformInput struct {
	PipelineToken
	SrcRepoName   string
	TransformName string
	DestRepoName  string
	Spec          *TransformSpec
}

func (t *CreateTransformInput) Validate() (err error) {
	if err = validateRepoName(t.SrcRepoName); err != nil {
		return
	}
	if err = validateRepoName(t.DestRepoName); err != nil {
		return
	}
	if t.TransformName == "" {
		err = reqerr.NewInvalidArgs("TransformName", "transform name should be empty")
		return
	}
	if t.SrcRepoName == t.DestRepoName {
		err = reqerr.NewInvalidArgs("DestRepoName", "dest repo name should be different to src repo name")
		return
	}
	return t.Spec.Validate()
}

type TransformDesc struct {
	TransformName string         `json:"name"`
	DestRepoName  string         `json:"to"`
	Spec          *TransformSpec `json:"spec"`
}

type GetTransformInput struct {
	PipelineToken
	RepoName      string
	TransformName string
}

type GetTransformOutput struct {
	TransformDesc
}

type DeleteTransformInput struct {
	PipelineToken
	RepoName      string
	TransformName string
}

type ListTransformsInput struct {
	PipelineToken
	RepoName string
}

type ListTransformsOutput struct {
	Transforms []TransformDesc `json:"transforms"`
}

type ExportFilter struct {
	Rules     map[string]map[string]string `json:"rules"`
	ToDefault bool                         `json:"toDefault"`
}

func (f *ExportFilter) Validate() (err error) {
	if len(f.Rules) == 0 {
		err = reqerr.NewInvalidArgs("ExportFilter", "rules in filter should be empty")
		return
	}
	return
}

type ExportTsdbSpec struct {
	DestRepoName string            `json:"destRepoName"`
	SeriesName   string            `json:"series"`
	Tags         map[string]string `json:"tags"`
	Fields       map[string]string `json:"fields"`
	Timestamp    string            `json:"timestamp,omitempty"`
	Filter       *ExportFilter     `json:"filter,omitempty"`
}

func (s *ExportTsdbSpec) Validate() (err error) {
	if s.DestRepoName == "" {
		err = reqerr.NewInvalidArgs("ExportSpec", "dest repo name should not be empty")
		return
	}
	if s.SeriesName == "" {
		err = reqerr.NewInvalidArgs("ExportSpec", "series name should not be empty")
		return
	}
	if s.Filter == nil {
		return
	}
	return s.Filter.Validate()
}

type ExportMongoSpec struct {
	Host      string                 `json:"host"`
	DbName    string                 `json:"dbName"`
	CollName  string                 `json:"collName"`
	Mode      string                 `json:"mode"`
	UpdateKey []string               `json:"updateKey,omitempty"`
	Doc       map[string]interface{} `json:"doc"`
	Version   string                 `json:"version,omitempty"`
	Filter    *ExportFilter          `json:"filter,omitempty"`
}

func (s *ExportMongoSpec) Validate() (err error) {
	if s.Host == "" {
		err = reqerr.NewInvalidArgs("ExportSpec", "host should not be empty")
		return
	}
	if s.DbName == "" {
		err = reqerr.NewInvalidArgs("ExportSpec", "dbname should not be empty")
		return
	}
	if s.CollName == "" {
		err = reqerr.NewInvalidArgs("ExportSpec", "collection name should not be empty")
		return
	}
	if s.Mode != "UPSERT" && s.Mode != "INSERT" && s.Mode != "UPDATE" {
		err = reqerr.NewInvalidArgs("ExportSpec", fmt.Sprintf("invalid mode: %s, mode should be one of \"UPSERT\", \"INSERT\" and \"UPDATE\"", s.Mode))
		return
	}
	if s.Filter == nil {
		return
	}
	return s.Filter.Validate()
}

type ExportLogDBSpec struct {
	DestRepoName string                 `json:"destRepoName"`
	Doc          map[string]interface{} `json:"doc"`
	Filter       *ExportFilter          `json:"filter,omitempty"`
}

func (s *ExportLogDBSpec) Validate() (err error) {
	if s.DestRepoName == "" {
		err = reqerr.NewInvalidArgs("ExportSpec", "dest repo name should not be empty")
		return
	}
	if s.Filter == nil {
		return
	}
	return s.Filter.Validate()
}

type ExportKodoSpec struct {
	Bucket         string            `json:"bucket"`
	KeyPrefix      string            `json:"keyPrefix"`
	Fields         map[string]string `json:"fields"`
	RotateInterval int               `json:"rotateInterval,omitempty"`
	Email          string            `json:"email"`
	AccessKey      string            `json:"accessKey"`
	Format         string            `json:"format"`
	Compress       bool              `json:"compress"`
	Retention      int               `json:"retention"`
	Filter         *ExportFilter     `json:"filter,omitempty"`
}

func (s *ExportKodoSpec) Validate() (err error) {
	if s.Bucket == "" {
		err = reqerr.NewInvalidArgs("ExportSpec", "bucket should not be empty")
		return
	}
	if s.Filter == nil {
		return
	}
	return s.Filter.Validate()
}

type ExportHttpSpec struct {
	Host string `json:"host"`
	Uri  string `json:"uri"`
}

func (s *ExportHttpSpec) Validate() (err error) {
	if s.Host == "" {
		err = reqerr.NewInvalidArgs("ExportSpec", "host should not be empty")
		return
	}
	if s.Uri == "" {
		err = reqerr.NewInvalidArgs("ExportSpec", "uri should not be empty")
		return
	}
	return
}

type CreateExportInput struct {
	PipelineToken
	RepoName   string      `json:"-"`
	ExportName string      `json:"-"`
	Type       string      `json:"type"`
	Spec       interface{} `json:"spec"`
	Whence     string      `json:"whence,omitempty"`
}

func (e *CreateExportInput) Validate() (err error) {
	if err = validateRepoName(e.RepoName); err != nil {
		return
	}
	if e.ExportName == "" {
		err = reqerr.NewInvalidArgs("ExportSpec", "export name should not be empty")
		return
	}
	if e.Spec == nil {
		err = reqerr.NewInvalidArgs("ExportSpec", "spec should not be nil")
		return
	}
	if e.Whence != "" && e.Whence != "oldest" && e.Whence != "newest" {
		err = reqerr.NewInvalidArgs("ExportSpec", "whence must be empty, \"oldest\" or \"newest\"")
		return
	}

	switch t := e.Spec.(type) {
	case *ExportTsdbSpec, ExportTsdbSpec:
		e.Type = "tsdb"
	case *ExportMongoSpec, ExportMongoSpec:
		e.Type = "mongo"
	case *ExportLogDBSpec, ExportLogDBSpec:
		e.Type = "logdb"
	case *ExportKodoSpec, ExportKodoSpec:
		e.Type = "kodo"
	case *ExportHttpSpec, ExportHttpSpec:
		e.Type = "http"
	default:
		err = reqerr.NewInvalidArgs("ExportSpec", fmt.Sprintf("invalid export spec type: %v", t))
		return
	}

	vv, ok := e.Spec.(base.Validator)
	if !ok {
		err = reqerr.NewInvalidArgs("ExportSpec", "export spec cannot cast to validator")
		return
	}
	return vv.Validate()
}

type ExportDesc struct {
	Name   string                 `json:"name,omitempty"`
	Type   string                 `json:"type"`
	Spec   map[string]interface{} `json:"spec"`
	Whence string                 `json:"whence,omitempty"`
}

type GetExportInput struct {
	PipelineToken
	RepoName   string
	ExportName string
}

type GetExportOutput struct {
	ExportDesc
}

type ListExportsInput struct {
	PipelineToken
	RepoName string
}

type ListExportsOutput struct {
	Exports []ExportDesc `json:"exports"`
}

type DeleteExportInput struct {
	PipelineToken
	RepoName     string
	ExportName   string
	DeleteOffset bool
}

type VerifyTransformInput struct {
	PipelineToken
	Schema []RepoSchemaEntry `json:"schema"`
	Spec   *TransformSpec    `json:"spec"`
}

func (v *VerifyTransformInput) Validate() (err error) {
	if v.Schema == nil || len(v.Schema) == 0 {
		err = reqerr.NewInvalidArgs("Schema", "schema should not be empty")
		return
	}
	for _, item := range v.Schema {
		if err = item.Validate(); err != nil {
			return
		}
	}

	return v.Spec.Validate()
}

type VerifyTransformOutput struct {
	Schema []RepoSchemaEntry `json:"schema"`
}

type VerifyExportInput struct {
	PipelineToken
	Schema []RepoSchemaEntry `json:"schema"`
	Type   string            `json:"type"`
	Spec   interface{}       `json:"spec"`
	Whence string            `json:"whence,omitempty"`
}

func (v *VerifyExportInput) Validate() (err error) {
	if v.Schema == nil || len(v.Schema) == 0 {
		err = reqerr.NewInvalidArgs("VerifyExportSpec", "schema should not be empty")
		return
	}
	for _, item := range v.Schema {
		if err = item.Validate(); err != nil {
			return
		}
	}

	if v.Spec == nil {
		err = reqerr.NewInvalidArgs("ExportSpec", "spec should not be nil")
		return
	}

	if v.Whence != "" && v.Whence != "oldest" && v.Whence != "newest" {
		err = reqerr.NewInvalidArgs("ExportSpec", "whence must be empty, \"oldest\" or \"newest\"")
		return
	}

	switch t := v.Spec.(type) {
	case *ExportTsdbSpec, ExportTsdbSpec:
		v.Type = "tsdb"
	case *ExportMongoSpec, ExportMongoSpec:
		v.Type = "mongo"
	case *ExportLogDBSpec, ExportLogDBSpec:
		v.Type = "logdb"
	case *ExportKodoSpec, ExportKodoSpec:
		v.Type = "kodo"
	case *ExportHttpSpec, ExportHttpSpec:
		v.Type = "http"
	default:
		err = reqerr.NewInvalidArgs("ExportSpec", fmt.Sprintf("invalid export spec type: %v", t))
		return
	}

	vv, ok := v.Spec.(base.Validator)
	if !ok {
		err = reqerr.NewInvalidArgs("ExportSpec", "export spec cannot cast to validator")
		return
	}
	return vv.Validate()
}
