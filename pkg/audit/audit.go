package audit

import (
	"context"
	"fmt"
	"github.com/olivere/elastic"
	"github.com/olivere/elastic/config"
	"time"
)

type AuditingInterface interface {
	ElkClient(url string, username string, password string) (*elastic.Client, error)
	DoAudit(ctx context.Context, client *elastic.Client, username string, operation string, values map[string]string)
}

type AuditingImpl struct {
}

//Document structure
type Document struct {
	username  string
	operation string
	CreatedAt time.Time
}

//ElkClient return a new ElkClient
func (a AuditingImpl) ElkClient(url string, username string, password string) (*elastic.Client, error) {

	config := &config.Config{
		URL:      url,
		Username: username,
		Password: password,
		Index:    "tenkai1audit",
	}

	elasticClient, err := elastic.NewClientFromConfig(config)
	if err != nil {
		return nil, err
	}

	return elasticClient, nil

}

//DoAudit create new audit log into elk
func (a AuditingImpl) DoAudit(ctx context.Context, client *elastic.Client, username string, operation string, values map[string]string) {

	if client != nil {

		bulk := client.Bulk().Index("tenkai1audit").Type("audit")

		doc := make(map[string]string)

		doc["username"] = username
		doc["operation"] = operation
		doc["createat"] = time.Now().UTC().Format("2006-01-02T15:04:05-0700")

		for k, v := range values {
			doc[k] = v
		}

		_, err := bulk.Add(elastic.NewBulkIndexRequest().Doc(doc)).Do(ctx)
		if err != nil {
			fmt.Println("error doing audit")
		}

	}

}
