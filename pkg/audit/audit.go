package audit

import (
	"context"
	"time"

	"github.com/olivere/elastic"
	"github.com/softplan/tenkai-api/pkg/global"
)

//AuditingInterface AuditingInterface
type AuditingInterface interface {
	ElkClient(url string, username string, password string) (*elastic.Client, error)
	DoAudit(ctx context.Context, client *elastic.Client, username string, operation string, values map[string]string)
}

//ElkInterface ElkInterface
type ElkInterface interface {
	NewClient(url string, username string, password string) (*elastic.Client, error)
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
func (a ElkImpl) NewClient(url string, username string, password string) (*elastic.Client, error) {
	elasticClient, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetURL(url),
		elastic.SetBasicAuth(username, password),
	)

	if err != nil {
		return nil, err
	}
	return elasticClient, nil
}

//ElkClient return a new ElkClient
func (a AuditingImpl) ElkClient(url string, username string, password string) (*elastic.Client, error) {
	return a.Elk.NewClient(url, username, password)

}

//DoAudit create new audit log into elk
func (a AuditingImpl) DoAudit(ctx context.Context, client *elastic.Client, username string, operation string, values map[string]string) {
	if client != nil {
		bulk := client.Bulk().Index("tenkai1audit").Type("audit")
		doc := buildDoc(username, operation, values)
		if _, err := bulk.Add(elastic.NewBulkIndexRequest().Doc(doc)).Do(ctx); err != nil {
			global.Logger.Error(global.AppFields{global.Function: "DoAudit"}, err.Error())
		}
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
