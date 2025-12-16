package service

import (
    "time"
    "errors"
    "os"
    "fmt"
    "path/filepath"
    "math"
    modelMongo "student-performance-report/app/models/mongodb"
    modelPg "student-performance-report/app/models/postgresql"
    repoMongo "student-performance-report/app/repository/mongodb"
    repoPg "student-performance-report/app/repository/postgresql"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "student-performance-report/middleware"
)

type AchievementService struct {
    mongoRepo repoMongo.AchievementRepository
    pgRepo    repoPg.AchievementRepoPostgres
    lecturer   repoPg.LecturerRepository
}

func NewAchievementService(m repoMongo.AchievementRepository, p repoPg.AchievementRepoPostgres, l repoPg.LecturerRepository) *AchievementService {
    return &AchievementService{mongoRepo: m, pgRepo: p, lecturer: l}
}

func getUserIDFromToken(c *fiber.Ctx) (uuid.UUID, error) {
    userIDRaw := c.Locals("user_id")
    
    if userIDRaw == nil {
        return uuid.Nil, errors.New("unauthorized: user_id missing in context")
    }

    if uid, ok := userIDRaw.(uuid.UUID); ok {
        return uid, nil
    }

    if uidStr, ok := userIDRaw.(string); ok {
        return uuid.Parse(uidStr)
    }

    return uuid.Nil, errors.New("server error: user_id format invalid (expected string or uuid)")
}


// CreateAchievement godoc
// @Summary Create New Achievement Draft
// @Description Create a new achievement draft (Student only)
// @Tags Achievements
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body modelMongo.Achievement true "Achievement Data"
// @Success 201 {object} map[string]interface{}
// @Failure 400,401,404,500 {object} map[string]interface{}
// @Router /achievements [post]
func (s *AchievementService) CreateAchievement(c *fiber.Ctx) error {
    ctx := c.Context()
    if !middleware.HasPermission(c, "achievement:create") {
    return fiber.ErrForbidden
    }

    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": err.Error()})
    }

    studentID, err := s.pgRepo.GetStudentByUserID(ctx, userID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Student profile not found. Are you registered as a student?"})
    }

    var req modelMongo.Achievement
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
    }

    req.Attachments = make([]modelMongo.Attachment, 0)
    req.StudentID = studentID.String()
    req.Points = 0 
    req.CreatedAt = time.Now()
    req.UpdatedAt = time.Now()
    mongoID, err := s.mongoRepo.InsertOne(ctx, req)

    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to save achievement details"})
    }

    ref := modelPg.AchievementReference{
        StudentID:          studentID,
        MongoAchievementID: mongoID,
        Status:             "draft", 
        CreatedAt:          time.Now(),
    }
    
    newID, err := s.pgRepo.Create(ctx, ref)
    if err != nil {
        _ = s.mongoRepo.DeleteAchievement(ctx, mongoID)
        
        return c.Status(500).JSON(fiber.Map{"error": "Failed to save achievement reference: " + err.Error()})
    }

    return c.Status(201).JSON(fiber.Map{
        "message": "Achievement created successfully",
        "id": newID,
        "status": "draft",
    })
}

// GetAllAchievements godoc
// @Summary Get List of Achievements
// @Description Get paginated list of achievements. Filter logic depends on role (Student: own data, Lecturer: advisees data).
// @Tags Achievements
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Param status query string false "Filter by status (draft, submitted, verified, rejected)"
// @Param sort query string false "Sort direction"
// @Success 200 {object} modelPg.PaginatedResponse
// @Failure 401,500 {object} map[string]interface{}
// @Router /achievements [get]
func (s *AchievementService) GetAllAchievements(c *fiber.Ctx) error {
    ctx := c.Context()
    if !middleware.HasPermission(c, "achievement:read") {
    return fiber.ErrForbidden
    }

    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": err.Error()}) 
    }
   

    var query modelPg.PaginationQuery
    if err := c.QueryParser(&query); err != nil {
        query.Page = 1
        query.Limit = 10
    }
    
    if query.Page <= 0 { query.Page = 1 }
    if query.Limit <= 0 { query.Limit = 10 }
    if query.Limit > 100 { query.Limit = 100 } 

    offset := (query.Page - 1) * query.Limit

    filters := make(map[string]interface{})

    if studentID, err := s.pgRepo.GetStudentByUserID(ctx, userID); err == nil {
        filters["student_id"] = studentID
        if query.Status != "" {
            filters["status"] = query.Status
        }
    }
    
    lecturerID, err := s.lecturer.GetLecturerByUserID(ctx, userID)
    if  err == nil { 

        advisees, err:= s.lecturer.GetAdvisees(lecturerID)
        if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "failed to fetch advisees",
        })
        }

        var studentIDs []uuid.UUID
        for _, mhs := range advisees {
            studentIDs = append(studentIDs, mhs.ID)
        }
        
        if len(studentIDs) == 0 {
            return c.JSON(modelPg.PaginatedResponse{
                Data: []interface{}{},
                Meta: modelPg.PaginationMeta{
                    CurrentPage: query.Page, Limit: query.Limit, TotalData: 0, TotalPage: 0,
                },
            })
        }

        filters["student_ids"] = studentIDs

        if query.Status != "" {
            filters["status"] = query.Status
        } else {
            filters["status"] = []string{"submitted", "verified"} 
        }
    }

    refs, totalData, err := s.pgRepo.GetAllReferences(ctx, filters, query.Limit, offset, query.Sort)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Database error: " + err.Error()})
    }

    if len(refs) == 0 {
        return c.JSON(modelPg.PaginatedResponse{
            Data: []interface{}{},
            Meta: modelPg.PaginationMeta{
                CurrentPage: query.Page, Limit: query.Limit, TotalData: 0, TotalPage: 0,
            },
        })
    }

    var mongoIDs []string
    refMap := make(map[string]modelPg.AchievementReference)
    for _, r := range refs {
        mongoIDs = append(mongoIDs, r.MongoAchievementID)
        refMap[r.MongoAchievementID] = r
    }

    details, _ := s.mongoRepo.FindAllDetails(ctx, mongoIDs)

    var data []interface{}
    for _, d := range details {
        mongoIDHex := d.ID.Hex()
        if ref, exists := refMap[mongoIDHex]; exists {
            data = append(data, map[string]interface{}{
                "id":             ref.ID,
                "status":         ref.Status,
                "submittedAt":    ref.SubmittedAt,
                "title":          d.Title,
                "type":           d.AchievementType,
                "points":         d.Points,
                "createdAt":      ref.CreatedAt,
                "studentId":      ref.StudentID,
            })
        }
    }

    totalPages := int(math.Ceil(float64(totalData) / float64(query.Limit)))
    
    return c.JSON(modelPg.PaginatedResponse{
        Data: data,
        Meta: modelPg.PaginationMeta{
            CurrentPage: query.Page,
            TotalPage:   totalPages,
            TotalData:   int(totalData),
            Limit:       query.Limit,
        },
    })
}

// GetAchievementDetail godoc
// @Summary Get Achievement Detail
// @Description Get full details of a specific achievement including MongoDB data
// @Tags Achievements
// @Security BearerAuth
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400,401,403,404,500 {object} map[string]interface{}
// @Router /achievements/{id} [get]
func (s *AchievementService) GetAchievementDetail(c *fiber.Ctx) error {
    ctx := c.Context()
    if !middleware.HasPermission(c, "achievement:read") {
    return fiber.ErrForbidden
    }

    achievementID, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
    }

    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
    }

    ref, err := s.pgRepo.GetReferenceByID(ctx, achievementID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
    }

    currentStudentID, err := s.pgRepo.GetStudentByUserID(ctx, userID)
    if err == nil {
        if ref.StudentID != currentStudentID {
            return c.Status(403).JSON(fiber.Map{"error": "Forbidden: You cannot view this achievement"})
        }
    } 
    lecturerID, err := s.lecturer.GetLecturerByUserID(ctx, userID)
    if err ==nil{
        advisees, err := s.lecturer.GetAdvisees(lecturerID)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to check advisee relationship"})
        }

        isAdvisee := false
        for _, mhs := range advisees {
            if mhs.ID == ref.StudentID {
                isAdvisee = true
                break
            }
        }

        if !isAdvisee {
            return c.Status(403).JSON(fiber.Map{"error": "Forbidden: This student is not your advisee"})
        }

        if ref.Status == "draft" {
            return c.Status(403).JSON(fiber.Map{"error": "Forbidden: You cannot view draft achievements of your advisees"})
        }
    }

    detail, err := s.mongoRepo.FindOne(ctx, ref.MongoAchievementID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch achievement details"})
    }

    response := map[string]interface{}{
        "id":            ref.ID,
        "status":        ref.Status,
        "rejectionNote": ref.RejectionNote,
        "details":       detail, 
        "createdAt":     ref.CreatedAt,
    }

    return c.JSON(response)
}

// SubmitAchievement godoc
// @Summary Submit Achievement
// @Description Submit a draft achievement for verification (Student only)
// @Tags Achievements
// @Security BearerAuth
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} map[string]string
// @Failure 400,401,403,404,500 {object} map[string]interface{}
// @Router /achievements/{id}/submit [post]
func (s *AchievementService) SubmitAchievement(c *fiber.Ctx) error {
    ctx := c.Context()
    if !middleware.HasPermission(c, "achievement:create") {
    return fiber.ErrForbidden
    }

    achievementID, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"}) 
    }

    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"}) 
    }

    studentID, err := s.pgRepo.GetStudentByUserID(ctx, userID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Student profile not found"}) 
    }

    ref, err := s.pgRepo.GetReferenceByID(ctx, achievementID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"}) 
    }

    if ref.StudentID != studentID {
        return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
    }

    if ref.Status != "draft" {
        return c.Status(400).JSON(fiber.Map{"error": "Only draft achievements can be submitted"})
    }

    err = s.pgRepo.SubmitReference(ctx, achievementID) 
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to submit achievement"+ err.Error(),})
    }

    return c.JSON(fiber.Map{"status": "success", "message": "Achievement submitted for verification"})
}

// DeleteAchievement godoc
// @Summary Delete Achievement
// @Description Delete an achievement (Only allowed if status is draft)
// @Tags Achievements
// @Security BearerAuth
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} map[string]string
// @Failure 400,401,403,404,500 {object} map[string]interface{}
// @Router /achievements/{id} [delete]
func (s *AchievementService) DeleteAchievement(c *fiber.Ctx) error {
    ctx := c.Context()
    if !middleware.HasPermission(c, "achievement:delete") {
    return fiber.ErrForbidden
    }

    
    achievementID, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
    }

    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": err.Error()}) 
    }

    studentID, err := s.pgRepo.GetStudentByUserID(ctx, userID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Student profile not found"}) 
    }

    ref, err := s.pgRepo.GetReferenceByID(ctx, achievementID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
    }
    if ref.StudentID != studentID {
        return c.Status(403).JSON(fiber.Map{"error": "Forbidden: You do not own this data"})
    }

    if ref.Status != "draft" {
        return c.Status(400).JSON(fiber.Map{"error": "Only draft achievements can be deleted"})
    }

    if err := s.pgRepo.DeleteReference(ctx, achievementID); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to delete reference"})
    }

    _ = s.mongoRepo.DeleteAchievement(ctx, ref.MongoAchievementID)

    return c.JSON(fiber.Map{"message": "Achievement deleted successfully"})
}

// VerifyAchievement godoc
// @Summary Verify Achievement
// @Description Approve a submitted achievement (Lecturer/Dosen Wali only)
// @Tags Achievements
// @Security BearerAuth
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} map[string]string
// @Failure 400,401,403,404,500 {object} map[string]interface{}
// @Router /achievements/{id}/verify [post]
func (s *AchievementService) VerifyAchievement(c *fiber.Ctx) error {
    ctx := c.Context()

    if !middleware.HasPermission(c, "achievement:verify") {
        return fiber.ErrForbidden
    }

    achievementID, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
    }

    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": err.Error()})
    }

    lecturerID, err := s.lecturer.GetLecturerByUserID(ctx, userID)
    if err != nil {
        return c.Status(403).JSON(fiber.Map{"error": "User is not a lecturer"})
    }

    var req struct {
        Points int `json:"points"`
    }

    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
    }

    if req.Points <= 0 {
        return c.Status(400).JSON(fiber.Map{
            "error": "Points must be greater than 0",
        })
    }

    ref, err := s.pgRepo.GetReferenceByID(ctx, achievementID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
    }

    if ref.Status != "submitted" {
        return c.Status(400).JSON(fiber.Map{
            "error": "Achievement must be in 'submitted' status to be verified",
        })
    }

    
    err = s.mongoRepo.UpdatePoints(ctx, ref.MongoAchievementID, req.Points)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to update achievement points",
        })
    }

    // ✅ update status di Postgres
    err = s.pgRepo.UpdateStatus(ctx, achievementID, "verified", &lecturerID, "")
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to verify achievement",
        })
    }

    return c.JSON(fiber.Map{
        "status":  "success",
        "message": "Achievement verified",
        "points":  req.Points,
    })
}


// RejectAchievement godoc
// @Summary Reject Achievement
// @Description Reject a submitted achievement with a note (Lecturer/Dosen Wali only)
// @Tags Achievements
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Param request body object{note=string} true "Rejection Note"
// @Success 200 {object} map[string]string
// @Failure 400,401,403,500 {object} map[string]interface{}
// @Router /achievements/{id}/reject [post]
func (s *AchievementService) RejectAchievement(c *fiber.Ctx) error {
    ctx := c.Context()
    if !middleware.HasPermission(c, "achievement:verify") {
    return fiber.ErrForbidden
    }

    achievementID, _ := uuid.Parse(c.Params("id"))

    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"}) 
    }

    _, err = s.lecturer.GetLecturerByUserID(ctx, userID)
    if err != nil {
        return c.Status(403).JSON(fiber.Map{"error": "User is not a lecturer"}) 
    }

    var req struct { Note string `json:"note"` }
    if err := c.BodyParser(&req); err != nil || req.Note == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Rejection note is required"})
    }

    err = s.pgRepo.UpdateStatus(ctx, achievementID, "rejected", &userID, req.Note)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to reject"}) 
    }

    return c.JSON(fiber.Map{"status": "success", "message": "Rejected"})
}

// UpdateAchievement godoc
// @Summary Update Achievement
// @Description Update achievement data (Only allowed if status is draft)
// @Tags Achievements
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Param request body modelMongo.Achievement true "Updated Data"
// @Success 200 {object} map[string]string
// @Failure 400,401,403,404,500 {object} map[string]interface{}
// @Router /achievements/{id} [put]
func (s *AchievementService) UpdateAchievement(c *fiber.Ctx) error {
    ctx := c.Context()
    if !middleware.HasPermission(c, "achievement:update") {
    return fiber.ErrForbidden
    }

    achievementID, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"}) 
    }

    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"}) 
    }

    studentID, err := s.pgRepo.GetStudentByUserID(ctx, userID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Student profile not found"}) 
    }

    ref, err := s.pgRepo.GetReferenceByID(ctx, achievementID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"}) 
    }

    if ref.StudentID != studentID {
        return c.Status(403).JSON(fiber.Map{"error": "Forbidden: You do not own this data"})
    }

    if ref.Status != "draft" {
        return c.Status(400).JSON(fiber.Map{"error": "Only draft achievements can be updated"})
    }

    var req modelMongo.Achievement
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid body","details": err.Error(),})
    }

    err = s.mongoRepo.UpdateOne(ctx, ref.MongoAchievementID, req)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to update achievement"})
    }

    return c.JSON(fiber.Map{"message": "Achievement updated successfully"})
}

// GetAchievementHistory godoc
// @Summary Get Achievement History
// @Description Get status history log of an achievement
// @Tags Achievements
// @Security BearerAuth
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {array} map[string]interface{}
// @Failure 400,401,404 {object} map[string]interface{}
// @Router /achievements/{id}/history [get]
func (s *AchievementService) GetAchievementHistory(c *fiber.Ctx) error {
    ctx := c.Context()
    if !middleware.HasPermission(c, "achievement:read") {
    return fiber.ErrForbidden
    }

    achievementID, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"}) 
    }

    ref, err := s.pgRepo.GetReferenceByID(ctx, achievementID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"}) 
    }

    var history []map[string]interface{}

    history = append(history, map[string]interface{}{
        "status":    "created",
        "timestamp": ref.CreatedAt,
        "note":      "Achievement draft created",
    })

    if ref.SubmittedAt != nil {
        history = append(history, map[string]interface{}{
            "status":    "submitted",
            "timestamp": ref.SubmittedAt,
            "note":      "Submitted for verification",
        })
    }

    if ref.VerifiedAt != nil {
        action := "verified"
        if ref.Status == "rejected" {
            action = "rejected"
        }
        
        item := map[string]interface{}{
            "status":    action,
            "timestamp": ref.VerifiedAt,
            "by":        ref.VerifiedBy,
        }

        if ref.RejectionNote != nil && *ref.RejectionNote != "" {
            item["note"] = *ref.RejectionNote
        }

        history = append(history, item)
    }

    return c.JSON(history)
}

// UploadAttachments godoc
// @Summary Upload Attachment
// @Description Upload a file attachment for an achievement (Draft only)
// @Tags Achievements
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Param file formData file true "File to upload"
// @Success 200 {object} map[string]interface{}
// @Failure 400,401,403,404,500 {object} map[string]interface{}
// @Router /achievements/{id}/attachments [post]
func (s *AchievementService) UploadAttachments(c *fiber.Ctx) error {
    ctx := c.Context()
    if !middleware.HasPermission(c, "achievement:create") {
    return fiber.ErrForbidden
    }

    achievementID, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"}) 
    }

    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"}) 
    }

    studentID, err := s.pgRepo.GetStudentByUserID(ctx, userID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Student profile not found"}) 
    }

    ref, err := s.pgRepo.GetReferenceByID(ctx, achievementID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"}) 
    }

    if ref.StudentID != studentID {
        return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
    }
    if ref.Status != "draft" {
        return c.Status(400).JSON(fiber.Map{"error": "Cannot upload files to submitted/verified achievements"})
    }

    file, err := c.FormFile("file")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "No file uploaded"})
    }

    uploadDir := "./uploads"
    if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
        os.Mkdir(uploadDir, 0755)
    }

    filename := uuid.New().String() + filepath.Ext(file.Filename)
    filePath := fmt.Sprintf("%s/%s", uploadDir, filename)

    if err := c.SaveFile(file, filePath); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to save file to disk"})
    }

    fileURL := fmt.Sprintf("/uploads/%s", filename)
    
    attachment := modelMongo.Attachment{
        FileName:   file.Filename,
        FileURL:    fileURL,
        FileType:   file.Header.Get("Content-Type"),
        UploadedAt: time.Now(),
    }

    err = s.mongoRepo.AddAttachment(ctx, ref.MongoAchievementID, attachment)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to update database info", "details": err.Error()})
    }

    return c.JSON(fiber.Map{
        "message": "File uploaded successfully", 
        "data": attachment,
    })
}