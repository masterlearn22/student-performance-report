package repository

import (
    "context"
    "database/sql"
    "errors"
    models "student-performance-report/app/models/postgresql"
    "github.com/google/uuid"
    "github.com/lib/pq"
)

type StudentRepository interface {
    GetAllStudents(ctx context.Context) ([]models.Student, error)
    GetStudentByID(ctx context.Context, id uuid.UUID) (*models.Student, error)
    UpdateAdvisor(ctx context.Context, studentID, lecturerID uuid.UUID) error
    GetStudentsByIDs(ctx context.Context, ids []string) ([]models.StudentWithUser, error)
}

type studentRepository struct {
    pg *sql.DB
}

func NewStudentRepository(pg *sql.DB) StudentRepository {
    return &studentRepository{pg: pg}
}

func (r *studentRepository) GetAllStudents(ctx context.Context) ([]models.Student, error) {
    query := `
        SELECT s.id, s.user_id, s.student_id, u.full_name, s.program_study, s.academic_year, s.advisor_id, s.created_at
        FROM students s
        JOIN users u ON s.user_id = u.id
        ORDER BY s.created_at DESC
    `
    rows, err := r.pg.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var list []models.Student
    for rows.Next() {
        var s models.Student
        err := rows.Scan(
            &s.ID,
            &s.UserID, 
            &s.StudentID, 
            &s.FullName, 
            &s.ProgramStudy,
            &s.AcademicYear, 
            &s.AdvisorID, 
            &s.CreatedAt,
        )
        if err != nil {
            return nil, err 
        } 
        list = append(list, s)
    }
    return list, nil
}

func (r *studentRepository) GetStudentByID(ctx context.Context, id uuid.UUID) (*models.Student, error) {
    var s models.Student
    query := `
        SELECT s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at, u.full_name
        FROM students s
        JOIN users u ON s.user_id = u.id
        WHERE s.id = $1
    `

    var advisorID sql.NullString 

    err := r.pg.QueryRowContext(ctx, query, id).Scan(
        &s.ID, &s.UserID,
        &s.StudentID,
        &s.ProgramStudy,
        &s.AcademicYear,
        &advisorID, 
        &s.CreatedAt,
        &s.FullName, 
    )

    if err == sql.ErrNoRows {
        return nil, errors.New("student not found")
    } else if err != nil {
        return nil, err
    }

    // Convert NullString kembali ke *UUID
    if advisorID.Valid {
        uid, _ := uuid.Parse(advisorID.String)
        s.AdvisorID = &uid
    } else {
        s.AdvisorID = nil
    }

    return &s, nil
}

func (r *studentRepository) UpdateAdvisor(ctx context.Context, studentID, lecturerID uuid.UUID) error {
    query := `UPDATE students SET advisor_id=$1 WHERE id=$2`
    _, err := r.pg.ExecContext(ctx, query, lecturerID, studentID)
    return err
}

func (r *studentRepository) GetStudentsByIDs(ctx context.Context, ids []string) ([]models.StudentWithUser, error) {
    if len(ids) == 0 {
        return []models.StudentWithUser{}, nil
    }

    query := `
        SELECT s.id, u.full_name, s.program_study
        FROM students s
        JOIN users u ON s.user_id = u.id
        WHERE s.id::text = ANY($1)
    `

    rows, err := r.pg.QueryContext(ctx, query, pq.Array(ids))
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var results []models.StudentWithUser
    for rows.Next() {
        var data models.StudentWithUser
        if err := rows.Scan(
            &data.ID, 
            &data.FullName, 
            &data.ProgramStudy,
            ); err != nil {
            return nil, err
        }
        results = append(results, data)
    }
    
    return results, nil
}