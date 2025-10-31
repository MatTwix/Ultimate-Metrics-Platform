package database

import (
	"context"
	"fmt"

	"github.com/MatTwix/Ultimate-Metrics-Platform/servises/analytics-service/pkg/analitycs"
	"github.com/MatTwix/Ultimate-Metrics-Platform/servises/analytics-service/pkg/models"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoAnalyticsWriter struct {
	collection *mongo.Collection
}

func NewMongoAnalyticsWriter(client *mongo.Client, dbName, collectionName string) analitycs.AnalyticsWriter {
	collection := client.Database(dbName).Collection(collectionName)
	return &MongoAnalyticsWriter{collection: collection}
}

func (w *MongoAnalyticsWriter) SaveAggregated(ctx context.Context, agg models.AggregatedMetric) error {
	_, err := w.collection.InsertOne(ctx, agg)
	if err != nil {
		return fmt.Errorf("failed to insert aggregated metric: %w", err)
	}

	return nil
}
