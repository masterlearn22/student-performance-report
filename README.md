# Student Performance Reporting System
**A robust backend REST API service for managing and tracking student academic achievements**

---

## 📖 Overview

The Student Performance Reporting System is a comprehensive backend solution designed to streamline the management of student achievements including competitions, seminars, organizational activities, and other academic accomplishments. Built with scalability and security in mind, it implements a hybrid database architecture combining PostgreSQL's relational integrity with MongoDB's flexible document storage.

### Why This System?

- **Dual-Database Architecture**: Leverages the strengths of both SQL and NoSQL databases
- **Enterprise-Grade Security**: JWT authentication with role-based access control (RBAC)
- **Workflow Management**: Complete achievement lifecycle from draft to verification
- **Clean Architecture**: Maintainable, testable, and scalable codebase
- **Real-time Analytics**: Comprehensive reporting and statistics

---

## ✨ Features

### Core Functionality

- **🔐 Authentication & Authorization**
  - JWT-based authentication with refresh token support
  - Role-based access control (Student, Lecturer, Admin)
  - Secure password hashing

- **📊 Achievement Management**
  - Full CRUD operations for student achievements
  - Multi-stage workflow: Draft → Submitted → Verified/Rejected
  - Support for various achievement types (competitions, seminars, organizations)
  - File attachment support for certificates and evidence

- **👥 User Management**
  - Comprehensive user administration
  - Student-advisor relationship management
  - Role assignment and permission management

- **📈 Reporting & Analytics**
  - Global achievement statistics
  - Individual student performance reports
  - MongoDB aggregation pipelines for complex queries
  - Historical tracking of achievement status changes

- **🛡️ Data Integrity**
  - Dual-write mechanism ensuring data synchronization
  - Soft delete strategy for data recovery
  - Transaction support for critical operations

---

## 🏗️ Architecture

### Technology Stack

| Component | Technology |
|-----------|-----------|
| **Language** | Go 1.20+ |
| **Web Framework** | Fiber v2 |
| **SQL Database** | PostgreSQL |
| **NoSQL Database** | MongoDB |
| **Authentication** | JWT |
| **Database Drivers** | lib/pq, mongo-driver |

### Database Design

**PostgreSQL** - Relational Data
- User accounts and authentication
- Role definitions and assignments
- Student and lecturer profiles
- Achievement reference tracking and status

**MongoDB** - Document Storage
- Dynamic achievement details
- Flexible schemas for various achievement types
- Attachment metadata
- Audit logs and history

### Project Structure

```
student-performance-report
├── app       
│   ├── models           # Data Structures / Entities
│   │   ├── mongodb      # Structs for MongoDB collections (Achievement details)
│   │   └── postgresql   # Structs for SQL tables (Users, Roles, References)
│   ├── repository       # Data Access Layer (Database Queries)
│   │   ├── mongodb      # Implementation of MongoDB operations
│   │   └── postgresql   # Implementation of PostgreSQL operations
│   └── service          # Business Logic Layer
│       ├── mongodb      # Services handling MongoDB logic (Achievement, Reports)
│       └── postgresql   # Services handling SQL logic (Auth, Admin, Student)     
├── config               # Configuration setup (Env, JWT)
├── database             # Connection logic for MongoDB & PostgreSQL
├── docs                 # API Documentation files
├── fiber                # Fiber app configuration
├── middleware           # Auth & Role-Based Access Control (RBAC)
├── pwhash               # Password hashing utilities
├── route                # API Endpoint definitions
├── uploads              # Directory for static file storage (attachments)
├── utils                # Helper functions (Token generators, Validators)
├── .env                 # Environment variables configuration
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
└── main.go              # Application entry point
```

---

## 📡 API Documentation

### Authentication

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "student123",
  "password": "securepassword"
}
```

#### Get Profile
```http
GET /api/v1/auth/profile
Authorization: Bearer <token>
```

### Achievement Management

#### Create Achievement (Student)
```http
POST /api/v1/achievements
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "First Place - National Programming Competition",
  "type": "competition",
  "achievement_date": "2024-11-15",
  "description": "Achieved first place in national programming competition",
  "details": {
    "event_name": "CodeFest 2024",
    "organizer": "Indonesia Computer Society",
    "level": "national"
  }
}
```

#### Submit for Verification
```http
POST /api/v1/achievements/:id/submit
Authorization: Bearer <token>
```

#### Verify Achievement (Lecturer)
```http
POST /api/v1/achievements/:id/verify
Authorization: Bearer <token>
Content-Type: application/json

{
  "notes": "Achievement verified with supporting documents"
}
```

#### List Achievements
```http
GET /api/v1/achievements?status=verified&type=competition
Authorization: Bearer <token>
```

### User Management (Admin)

#### Create User
```http
POST /api/v1/users
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "newstudent",
  "email": "student@university.edu",
  "password": "securepassword",
  "role_id": 1
}
```

### Complete API Reference

| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| **Authentication** |
| POST | `/api/v1/auth/login` | User login & JWT generation | Public |
| POST | `/api/v1/auth/refresh` | Refresh access token | Public |
| POST | `/api/v1/auth/logout` | Logout session | Authenticated |
| GET | `/api/v1/auth/profile` | Get current user profile | Authenticated |
| **Users** |
| GET | `/api/v1/users` | List all users | Admin |
| GET | `/api/v1/users/:id` | Get user by ID | Admin |
| POST | `/api/v1/users` | Create new user | Admin |
| PUT | `/api/v1/users/:id` | Update user | Admin |
| DELETE | `/api/v1/users/:id` | Delete user | Admin |
| PUT | `/api/v1/users/:id/role` | Assign role | Admin |
| **Achievements** |
| GET | `/api/v1/achievements` | List achievements | All |
| GET | `/api/v1/achievements/:id` | Get achievement detail | All |
| POST | `/api/v1/achievements` | Create achievement | Student |
| PUT | `/api/v1/achievements/:id` | Update achievement | Student |
| DELETE | `/api/v1/achievements/:id` | Delete achievement | Student |
| POST | `/api/v1/achievements/:id/submit` | Submit for verification | Student |
| POST | `/api/v1/achievements/:id/verify` | Verify achievement | Lecturer |
| POST | `/api/v1/achievements/:id/reject` | Reject achievement | Lecturer |
| GET | `/api/v1/achievements/:id/history` | View status history | All |
| POST | `/api/v1/achievements/:id/attachments` | Upload attachments | Student |
| **Students & Lecturers** |
| GET | `/api/v1/students` | List students | Authorized |
| GET | `/api/v1/students/:id` | Get student profile | Authorized |
| GET | `/api/v1/students/:id/achievements` | Get student achievements | Authorized |
| PUT | `/api/v1/students/:id/advisor` | Assign advisor | Admin |
| GET | `/api/v1/lecturers` | List lecturers | Authorized |
| GET | `/api/v1/lecturers/:id/advisees` | Get advisees | Lecturer/Admin |
| **Reports** |
| GET | `/api/v1/reports/statistics` | Global statistics | Admin |
| GET | `/api/v1/reports/student/:id` | Student performance report | Admin/Lecturer/Owner |

---

## 🔒 Security

### Authentication Flow

1. User submits credentials via `/api/v1/auth/login`
2. Server validates credentials and generates JWT access token
3. Client includes token in `Authorization: Bearer <token>` header
4. Middleware validates token and extracts user information
5. RBAC middleware checks user permissions for requested resource

### Role-Based Access Control

| Role | Permissions |
|------|------------|
| **Student (Mahasiswa)** | Create and manage own achievements, view own profile, submit for verification |
| **Lecturer (Dosen Wali)** | Verify/reject advisee achievements, view advisee data and reports |
| **Admin** | Full system access, user management, global statistics, all CRUD operations |

### Security Best Practices

- ✅ Password hashing using bcrypt
- ✅ JWT token expiration and refresh mechanism
- ✅ Input validation and sanitization
- ✅ SQL injection prevention through prepared statements
- ✅ NoSQL injection prevention through driver validation
- ✅ Role-based endpoint protection
- ✅ Soft delete for data recovery

---

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./app/service/...
```

---

## 👥 Authors

**Surya Dwi Satria** - *Initial work* - [masterlearn22](https://github.com/]masterlearn22)

---

## 🙏 Acknowledgments

- [Fiber](https://gofiber.io/) - Web framework
- [PostgreSQL](https://www.postgresql.org/) - SQL database
- [MongoDB](https://www.mongodb.com/) - NoSQL database
- Go community for excellent libraries and tools
