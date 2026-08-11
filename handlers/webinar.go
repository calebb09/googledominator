package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"googledominator-backend/db"
)

type CreateWebinarRegistrationInput struct {
	FirstName        string `json:"first_name" binding:"required"`
	LastName         string `json:"last_name" binding:"required"`
	Email            string `json:"email" binding:"required,email"`
	Phone            string `json:"phone" binding:"required"`
	RegistrationType string `json:"registration_type" binding:"required"`
	AdditionalInfo   string `json:"additional_info"`
	Agreed           *bool  `json:"agreed" binding:"required"`
}

type WebinarRegistrationResponse struct {
	ID               string    `json:"id"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	RegistrationType string    `json:"registration_type"`
	AdditionalInfo   string    `json:"additional_info,omitempty"`
	Agreed           bool      `json:"agreed"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

var (
	mockWebinarLock         sync.RWMutex
	mockWebinarRegistrations = []WebinarRegistrationResponse{
		{
			ID:               "webinar-reg-001",
			FirstName:        "Alex",
			LastName:         "Morgan",
			Email:            "alex.morgan@example.com",
			Phone:            "+1-555-012-3456",
			RegistrationType: "tax professional",
			AdditionalInfo:   "Interested in CPE credit for tax software automation.",
			Agreed:           true,
			Status:           "REGISTERED",
			CreatedAt:        time.Now().Add(-12 * time.Hour),
			UpdatedAt:        time.Now().Add(-12 * time.Hour),
		},
		{
			ID:               "webinar-reg-002",
			FirstName:        "Sarah",
			LastName:         "Connor",
			Email:            "sarah@cyberbiz.com",
			Phone:            "+1-555-987-6543",
			RegistrationType: "small business",
			AdditionalInfo:   "Looking to scale local tax services marketing.",
			Agreed:           true,
			Status:           "REGISTERED",
			CreatedAt:        time.Now().Add(-2 * time.Hour),
			UpdatedAt:        time.Now().Add(-2 * time.Hour),
		},
	}
)

var allowedRegistrationTypes = map[string]string{
	"tax professional": "tax professional",
	"small business":   "small business",
	"large enterprise": "large enterprise",
}

func isValidRegistrationType(regType string) (string, bool) {
	norm := strings.TrimSpace(strings.ToLower(regType))
	val, ok := allowedRegistrationTypes[norm]
	return val, ok
}

func CreateWebinarRegistration(c *gin.Context) {
	var input CreateWebinarRegistrationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Validate consent / agreed checkbox
	if input.Agreed == nil || !*input.Agreed {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "You must agree to the terms and conditions (agreed must be true)",
		})
		return
	}

	// Validate registration type
	canonicalType, valid := isValidRegistrationType(input.RegistrationType)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Invalid registration_type '%s'. Allowed values are: ['tax professional', 'small business', 'large enterprise']", input.RegistrationType),
		})
		return
	}

	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		opts := []db.WebinarRegistrationSetParam{
			db.WebinarRegistration.Agreed.Set(*input.Agreed),
		}
		if input.AdditionalInfo != "" {
			opts = append(opts, db.WebinarRegistration.AdditionalInfo.Set(input.AdditionalInfo))
		}

		reg, err := db.Instance.Prisma.WebinarRegistration.CreateOne(
			db.WebinarRegistration.FirstName.Set(input.FirstName),
			db.WebinarRegistration.LastName.Set(input.LastName),
			db.WebinarRegistration.Email.Set(input.Email),
			db.WebinarRegistration.Phone.Set(input.Phone),
			db.WebinarRegistration.RegistrationType.Set(canonicalType),
			opts...,
		).Exec(ctx)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "Webinar registration completed successfully",
			"data":    reg,
		})
		return
	}

	// Fallback mock memory mode
	mockWebinarLock.Lock()
	now := time.Now()
	newReg := WebinarRegistrationResponse{
		ID:               fmt.Sprintf("webinar-reg-%d", now.UnixNano()),
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		Email:            input.Email,
		Phone:            input.Phone,
		RegistrationType: canonicalType,
		AdditionalInfo:   input.AdditionalInfo,
		Agreed:           *input.Agreed,
		Status:           "REGISTERED",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	mockWebinarRegistrations = append(mockWebinarRegistrations, newReg)
	mockWebinarLock.Unlock()

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Webinar registration completed successfully (mock memory mode)",
		"data":    newReg,
	})
}

func GetWebinarRegistrations(c *gin.Context) {
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		regs, err := db.Instance.Prisma.WebinarRegistration.FindMany().Exec(ctx)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    regs,
			})
			return
		}
	}

	mockWebinarLock.RLock()
	defer mockWebinarLock.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    mockWebinarRegistrations,
	})
}

func GetWebinarRegistrationByID(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		reg, err := db.Instance.Prisma.WebinarRegistration.FindUnique(
			db.WebinarRegistration.ID.Equals(id),
		).Exec(ctx)

		if err == nil && reg != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    reg,
			})
			return
		}
	}

	mockWebinarLock.RLock()
	defer mockWebinarLock.RUnlock()

	for _, r := range mockWebinarRegistrations {
		if r.ID == id {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    r,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Webinar registration not found",
	})
}

type UpdateWebinarRegistrationInput struct {
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	Email            string `json:"email"`
	Phone            string `json:"phone"`
	RegistrationType string `json:"registration_type"`
	AdditionalInfo   string `json:"additional_info"`
	Agreed           *bool  `json:"agreed"`
	Status           string `json:"status"`
}

func UpdateWebinarRegistration(c *gin.Context) {
	id := c.Param("id")
	var input UpdateWebinarRegistrationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if input.RegistrationType != "" {
		canonicalType, valid := isValidRegistrationType(input.RegistrationType)
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Invalid registration_type '%s'. Allowed values are: ['tax professional', 'small business', 'large enterprise']", input.RegistrationType),
			})
			return
		}
		input.RegistrationType = canonicalType
	}

	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		var updates []db.WebinarRegistrationSetParam
		if input.FirstName != "" {
			updates = append(updates, db.WebinarRegistration.FirstName.Set(input.FirstName))
		}
		if input.LastName != "" {
			updates = append(updates, db.WebinarRegistration.LastName.Set(input.LastName))
		}
		if input.Email != "" {
			updates = append(updates, db.WebinarRegistration.Email.Set(input.Email))
		}
		if input.Phone != "" {
			updates = append(updates, db.WebinarRegistration.Phone.Set(input.Phone))
		}
		if input.RegistrationType != "" {
			updates = append(updates, db.WebinarRegistration.RegistrationType.Set(input.RegistrationType))
		}
		if input.AdditionalInfo != "" {
			updates = append(updates, db.WebinarRegistration.AdditionalInfo.Set(input.AdditionalInfo))
		}
		if input.Agreed != nil {
			updates = append(updates, db.WebinarRegistration.Agreed.Set(*input.Agreed))
		}
		if input.Status != "" {
			updates = append(updates, db.WebinarRegistration.Status.Set(input.Status))
		}

		reg, err := db.Instance.Prisma.WebinarRegistration.FindUnique(
			db.WebinarRegistration.ID.Equals(id),
		).Update(updates...).Exec(ctx)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Webinar registration updated successfully",
			"data":    reg,
		})
		return
	}

	mockWebinarLock.Lock()
	defer mockWebinarLock.Unlock()

	for i, r := range mockWebinarRegistrations {
		if r.ID == id {
			if input.FirstName != "" {
				mockWebinarRegistrations[i].FirstName = input.FirstName
			}
			if input.LastName != "" {
				mockWebinarRegistrations[i].LastName = input.LastName
			}
			if input.Email != "" {
				mockWebinarRegistrations[i].Email = input.Email
			}
			if input.Phone != "" {
				mockWebinarRegistrations[i].Phone = input.Phone
			}
			if input.RegistrationType != "" {
				mockWebinarRegistrations[i].RegistrationType = input.RegistrationType
			}
			if input.AdditionalInfo != "" {
				mockWebinarRegistrations[i].AdditionalInfo = input.AdditionalInfo
			}
			if input.Agreed != nil {
				mockWebinarRegistrations[i].Agreed = *input.Agreed
			}
			if input.Status != "" {
				mockWebinarRegistrations[i].Status = input.Status
			}
			mockWebinarRegistrations[i].UpdatedAt = time.Now()

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Webinar registration updated (mock memory mode)",
				"data":    mockWebinarRegistrations[i],
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Webinar registration not found",
	})
}

func DeleteWebinarRegistration(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		_, err := db.Instance.Prisma.WebinarRegistration.FindUnique(
			db.WebinarRegistration.ID.Equals(id),
		).Delete().Exec(ctx)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Webinar registration deleted successfully",
		})
		return
	}

	mockWebinarLock.Lock()
	defer mockWebinarLock.Unlock()

	for i, r := range mockWebinarRegistrations {
		if r.ID == id {
			mockWebinarRegistrations = append(mockWebinarRegistrations[:i], mockWebinarRegistrations[i+1:]...)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Webinar registration deleted (mock memory mode)",
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Webinar registration not found",
	})
}
