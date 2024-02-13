package post

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Post struct {
	PK             string    `dynamodbav:"PK" json:"-"`
	SK             string    `dynamodbav:"SK" json:"-"`
	GSI1PK         string    `dynamodbav:"GSI1PK" json:"-"`
	GSI1SK         string    `dynamodbav:"GSI1SK" json:"-"`
	ID             string    `dynamodbav:"id" json:"id"`
	Caption        string    `dynamodbav:"caption" json:"caption"`
	Image          string    `dynamodbav:"image" json:"image"`
	CreatedAt      time.Time `dynamodbav:"created,omitempty" json:"created_at"`
	Creator        string    `dynamodbav:"creator" json:"creator"`
	LatestComments []comment `dynamodbav:"latest_comments" json:"latest_comments"`
	TotalComments  int       `dynamodbav:"total_comments" json:"total_comments"`
}

type comment struct {
	ID        string    `dynamodbav:"id" json:"id"`
	Content   string    `dynamodbav:"content" json:"content"`
	CreatedAt time.Time `dynamodbav:"created,omitempty" json:"created_at"`
	Creator   string    `dynamodbav:"creator" json:"creator"`
}

func FromItem(item map[string]*dynamodb.AttributeValue) (Post, error) {
	e := Post{}
	err := dynamodbattribute.UnmarshalMap(item, &e)
	if err != nil {
		return e, err
	}
	return e, nil
}

func (p *Post) ToItem() (map[string]*dynamodb.AttributeValue, error) {
	p.generateKey()
	p.generateGSI1Key()
	av, err := dynamodbattribute.MarshalMap(p)

	return av, err
}

func (p *Post) ToKey() map[string]*dynamodb.AttributeValue {
	p.generateKey()
	key := map[string]*dynamodb.AttributeValue{
		"PK": {
			S: aws.String(p.PK),
		},
		"SK": {
			S: aws.String(p.SK),
		},
	}
	return key
}
func (p *Post) generateKey() {
	p.PK = "USER#" + p.Creator
	p.SK = "POST#" + p.ID
}

func (p *Post) generateGSI1Key() {
	paddedInt := fmt.Sprintf("%010d", p.TotalComments)
	p.GSI1PK = "POST"
	p.GSI1SK = fmt.Sprintf("%v#POST#%v", paddedInt, p.ID)
}
