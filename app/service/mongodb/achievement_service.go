package service

import (
    "time"
    "errors"
    "os"
    "fmt"
    "path/filepath"
    "math"

    
    // Import Models
    modelMongo "student-performance-report/app/models/mongodb"
    modelPg "student-performance-report/app/models/postgresql"
    
    // Import Repositories
    repoMongo "student-performance-report/app/repository/mongodb"
    repoPg "student-performance-report/app/repository/postgresql"
    
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    // "github.com/golang-jwt/jwt/v5" // Kita tidak butuh ini lagi di service karena middleware sudah handle parsing
)

type AchievementService struct {
    mongoRepo repoMongo.AchievementRepository
    pgRepo    repoPg.AchievementRepoPostgres
    lecturer   repoPg.LecturerRepository
}

func NewAchievementService(m repoMongo.AchievementRepository, p repoPg.AchievementRepoPostgres, l repoPg.LecturerRepository) *AchievementService {
    return &AchievementService{mongoRepo: m, pgRepo: p, lecturer: l}
}

// === HELPER DIPERBAIKI (Sesuaikan dengan Middleware) ===

// Helper untuk ambil User ID dari c.Locals("user_id")
func getUserIDFromToken(c *fiber.Ctx) (uuid.UUID, error) {
    // 1. Ambil dari Locals
    userIDRaw := c.Locals("user_id")
    
    if userIDRaw == nil {
        return uuid.Nil, errors.New("unauthorized: user_id missing in context")
    }

    // 2. Cek Tipe Data: Apakah dia uuid.UUID langsung? (Skenario Paling Mungkin)
    if uid, ok := userIDRaw.(uuid.UUID); ok {
        return uid, nil
    }

    // 3. Cek Tipe Data: Apakah dia String? (Fallback)
    if uidStr, ok := userIDRaw.(string); ok {
        return uuid.Parse(uidStr)
    }

    // 4. Jika bukan keduanya, berarti format tidak dikenali
    return uuid.Nil, errors.New("server error: user_id format invalid (expected string or uuid)")
}

// Helper untuk ambil Role dari c.Locals("role_name")
func getUserRoleFromToken(c *fiber.Ctx) string {
    // 1. Ambil dari Locals dengan key "role_name" (sesuai middleware line 27)
    roleRaw := c.Locals("role_name")
    
    if roleRaw == nil {
        return ""
    }

    // 2. Cast ke string
    if roleStr, ok := roleRaw.(string); ok {
        return roleStr
    }
    return ""
}

// === Endpoint Logic: CREATE ===
func (s *AchievementService) CreateAchievement(c *fiber.Ctx) error {
    ctx := c.Context()
    
    // 1. Ambil User ID
    userID, err := getUserIDFromToken(c)
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": err.Error()})
    }

    // 2. Ambil Student ID dari Database Postgres
    studentID, err := s.pgRepo.GetStudentByUserID(ctx, userID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Student profile not found. Are you registered as a student?"})
    }

    // 3. Parse Body ke Model Mongo
    var req modelMongo.Achievement
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
    }

    // =================================================================
    // [PERBAIKAN PENTING]: Inisialisasi Attachments sebagai Array Kosong
    // Agar di MongoDB tersimpan [], bukan null. 
    // Ini mencegah error saat Upload File ($push ke null).
    // =================================================================
    req.Attachments = make([]modelMongo.Attachment, 0)

    // 4. Set Data Mongo
    req.StudentID = studentID.String()
    req.CreatedAt = time.Now()
    req.UpdatedAt = time.Now()

    // 5. Simpan ke Mongo
    mongoID, err := s.mongoRepo.InsertOne(ctx, req)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to save achievement details"})
    }

    // 6. Simpan Referensi ke Postgres
    ref := modelPg.AchievementReference{
        StudentID:          studentID,
        MongoAchievementID: mongoID,
        Status:             "draft", // Pastikan status terisi
        CreatedAt:          time.Now(),
    }
    
    newID, err := s.pgRepo.Create(ctx, ref)
    if err != nil {
        // Jika simpan ke Postgres gagal, kita coba hapus data Mongo yang terlanjur masuk (Rollback sederhana)
        _ = s.mongoRepo.DeleteAchievement(ctx, mongoID)
        
        return c.Status(500).JSON(fiber.Map{"error": "Failed to save achievement reference: " + err.Error()})
    }

    return c.Status(201).JSON(fiber.Map{
        "message": "Achievement created successfully",
        "id": newID,
        "status": "draft",
    })
}

// === Endpoint Logic: GET ALL (List) ===
// ... imports

// === Endpoint Logic: GET ALL (List with Pagination) ===
func (s *AchievementService) GetAllAchievements(c *fiber.Ctx) error {
    ctx := c.Context()
    userID, err := getUserIDFromToken(c)
    if err != nil { return c.Status(401).JSON(fiber.Map{"error": err.Error()}) }
    role := getUserRoleFromToken(c)

    // 1. PARSE QUERY PARAMS (Page, Limit, Status, Sort)
    var query modelPg.PaginationQuery
    if err := c.QueryParser(&query); err != nil {
        // Default values jika parsing gagal
        query.Page = 1
        query.Limit = 10
    }
    
    // Set Defaults
    if query.Page <= 0 { query.Page = 1 }
    if query.Limit <= 0 { query.Limit = 10 }
    if query.Limit > 100 { query.Limit = 100 } // Max limit guard

    // Calculate Offset
    offset := (query.Page - 1) * query.Limit

    // 2. SIAPKAN FILTER (Logic Role sama seperti sebelumnya)
    filters := make(map[string]interface{})

    if role == "mahasiswa" || role == "Mahasiswa" {
        studentID, _ := s.pgRepo.GetStudentByUserID(ctx, userID)
        filters["student_id"] = studentID
        // Mahasiswa bisa filter status sendiri via query param (misal ?status=draft)
        if query.Status != "" {
            filters["status"] = query.Status
        }
    } 
    
    if role == "dosen_wali" || role == "Dosen Wali" {
        lecturerID, _ := s.lecturer.GetLecturerByUserID(ctx, userID)
        advisees, _ := s.lecturer.GetAdvisees(lecturerID)
        var studentIDs []uuid.UUID
        for _, mhs := range advisees {
            studentIDs = append(studentIDs, mhs.ID)
        }
        
        if len(studentIDs) == 0 {
            // Return empty paginated response
            return c.JSON(modelPg.PaginatedResponse{
                Data: []interface{}{},
                Meta: modelPg.PaginationMeta{
                    CurrentPage: query.Page, Limit: query.Limit, TotalData: 0, TotalPage: 0,
                },
            })
        }
        filters["student_ids"] = studentIDs
        
        // Dosen default lihat submitted/verified, tapi bisa filter spesifik via query param
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

    // 4. AMBIL DETAIL MONGO (Tetap sama)
    var mongoIDs []string
    refMap := make(map[string]modelPg.AchievementReference)
    for _, r := range refs {
        mongoIDs = append(mongoIDs, r.MongoAchievementID)
        refMap[r.MongoAchievementID] = r
    }

    details, _ := s.mongoRepo.FindAllDetails(ctx, mongoIDs)

    // 5. GABUNGKAN DATA
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

    // 6. RETURN PAGINATED RESPONSE
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

// === Endpoint Logic: GET DETAIL ===
func (s *AchievementService) GetAchievementDetail(c *fiber.Ctx) error {
    ctx := c.Context()
    
    // 1. Ambil ID Prestasi dari URL parameter
    achievementID, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
    }

    // 2. Ambil User Login untuk validasi akses
    userID, err := getUserIDFromToken(c) // Helper yang sudah kita perbaiki
    if err != nil {
        return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
    }
    role := getUserRoleFromToken(c)

    // 3. Ambil Referensi dari Postgres
    ref, err := s.pgRepo.GetReferenceByID(ctx, achievementID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
    }

    // 4. Validasi Kepemilikan (PENTING untuk keamanan)
    if role == "mahasiswa" {
        // Ambil studentID user yang login
        currentStudentID, err := s.pgRepo.GetStudentByUserID(ctx, userID)
        if err != nil {
             return c.Status(500).JSON(fiber.Map{"error": "Student profile error"})
        }
        
        // Bandingkan StudentID di Prestasi vs StudentID User Login
        if ref.StudentID != currentStudentID {
            return c.Status(403).JSON(fiber.Map{"error": "Forbidden: You cannot view this achievement"})
        }
    } else if role == "dosen_wali" {
        // 1. Ambil ID Dosen
        lecturerID, err := s.lecturer.GetLecturerByUserID(ctx, userID)
        if err != nil {
            return c.Status(403).JSON(fiber.Map{"error": "Lecturer profile not found"})
        }

        // 2. Ambil Daftar Mahasiswa Bimbingan
        // (Kita gunakan method yang sudah ada di LecturerRepository)
        advisees, err := s.lecturer.GetAdvisees(lecturerID)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to check advisee relationship"})
        }

        // 3. Cek apakah pemilik prestasi ada di daftar bimbingan
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

    // 5. Ambil Detail dari Mongo
    detail, err := s.mongoRepo.FindOne(ctx, ref.MongoAchievementID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch achievement details"})
    }

    // 6. Return Data Gabungan
    response := map[string]interface{}{
        "id":            ref.ID,
        "status":        ref.Status,
        "rejectionNote": ref.RejectionNote,
        "details":       detail, 
        "createdAt":     ref.CreatedAt,
    }

    return c.JSON(response)
}

// === FR-004: SUBMIT ACHIEVEMENT (Draft -> Submitted) ===
func (s *AchievementService) SubmitAchievement(c *fiber.Ctx) error {
    ctx := c.Context()
    achievementID, err := uuid.Parse(c.Params("id"))
    if err != nil { return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"}) }

    // 1. Validasi User
    userID, err := getUserIDFromToken(c)
    if err != nil { return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"}) }

    studentID, err := s.pgRepo.GetStudentByUserID(ctx, userID)
    if err != nil { return c.Status(404).JSON(fiber.Map{"error": "Student profile not found"}) }

    // 2. Cek Data
    ref, err := s.pgRepo.GetReferenceByID(ctx, achievementID)
    if err != nil { return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"}) }

    // 3. Validasi Kepemilikan
    if ref.StudentID != studentID {
        return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
    }

    // 4. Validasi Status (Hanya Draft yang bisa di-submit) [cite: 190]
    if ref.Status != "draft" {
        return c.Status(400).JSON(fiber.Map{"error": "Only draft achievements can be submitted"})
    }

    // 5. Update Status jadi 'submitted' [cite: 195]

    err = s.pgRepo.SubmitReference(ctx, achievementID) 
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to submit achievement"+ err.Error(),})
    }

    return c.JSON(fiber.Map{"status": "success", "message": "Achievement submitted for verification"})
}

func (s *AchievementService) DeleteAchievement(c *fiber.Ctx) error {
    ctx := c.Context()
    
    // 1. Parse ID Prestasi
    achievementID, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
    }

    // 2. Ambil User ID & Student ID
    userID, err := getUserIDFromToken(c)
    if err != nil { return c.Status(401).JSON(fiber.Map{"error": err.Error()}) }

    studentID, err := s.pgRepo.GetStudentByUserID(ctx, userID)
    if err != nil { return c.Status(404).JSON(fiber.Map{"error": "Student profile not found"}) }

    // 3. Cek Data Eksisting
    ref, err := s.pgRepo.GetReferenceByID(ctx, achievementID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
    }

    // 4. Validasi: Harus Milik Sendiri
    if ref.StudentID != studentID {
        return c.Status(403).JSON(fiber.Map{"error": "Forbidden: You do not own this data"})
    }

    // 5. Validasi: Status Harus Draft 
    if ref.Status != "draft" {
        return c.Status(400).JSON(fiber.Map{"error": "Only draft achievements can be deleted"})
    }

    // 6. Hapus dari Postgres (Referensi) [cite: 204]
    if err := s.pgRepo.DeleteReference(ctx, achievementID); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to delete reference"})
    }

    // 7. Hapus dari MongoDB (Detail) [cite: 203]
    // Note: Walaupun Postgres sudah hapus, kita tetap coba hapus Mongo agar bersih
    _ = s.mongoRepo.DeleteAchievement(ctx, ref.MongoAchievementID)

    return c.JSON(fiber.Map{"message": "Achievement deleted successfully"})
}

// === FR-007: VERIFY ACHIEVEMENT ===
func (s *AchievementService) VerifyAchievement(c *fiber.Ctx) error {
    ctx := c.Context()
    achievementID, _ := uuid.Parse(c.Params("id"))

    // 1. Ambil Data Dosen yang Login
    userID, err := getUserIDFromToken(c)
    if err != nil { return c.Status(401).JSON(fiber.Map{"error": err.Error()}) }

    // [FIX BUG LOGIKA] Panggil repository lecturer untuk memastikan user adalah Dosen
    
    _, err = s.lecturer.GetLecturerByUserID(ctx, userID) 
    if err != nil { return c.Status(403).JSON(fiber.Map{"error": "User is not a lecturer"}) }

    // 2. Cek Data Prestasi
    ref, err := s.pgRepo.GetReferenceByID(ctx, achievementID)
    if err != nil { return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"}) }

   
    if ref.Status != "submitted" {
        return c.Status(400).JSON(fiber.Map{"error": "Achievement must be in 'submitted' status to be verified"})
    }

    // 3. Update Status jadi 'verified'
    // UserID dosen disimpan sebagai verified_by
    err = s.pgRepo.UpdateStatus(ctx, achievementID, "verified", &userID, "") 
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to verify achievement"})
    }

    return c.JSON(fiber.Map{"status": "success", "message": "Achievement verified"})
}

// === FR-008: REJECT ACHIEVEMENT ===
func (s *AchievementService) RejectAchievement(c *fiber.Ctx) error {
    ctx := c.Context()
    achievementID, _ := uuid.Parse(c.Params("id"))

    userID, err := getUserIDFromToken(c)
    if err != nil { return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"}) }

    // [FIX] Validasi role lecturer
    _, err = s.lecturer.GetLecturerByUserID(ctx, userID)
    if err != nil { return c.Status(403).JSON(fiber.Map{"error": "User is not a lecturer"}) }

    var req struct { Note string `json:"note"` }
    if err := c.BodyParser(&req); err != nil || req.Note == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Rejection note is required"})
    }

    err = s.pgRepo.UpdateStatus(ctx, achievementID, "rejected", &userID, req.Note)
    if err != nil { return c.Status(500).JSON(fiber.Map{"error": "Failed to reject"}) }

    return c.JSON(fiber.Map{"status": "success", "message": "Rejected"})
}

// === UPDATE ACHIEVEMENT (Draft Only) ===
func (s *AchievementService) UpdateAchievement(c *fiber.Ctx) error {
    ctx := c.Context()
    achievementID, err := uuid.Parse(c.Params("id"))
    if err != nil { return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"}) }

    // 1. Validasi User
    userID, err := getUserIDFromToken(c)
    if err != nil { return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"}) }

    studentID, err := s.pgRepo.GetStudentByUserID(ctx, userID)
    if err != nil { return c.Status(404).JSON(fiber.Map{"error": "Student profile not found"}) }

    // 2. Cek Data & Status
    ref, err := s.pgRepo.GetReferenceByID(ctx, achievementID)
    if err != nil { return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"}) }

    // 3. Validasi Kepemilikan
    if ref.StudentID != studentID {
        return c.Status(403).JSON(fiber.Map{"error": "Forbidden: You do not own this data"})
    }

    // 4. Validasi Status (Hanya Draft yang bisa diedit)
    if ref.Status != "draft" {
        return c.Status(400).JSON(fiber.Map{"error": "Only draft achievements can be updated"})
    }

    // 5. Parse Body
    var req modelMongo.Achievement
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid body","details": err.Error(),})
    }

    // 6. Update ke Mongo
    err = s.mongoRepo.UpdateOne(ctx, ref.MongoAchievementID, req)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to update achievement"})
    }

    return c.JSON(fiber.Map{"message": "Achievement updated successfully"})
}

// === GET HISTORY (Status Tracking) ===
func (s *AchievementService) GetAchievementHistory(c *fiber.Ctx) error {
    ctx := c.Context()
    achievementID, err := uuid.Parse(c.Params("id"))
    if err != nil { return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"}) }

    // 1. Ambil Referensi Postgres
    ref, err := s.pgRepo.GetReferenceByID(ctx, achievementID)
    if err != nil { return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"}) }

    // 2. Bangun Timeline History berdasarkan Timestamp yang ada
    var history []map[string]interface{}

    // Event: Dibuat
    history = append(history, map[string]interface{}{
        "status":    "created",
        "timestamp": ref.CreatedAt,
        "note":      "Achievement draft created",
    })

    // Event: Diajukan (Submitted)
    if ref.SubmittedAt != nil {
        history = append(history, map[string]interface{}{
            "status":    "submitted",
            "timestamp": ref.SubmittedAt,
            "note":      "Submitted for verification",
        })
    }

    // Event: Diverifikasi / Ditolak
    if ref.VerifiedAt != nil {
        action := "verified"
        if ref.Status == "rejected" {
            action = "rejected"
        }
        
        item := map[string]interface{}{
            "status":    action,
            "timestamp": ref.VerifiedAt,
            "by":        ref.VerifiedBy, // UUID Dosen
        }

        if ref.RejectionNote != nil && *ref.RejectionNote != "" {
            item["note"] = *ref.RejectionNote
        }

        history = append(history, item)
    }

    return c.JSON(history)
}

// === UPLOAD ATTACHMENTS ===
func (s *AchievementService) UploadAttachments(c *fiber.Ctx) error {
    ctx := c.Context()
    achievementID, err := uuid.Parse(c.Params("id"))
    if err != nil { return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"}) }

    // 1. Validasi User & Student
    userID, err := getUserIDFromToken(c)
    if err != nil { return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"}) }

    studentID, err := s.pgRepo.GetStudentByUserID(ctx, userID)
    if err != nil { return c.Status(404).JSON(fiber.Map{"error": "Student profile not found"}) }

    // 2. Cek Referensi & Status
    ref, err := s.pgRepo.GetReferenceByID(ctx, achievementID)
    if err != nil { return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"}) }

    if ref.StudentID != studentID {
        return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
    }
    if ref.Status != "draft" {
        return c.Status(400).JSON(fiber.Map{"error": "Cannot upload files to submitted/verified achievements"})
    }

    // 3. Ambil File dari Request
    file, err := c.FormFile("file")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "No file uploaded"})
    }

    // 4. Simpan File ke Server (Local Storage)
    // Pastikan folder "uploads" sudah dibuat: mkdir uploads
    uploadDir := "./uploads"
    if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
        os.Mkdir(uploadDir, 0755)
    }

    // Generate nama file unik agar tidak tertimpa
    filename := uuid.New().String() + filepath.Ext(file.Filename)
    filePath := fmt.Sprintf("%s/%s", uploadDir, filename)

    if err := c.SaveFile(file, filePath); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to save file to disk"})
    }

    // 5. Update MongoDB
    // Buat URL file (misal: static path)
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