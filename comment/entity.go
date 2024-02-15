package comment

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Comment struct {
	PK string `dynamodbav:"PK" json:"-"`
	SK string `dynamodbav:"SK" json:"-"`
	// GSI1PK    string    `dynamodbav:"GSI1PK" json:"-"` // No need for comments at this point.
	// GSI1SK    string    `dynamodbav:"GSI1SK" json:"-"`
	ID        string    `dynamodbav:"id" json:"id"`
	PostID    string    `dynamodbav:"post_id" json:"post_id"`
	Content   string    `dynamodbav:"content" json:"content"`
	CreatedAt time.Time `dynamodbav:"created,omitempty" json:"created_at"`
	Creator   string    `dynamodbav:"creator" json:"creator"`
}

func FromItem(item map[string]*dynamodb.AttributeValue) (Comment, error) {
	e := Comment{}
	err := dynamodbattribute.UnmarshalMap(item, &e)
	if err != nil {
		return e, err
	}
	return e, nil
}

func (c *Comment) ToItem() (map[string]*dynamodb.AttributeValue, error) {
	c.generateKey()
	av, err := dynamodbattribute.MarshalMap(c)

	return av, err
}

func (c *Comment) ToKey() map[string]*dynamodb.AttributeValue {
	c.generateKey()
	key := map[string]*dynamodb.AttributeValue{
		"PK": {
			S: aws.String(c.PK),
		},
		"SK": {
			S: aws.String(c.SK),
		},
	}
	return key
}

func FromKey(pk string, sk string) Comment {
	// parse pk
	pkSplit := strings.Split(pk, "#")
	skSplit := strings.Split(sk, "#")
	return Comment{
		PK:     pk,
		SK:     sk,
		PostID: pkSplit[1],
		ID:     skSplit[1],
	}
}

func (c *Comment) generateKey() {
	c.PK = fmt.Sprintf("%s#%s", GetPKPrefix(), c.PostID)
	c.SK = fmt.Sprintf("%s#%s", GetSKPrefix(), c.ID)
}

func GetPKPrefix() string {
	return "POST"
}

func GetSKPrefix() string {
	return "COMMENT"
}
