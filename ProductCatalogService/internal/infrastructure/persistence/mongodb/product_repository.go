package mongodb

import (
	"context"
	"fmt"

	"github.com/egannguyen/go-kafka-ecommerce/product-catalog-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type productRepository struct {
	db *mongo.Database
}

func NewProductRepository(db *mongo.Database) domain.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) FindAll(ctx context.Context) ([]domain.Product, error) {
	coll := r.db.Collection("products")
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer cursor.Close(ctx)

	var products []domain.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, fmt.Errorf("failed to decode products: %w", err)
	}
	return products, nil
}

func (r *productRepository) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	coll := r.db.Collection("products")
	var p domain.Product
	err := coll.FindOne(ctx, bson.M{"id": id}).Decode(&p)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product by id: %w", err)
	}
	return &p, nil
}

func (r *productRepository) Update(ctx context.Context, p *domain.Product) error {
	coll := r.db.Collection("products")
	_, err := coll.UpdateOne(ctx,
		bson.M{"id": p.ID},
		bson.M{"$set": p},
	)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	return nil
}
