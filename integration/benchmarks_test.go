// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/v2/internal/util/xiter"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
	"github.com/FerretDB/FerretDB/v2/integration/shareddata"
)

func BenchmarkFind(b *testing.B) {
	provider := shareddata.BenchmarkSmallDocuments

	b.Run(provider.Name(), func(b *testing.B) {
		s := setup.SetupWithOpts(b, &setup.SetupOpts{
			BenchmarkProvider: provider,
		})

		for name, bc := range map[string]struct {
			filter bson.D
		}{
			"Int32ID": {
				filter: bson.D{{"_id", int32(42)}},
			},
			"Int32One": {
				filter: bson.D{{"id", int32(42)}},
			},
			"Int32Many": {
				filter: bson.D{{"v", int32(42)}},
			},
			"Int32ManyDotNotation": {
				filter: bson.D{{"v.foo", int32(42)}},
			},
		} {
			b.Run(name, func(b *testing.B) {
				var firstDocs, docs int

				for range b.N {
					cursor, err := s.Collection.Find(s.Ctx, bc.filter)
					require.NoError(b, err)

					docs = 0
					for cursor.Next(s.Ctx) {
						docs++
					}

					require.NoError(b, cursor.Close(s.Ctx))
					require.NoError(b, cursor.Err())
					require.Positive(b, docs)

					if firstDocs == 0 {
						firstDocs = docs
					}
				}

				b.StopTimer()

				require.Equal(b, firstDocs, docs)

				b.ReportMetric(float64(docs), "docs-returned")
			})
		}
	})
}

func BenchmarkReplaceOne(b *testing.B) {
	provider := shareddata.BenchmarkSettingsDocuments

	s := setup.SetupWithOpts(b, &setup.SetupOpts{
		BenchmarkProvider: provider,
	})
	ctx, collection := s.Ctx, s.Collection

	cursor, err := collection.Find(ctx, bson.D{})
	require.NoError(b, err)

	var lastRaw bson.Raw
	for cursor.Next(ctx) {
		lastRaw = cursor.Current
	}
	require.NoError(b, cursor.Err())
	require.NoError(b, cursor.Close(ctx))

	var doc bson.D
	require.NoError(b, bson.Unmarshal(lastRaw, &doc))
	require.Equal(b, "_id", doc[0].Key)
	require.NotEmpty(b, doc[0].Value)
	require.NotZero(b, doc[1].Value)

	b.Run(provider.Name(), func(b *testing.B) {
		filter := bson.D{{"_id", doc[0].Value}}
		var res *mongo.UpdateResult

		for i := range b.N {
			doc[1].Value = int64(i + 1)

			res, err = collection.ReplaceOne(ctx, filter, doc)
			require.NoError(b, err)
			require.Equal(b, int64(1), res.MatchedCount)
			require.Equal(b, int64(1), res.ModifiedCount)
		}

		b.StopTimer()

		var actual bson.D
		err = collection.FindOne(ctx, filter).Decode(&actual)
		require.NoError(b, err)
		AssertEqualDocuments(b, doc, actual)
	})
}

func BenchmarkInsertMany(b *testing.B) {
	ctx, collection := setup.Setup(b)

	for _, provider := range shareddata.AllBenchmarkProviders() {
		var total int
		for range provider.NewIter() {
			total++
		}

		var batchSizes []int
		for _, batchSize := range []int{1, 10, 100, 1000} {
			if batchSize <= total {
				batchSizes = append(batchSizes, batchSize)
			}
		}

		for _, batchSize := range batchSizes {
			b.Run(fmt.Sprintf("%s/Batch%d", provider.Name(), batchSize), func(b *testing.B) {
				b.StopTimer()

				for range b.N {
					require.NoError(b, collection.Drop(ctx))

					for docs := range xiter.Chunk(provider.NewIter(), batchSize) {
						insertDocs := make([]any, len(docs))
						for i := range insertDocs {
							insertDocs[i] = docs[i]
						}

						b.StartTimer()

						_, err := collection.InsertMany(ctx, insertDocs)
						require.NoError(b, err)

						b.StopTimer()
					}
				}
			})
		}
	}
}
