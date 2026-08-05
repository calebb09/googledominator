package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"googledominator-backend/config"
	"googledominator-backend/db"
	"googledominator-backend/middleware"
)

type AdminRegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type AdminLoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AdminResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type MockAdminUser struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

var (
	mockAdminsLock sync.RWMutex
	mockAdmins     = map[string]*MockAdminUser{}
)

func init() {
	// Seed default admin in mock store (password: "Admin123!")
	hashedPass, _ := bcrypt.GenerateFromPassword([]byte("Admin123!"), bcrypt.DefaultCost)
	mockAdmins["admin@googledominator.co"] = &MockAdminUser{
		ID:           "admin-default-001",
		Name:         "GoogleDominator Admin",
		Email:        "admin@googledominator.co",
		PasswordHash: string(hashedPass),
		Role:         "ADMIN",
		CreatedAt:    time.Now(),
	}
}

func AdminRegister(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input AdminRegisterInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to process password",
			})
			return
		}

		ctx := c.Request.Context()

		if db.Instance != nil && db.Instance.IsConnected {
			existing, _ := db.Instance.Prisma.Admin.FindUnique(
				db.Admin.Email.Equals(input.Email),
			).Exec(ctx)

			if existing != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "Admin user with this email already exists",
				})
				return
			}

			admin, err := db.Instance.Prisma.Admin.CreateOne(
				db.Admin.Email.Set(input.Email),
				db.Admin.Password.Set(string(hashedPassword)),
				db.Admin.Name.Set(input.Name),
				db.Admin.Role.Set("ADMIN"),
			).Exec(ctx)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Failed to create admin record: " + err.Error(),
				})
				return
			}

			token, err := middleware.GenerateToken(admin.ID, admin.Email, admin.Role, cfg.JWTSecret)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Failed to generate authentication token",
				})
				return
			}

			c.JSON(http.StatusCreated, gin.H{
				"success": true,
				"message": "Admin registered successfully",
				"data": AdminResponse{
					ID:        admin.ID,
					Name:      admin.Name,
					Email:     admin.Email,
					Role:      admin.Role,
					CreatedAt: admin.CreatedAt,
				},
				"token": token,
			})
			return
		}

		mockAdminsLock.Lock()
		defer mockAdminsLock.Unlock()

		if _, exists := mockAdmins[input.Email]; exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Admin user with this email already exists",
			})
			return
		}

		newAdmin := &MockAdminUser{
			ID:           "admin-" + time.Now().Format("20060102150405"),
			Name:         input.Name,
			Email:        input.Email,
			PasswordHash: string(hashedPassword),
			Role:         "ADMIN",
			CreatedAt:    time.Now(),
		}
		mockAdmins[input.Email] = newAdmin

		token, err := middleware.GenerateToken(newAdmin.ID, newAdmin.Email, newAdmin.Role, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to generate token",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "Admin registered successfully (mock memory mode)",
			"data": AdminResponse{
				ID:        newAdmin.ID,
				Name:      newAdmin.Name,
				Email:     newAdmin.Email,
				Role:      newAdmin.Role,
				CreatedAt: newAdmin.CreatedAt,
			},
			"token": token,
		})
	}
}

func AdminLogin(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input AdminLoginInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		ctx := c.Request.Context()

		if db.Instance != nil && db.Instance.IsConnected {
			admin, err := db.Instance.Prisma.Admin.FindUnique(
				db.Admin.Email.Equals(input.Email),
			).Exec(ctx)

			if err != nil || admin == nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "Invalid email or password",
				})
				return
			}

			if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(input.Password)); err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "Invalid email or password",
				})
				return
			}

			token, err := middleware.GenerateToken(admin.ID, admin.Email, admin.Role, cfg.JWTSecret)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Failed to generate authentication token",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Admin logged in successfully",
				"data": AdminResponse{
					ID:        admin.ID,
					Name:      admin.Name,
					Email:     admin.Email,
					Role:      admin.Role,
					CreatedAt: admin.CreatedAt,
				},
				"token": token,
			})
			return
		}

		mockAdminsLock.RLock()
		admin, exists := mockAdmins[input.Email]
		mockAdminsLock.RUnlock()

		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid email or password",
			})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(input.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid email or password",
			})
			return
		}

		token, err := middleware.GenerateToken(admin.ID, admin.Email, admin.Role, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to generate authentication token",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Admin logged in successfully (mock memory mode)",
			"data": AdminResponse{
				ID:        admin.ID,
				Name:      admin.Name,
				Email:     admin.Email,
				Role:      admin.Role,
				CreatedAt: admin.CreatedAt,
			},
			"token": token,
		})
	}
}
