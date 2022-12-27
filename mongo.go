package main

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Response body for DB-related operations
type DbResponse struct {
	Code    int      `json:"code"`
	Msg     string   `json:"msg"`
	Records []Record `json:"records"`
}

// Struct for the relevant fields of the documents present in DB
type Record struct {
	CreatedAt  time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	Key        string    `json:"key,omitempty" bson:"key,omitempty"`
	TotalCount int       `json:"totalCount,omitempty" bson:"totalCount,omitempty"`
}

func fetchDB(startDate time.Time, endDate time.Time, minCount int, maxCount int) (*DbResponse, error) {

	uri := "mongodb+srv://challengeUser:WUMglwNBaydH8Yvu@challenge-xzwqd.mongodb.net/getircase-study?retryWrites=true"

	// Creates a context and defers a function to release its resources if the context expires before the deadline
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Creates a connection with MongoDB and defers a disconnect operation until the surrounding fetchDB function returns
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	// Pings our DB to check for connection. In the case of no connection, returns error
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	coll := client.Database("getircase-study").Collection("records")

	// Query for selecting the documents with totalCount fields in between the given min/maxCount params
	// and the ones that are created during the period marked by the given start/endDate params
	filter := bson.D{
		{"$and",
			bson.A{
				bson.D{{"totalCount",
					bson.D{{"$gte", minCount}},
				}},
				bson.D{{"totalCount",
					bson.D{{"$lte", maxCount}},
				}},
			},
		},
		// Need to cast time.Time objects startDate, endDate into string. Else, it doesn't work
		{"$and",
			bson.A{
				bson.D{{"createdAt",
					bson.D{{"$gte", startDate.String()}},
				}},
				bson.D{{"createdAt",
					bson.D{{"$lte", endDate.String()}},
				}},
			},
		},
	}

	// Cursor object holds the documents that are found as a result of our query
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	response := &DbResponse{}

	// Iterates over the found documents and appends them to the Records array of our DbReponse instance
	for cursor.Next(context.TODO()) {
		var rec Record
		if err := cursor.Decode(&rec); err != nil {
			return nil, err
		}
		response.Records = append(response.Records, rec)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	response.Code = 0
	response.Msg = "Success"

	return response, nil
}
