package pipeline

import (
	"bytes"
	"os"
	"reflect"
	"testing"
	"time"

	. "qiniu.com/pandora/base"
	"qiniu.com/pandora/base/config"
	"qiniu.com/pandora/base/reqerr"
	. "qiniu.com/pandora/pipeline"
)

var (
	cfg               *config.Config
	client            PipelineAPI
	region            = os.Getenv("REGION")
	endpoint          = os.Getenv("PIPELINE_HOST")
	ak                = os.Getenv("ACCESS_KEY")
	sk                = os.Getenv("SECRET_KEY")
	logger            Logger
	defaultRepoSchema []RepoSchemaEntry
	defaultContainer  *Container
)

func init() {
	var err error
	logger = NewDefaultLogger()
	cfg = NewConfig().
		WithEndpoint(endpoint).
		WithAccessKeySecretKey(ak, sk).
		WithLogger(logger).
		WithLoggerLevel(LogDebug)

	client, err = New(cfg)
	if err != nil {
		logger.Error("new pipeline client failed, err: %v", err)
	}

	defaultRepoSchema = []RepoSchemaEntry{
		RepoSchemaEntry{
			Key:       "f1",
			ValueType: "string",
			Required:  true,
		},
		RepoSchemaEntry{
			Key:       "f2",
			ValueType: "float",
			Required:  true,
		},
	}
	defaultContainer = &Container{
		Type:  "M16C4",
		Count: 1,
	}

	cleanupEnv()
}

func cleanupEnv() {
	output, err := client.ListRepos(&ListReposInput{})
	if err != nil {
		logger.Errorf("list repos failed: %v", err)
		return
	}
	for _, repo := range output.Repos {
		if repo.DerivedFrom == "" {
			for _, repo := range output.Repos {
				if err = deleteRepo(repo.RepoName); err != nil {
					return
				}
			}
		}
	}

	logger.Info("cleanupEnv success")
}

func deleteRepo(repoName string) (err error) {
	output1, err := client.ListExports(&ListExportsInput{RepoName: repoName})
	if err != nil {
		logger.Errorf("list exports failed: %v", err)
		return
	}
	for _, export := range output1.Exports {
		err = client.DeleteExport(&DeleteExportInput{RepoName: repoName, ExportName: export.Name})
		if err != nil {
			logger.Errorf("delete export failed: %v", err)
			return
		}
	}

	output2, err := client.ListTransforms(&ListTransformsInput{RepoName: repoName})
	if err != nil {
		logger.Errorf("list transforms failed: %v", err)
		return
	}
	for _, transform := range output2.Transforms {
		err = deleteRepo(transform.DestRepoName)
		if err != nil {
			logger.Error("delete repo cascading failed")
			return
		}
		err = client.DeleteTransform(&DeleteTransformInput{RepoName: repoName, TransformName: transform.TransformName})
		if err != nil {
			logger.Errorf("delete transform failed: %v", err)
			return
		}
	}

	return client.DeleteRepo(&DeleteRepoInput{RepoName: repoName})
}

func writeToFile(path string, content []byte, t *testing.T) {
	file, err := os.Create(path)
	if err != nil {
		t.Error(err)
	}
	if _, err = file.Write(content); err != nil {
		t.Error(err)
	}
	if err = file.Close(); err != nil {
		t.Error(err)
	}
}

func TestGroup(t *testing.T) {
	groupName := "group"
	createInput := &CreateGroupInput{
		GroupName:       groupName,
		Container:       defaultContainer,
		Region:          region,
		AllocateOnStart: false,
	}
	err := client.CreateGroup(createInput)
	if err != nil {
		t.Error(err)
	}

	getOutput, err := client.GetGroup(&GetGroupInput{GroupName: groupName})
	if err != nil {
		t.Error(err)
	}
	if getOutput == nil {
		t.Errorf("get output should not be empty")
	}

	if getOutput.Region != region {
		t.Errorf("region of group should be nb")
	}
	if getOutput.Container.Count != defaultContainer.Count ||
		getOutput.Container.Type != defaultContainer.Type {
		t.Errorf("container of group %v should equal to default container %v", getOutput.Container, defaultContainer)
	}
	if getOutput.CreateTime == "" || getOutput.UpdateTime == "" {
		t.Errorf("create time and update time should not be empty")
	}

	listOutput, err := client.ListGroups(&ListGroupsInput{})
	if err != nil {
		t.Error(err)
	}
	if listOutput == nil {
		t.Error("listOutput should not be empty")
	}
	if len(listOutput.Groups) != 1 {
		t.Errorf("group count should be 1 but %d", len(listOutput.Groups))
	}

	err = client.StartGroupTask(&StartGroupTaskInput{GroupName: groupName})
	if err != nil {
		t.Error(err)
	}

	err = client.StopGroupTask(&StopGroupTaskInput{GroupName: groupName})
	if err != nil {
		t.Error(err)
	}

	err = client.DeleteGroup(&DeleteGroupInput{GroupName: groupName})
	if err != nil {
		t.Error(err)
	}
}

func TestRepo(t *testing.T) {
	repoName := "repo"
	createInput := &CreateRepoInput{
		RepoName:  repoName,
		Region:    region,
		Schema:    defaultRepoSchema,
		GroupName: "",
	}

	err := client.CreateRepo(createInput)
	if err != nil {
		t.Error(err)
	}

	getOutput, err := client.GetRepo(&GetRepoInput{RepoName: repoName})
	if err != nil {
		t.Error(err)
	}
	if getOutput == nil {
		t.Error("schema ret is empty")
	}
	if "nb" != getOutput.Region ||
		"" != getOutput.GroupName ||
		!reflect.DeepEqual(defaultRepoSchema, getOutput.Schema) {
		t.Error("spec is different to default spec")
	}

	listOutput, err := client.ListRepos(&ListReposInput{})
	if err != nil {
		t.Error(err)
	}
	if listOutput == nil {
		t.Error("repo list should not be empty")
	}
	if len(listOutput.Repos) != 1 {
		t.Errorf("repo count should be 1 but %d", len(listOutput.Repos))
	}
	if listOutput.Repos[0].RepoName != "repo" {
		t.Error("repo name is different to origin name")
	}

	err = client.DeleteRepo(&DeleteRepoInput{RepoName: repoName})
	if err != nil {
		t.Error(err)
	}
}

func TestRepo_InvalidSpec(t *testing.T) {
	var tests = []struct {
		input *CreateRepoInput
	}{
		{
			input: &CreateRepoInput{
				RepoName:  "...",
				GroupName: "group",
				Schema: []RepoSchemaEntry{
					RepoSchemaEntry{
						Key:       "f1",
						ValueType: "string",
						Required:  true,
					},
				},
			},
		},
		{
			input: &CreateRepoInput{
				RepoName:  "repo",
				GroupName: "...",
				Schema: []RepoSchemaEntry{
					RepoSchemaEntry{
						Key:       "f1",
						ValueType: "string",
						Required:  true,
					},
				},
			},
		},
		{
			input: &CreateRepoInput{
				RepoName:  "repo",
				GroupName: "",
				Schema:    []RepoSchemaEntry{},
			},
		},
		{
			input: &CreateRepoInput{
				RepoName:  "repo",
				GroupName: "",
				Schema: []RepoSchemaEntry{
					RepoSchemaEntry{
						Key:       "...",
						ValueType: "string",
						Required:  true,
					},
				},
			},
		},
		{
			input: &CreateRepoInput{
				RepoName:  "repo",
				GroupName: "",
				Schema: []RepoSchemaEntry{
					RepoSchemaEntry{
						Key:       "f1",
						ValueType: "map",
						Required:  true,
					},
				},
			},
		},
	}

	for i, tt := range tests {
		err := client.CreateRepo(tt.input)
		if err == nil {
			t.Error("index: %s, create repo should failed for invalid input", i)
		}
		if e, ok := err.(*reqerr.RequestError); !ok || e.ErrorType != reqerr.InvalidArgs {
			t.Errorf("index: %d, got err msg: %s", i, err.Error())
		}
	}
}

func TestPlugin(t *testing.T) {
	pluginName := "plugin"
	pluginInput := &UploadPluginInput{
		PluginName: pluginName,
		Buffer:     bytes.NewBufferString("plugin content"),
	}
	err := client.UploadPlugin(pluginInput)
	if err != nil {
		t.Error(err)
	}

	getOutput, err := client.GetPlugin(&GetPluginInput{PluginName: pluginName})
	if err != nil {
		t.Error(err)
	}
	if getOutput.PluginName != pluginName {
		t.Error("plugin name is different to orign plugin name")
	}

	listOutput, err := client.ListPlugins(&ListPluginsInput{})
	if err != nil {
		t.Error(err)
	}
	if len(listOutput.Plugins) != 1 {
		t.Errorf("plugin count should be 1 but %d", len(listOutput.Plugins))
	}
	if listOutput.Plugins[0].PluginName != pluginName {
		t.Errorf("plugin name is different to origin name")
	}

	if err = client.DeletePlugin(&DeletePluginInput{PluginName: pluginName}); err != nil {
		t.Error(err)
	}

	path := "/tmp/plugin.jar"
	filePluginInput := &UploadPluginFromFileInput{
		PluginName: "plugin",
		FilePath:   path,
	}
	content := []byte("local file plugin content")
	writeToFile(path, content, t)

	if err = client.UploadPluginFromFile(filePluginInput); err != nil {
		t.Error(err)
	}
	if err = client.DeletePlugin(&DeletePluginInput{PluginName: pluginName}); err != nil {
		t.Error(err)
	}
}

func TestPostData(t *testing.T) {
	repoName := "repo_post_data"
	createRepoInput := &CreateRepoInput{
		RepoName: repoName,
		Schema:   defaultRepoSchema,
		Region:   "nb",
	}
	err := client.CreateRepo(createRepoInput)
	if err != nil {
		t.Error(err)
	}

	postDataInput := &PostDataInput{
		RepoName: repoName,
		Points: Points{
			Point{
				[]PointField{
					PointField{
						Key:   "f1",
						Value: "12.7",
					},
					PointField{
						Key:   "f2",
						Value: 1.0,
					},
				},
			},
		},
	}
	err = client.PostData(postDataInput)
	if err != nil {
		t.Error(err)
	}

	buf := []byte("f1=\"12.7\"\tf2=3.14\nf1=\"dang\"\tf2=1024.0")
	postDataFromBytesInput := &PostDataFromBytesInput{
		RepoName: repoName,
		Buffer:   buf,
	}
	err = client.PostDataFromBytes(postDataFromBytesInput)
	if err != nil {
		t.Error(err)
	}

	postDataFromReaderInput := &PostDataFromReaderInput{
		RepoName: repoName,
		Reader:   bytes.NewReader(buf),
	}
	err = client.PostDataFromReader(postDataFromReaderInput)
	if err != nil {
		t.Error(err)
	}

	path := "/tmp/postdata"
	writeToFile(path, buf, t)
	postDataFromFileInput := &PostDataFromFileInput{
		RepoName: repoName,
		FilePath: path,
	}
	err = client.PostDataFromFile(postDataFromFileInput)
	if err != nil {
		t.Error(err)
	}

	err = client.DeleteRepo(&DeleteRepoInput{RepoName: repoName})
	if err != nil {
		t.Error(err)
	}
}

func TestPostData_WithEscapeCharacters(t *testing.T) {
	repoName := "repo_post_data_with_escape"
	createRepoInput := &CreateRepoInput{
		RepoName: repoName,
		Schema:   defaultRepoSchema,
		Region:   "nb",
	}
	err := client.CreateRepo(createRepoInput)
	if err != nil {
		t.Error(err)
	}

	postDataInput := &PostDataInput{
		RepoName: repoName,
		Points: Points{
			Point{
				[]PointField{
					PointField{
						Key:   "f1",
						Value: "tab: \t xxxxx",
					},
					PointField{
						Key:   "f2",
						Value: 1.0,
					},
				},
			},
			Point{
				[]PointField{
					PointField{
						Key:   "f1",
						Value: "newline: \n yyyy",
					},
					PointField{
						Key:   "f2",
						Value: 1.0,
					},
				},
			},
			Point{
				[]PointField{
					PointField{
						Key:   "f1",
						Value: "tab: \t and newline: \n zzzz",
					},
					PointField{
						Key:   "f2",
						Value: 1.0,
					},
				},
			},
		},
	}
	err = client.PostData(postDataInput)
	if err != nil {
		t.Error(err)
	}

	err = client.DeleteRepo(&DeleteRepoInput{RepoName: repoName})
	if err != nil {
		t.Error(err)
	}
}

func TestTransform(t *testing.T) {
	createRepoInput := &CreateRepoInput{
		RepoName: "src_repo",
		Schema:   defaultRepoSchema,
		Region:   "nb",
	}
	err := client.CreateRepo(createRepoInput)
	if err != nil {
		t.Error(err)
	}

	spec := &TransformSpec{
		Mode:      "sql",
		Code:      "select * from stream",
		Interval:  "5m",
		Container: defaultContainer,
	}
	createTransInput := &CreateTransformInput{
		SrcRepoName:   "src_repo",
		DestRepoName:  "dest_repo",
		TransformName: "transform",
		Spec:          spec,
	}
	err = client.CreateTransform(createTransInput)
	if err != nil {
		t.Error(err)
	}

	listTransOutput, err := client.ListTransforms(&ListTransformsInput{RepoName: "src_repo"})
	if err != nil {
		t.Error(err)
	}
	if listTransOutput == nil {
		t.Error("listTransOutput should not be empty")
	}

	getTransOutput, err := client.GetTransform(&GetTransformInput{RepoName: "src_repo", TransformName: "transform"})
	if err != nil {
		t.Error(err)
	}
	if getTransOutput == nil {
		t.Errorf("getTransInput should be empty")
	}
	if getTransOutput.TransformName != "transform" {
		t.Errorf("transformName should be \"transform\"")
	}
	if getTransOutput.DestRepoName != "dest_repo" {
		t.Errorf("destRepoName should be \"dest_repo\"")
	}
	if getTransOutput.Spec == nil {
		t.Error("spec in getTransOutput should not be empty")
	}
	if getTransOutput.Spec.Mode != "sql" {
		t.Errorf("Mode should be \"sql\"")
	}
	if getTransOutput.Spec.Code != "select * from stream" {
		t.Errorf("Code should be \"select * from stream\"")
	}
	if getTransOutput.Spec.Interval != "5m" {
		t.Errorf("Interval should be \"5m\"")
	}
	if !reflect.DeepEqual(getTransOutput.Spec.Container, defaultContainer) {
		t.Errorf("spec in getTransOutput %v is not equal spec %v", getTransOutput.Spec, spec)
	}
	err = client.DeleteTransform(&DeleteTransformInput{RepoName: "src_repo", TransformName: "transform"})
	if err != nil {
		t.Error(err)
	}
	err = client.DeleteRepo(&DeleteRepoInput{RepoName: "src_repo"})
	if err != nil {
		t.Error(err)
	}
}

func TestExport(t *testing.T) {
	repoName := "repo_for_export"
	createRepoInput := &CreateRepoInput{
		RepoName: repoName,
		Schema:   defaultRepoSchema,
		Region:   "nb",
	}
	err := client.CreateRepo(createRepoInput)
	if err != nil {
		t.Error(err)
	}

	exports := []CreateExportInput{
		CreateExportInput{
			RepoName:   repoName,
			ExportName: "tsdb_export",
			Spec: &ExportTsdbSpec{
				DestRepoName: "tsdb_dest_repo",
				SeriesName:   "series",
				Tags:         map[string]string{"tag1": "#f1"},
				Fields:       map[string]string{"field1": "#f1", "field2": "#f2"},
			},
			Whence: "oldest",
		},
		CreateExportInput{
			RepoName:   repoName,
			ExportName: "lg_export",
			Spec: &ExportLogDBSpec{
				DestRepoName: "lg_dest_repo",
				Doc:          map[string]interface{}{"f1": "#f1"},
			},
			Whence: "newest",
		},
		CreateExportInput{
			RepoName:   repoName,
			ExportName: "mongo_export",
			Spec: &ExportMongoSpec{
				Host:     "10.200.20.23:27017",
				DbName:   "test",
				CollName: "my_coll",
				Mode:     "UPSERT",
				Doc:      map[string]interface{}{"f1": "#f1"},
			},
		},
		CreateExportInput{
			RepoName:   repoName,
			ExportName: "kodo_export",
			Spec: &ExportKodoSpec{
				Bucket:         "mybucket",
				KeyPrefix:      "export_prefix_",
				Fields:         map[string]string{"field1": "#f1", "field2": "#f2"},
				RotateInterval: 30,
				Email:          "pipeline@qiniu.com",
				AccessKey:      "ak",
				Format:         "text",
				Compress:       true,
				Retention:      3,
			},
		},
		CreateExportInput{
			RepoName:   repoName,
			ExportName: "http_export",
			Spec: &ExportHttpSpec{
				Host: "http://qiniu.com",
				Uri:  "/resource",
			},
		},
	}

	for _, export := range exports {
		err = client.CreateExport(&export)
		if err != nil {
			t.Errorf("export: %s create failed, err: %v", export.ExportName, err)
		}

		getExportInput := &GetExportInput{
			RepoName:   export.RepoName,
			ExportName: export.ExportName,
		}
		getExportOutput, err := client.GetExport(getExportInput)
		if err != nil {
			t.Error(err)
		}
		if getExportOutput == nil {
			t.Error("getExportInput should not be nil")
		}
		if getExportOutput.Type != export.Type {
			t.Errorf("type %s is different to expected type %s", getExportOutput.Type, export.Type)
		}
		if getExportOutput.Whence != export.Whence {
			t.Errorf("whence %s is different to expected whence %s", getExportOutput.Whence, export.Whence)
		}
	}

	listExportsOutput, err := client.ListExports(&ListExportsInput{RepoName: repoName})
	if err != nil {
		t.Error(err)
	}
	if listExportsOutput == nil {
		t.Error("listExportsOutput should not be nil")
	}
	if len(listExportsOutput.Exports) != len(exports) {
		t.Errorf("list export count %d should be equal to %d", len(listExportsOutput.Exports), len(exports))
	}

	for _, export := range exports {
		deleteExportInput := &DeleteExportInput{
			RepoName:   export.RepoName,
			ExportName: export.ExportName,
		}
		err = client.DeleteExport(deleteExportInput)
		if err != nil {
			t.Error(err)
		}
	}

	err = client.DeleteRepo(&DeleteRepoInput{RepoName: repoName})
	if err != nil {
		t.Error(err)
	}
}

func TestPostDataWithToken(t *testing.T) {
	repoName := "repo_post_data_with_token"
	createInput := &CreateRepoInput{
		RepoName: repoName,
		Region:   "nb",
		Schema:   defaultRepoSchema,
	}

	err := client.CreateRepo(createInput)
	if err != nil {
		t.Error(err)
	}

	td := &TokenDesc{}
	td.Expires = time.Now().Unix() + 10
	td.Method = "POST"
	td.Url = "/v2/repos/repo_post_data_with_token/data"
	td.ContentType = "text/plain"

	token, err := client.MakeToken(td)
	if err != nil {
		t.Error(err)
	}

	cfg2 := NewConfig().WithEndpoint(endpoint)

	client2, err2 := New(cfg2)
	if err2 != nil {
		logger.Error("new pipeline client failed, err: %v", err2)
	}
	postDataInput := &PostDataInput{
		RepoName: repoName,
		Points: Points{
			Point{
				[]PointField{
					PointField{
						Key:   "f1",
						Value: 12.7,
					},
					PointField{
						Key:   "f2",
						Value: 1.0,
					},
				},
			},
		},
		PipelineToken: PipelineToken{
			Token: token,
		},
	}
	err = client2.PostData(postDataInput)
	if err != nil {
		t.Error(err)
	}

	time.Sleep(15 * time.Second)

	err = client2.PostData(postDataInput)
	if err == nil {
		t.Errorf("expired token: %s, expires: %d, now: %d", token, td.Expires, time.Now().Unix())
	}

	v, ok := err.(*reqerr.RequestError)
	if !ok {
		t.Errorf("cast err to RequestError fail, err: %v", err)
	}

	if v.ErrorType != reqerr.UnauthorizedError {
		t.Errorf("got errorType: %d, expected errorType: %d", v.ErrorType, reqerr.UnauthorizedError)
	}

	if v.StatusCode != 401 {
		t.Errorf("expires token, expires: %d, now: %d", td.Expires, time.Now().Unix())
	}

	err = client.DeleteRepo(&DeleteRepoInput{RepoName: repoName})
	if err != nil {
		t.Error(err)
	}
}

func TestVerifyTransform(t *testing.T) {
	var tests = []struct {
		srcSchema  []RepoSchemaEntry
		spec       *TransformSpec
		err        error
		destSchema []RepoSchemaEntry
	}{
		{
			srcSchema: []RepoSchemaEntry{
				RepoSchemaEntry{
					Key:       "f1",
					ValueType: "string",
					Required:  true,
				},
			},
			spec: &TransformSpec{
				Mode:      "sql",
				Code:      "select * from stream",
				Interval:  "5m",
				Container: defaultContainer,
			},
			destSchema: []RepoSchemaEntry{
				RepoSchemaEntry{
					Key:       "f1",
					ValueType: "string",
					Required:  true,
				},
			},
			err: nil,
		},
	}

	for i, test := range tests {
		input := &VerifyTransformInput{
			Schema: test.srcSchema,
			Spec:   test.spec,
		}
		output, err := client.VerifyTransform(input)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(output.Schema, test.destSchema) {
			t.Errorf("got schema is different to expected one, got: %#v, expected: %#v, index: %d", output.Schema, test.destSchema, i)
		}
	}
}

func TestVerifyExport(t *testing.T) {
	var tests = []struct {
		schema []RepoSchemaEntry
		spec   interface{}
		err    error
	}{
		{
			schema: []RepoSchemaEntry{
				RepoSchemaEntry{
					Key:       "f1",
					ValueType: "string",
					Required:  true,
				},
				RepoSchemaEntry{
					Key:       "f2",
					ValueType: "float",
					Required:  true,
				},
			},
			spec: &ExportTsdbSpec{
				DestRepoName: "repo",
				SeriesName:   "series",
				Tags:         map[string]string{"tag1": "#f1"},
				Fields:       map[string]string{"field1": "#f1", "field2": "#f2"},
				Timestamp:    "#f1",
			},
			err: nil,
		},
	}

	for i, test := range tests {
		input := &VerifyExportInput{
			Schema: test.schema,
			Spec:   test.spec,
		}
		err := client.VerifyExport(input)
		if err != test.err {
			t.Errorf("verify export %d failed: %v", i, err)
		}
	}
}
