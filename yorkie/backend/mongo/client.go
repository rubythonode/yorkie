/*
 * Copyright 2020 The Yorkie Authors. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mongo

import (
	"context"
	"errors"
	"github.com/yorkie-team/yorkie/api/converter"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/yorkie-team/yorkie/pkg/document"
	"github.com/yorkie-team/yorkie/pkg/document/change"
	"github.com/yorkie-team/yorkie/pkg/log"
	"github.com/yorkie-team/yorkie/yorkie/types"
)

var (
	// ErrClientNotFound is returned when the client could not be found.
	ErrClientNotFound = errors.New("fail to find the client")

	// ErrDocumentNotFound is returned when the document could not be found.
	ErrDocumentNotFound = errors.New("fail to find the document")

	// ErrSnapshotNotFound is returned when the snapshot could not be found.
	ErrSnapshotNotFound = errors.New("fail to find the snapshot")
)

// Config is the configuration for creating a Client instance.
type Config struct {
	ConnectionTimeoutSec time.Duration `json:"ConnectionTimeOutSec"`
	ConnectionURI        string        `json:"ConnectionURI"`
	YorkieDatabase       string        `json:"YorkieDatabase"`
	PingTimeoutSec       time.Duration `json:"PingTimeoutSec"`
}

type Client struct {
	config *Config
	client *mongo.Client
}

func NewClient(conf *Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		conf.ConnectionTimeoutSec*time.Second,
	)
	defer cancel()

	client, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI(conf.ConnectionURI),
	)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}

	ctxPing, cancel := context.WithTimeout(ctx, conf.PingTimeoutSec*time.Second)
	defer cancel()

	if err := client.Ping(ctxPing, readpref.Primary()); err != nil {
		log.Logger.Errorf("fail to connect to %s in %d sec", conf.ConnectionURI, conf.PingTimeoutSec)
		return nil, err
	}

	if err := ensureIndexes(ctx, client.Database(conf.YorkieDatabase)); err != nil {
		log.Logger.Error(err)
		return nil, err
	}

	log.Logger.Infof("connected, URI: %s, DB: %s", conf.ConnectionURI, conf.YorkieDatabase)

	return &Client{
		config: conf,
		client: client,
	}, nil
}

func (c *Client) Close() error {
	if err := c.client.Disconnect(context.Background()); err != nil {
		log.Logger.Error(err)
		return err
	}

	return nil
}

func (c *Client) ActivateClient(ctx context.Context, key string) (*types.ClientInfo, error) {
	clientInfo := types.ClientInfo{}
	if err := c.withCollection(ColClientInfos, func(col *mongo.Collection) error {
		now := time.Now()
		res, err := col.UpdateOne(ctx, bson.M{
			"key": key,
		}, bson.M{
			"$set": bson.M{
				"status":     types.ClientActivated,
				"updated_at": now,
			},
		}, options.Update().SetUpsert(true))
		if err != nil {
			log.Logger.Error(err)
			return err
		}

		var result *mongo.SingleResult
		if res.UpsertedCount > 0 {
			result = col.FindOneAndUpdate(ctx, bson.M{
				"_id": res.UpsertedID,
			}, bson.M{
				"$set": bson.M{
					"created_at": now,
				},
			})
		} else {
			result = col.FindOne(ctx, bson.M{
				"key": key,
			})
		}

		if err := result.Decode(&clientInfo); err != nil {
			log.Logger.Error(err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &clientInfo, nil
}

func (c *Client) DeactivateClient(ctx context.Context, clientID string) (*types.ClientInfo, error) {
	clientInfo := types.ClientInfo{}
	if err := c.withCollection(ColClientInfos, func(col *mongo.Collection) error {
		id, err := primitive.ObjectIDFromHex(clientID)
		if err != nil {
			log.Logger.Error(err)
			return err
		}
		res := col.FindOneAndUpdate(ctx, bson.M{
			"_id": id,
		}, bson.M{
			"$set": bson.M{
				"status":     types.ClientDeactivated,
				"updated_at": time.Now(),
			},
		})

		if err := res.Decode(&clientInfo); err != nil {
			if err == mongo.ErrNoDocuments {
				return ErrClientNotFound
			}

			log.Logger.Error(err)
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &clientInfo, nil
}

func (c *Client) FindClientInfoByID(ctx context.Context, clientID string) (*types.ClientInfo, error) {
	var client types.ClientInfo

	if err := c.withCollection(ColClientInfos, func(col *mongo.Collection) error {
		id, err := primitive.ObjectIDFromHex(clientID)
		if err != nil {
			log.Logger.Error(err)
			return err
		}
		result := col.FindOne(ctx, bson.M{
			"_id": id,
		})

		if err := result.Decode(&client); err != nil {
			if err == mongo.ErrNoDocuments {
				return ErrClientNotFound
			}
			log.Logger.Error(err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &client, nil
}

func (c *Client) UpdateClientInfoAfterPushPull(
	ctx context.Context,
	clientInfo *types.ClientInfo,
	docInfo *types.DocInfo,
) error {
	return c.withCollection(ColClientInfos, func(col *mongo.Collection) error {
		result := col.FindOneAndUpdate(ctx, bson.M{
			"key": clientInfo.Key,
		}, bson.M{
			"$set": bson.M{
				"documents." + docInfo.ID.Hex(): clientInfo.Documents[docInfo.ID.Hex()],
				"updated_at":                    clientInfo.UpdatedAt,
			},
		})

		if result.Err() != nil {
			if result.Err() == mongo.ErrNoDocuments {
				return ErrClientNotFound
			}
			log.Logger.Error(result.Err())
			return result.Err()
		}

		return nil
	})
}

func (c *Client) FindDocInfoByKey(
	ctx context.Context,
	clientInfo *types.ClientInfo,
	bsonDocKey string,
	createDocIfNotExist bool,
) (*types.DocInfo, error) {
	docInfo := types.DocInfo{}

	if err := c.withCollection(ColDocInfos, func(col *mongo.Collection) error {
		now := time.Now()
		res, err := col.UpdateOne(ctx, bson.M{
			"key": bsonDocKey,
		}, bson.M{
			"$set": bson.M{
				"accessed_at": now,
			},
		}, options.Update().SetUpsert(createDocIfNotExist))
		if err != nil {
			log.Logger.Error(err)
			return err
		}

		var result *mongo.SingleResult
		if res.UpsertedCount > 0 {
			result = col.FindOneAndUpdate(ctx, bson.M{
				"_id": res.UpsertedID,
			}, bson.M{
				"$set": bson.M{
					"owner":      clientInfo.ID,
					"created_at": now,
				},
			})
		} else {
			result = col.FindOne(ctx, bson.M{
				"key": bsonDocKey,
			})
			if result.Err() == mongo.ErrNoDocuments {
				return ErrDocumentNotFound
			}
			if result.Err() != nil {
				log.Logger.Error(result.Err())
				return result.Err()
			}
		}

		if err := result.Decode(&docInfo); err != nil {
			log.Logger.Error(err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &docInfo, nil
}

func (c *Client) CreateChangeInfos(
	ctx context.Context,
	docID primitive.ObjectID,
	changes []*change.Change,
) error {
	if len(changes) == 0 {
		return nil
	}

	return c.withCollection(ColChanges, func(col *mongo.Collection) error {
		var modelChanges []mongo.WriteModel

		for _, c := range changes {
			modelChanges = append(modelChanges, mongo.NewUpdateOneModel().SetFilter(bson.M{
				"doc_id":     docID,
				"server_seq": c.ServerSeq(),
			}).SetUpdate(bson.M{"$set": bson.M{
				"actor":      types.EncodeActorID(c.ID().Actor()),
				"client_seq": c.ID().ClientSeq(),
				"lamport":    c.ID().Lamport(),
				"message":    c.Message(),
				"operations": types.EncodeOperation(c.Operations()),
			}}).SetUpsert(true))
		}

		_, err := col.BulkWrite(ctx, modelChanges, options.BulkWrite().SetOrdered(true))
		if err != nil {
			log.Logger.Error(err)
			return err
		}

		return nil
	})
}

func (c *Client) CreateSnapshotInfo(
	ctx context.Context,
	docID primitive.ObjectID,
	doc *document.Document,
) error {
	return c.withCollection(ColSnapshots, func(col *mongo.Collection) error {
		if _, err := col.InsertOne(ctx, bson.M{
			"doc_id": docID,
			"server_seq": doc.Checkpoint().ServerSeq,
			"snapshot": converter.ObjectToBytes(doc.RootObject()),
			"created_at": time.Now(),
		}); err != nil {
			log.Logger.Error(err)
			return err
		}

		return nil
	})
}

func (c *Client) UpdateDocInfo(
	ctx context.Context,
	docInfo *types.DocInfo,
) error {
	return c.withCollection(ColDocInfos, func(col *mongo.Collection) error {
		now := time.Now()
		_, err := col.UpdateOne(ctx, bson.M{
			"_id": docInfo.ID,
		}, bson.M{
			"$set": bson.M{
				"server_seq": docInfo.ServerSeq,
				"updated_at": now,
			},
		})

		if err != nil {
			if err == mongo.ErrNoDocuments {
				return ErrDocumentNotFound
			}

			log.Logger.Error(err)
			return err
		}

		return nil
	})
}

func (c *Client) FindChangeInfosBetweenServerSeqs(
	ctx context.Context,
	docID primitive.ObjectID,
	from uint64,
	to uint64,
) ([]*change.Change, error) {
	var changes []*change.Change

	if err := c.withCollection(ColChanges, func(col *mongo.Collection) error {
		cursor, err := col.Find(ctx, bson.M{
			"doc_id": docID,
			"server_seq": bson.M{
				"$gte": from,
				"$lte": to,
			},
		}, options.Find())
		if err != nil {
			log.Logger.Error(err)
			return err
		}

		defer func() {
			if err := cursor.Close(ctx); err != nil {
				log.Logger.Error(err)
			}
		}()

		for cursor.Next(ctx) {
			var changeInfo types.ChangeInfo
			if err := cursor.Decode(&changeInfo); err != nil {
				log.Logger.Error(err)
				return err
			}

			c, err := changeInfo.ToChange()
			if err != nil {
				return err
			}
			changes = append(changes, c)
		}

		if cursor.Err() != nil {
			log.Logger.Error(cursor.Err())
			return cursor.Err()
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return changes, nil
}

func (c *Client) withCollection(
	collection string,
	callback func(collection *mongo.Collection) error,
) error {
	col := c.client.Database(c.config.YorkieDatabase).Collection(collection)
	return callback(col)
}

func (c *Client) FindLastSnapshotInfo(
	ctx context.Context,
	docID primitive.ObjectID,
) (*types.SnapshotInfo, error) {
	snapshotInfo := types.SnapshotInfo{}

	if err := c.withCollection(ColSnapshots, func(col *mongo.Collection) error {
		result := col.FindOne(ctx, bson.M{
			"doc_id": docID,
		}, options.FindOne().SetSort(bson.M{
			"server_seq": 1,
		}))
		if result.Err() == mongo.ErrNoDocuments {
			return nil
		}
		if result.Err() != nil {
			log.Logger.Error(result.Err())
			return result.Err()
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &snapshotInfo, nil
}
