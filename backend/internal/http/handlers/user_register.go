package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	sqlite3 "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	modernSQLite "modernc.org/sqlite"
	sqliteLib "modernc.org/sqlite/lib"

	"github.com/ngocan-dev/mangahub/backend/domain/user"
	"github.com/ngocan-dev/mangahub/backend/internal/security"
)

type UserHandler struct {
	DB *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{DB: db}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req user.RegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("register: invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "username, email, and password are required",
		})
		return
	}

	if err := security.ValidateUsername(req.Username); err != nil {
		log.Printf("register: invalid username: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// A2
	if err := user.ValidateEmail(req.Email); err != nil {
		log.Printf("register: invalid email: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// A3
	if err := user.ValidatePassword(req.Password); err != nil {
		log.Printf("register: weak password: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := user.Register(c.Request.Context(), h.DB, req)
	if err != nil {
		// A1
		if errors.Is(err, user.ErrDuplicateUsername) {
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
			return
		}
		if errors.Is(err, user.ErrDuplicateEmail) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}

		// Validation or data errors
		if isValidationError(err) {
			log.Printf("register: validation failed: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if errors.Is(err, user.ErrPasswordHashing) || errors.Is(err, bcrypt.ErrPasswordTooShort) {
			log.Printf("register: password hashing failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
			return
		}

		if isSQLiteConstraintError(err) {
			log.Printf("register: constraint violation: %v", err)
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
			return
		}

		log.Printf("register: unexpected error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Step 5: success
	c.JSON(http.StatusCreated, u)
}

func isValidationError(err error) bool {
	return errors.Is(err, security.ErrInvalidFormat) ||
		errors.Is(err, security.ErrInputTooShort) ||
		errors.Is(err, security.ErrInputTooLong) ||
		errors.Is(err, security.ErrContainsSQLInjection) ||
		errors.Is(err, security.ErrContainsXSS)
}

func isSQLiteConstraintError(err error) bool {
	var sqliteErr *sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.Code == sqlite3.ErrConstraint ||
			sqliteErr.ExtendedCode == sqlite3.ErrConstraint ||
			sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique ||
			sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey
	}

	var modernErr *modernSQLite.Error
	if errors.As(err, &modernErr) {
		code := modernErr.Code()
		return code == sqliteLib.SQLITE_CONSTRAINT ||
			code == sqliteLib.SQLITE_CONSTRAINT_UNIQUE ||
			code == sqliteLib.SQLITE_CONSTRAINT_COMMITHOOK ||
			code == sqliteLib.SQLITE_CONSTRAINT_FOREIGNKEY ||
			code == sqliteLib.SQLITE_CONSTRAINT_PRIMARYKEY
	}

	return strings.Contains(strings.ToLower(err.Error()), "constraint")
}
