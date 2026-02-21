package http

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UploadHandler struct {
	storagePath string
}

func NewUploadHandler(storagePath string) *UploadHandler {
	// Ensure storage directory exists
	os.MkdirAll(storagePath, 0755)

	return &UploadHandler{
		storagePath: storagePath,
	}
}

func (h *UploadHandler) UploadRecordingChunk(c *fiber.Ctx) error {

	recordingID := c.FormValue("recordingId")
	if recordingID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Recording ID is required",
		})
	}

	file, err := c.FormFile("chunk")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	chunkID := uuid.New().String()
	filename := fmt.Sprintf("%s_%s.webm", recordingID, chunkID)
	filepath := filepath.Join(h.storagePath, filename)

	// Save file
	if err := c.SaveFile(file, filepath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save file",
		})
	}

	// Return chunk URL
	chunkURL := fmt.Sprintf("/storage/%s", filename)

	return c.JSON(fiber.Map{
		"chunkId":  chunkID,
		"chunkUrl": chunkURL,
		"size":     file.Size,
	})
}

// ServeRecording serves a recording file
func (h *UploadHandler) ServeRecording(c *fiber.Ctx) error {
	filename := c.Params("filename")
	filepath := filepath.Join(h.storagePath, filename)

	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File not found",
		})
	}

	return c.SendFile(filepath)
}
