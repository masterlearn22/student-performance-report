package repository

import (
	"database/sql"
	"errors"

	models "student-performance-report/app/models/postgresql"
	"github.com/google/uuid"
)

type AdminRepository interface {
	// User CRUD
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	DeleteUser(id uuid.UUID) error
	GetUserByID(id uuid.UUID) (*models.User, error)
	GetAllUsers() ([]models.User, error)

	// Roles
	AssignRole(userID uuid.UUID, roleID uuid.UUID) error

	// Student / Lecturer Profile
	SetStudentProfile(profile *models.Student) error
	SetLecturerProfile(profile *models.Lecturer) error
	SetAdvisor(studentID, lecturerID uuid.UUID) error
}

type adminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) AdminRepository {
	return &adminRepository{db: db}
}

//////////////////////////////////////////////////
// USER CRUD
//////////////////////////////////////////////////

func (r *adminRepository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7, NOW(), NOW())
	`
	_, err := r.db.Exec(query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.RoleID,
		user.IsActive,
	)
	return err
}

func (r *adminRepository) UpdateUser(user *models.User) error {
	query := `
		UPDATE users SET
			username=$1, email=$2, full_name=$3,
			role_id=$4, is_active=$5, updated_at=NOW()
		WHERE id=$6
	`

	_, err := r.db.Exec(query,
		user.Username,
		user.Email,
		user.FullName,
		user.RoleID,
		user.IsActive,
		user.ID,
	)
	return err
}

func (r *adminRepository) DeleteUser(id uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id=$1`, id)
	return err
}

func (r *adminRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User

	query := `
		SELECT id, username, email, password_hash, full_name, role_id, is_active,
		       created_at, updated_at
		FROM users
		WHERE id=$1
	`

	row := r.db.QueryRow(query, id)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.RoleID,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (r *adminRepository) GetAllUsers() ([]models.User, error) {
	query := `
		SELECT id, username, email, full_name, role_id, is_active, created_at
		FROM users ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.User

	for rows.Next() {
		var u models.User
		err = rows.Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.FullName,
			&u.RoleID,
			&u.IsActive,
			&u.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, u)
	}

	return list, nil
}

//////////////////////////////////////////////////
// ROLE ASSIGN
//////////////////////////////////////////////////

func (r *adminRepository) AssignRole(userID uuid.UUID, roleID uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE users SET role_id=$1 WHERE id=$2`, roleID, userID)
	return err
}

//////////////////////////////////////////////////
// STUDENT PROFILE
//////////////////////////////////////////////////

func (r *adminRepository) SetStudentProfile(s *models.Student) error {
	query := `
		INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id, created_at)
		VALUES ($1,$2,$3,$4,$5,$6, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			student_id=$3, program_study=$4, academic_year=$5, advisor_id=$6
	`
	_, err := r.db.Exec(query,
		s.ID,
		s.UserID,
		s.StudentID,
		s.ProgramStudy,
		s.AcademicYear,
		s.AdvisorID,
	)
	return err
}

//////////////////////////////////////////////////
// LECTURER PROFILE
//////////////////////////////////////////////////

func (r *adminRepository) SetLecturerProfile(l *models.Lecturer) error {
	query := `
		INSERT INTO lecturers (id, user_id, lecturer_id, department, created_at)
		VALUES ($1,$2,$3,$4, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			lecturer_id=$3, department=$4
	`
	_, err := r.db.Exec(query,
		l.ID,
		l.UserID,
		l.LecturerID,
		l.Department,
	)
	return err
}

//////////////////////////////////////////////////
// SET ADVISOR
//////////////////////////////////////////////////

func (r *adminRepository) SetAdvisor(studentID, lecturerID uuid.UUID) error {
	query := `UPDATE students SET advisor_id=$1 WHERE user_id=$2`
	_, err := r.db.Exec(query, lecturerID, studentID)
	return err
}
