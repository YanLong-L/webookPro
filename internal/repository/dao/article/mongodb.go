package article

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
	"webookpro/internal/domain"
)

type MongoDBArticleDAO struct {
	client  *mongo.Client
	col     *mongo.Collection // 代表制作库
	liveCol *mongo.Collection // 代表线上库
	node    *snowflake.Node
}

type MongoDBArticleDAOV1 struct {
	col     *mongo.Collection // 代表制作库
	liveCol *mongo.Collection // 代表线上库
	node    *snowflake.Node
	idGen   IDGenerator
}

func InitCollections(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	index := []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{bson.E{Key: "author_id", Value: 1},
				bson.E{Key: "ctime", Value: 1},
			},
			Options: options.Index(),
		},
	}
	_, err := db.Collection("articles").Indexes().
		CreateMany(ctx, index)
	if err != nil {
		return err
	}
	_, err = db.Collection("published_articles").Indexes().
		CreateMany(ctx, index)
	return err
}

func NewMongoDBArticleDAO(client *mongo.Client, db *mongo.Database, node *snowflake.Node) *MongoDBArticleDAO {
	return &MongoDBArticleDAO{
		client:  client,
		col:     db.Collection("articles"),
		liveCol: db.Collection("published_articles"),
		node:    node,
	}
}

func (m MongoDBArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	// 通过雪花算法解决mongo中article的id问题
	id := m.node.Generate().Int64()
	art.Id = id
	_, err := m.col.InsertOne(ctx, art)
	return id, err
}

func (m MongoDBArticleDAO) UpdateById(ctx context.Context, art Article) error {
	// 操作的是制作库
	filter := bson.M{"id": art.Id, "author_id": art.AuthorId}
	update := bson.D{bson.E{"$set", bson.M{
		"title":   art.Title,
		"content": art.Content,
		"utime":   time.Now().UnixMilli(),
		"status":  art.Status,
	}}}
	res, err := m.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("更新数据失败")
	}
	return nil
}

func (m MongoDBArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	// 在这里同步制作库和线上库没有办法做到类似事务的概念
	// 制作库新建或更新
	now := time.Now().UnixMilli()
	var (
		artId = art.Id
		err   error
	)
	if artId > 0 {
		// 更新到制作库
		err = m.UpdateById(ctx, art)
	} else {
		// 新建到制作库
		artId, err = m.Insert(ctx, art)
	}
	// upsert到线上库
	art.Utime = now
	update := bson.M{
		// 更新，如果不存在，就是插入，
		"$set": PublishedArticle{Article: art},
		// 在插入的时候，要插入 ctime
		"$setOnInsert": bson.M{"ctime": now},
	}
	filter := bson.M{"id": art.Id}
	_, err = m.liveCol.UpdateOne(ctx, filter, update,
		options.Update().SetUpsert(true))

	return artId, err
}

func (m MongoDBArticleDAO) Upsert(ctx context.Context, article PublishedArticle) error {
	//TODO implement me
	panic("implement me")
}

// SyncStatus 撤回帖子接口，同步制作库线上库帖子状态
func (m MongoDBArticleDAO) SyncStatus(ctx context.Context, art Article, status domain.ArticleStatus) error {
	// 开始事务
	session, err := m.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)
	err = session.StartTransaction()
	if err != nil {
		return err
	}
	filter := bson.M{"id": art.Id, "author_id": art.AuthorId}
	update := bson.D{bson.E{"$set", bson.M{
		"status": status.ToUint8(),
	}}}
	res, err := m.col.UpdateOne(ctx, filter, update)
	if err != nil || res.ModifiedCount == 0 {
		err = session.AbortTransaction(ctx) // 回滚事务
		if err != nil {
			return fmt.Errorf("更新到制作库失败 %#v ModifiedCount%d，回滚也失败",
				err, res.ModifiedCount)
		}
		return fmt.Errorf("更新到制作库失败 %#v ModifiedCount%d", err, res.ModifiedCount)
	}
	res, err = m.liveCol.UpdateOne(ctx, filter, update)
	if err != nil || res.ModifiedCount == 0 {
		err = session.AbortTransaction(ctx) // 回滚事务
		if err != nil {
			return fmt.Errorf("更新到线上库库失败 %#v ModifiedCount%d，回滚也失败",
				err, res.ModifiedCount)
		}
		return fmt.Errorf("更新到线上库失败 %#v ModifiedCount%d", err, res.ModifiedCount)
	}
	err = session.CommitTransaction(ctx)
	if err != nil {
		return errors.New("MongoDBArticleDAO SyncStatus 提交事务失败")

	}
	return err
}

type IDGenerator func() int64
