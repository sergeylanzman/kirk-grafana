package logdb

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"qiniu.com/pandora/base/reqerr"
)

type LogdbToken struct {
	Token string `json:"-"`
}

const (
	schemaKeyPattern = "^[a-zA-Z_][a-zA-Z0-9_]{0,99}$"
	groupNamePattern = "^[a-zA-Z_][a-zA-Z0-9_]{2,127}$"
	repoNamePattern  = "^[a-zA-Z_][a-zA-Z0-9_]{2,127}$"
	retentionPattern = "^(-1|0|[1-9][0-9]*)d$"
)

const (
	minRetentionDay = 1
	maxRetentionDay = 30
)

var schemaTypes = map[string]bool{
	"float":  true,
	"string": true,
	"long":   true,
	"date":   true,
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

type RepoSchemaEntry struct {
	Key       string `json:"key"`
	ValueType string `json:"valtype"`
}

func (e RepoSchemaEntry) String() string {
	bytes, _ := json.Marshal(e)
	return string(bytes)
}

func (e *RepoSchemaEntry) Validate() (err error) {
	matched, err := regexp.MatchString(schemaKeyPattern, e.Key)
	if err != nil {
		return reqerr.NewInvalidArgs("Schema", err.Error())
	}
	if !matched {
		return reqerr.NewInvalidArgs("Schema", fmt.Sprintf("invalid field key: %s", e.Key))
	}
	if !schemaTypes[e.ValueType] {
		return reqerr.NewInvalidArgs("Schema", fmt.Sprintf("invalid field type: %s, invalid field type should be one of \"float\", \"string\", \"date\" and \"long\"", e.ValueType))
	}

	return
}

type CreateRepoInput struct {
	LogdbToken
	RepoName  string
	Region    string            `json:"region"`
	Retention string            `json:"retention"`
	Schema    []RepoSchemaEntry `json:"schema"`
}

func (r *CreateRepoInput) Validate() (err error) {
	if err = validateRepoName(r.RepoName); err != nil {
		return
	}

	if r.Schema == nil || len(r.Schema) == 0 {
		return reqerr.NewInvalidArgs("Schema", "schema should not be empty")
	}
	for _, item := range r.Schema {
		if err = item.Validate(); err != nil {
			return
		}
	}

	return checkRetention(r.Retention)
}

func checkRetention(retention string) error {
	matched, err := regexp.MatchString(retentionPattern, retention)
	if err != nil {
		return reqerr.NewInvalidArgs("Retention", "parse retention time failed")
	}
	if !matched {
		return reqerr.NewInvalidArgs("Retention", "invalid retention time format")
	}
	retentionInt, err := strconv.Atoi(strings.Replace(retention, "d", "", -1))
	if err != nil {
		return reqerr.NewInvalidArgs("Retention", "invalid retention time format")
	}

	if retentionInt > maxRetentionDay || retentionInt < minRetentionDay {
		return reqerr.NewInvalidArgs("Retention", "invalid retention range")
	}
	return nil
}

type UpdateRepoInput struct {
	LogdbToken
	RepoName  string
	Region    string            `json:"region"`
	Retention string            `json:"retention"`
	Schema    []RepoSchemaEntry `json:"schema"`
}

func (r *UpdateRepoInput) Validate() (err error) {
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

	return checkRetention(r.Retention)
}

type GetRepoInput struct {
	LogdbToken
	RepoName string
}

type GetRepoOutput struct {
	Region     string            `json:"region"`
	Retention  string            `json:"retention"`
	Schema     []RepoSchemaEntry `json:"schema"`
	CreateTime string            `json:"createTime"`
	UpdateTime string            `json:"updateTime"`
}

type RepoDesc struct {
	RepoName   string `json:"name"`
	Region     string `json:"region"`
	Retention  string `json:"retention"`
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
}

type ListReposInput struct {
	LogdbToken
}

type ListReposOutput struct {
	Repos []RepoDesc `json:"repos"`
}

type DeleteRepoInput struct {
	LogdbToken
	RepoName string
}

type Log map[string]interface{}

type Logs []Log

func (ls Logs) Buf() (buf []byte, err error) {
	buf, err = json.Marshal(ls)
	if err != nil {
		return
	}
	return
}

type SendLogInput struct {
	LogdbToken
	RepoName       string `json:"-"`
	OmitInvalidLog bool   `json:"-"`
	Logs           Logs
}

type SendLogOutput struct {
	Success int `json:"success"`
	Failed  int `json:"failed"`
	Total   int `json:"total"`
}

type Highlight struct {
	PreTags           []string               `json:"pre_tags"`
	PostTags          []string               `json:"post_tags"`
	Fields            map[string]interface{} `json:"fields"`
	RequireFieldMatch bool                   `json:"require_field_match"`
	FragmentSize      int                    `json:"fragment_size"`
}

func (h *Highlight) Validate() error {
	return nil
}

type QueryLogInput struct {
	LogdbToken
	RepoName  string
	Query     string
	Sort      string
	From      int
	Size      int
	Highlight *Highlight
}

type QueryLogOutput struct {
	Total          int                      `json:"total"`
	PartialSuccess bool                     `json:"partialSuccess"`
	Data           []map[string]interface{} `json:"data"`
}

type QueryHistogramLogInput struct {
	LogdbToken
	RepoName string
	Query    string
	Field    string
	From     int64
	To       int64
}

type LogHistogramDesc struct {
	Key   int64 `json:"key"`
	Count int64 `json:"count"`
}

type QueryHistogramLogOutput struct {
	Total          int                `json:"total"`
	PartialSuccess bool               `json:"partialSuccess"`
	Buckets        []LogHistogramDesc `json:"buckets"`
}

type PutRepoConfigInput struct {
	LogdbToken
	RepoName      string
	TimeFieldName string `json:"timeFieldName"`
}

func (r *PutRepoConfigInput) Validate() (err error) {
	return nil
}

type GetRepoConfigInput struct {
	LogdbToken
	RepoName string
}

type GetRepoConfigOutput struct {
	TimeFieldName string `json:"timeFieldName"`
}
