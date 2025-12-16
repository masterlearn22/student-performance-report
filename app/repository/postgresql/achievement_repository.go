package repository

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    models "student-performance-report/app/models/postgresql"
    "github.com/google/uuid"
    "github.com/lib/pq"
)

type AchievementRepoPostgres interface {
    Create(ctx context.Context, ref models.AchievementReference) (uuid.UUID, error)
    GetStudentByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
    GetAllReferences(ctx context.Context, filter map[string]interface{}, limit, offset int, sort string) ([]models.AchievementReference, int64, error)
    GetReferenceByID(ctx context.Context, id uuid.UUID) (models.AchievementReference, error)
    DeleteReference(ctx context.Context, id uuid.UUID) error
    UpdateStatus(ctx context.Context, id uuid.UUID, status string, verifiedBy *uuid.UUID, note string) error
    SubmitReference(ctx context.Context, id uuid.UUID) error
}

type achievementRepoPostgres struct {
    db *sql.DB
}

func NewAchievementRepoPostgres(db *sql.DB) AchievementRepoPostgres {
    return &achievementRepoPostgres{db: db}
}

func (r *achievementRepoPostgres) GetStudentByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
    query := `
                SELECT id 
                FROM students
                WHERE user_id = $1
    `
    var studentID uuid.UUID
    err := r.db.QueryRowContext(ctx, query, userID).Scan(&studentID)
    return studentID, err
}

func (r *achievementRepoPostgres) Create(ctx context.Context, ref models.AchievementReference) (uuid.UUID, error) {
    query := `
        INSERT INTO achievement_references (
            student_id, mongo_achievement_id, status, created_at, updated_at
        ) VALUES ($1, $2, $3, NOW(), NOW())
        RETURNING id
    `
    var newID uuid.UUID
    err := r.db.QueryRowContext(ctx, query, 
        ref.StudentID, 
        ref.MongoAchievementID, 
        ref.Status, 
    ).Scan(&newID)

    return newID, err
}

func (r *achievementRepoPostgres) GetAllReferences(ctx context.Context, filter map[string]interface{}, limit, offset int, sort string) ([]models.AchievementReference, int64, error) {
    whereClause := " WHERE status != 'deleted'"
    var args []interface{}
    argCount := 1

    if val, ok := filter["student_id"]; ok {
        whereClause += fmt.Sprintf(" AND student_id = $%d", argCount)
        args = append(args, val)
        argCount++
    }

    if val, ok := filter["student_ids"]; ok {
        whereClause += fmt.Sprintf(" AND student_id = ANY($%d)", argCount)
        args = append(args, pq.Array(val))
        argCount++
    }

    if val, ok := filter["status"]; ok {
        if statuses, isSlice := val.([]string); isSlice {
            whereClause += fmt.Sprintf(" AND status = ANY($%d)", argCount)
            args = append(args, pq.Array(statuses))
        } else {
            whereClause += fmt.Sprintf(" AND status = $%d", argCount)
            args = append(args, val)
        }
        argCount++
    }

    var totalCount int64
    countQuery := `
                    SELECT COUNT(*) 
                    FROM achievement_references
    ` + whereClause
    
    err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
    if err != nil {
        return nil, 0, err
    }

    query := `
        SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, created_at 
        FROM achievement_references 
    ` + whereClause

    if sort == "oldest" {
        query += ` ORDER BY created_at ASC`
    } else {
        query += ` ORDER BY created_at DESC`
    }

    if limit > 0 {
        query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
        args = append(args, limit, offset)
    }

    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var results []models.AchievementReference
    for rows.Next() {
        var ref models.AchievementReference
        err := rows.Scan(
            &ref.ID, 
            &ref.StudentID, 
            &ref.MongoAchievementID, 
            &ref.Status, 
            &ref.SubmittedAt, 
            &ref.VerifiedAt,
            &ref.CreatedAt,
        )
        if err != nil {
            return nil, 0, err
        }
        results = append(results, ref)
    }

    return results, totalCount, nil
}

func (r *achievementRepoPostgres) GetReferenceByID(ctx context.Context, id uuid.UUID) (models.AchievementReference, error) {
    query := `
        SELECT 
            id, student_id, mongo_achievement_id, status, rejection_note, 
            created_at, submitted_at, verified_at, verified_by 
        FROM achievement_references 
        WHERE status != 'deleted' AND id = $1
    `
    
    var ref models.AchievementReference
    var rejectionNote sql.NullString

    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &ref.ID, 
        &ref.StudentID, 
        &ref.MongoAchievementID, 
        &ref.Status, 
        &rejectionNote,
        &ref.CreatedAt,
        &ref.SubmittedAt, 
        &ref.VerifiedAt,  
        &ref.VerifiedBy,  
    )

    if rejectionNote.Valid {
        note := rejectionNote.String
        ref.RejectionNote = &note
    }

    return ref, err
}

func (r *achievementRepoPostgres) DeleteReference(ctx context.Context, id uuid.UUID) error {
    query := `
        UPDATE achievement_references 
        SET status = 'deleted', updated_at = NOW() 
        WHERE id = $1
    `
    _, err := r.db.ExecContext(ctx, query, id)
    return err
}



func (r *achievementRepoPostgres) UpdateStatus(ctx context.Context, id uuid.UUID, status string, verifiedBy *uuid.UUID, note string) error {
    query := `
        UPDATE achievement_references 
        SET status = $1, verified_by = $2, verified_at = $3, rejection_note = $4, updated_at = NOW()
        WHERE id = $5
    `
    _, err := r.db.ExecContext(ctx, query, status, verifiedBy, time.Now(), note, id)
    return err
}

func (r *achievementRepoPostgres) SubmitReference(ctx context.Context, id uuid.UUID) error {
    query := `
        UPDATE achievement_references 
        SET status = 'submitted', 
            submitted_at = NOW(), 
            updated_at = NOW()
        WHERE id = $1
    `
    _, err := r.db.ExecContext(ctx, query, id)
    return err
}