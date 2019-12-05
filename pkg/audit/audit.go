package audit

import (
	"context"
	"github.com/olivere/elastic"
	"github.com/olivere/elastic/config"
	"time"
)

//AuditingInterface AuditingInterface
type AuditingInterface interface {
	ElkClient(url string, username string, password string) (*elastic.Client, error)
	DoAudit(ctx context.Context, client *elastic.Client, username string, operation string, values map[string]string)
}

//ElkInterface ElkInterface
type ElkInterface interface {
	NewClient(config *config.Config) (*elastic.Client, error)
}

//ElkImpl ElkImpl
type ElkImpl struct {
}

//AuditingImpl AuditingImpl
type AuditingImpl struct {
	Elk ElkInterface
}

//Document structure
type Document struct {
	username  string
	operation string
	CreatedAt time.Time
}

//AuditingBuilder AuditingBuilder
func AuditingBuilder() AuditingInterface {
	result := &AuditingImpl{}
	result.Elk = &ElkImpl{}
	return result
}

//NewClient NewClient
func (a ElkImpl) NewClient(config *config.Config) (*elastic.Client, error) {
	elasticClient, err := elastic.NewClientFromConfig(config)
	if err != nil {
		return nil, err
	}
	return elasticClient, nil
}

//ElkClient return a new ElkClient
func (a AuditingImpl) ElkClient(url string, username string, password string) (*elastic.Client, error) {
	config := buildConfig(url, username, password)
	return a.Elk.NewClient(config)
}

//DoAudit create new audit log into elk
func (a AuditingImpl) DoAudit(ctx context.Context, client *elastic.Client, username string, operation string, values map[string]string) {
	if client != nil {
		bulk := client.Bulk().Index("tenkai1audit").Type("audit")
		doc := buildDoc(username, operation, values)
		bulk.Add(elastic.NewBulkIndexRequest().Doc(doc)).Do(ctx)
	}
}

func buildDoc(username string, operation string, values map[string]string) map[string]string {
	doc := make(map[string]string)
	doc["username"] = username
	doc["operation"] = operation
	doc["createat"] = time.Now().UTC().Format("2006-01-02T15:04:05-0700")
	for k, v := range values {
		doc[k] = v
	}
	return doc
}

func buildConfig(url string, username string, password string) *config.Config {
	config := &config.Config{
		URL:      url,
		Username: username,
		Password: password,
		Index:    "tenkai1audit",
	}
	return config
}
