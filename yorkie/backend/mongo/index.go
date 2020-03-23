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

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"

	"github.com/yorkie-team/yorkie/pkg/log"
)

var (
	ColClientInfos = "clients"
	idxClientInfos = []mongo.IndexModel{{
		Keys:    bsonx.Doc{{Key: "key", Value: bsonx.Int32(1)}},
		Options: options.Index().SetUnique(true),
	}}

	ColDocInfos = "documents"
	idxDocInfos = []mongo.IndexModel{{
		Keys:    bsonx.Doc{{Key: "key", Value: bsonx.Int32(1)}},
		Options: options.Index().SetUnique(true),
	}}

	ColChanges = "changes"
	idxChanges = []mongo.IndexModel{{
		Keys: bsonx.Doc{
			{Key: "doc_id", Value: bsonx.Int32(1)},
			{Key: "server_seq", Value: bsonx.Int32(1)},
		},
		Options: options.Index().SetUnique(true),
	}}

	ColSnapshots = "snapshots"
	idxSnapshots = []mongo.IndexModel{{
		Keys: bsonx.Doc{
			{Key: "doc_id", Value: bsonx.Int32(1)},
			{Key: "server_seq", Value: bsonx.Int32(1)},
		},
		Options: options.Index().SetUnique(true),
	}}
)

func ensureIndexes(ctx context.Context, db *mongo.Database) error {
	if _, err := db.Collection(ColClientInfos).Indexes().CreateMany(
		ctx,
		idxClientInfos,
	); err != nil {
		log.Logger.Error(err)
		return err
	}

	if _, err := db.Collection(ColDocInfos).Indexes().CreateMany(
		ctx,
		idxDocInfos,
	); err != nil {
		log.Logger.Error(err)
		return err
	}

	if _, err := db.Collection(ColChanges).Indexes().CreateMany(
		ctx,
		idxChanges,
	); err != nil {
		log.Logger.Error(err)
		return err
	}

	if _, err := db.Collection(ColSnapshots).Indexes().CreateMany(
		ctx,
		idxSnapshots,
	); err != nil {
		log.Logger.Error(err)
		return err
	}

	return nil
}
