package pipeline

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	. "qiniu.com/pandora/base"
	"qiniu.com/pandora/base/config"
	"qiniu.com/pandora/base/request"
)

var builder errBuilder

type Pipeline struct {
	Config     *config.Config
	HTTPClient *http.Client
}

func NewConfig() *config.Config {
	return config.NewConfig()
}

func New(c *config.Config) (PipelineAPI, error) {
	return newClient(c)
}

func newClient(c *config.Config) (p *Pipeline, err error) {
	if !strings.HasPrefix(c.Endpoint, "http://") && !strings.HasPrefix(c.Endpoint, "https://") {
		err = fmt.Errorf("endpoint should start with 'http://' or 'https://'")
		return
	}
	if strings.HasSuffix(c.Endpoint, "/") {
		err = fmt.Errorf("endpoint should not end with '/'")
		return
	}

	var t = &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   c.DialTimeout,
			KeepAlive: 30 * time.Second,
		}).Dial,
		ResponseHeaderTimeout: c.ResponseTimeout,
	}

	p = &Pipeline{
		Config:     c,
		HTTPClient: &http.Client{Transport: t},
	}

	return
}

func (c *Pipeline) newRequest(op *request.Operation, token string, v interface{}) *request.Request {
	req := request.New(c.Config, c.HTTPClient, op, token, builder, v)
	req.Data = v
	return req
}

func (c *Pipeline) newOperation(opName string, args ...interface{}) *request.Operation {
	var method, urlTmpl string
	switch opName {
	case OpCreateGroup:
		method, urlTmpl = MethodPost, "/v2/groups/%s"
	case OpUpdateGroup:
		method, urlTmpl = MethodPut, "/v2/groups/%s"
	case OpStartGroupTask:
		method, urlTmpl = MethodPost, "/v2/groups/%s/actions/start"
	case OpStopGroupTask:
		method, urlTmpl = MethodPost, "/v2/groups/%s/actions/stop"
	case OpListGroups:
		method, urlTmpl = MethodGet, "/v2/groups"
	case OpGetGroup:
		method, urlTmpl = MethodGet, "/v2/groups/%s"
	case OpDeleteGroup:
		method, urlTmpl = MethodDelete, "/v2/groups/%s"
	case OpCreateRepo:
		method, urlTmpl = MethodPost, "/v2/repos/%s"
	case OpListRepos:
		method, urlTmpl = MethodGet, "/v2/repos"
	case OpGetRepo:
		method, urlTmpl = MethodGet, "/v2/repos/%s"
	case OpDeleteRepo:
		method, urlTmpl = MethodDelete, "/v2/repos/%s"
	case OpPostData:
		method, urlTmpl = MethodPost, "/v2/repos/%s/data"
	case OpCreateTransform:
		method, urlTmpl = MethodPost, "/v2/repos/%s/transforms/%s/to/%s"
	case OpListTransforms:
		method, urlTmpl = MethodGet, "/v2/repos/%s/transforms"
	case OpGetTransform:
		method, urlTmpl = MethodGet, "/v2/repos/%s/transforms/%s"
	case OpDeleteTransform:
		method, urlTmpl = MethodDelete, "/v2/repos/%s/transforms/%s"
	case OpCreateExport:
		method, urlTmpl = MethodPost, "/v2/repos/%s/exports/%s"
	case OpListExports:
		method, urlTmpl = MethodGet, "/v2/repos/%s/exports"
	case OpGetExport:
		method, urlTmpl = MethodGet, "/v2/repos/%s/exports/%s"
	case OpDeleteExport:
		method, urlTmpl = MethodDelete, "/v2/repos/%s/exports/%s?delOffset=%s"
	case OpUploadPlugin:
		method, urlTmpl = MethodPost, "/v2/plugins/%s"
	case OpListPlugins:
		method, urlTmpl = MethodGet, "/v2/plugins"
	case OpGetPlugin:
		method, urlTmpl = MethodGet, "/v2/plugins/%s"
	case OpDeletePlugin:
		method, urlTmpl = MethodDelete, "/v2/plugins/%s"
	case OpVerifyTransform:
		method, urlTmpl = MethodPost, "/v2/verify/transform"
	case OpVerifyExport:
		method, urlTmpl = MethodPost, "/v2/verify/export"
	default:
		c.Config.Logger.Errorf("unmatched operation name: %s", opName)
		return nil
	}

	return &request.Operation{
		Name:   opName,
		Method: method,
		Path:   fmt.Sprintf(urlTmpl, args...),
	}
}
