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
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
	"github.com/FerretDB/FerretDB/v2/integration/shareddata"
)

func TestQueryBadFindType(t *testing.T) {
	t.Parallel()
	s := setup.SetupWithOpts(t, nil)

	ctx, collection := s.Ctx, s.Collection

	for name, tc := range map[string]struct {
		value      any
		err        *mongo.CommandError
		altMessage string
	}{
		"Document": {
			value: bson.D{},
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Failed to parse namespace element",
			},
			altMessage: "collection name has invalid type object",
		},
		"Array": {
			value: primitive.A{},
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Failed to parse namespace element",
			},
			altMessage: "collection name has invalid type array",
		},
		"Double": {
			value: 3.14,
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Failed to parse namespace element",
			},
			altMessage: "collection name has invalid type double",
		},
		"Binary": {
			value: primitive.Binary{},
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Failed to parse namespace element",
			},
			altMessage: "collection name has invalid type binData",
		},
		"ObjectID": {
			value: primitive.ObjectID{},
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Failed to parse namespace element",
			},
			altMessage: "collection name has invalid type objectId",
		},
		"Bool": {
			value: true,
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Failed to parse namespace element",
			},
			altMessage: "collection name has invalid type bool",
		},
		"Date": {
			value: time.Now(),
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Failed to parse namespace element",
			},
			altMessage: "collection name has invalid type date",
		},
		"Null": {
			value: nil,
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Failed to parse namespace element",
			},
			altMessage: "collection name has invalid type null",
		},
		"Regex": {
			value: primitive.Regex{Pattern: "/foo/"},
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Failed to parse namespace element",
			},
			altMessage: "collection name has invalid type regex",
		},
		"Int": {
			value: int32(42),
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Failed to parse namespace element",
			},
			altMessage: "collection name has invalid type int",
		},
		"Timestamp": {
			value: primitive.Timestamp{},
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Failed to parse namespace element",
			},
			altMessage: "collection name has invalid type timestamp",
		},
		"Long": {
			value: int64(42),
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Failed to parse namespace element",
			},
			altMessage: "collection name has invalid type long",
		},
	} {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB-DocumentDB/issues/241")
			require.NotNil(t, tc.err, "err must not be nil")

			cmd := bson.D{
				{"find", tc.value},
			}

			var res bson.D
			err := collection.Database().RunCommand(ctx, cmd).Decode(&res)

			require.Nil(t, res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestQuerySortErrors(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		command bson.D // required, command to run

		err              *mongo.CommandError // required
		altMessage       string              // optional, alternative error message
		failsForFerretDB string
	}{
		"SortTypeDouble": {
			command: bson.D{
				{"find", collection.Name()},
				{"projection", bson.D{{"v", "some"}}},
				{"sort", 42.13},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "Expected field sortto be of type object",
			},
			altMessage: "BSON field 'sort' is the wrong type 'double', expected type 'object'",
		},
		"SortTypeString": {
			command: bson.D{
				{"find", collection.Name()},
				{"projection", bson.D{{"v", "some"}}},
				{"sort", "123"},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "Expected field sortto be of type object",
			},
			altMessage: "BSON field 'sort' is the wrong type 'string', expected type 'object'",
		},
		"SortStringValue": {
			command: bson.D{
				{"find", collection.Name()},
				{"projection", bson.D{{"v", 42}}},
				{"sort", bson.D{{"asc", "123"}}},
			},
			err: &mongo.CommandError{
				Code:    15974,
				Name:    "Location15974",
				Message: `Illegal key in $sort specification: asc: "123"`,
			},
			altMessage:       `Illegal key in $sort specification: asc: 123`,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/241",
		},
		"DoubleValue": {
			command: bson.D{
				{"find", collection.Name()},
				{"projection", bson.D{{"v", 42}}},
				{"sort", bson.D{{"asc", 42.12}}},
			},
			err: &mongo.CommandError{
				Code:    15975,
				Name:    "Location15975",
				Message: `$sort key ordering must be 1 (for ascending) or -1 (for descending)`,
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/241",
		},
		"IncorrectIntValue": {
			command: bson.D{
				{"find", collection.Name()},
				{"projection", bson.D{{"v", 42}}},
				{"sort", bson.D{{"asc", int32(12)}}},
			},
			err: &mongo.CommandError{
				Code:    15975,
				Name:    "Location15975",
				Message: `$sort key ordering must be 1 (for ascending) or -1 (for descending)`,
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/241",
		},
		"ExceedIntValue": {
			command: bson.D{
				{"find", collection.Name()},
				{"projection", bson.D{{"v", 42}}},
				{"sort", bson.D{{"asc", int64(math.MaxInt64)}}},
			},
			err: &mongo.CommandError{
				Code:    15975,
				Name:    "Location15975",
				Message: `$sort key ordering must be 1 (for ascending) or -1 (for descending)`,
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/241",
		},
	} {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testing.TB = tt

			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
			}

			require.NotNil(t, tc.command, "command must not be nil")
			require.NotNil(t, tc.err, "err must not be nil")

			var res bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&res)

			assert.Nil(t, res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestQueryMaxTimeMSErrors(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct {
		command bson.D // required, command to run

		err        *mongo.CommandError // required, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
	}{
		"BadMaxTimeMSTypeDouble": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", 43.15},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "maxTimeMS has non-integral value",
			},
			altMessage: "maxTimeMS has non-integral value",
		},
		"BadMaxTimeMSNegativeDouble": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", -14245345234123245.55},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "-14245345234123246 value for maxTimeMS is out of range " + shareddata.Int32Interval,
			},
			altMessage: "-14245345234123246 value for maxTimeMS is out of range",
		},
		"BadMaxTimeMSTypeString": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", "string"},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "maxTimeMS must be a number",
			},
		},
		"BadMaxTimeMSMaxInt64": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", math.MaxInt64},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "9223372036854775807 value for maxTimeMS is out of range " + shareddata.Int32Interval,
			},
			altMessage: "9223372036854775807 value for maxTimeMS is out of range",
		},
		"BadMaxTimeMSMinInt64": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", math.MinInt64},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "-9223372036854775808 value for maxTimeMS is out of range " + shareddata.Int32Interval,
			},
			altMessage: "-9223372036854775808 value for maxTimeMS is out of range",
		},
		"BadMaxTimeMSNull": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", nil},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "maxTimeMS must be a number",
			},
		},
		"BadMaxTimeMSArray": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", bson.A{int32(42), "foo", nil}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "maxTimeMS must be a number",
			},
		},
		"BadMaxTimeMSDocument": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "maxTimeMS must be a number",
			},
		},
		"BadMaxTimeMSTypeNegativeInt32": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", -1123123},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "-1123123 value for maxTimeMS is out of range " + shareddata.Int32Interval,
			},
			altMessage: "-1123123 value for maxTimeMS is out of range",
		},
	} {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB-DocumentDB/issues/243")

			require.NotNil(t, tc.command, "command must not be nil")
			require.NotNil(t, tc.err, "err must not be nil")

			var res bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&res)

			assert.Nil(t, res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestQueryMaxTimeMSAvailableValues(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		command any
	}{
		"Double": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", float64(10000)},
			},
		},
		"DoubleZero": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", float64(0)},
			},
		},
		"Int32": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", int32(10000)},
			},
		},
		"Int32Zero": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", int32(0)},
			},
		},
		"Int64": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", int64(10000)},
			},
		},
		"Int64Zero": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", int64(0)},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			require.NoError(t, err)
		})
	}
}

func TestQueryExactMatches(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{
			{"_id", "document-two-fields"},
			{"foo", "bar"},
			{"baz", int32(42)},
		},
		bson.D{
			{"_id", "document-value-two-fields"},
			{"v", bson.D{{"foo", "bar"}, {"baz", int32(42)}}},
		},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
	}{
		"Document": {
			filter:      bson.D{{"foo", "bar"}, {"baz", int32(42)}},
			expectedIDs: []any{"document-two-fields"},
		},
		"DocumentChangedFieldsOrder": {
			filter:      bson.D{{"baz", int32(42)}, {"foo", "bar"}},
			expectedIDs: []any{"document-two-fields"},
		},
		"DocumentValueFields": {
			filter:      bson.D{{"v", bson.D{{"foo", "bar"}, {"baz", int32(42)}}}},
			expectedIDs: []any{"document-value-two-fields"},
		},

		"Array": {
			filter:      bson.D{{"v", bson.A{int32(42), "foo", nil}}},
			expectedIDs: []any{"array-three"},
		},
		"ArrayChangedOrder": {
			filter:      bson.D{{"v", bson.A{int32(42), nil, "foo"}}},
			expectedIDs: []any{},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}

func TestDotNotation(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	_, err := collection.InsertMany(
		ctx,
		[]any{
			bson.D{
				{"_id", "document-deeply-nested"},
				{
					"foo",
					bson.D{
						{
							"bar",
							bson.D{{
								"baz",
								bson.D{{"qux", bson.D{{"quz", int32(42)}}}},
							}},
						},
						{
							"qaz",
							bson.A{bson.D{{"baz", int32(1)}}},
						},
					},
				},
				{
					"wsx",
					bson.A{bson.D{{"edc", bson.A{bson.D{{"rfv", int32(1)}}}}}},
				},
			},
			bson.D{{"_id", bson.D{{"foo", "bar"}}}},
		},
	)
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
	}{
		"DeeplyNested": {
			filter:      bson.D{{"foo.bar.baz.qux.quz", int32(42)}},
			expectedIDs: []any{"document-deeply-nested"},
		},
		"DottedField": {
			filter:      bson.D{{"foo.bar.baz", bson.D{{"qux.quz", int32(42)}}}},
			expectedIDs: []any{},
		},
		"FieldArrayField": {
			filter:      bson.D{{"foo.qaz.0.baz", int32(1)}},
			expectedIDs: []any{"document-deeply-nested"},
		},
		"FieldArrayFieldArrayField": {
			filter:      bson.D{{"wsx.0.edc.0.rfv", int32(1)}},
			expectedIDs: []any{"document-deeply-nested"},
		},
		"FieldID": {
			filter:      bson.D{{"_id.foo", "bar"}},
			expectedIDs: []any{bson.D{{"foo", "bar"}}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}

// TestQueryNonExistingCollection tests that a query to a non existing collection doesn't fail but returns an empty result.
func TestQueryNonExistingCollection(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	cursor, err := collection.Database().Collection("doesnotexist").Find(ctx, bson.D{})
	require.NoError(t, err)

	var actual []bson.D
	err = cursor.All(ctx, &actual)
	require.NoError(t, err)
	require.Len(t, actual, 0)
}

// TestQueryIDDoc checks that the order of fields in the _id document matters.
func TestQueryIDDoc(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	_, err := collection.InsertOne(ctx, bson.D{
		{"_id", bson.D{{"a", int32(1)}, {"z", int32(2)}}},
		{"v", int32(1)},
	})
	require.NoError(t, err)
	_, err = collection.InsertOne(ctx, bson.D{
		{"_id", bson.D{{"a", int32(3)}, {"z", int32(4)}}},
		{"v", int32(2)},
	})
	require.NoError(t, err)

	expected := []bson.D{{
		{"_id", bson.D{{"a", int32(3)}, {"z", int32(4)}}},
		{"v", int32(2)},
	}}
	actual := FilterAll(t, ctx, collection, bson.D{{"_id", bson.D{{"a", int32(3)}, {"z", int32(4)}}}})
	AssertEqualDocumentsSlice(t, expected, actual)

	expected = []bson.D{}
	actual = FilterAll(t, ctx, collection, bson.D{{"_id", bson.D{{"z", int32(4)}, {"a", int32(3)}}}})
	AssertEqualDocumentsSlice(t, expected, actual)
}

func TestQueryShowRecordID(t *testing.T) {
	t.Parallel()

	provider := shareddata.Scalars
	ctx, collection := setup.Setup(t, provider)

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		collection   *mongo.Collection
		showRecordID bool

		nonZeroRecordID  bool // if true, asserts recordID is not zero
		failsForFerretDB string
	}{
		"ShowRecordID": {
			showRecordID:     true,
			collection:       collection,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/242",
		},
		"ShowRecordIDFalse": {
			showRecordID: false,
			collection:   collection,
		},
	} {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testing.TB = tt

			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
			}

			require.NotNil(t, tc.collection, "collection must be set")

			// small batch size is set to ensure getMore sets recordID
			opts := options.Find().SetShowRecordID(tc.showRecordID).SetBatchSize(2)
			cursor, err := tc.collection.Find(ctx, bson.D{}, opts)
			require.NoError(t, err)

			if setup.IsYugabyteDB(t) {
				// TODO https://github.com/yugabyte/yugabyte-db/issues/27989
				t.Skip("https://github.com/yugabyte/yugabyte-db/issues/27989")
			}

			var res []bson.D
			err = cursor.All(ctx, &res)
			require.NoError(t, cursor.Close(ctx))
			require.NoError(t, err)

			for i, r := range res {
				var recordComparable bson.D

				expected := bson.D{
					{"_id", ""},
				}

				for _, field := range r {
					switch field.Key {
					case "$recordId":
						recordID := field.Value
						t.Logf("%dth document with recordID %v", i, recordID)

						if !tc.showRecordID {
							require.Nil(t, recordID)
						} else {
							require.NotNil(t, recordID)
						}

						if tc.nonZeroRecordID {
							require.NotZero(t, recordID)
						}

						recordComparable = append(recordComparable, bson.E{field.Key, ""})

					case "_id":
						assert.IsType(tt, "", field.Value)
						recordComparable = append(recordComparable, bson.E{field.Key, ""})

					case "v":
						// do nothing, this test is dedicated to $recordId logic

					default:
						recordComparable = append(recordComparable, field)
					}
				}

				if tc.showRecordID {
					expected = append(expected, bson.E{"$recordId", ""})
				}

				AssertEqualDocuments(t, expected, recordComparable)
			}
		})
	}
}

func TestQueryShowRecordIDErrors(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct {
		showRecordID any

		err        *mongo.CommandError // optional, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		skip       string              // optional, skip test with a specified reason
	}{
		"Nil": {
			showRecordID: nil,
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "Field 'showRecordId' should be a boolean value, but found: null",
			},
			altMessage: "BSON field 'find.showRecordId' is the wrong type 'null', expected type 'bool'",
		},
		"Int32": {
			showRecordID: int32(0),
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "Field 'showRecordId' should be a boolean value, but found: int",
			},
			altMessage: "BSON field 'find.showRecordId' is the wrong type 'int', expected type 'bool'",
		},
		"String": {
			showRecordID: "string",
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "Field 'showRecordId' should be a boolean value, but found: string",
			},
			altMessage: "BSON field 'find.showRecordId' is the wrong type 'string', expected type 'bool'",
		},
	} {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB-DocumentDB/issues/242")

			require.NotNil(t, tc.err, "err must not be nil")

			var res bson.D
			err := collection.Database().RunCommand(ctx, bson.D{
				{"find", collection.Name()},
				{"showRecordId", tc.showRecordID},
			}).Decode(&res)

			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
			require.Nil(t, res)
		})
	}
}
