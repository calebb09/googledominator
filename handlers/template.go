package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"googledominator-backend/db"
)

type TemplateResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Subtitle    string    `json:"subtitle,omitempty"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateTemplateJSONInput struct {
	Title       string `json:"title" binding:"required"`
	Subtitle    string `json:"subtitle"`
	Description string `json:"description" binding:"required"`
	Image       string `json:"image_url"`
}

type UpdateTemplateJSONInput struct {
	Title       string `json:"title"`
	Subtitle    string `json:"subtitle"`
	Description string `json:"description"`
	Image       string `json:"image_url"`
	IsActive    *bool  `json:"is_active"`
}

var (
	mockTemplatesLock sync.RWMutex
	mockTemplates     = []TemplateResponse{
		{
			ID:          "tpl-001-tax-pro",
			Title:       "Tax Practice Dominator Pro",
			Subtitle:    "High-converting turnkey tax practice website template",
			Description: "Optimized for CPA firms, tax preparers, and bookkeeping practices to capture local search leads.",
			Image:       "/uploads/templates/tax_practice_dominator.jpg",
			IsActive:    true,
			CreatedAt:   time.Now().Add(-72 * time.Hour),
			UpdatedAt:   time.Now().Add(-72 * time.Hour),
		},
		{
			ID:          "tpl-002-small-biz",
			Title:       "Small Business Growth Template",
			Subtitle:    "Clean modern web design for growing businesses",
			Description: "Built-in appointment scheduling, customer review widgets, and Google Business Profile sync.",
			Image:       "/uploads/templates/small_biz_growth.jpg",
			IsActive:    true,
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			UpdatedAt:   time.Now().Add(-48 * time.Hour),
		},
	}
)

const UploadDir = "./uploads/templates"

func init() {
	_ = os.MkdirAll(UploadDir, os.ModePerm)
}

// CreateTemplate handles POST /api/v1/templates (supports multipart/form-data & application/json)
func CreateTemplate(c *gin.Context) {
	ctx := c.Request.Context()
	var title, subtitle, description, imageURL string

	contentType := c.GetHeader("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		title = c.PostForm("title")
		subtitle = c.PostForm("subtitle")
		description = c.PostForm("description")
		imageURL = c.PostForm("image_url")

		file, err := c.FormFile("image")
		if err == nil && file != nil {
			_ = os.MkdirAll(UploadDir, os.ModePerm)
			ext := filepath.Ext(file.Filename)
			if ext == "" {
				ext = ".jpg"
			}
			filename := fmt.Sprintf("template_%d%s", time.Now().UnixNano(), ext)
			dst := filepath.Join(UploadDir, filename)

			if err := c.SaveUploadedFile(file, dst); err == nil {
				imageURL = "/uploads/templates/" + filename
			}
		}
	} else {
		var input CreateTemplateJSONInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		title = input.Title
		subtitle = input.Subtitle
		description = input.Description
		imageURL = input.Image
	}

	if strings.TrimSpace(title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Title is required",
		})
		return
	}

	if strings.TrimSpace(description) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Description is required",
		})
		return
	}

	if strings.TrimSpace(imageURL) == "" {
		imageURL = "/uploads/templates/default.jpg"
	}

	// Save to DB if connected
	if db.Instance != nil && db.Instance.IsConnected {
		var optionalParams []db.TemplateSetParam
		if subtitle != "" {
			optionalParams = append(optionalParams, db.Template.Subtitle.Set(subtitle))
		}
		optionalParams = append(optionalParams, db.Template.IsActive.Set(true))

		created, err := db.Instance.Prisma.Template.CreateOne(
			db.Template.Title.Set(title),
			db.Template.Description.Set(description),
			db.Template.Image.Set(imageURL),
			optionalParams...,
		).Exec(ctx)

		if err == nil && created != nil {
			sub, _ := created.Subtitle()
			c.JSON(http.StatusCreated, gin.H{
				"success": true,
				"message": "Template created successfully",
				"data": TemplateResponse{
					ID:          created.ID,
					Title:       created.Title,
					Subtitle:    sub,
					Description: created.Description,
					Image:       created.Image,
					IsActive:    created.IsActive,
					CreatedAt:   created.CreatedAt,
					UpdatedAt:   created.UpdatedAt,
				},
			})
			return
		}
	}

	// Fallback in-memory store
	mockTemplatesLock.Lock()
	newTpl := TemplateResponse{
		ID:          fmt.Sprintf("tpl-%d", time.Now().UnixNano()),
		Title:       title,
		Subtitle:    subtitle,
		Description: description,
		Image:       imageURL,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	mockTemplates = append(mockTemplates, newTpl)
	mockTemplatesLock.Unlock()

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Template created successfully",
		"data":    newTpl,
	})
}

// GetTemplates handles GET /api/v1/templates
func GetTemplates(c *gin.Context) {
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		templates, err := db.Instance.Prisma.Template.FindMany(
			db.Template.IsActive.Equals(true),
		).Exec(ctx)

		if err == nil {
			res := make([]TemplateResponse, 0, len(templates))
			for _, t := range templates {
				sub, _ := t.Subtitle()
				res = append(res, TemplateResponse{
					ID:          t.ID,
					Title:       t.Title,
					Subtitle:    sub,
					Description: t.Description,
					Image:       t.Image,
					IsActive:    t.IsActive,
					CreatedAt:   t.CreatedAt,
					UpdatedAt:   t.UpdatedAt,
				})
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    res,
			})
			return
		}
	}

	mockTemplatesLock.RLock()
	defer mockTemplatesLock.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    mockTemplates,
	})
}

// GetTemplateByID handles GET /api/v1/templates/:id
func GetTemplateByID(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		t, err := db.Instance.Prisma.Template.FindUnique(
			db.Template.ID.Equals(id),
		).Exec(ctx)

		if err == nil && t != nil {
			sub, _ := t.Subtitle()
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": TemplateResponse{
					ID:          t.ID,
					Title:       t.Title,
					Subtitle:    sub,
					Description: t.Description,
					Image:       t.Image,
					IsActive:    t.IsActive,
					CreatedAt:   t.CreatedAt,
					UpdatedAt:   t.UpdatedAt,
				},
			})
			return
		}
	}

	mockTemplatesLock.RLock()
	defer mockTemplatesLock.RUnlock()

	for _, t := range mockTemplates {
		if t.ID == id {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    t,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Template not found",
	})
}

// UpdateTemplate handles PUT /api/v1/templates/:id
func UpdateTemplate(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	var title, subtitle, description, imageURL string
	var isActive *bool

	contentType := c.GetHeader("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		title = c.PostForm("title")
		subtitle = c.PostForm("subtitle")
		description = c.PostForm("description")
		imageURL = c.PostForm("image_url")
		if act := c.PostForm("is_active"); act != "" {
			val := act == "true" || act == "1"
			isActive = &val
		}

		file, err := c.FormFile("image")
		if err == nil && file != nil {
			ext := filepath.Ext(file.Filename)
			if ext == "" {
				ext = ".jpg"
			}
			filename := fmt.Sprintf("template_%d%s", time.Now().UnixNano(), ext)
			dst := filepath.Join(UploadDir, filename)

			if err := c.SaveUploadedFile(file, dst); err == nil {
				imageURL = "/uploads/templates/" + filename
			}
		}
	} else {
		var input UpdateTemplateJSONInput
		if err := c.ShouldBindJSON(&input); err == nil {
			title = input.Title
			subtitle = input.Subtitle
			description = input.Description
			imageURL = input.Image
			isActive = input.IsActive
		}
	}

	if db.Instance != nil && db.Instance.IsConnected {
		var updates []db.TemplateSetParam
		if title != "" {
			updates = append(updates, db.Template.Title.Set(title))
		}
		if subtitle != "" {
			updates = append(updates, db.Template.Subtitle.Set(subtitle))
		}
		if description != "" {
			updates = append(updates, db.Template.Description.Set(description))
		}
		if imageURL != "" {
			updates = append(updates, db.Template.Image.Set(imageURL))
		}
		if isActive != nil {
			updates = append(updates, db.Template.IsActive.Set(*isActive))
		}

		if len(updates) > 0 {
			updated, err := db.Instance.Prisma.Template.FindUnique(
				db.Template.ID.Equals(id),
			).Update(updates...).Exec(ctx)

			if err == nil && updated != nil {
				sub, _ := updated.Subtitle()
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Template updated successfully",
					"data": TemplateResponse{
						ID:          updated.ID,
						Title:       updated.Title,
						Subtitle:    sub,
						Description: updated.Description,
						Image:       updated.Image,
						IsActive:    updated.IsActive,
						CreatedAt:   updated.CreatedAt,
						UpdatedAt:   updated.UpdatedAt,
					},
				})
				return
			}
		}
	}

	mockTemplatesLock.Lock()
	defer mockTemplatesLock.Unlock()

	for i, t := range mockTemplates {
		if t.ID == id {
			if title != "" {
				mockTemplates[i].Title = title
			}
			if subtitle != "" {
				mockTemplates[i].Subtitle = subtitle
			}
			if description != "" {
				mockTemplates[i].Description = description
			}
			if imageURL != "" {
				mockTemplates[i].Image = imageURL
			}
			if isActive != nil {
				mockTemplates[i].IsActive = *isActive
			}
			mockTemplates[i].UpdatedAt = time.Now()

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Template updated successfully",
				"data":    mockTemplates[i],
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Template not found",
	})
}

// DeleteTemplate handles DELETE /api/v1/templates/:id
func DeleteTemplate(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		_, err := db.Instance.Prisma.Template.FindUnique(
			db.Template.ID.Equals(id),
		).Delete().Exec(ctx)

		if err == nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Template deleted successfully",
			})
			return
		}
	}

	mockTemplatesLock.Lock()
	defer mockTemplatesLock.Unlock()

	for i, t := range mockTemplates {
		if t.ID == id {
			mockTemplates = append(mockTemplates[:i], mockTemplates[i+1:]...)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Template deleted successfully",
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Template not found",
	})
}
