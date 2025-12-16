package repository

import (
    "context"
	"time"
    models "student-performance-report/app/models/mongodb"
    "github.com/google/uuid"
    "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementRepository interface {
    GetStudentAchievements(studentId uuid.UUID) ([]models.Achievement, error)
    InsertOne(ctx context.Context, achievement models.Achievement) (string, error)
    FindAllDetails(ctx context.Context, mongoIDs []string) ([]models.Achievement, error)
	FindOne(ctx context.Context, mongoID string) (*models.Achievement, error)
	DeleteAchievement(ctx context.Context, mongoID string) error
	UpdateOne(ctx context.Context, mongoID string, data models.Achievement) error
	AddAttachment(ctx context.Context, mongoID string, attachment models.Attachment) error
    GetGlobalStats(ctx context.Context) (*models.GlobalStatistics, error) 
    GetStudentStats(ctx context.Context, studentID string) (*models.StudentStatistics, error) 
    UpdatePoints(ctx context.Context, mongoID string, points int) error
}

type achievementRepository struct {
    collection *mongo.Collection
}

func NewAchievementRepository(mongodb *mongo.Database) AchievementRepository {
    return &achievementRepository{
        collection: mongodb.Collection("achievements"),
    }
}

func (r *achievementRepository) GetStudentAchievements(studentId uuid.UUID) ([]models.Achievement, error) {
    ctx := context.Background()

    filter := bson.M{"studentId": studentId.String()}
    cursor, err := r.collection.Find(ctx, filter)
    if err != nil {
        return nil, err
    }

    var results []models.Achievement
    err = cursor.All(ctx, &results)
    return results, err
}

func (r *achievementRepository) InsertOne(ctx context.Context, achievement models.Achievement) (string, error) {
	collection := r.collection
	result, err := collection.InsertOne(ctx, achievement)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *achievementRepository) FindAllDetails(ctx context.Context, mongoIDs []string) ([]models.Achievement, error) {
	var objectIDs []primitive.ObjectID
	for _, id := range mongoIDs {
		oid, _ := primitive.ObjectIDFromHex(id)
		objectIDs = append(objectIDs, oid)
	}

	collection := r.collection
	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var achievements []models.Achievement
	if err = cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}
	return achievements, nil
}

func (r *achievementRepository) FindOne(ctx context.Context, mongoID string) (*models.Achievement, error) {
    oid, err := primitive.ObjectIDFromHex(mongoID)
    if err != nil {
        return nil, err
    }

    var result models.Achievement
    filter := bson.M{"_id": oid}
    
    err = r.collection.FindOne(ctx, filter).Decode(&result)
    if err != nil {
        return nil, err
    }
    
    return &result, nil
}

func (r *achievementRepository) DeleteAchievement(ctx context.Context, mongoID string) error {
    oid, err := primitive.ObjectIDFromHex(mongoID)
    if err != nil {
        return err
    }

    filter := bson.M{"_id": oid}
    _, err = r.collection.DeleteOne(ctx, filter)
    return err
}

func (r *achievementRepository) UpdateOne(ctx context.Context, mongoID string, data models.Achievement) error {
    oid, err := primitive.ObjectIDFromHex(mongoID)
    if err != nil {
         return err 
    }

    update := bson.M{
        "$set": bson.M{
            "title":           data.Title,
            "description":     data.Description,
            "achievementType": data.AchievementType,
            "details":         data.Details,
            "tags":            data.Tags,
            "points":          data.Points,
            "updatedAt":       time.Now(),
        },
    }

    _, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, update)
    return err
}

func (r *achievementRepository) AddAttachment(ctx context.Context, mongoID string, attachment models.Attachment) error {
    oid, err := primitive.ObjectIDFromHex(mongoID)
    if err != nil {
         return err 
    }

    update := bson.M{
        "$push": bson.M{"attachments": attachment},
        "$set":  bson.M{"updatedAt": time.Now()},
    }

    _, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, update)
    return err
}

func (r *achievementRepository) GetGlobalStats(ctx context.Context) (*models.GlobalStatistics, error) {
    stats := &models.GlobalStatistics{
        TypeDistribution:  make(map[string]int),
        LevelDistribution: make(map[string]int),
        TrendByYear:       make(map[string]int),
    }

    pipelineType := bson.A{
        bson.M{"$group": bson.M{"_id": "$achievementType", "count": bson.M{"$sum": 1}}},
    }
    cursor, _ := r.collection.Aggregate(ctx, pipelineType)
    var typeResults []struct { Id string `bson:"_id"`; Count int `bson:"count"` }
    cursor.All(ctx, &typeResults)
    for _, res := range typeResults {
        stats.TypeDistribution[res.Id] = res.Count 
    }

    pipelineLevel := bson.A{
        bson.M{"$match": bson.M{"details.competitionLevel": bson.M{"$exists": true}}},
        bson.M{"$group": bson.M{"_id": "$details.competitionLevel", "count": bson.M{"$sum": 1}}},
    }
    cursor, _ = r.collection.Aggregate(ctx, pipelineLevel)
    var levelResults []struct { Id string `bson:"_id"`; Count int `bson:"count"` }
    cursor.All(ctx, &levelResults)
    for _, res := range levelResults {
        stats.LevelDistribution[res.Id] = res.Count
    }

    pipelineTop := bson.A{
        bson.M{"$group": bson.M{"_id": "$studentId", "totalPoints": bson.M{"$sum": "$points"}}},
        bson.M{"$sort": bson.M{"totalPoints": -1}},
        bson.M{"$limit": 5},
    }
    cursor, _ = r.collection.Aggregate(ctx, pipelineTop)
    var topResults []struct { Id string `bson:"_id"`; TotalPoints int `bson:"totalPoints"` }
    cursor.All(ctx, &topResults)

    for _, res := range topResults {
        stats.PointsDistribution = append(stats.PointsDistribution, models.TopStudent{
            StudentID:   res.Id,
            TotalPoints: res.TotalPoints,
        })
    }

    return stats, nil
}

func (r *achievementRepository) GetStudentStats(ctx context.Context, studentID string) (*models.StudentStatistics, error) {
    stats := &models.StudentStatistics{ByType: make(map[string]int)}
    pipeline := bson.A{
        bson.M{"$match": bson.M{"studentId": studentID}},
        bson.M{"$group": bson.M{
            "_id": "$achievementType",
            "count": bson.M{"$sum": 1},
            "points": bson.M{"$sum": "$points"},
        }},
    }

    cursor, err := r.collection.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, err
    }

    var results []struct {
        Id     string `bson:"_id"`
        Count  int    `bson:"count"`
        Points int    `bson:"points"`
    }
    if err = cursor.All(ctx, &results); err != nil {
        return nil, err
    }

    for _, res := range results {
        stats.ByType[res.Id] = res.Count
        stats.TotalAchievements += res.Count
        stats.TotalPoints += res.Points
    }

    return stats, nil
}

func (r *achievementRepository) UpdatePoints(ctx context.Context,mongoID string, points int,) error {
    oid, err := primitive.ObjectIDFromHex(mongoID)
    if err != nil {
        return err
    }

    _, err = r.collection.UpdateOne(
        ctx,
        bson.M{"_id": oid},
        bson.M{
            "$set": bson.M{
                "points":     points,
                "updatedAt": time.Now(),
            },
        },
    )

    return err
}
