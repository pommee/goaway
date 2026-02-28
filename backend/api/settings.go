package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"goaway/backend/audit"
	"goaway/backend/settings"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func (api *API) registerSettingsRoutes() {
	api.routes.POST("/settings", api.updateSettings)
	api.routes.POST("/importDatabase", api.importDatabase)

	api.routes.GET("/settings", api.getSettings)
	api.routes.GET("/exportDatabase", api.exportDatabase)
}

func (api *API) updateSettings(c *gin.Context) {
	var updatedSettings settings.Config
	if err := c.BindJSON(&updatedSettings); err != nil {
		log.Warning("Could not save new settings, reason: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid settings data",
		})
		return
	}

	api.Config.Update(updatedSettings)
	settingsJSON, _ := json.MarshalIndent(updatedSettings, "", "  ")
	log.Debug("%s", string(settingsJSON))

	api.DNSServer.AuditService.CreateAudit(&audit.Entry{
		Topic:   audit.TopicSettings,
		Message: "Settings was updated",
	})

	log.Info("Settings have been updated")
	c.JSON(http.StatusOK, gin.H{
		"config": api.Config,
	})
}

func (api *API) getSettings(c *gin.Context) {
	c.JSON(http.StatusOK, api.Config)
}

func (api *API) exportDatabase(c *gin.Context) {
	log.Debug("Starting export of database")

	// Temporary filename for export the database into
	tempExport := "export_temp.db"

	// remove in case it already exists, otherwise VACUUM INTO will fail
	_ = os.Remove(tempExport)

	// Create a new connection to a temp file and vacuum into it
	if err := api.DBConn.Exec(fmt.Sprintf("VACUUM INTO '%s';", tempExport)).Error; err != nil {
		log.Error("Failed to write WAL to temp export: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare database for export"})
		return
	}

	file, err := os.Open(tempExport)
	if err != nil {
		log.Error("Error opening database export file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	defer func() {
		_ = file.Close()
		// remove the temporary export file after sending it
		_ = os.Remove(tempExport)
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		log.Error("Error getting file info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename=database.db")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Header("Cache-Control", "no-cache")

	c.Stream(func(w io.Writer) bool {
		buffer := make([]byte, 32*1024)
		n, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Error("Error reading file during stream: %v", err)
			}
			return false
		}

		_, writeErr := w.Write(buffer[:n])
		if writeErr != nil {
			log.Error("Error writing to response stream: %v", writeErr)
			return false
		}

		return n > 0
	})

	api.DNSServer.AuditService.CreateAudit(&audit.Entry{
		Topic:   audit.TopicDatabase,
		Message: "Database was exported",
	})
}

func validateSQLiteFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	go func() {
		_ = file.Close()
	}()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat file: %w", err)
	}

	if stat.Size() < 50 {
		return fmt.Errorf("file too small to be a valid SQLite database")
	}

	header := make([]byte, 16)
	_, err = file.Read(header)
	if err != nil {
		return fmt.Errorf("cannot read file header: %w", err)
	}

	expectedHeader := "SQLite format 3\x00"
	if string(header) != expectedHeader {
		return fmt.Errorf("invalid SQLite header - file may be corrupted or not a SQLite database")
	}

	return nil
}

func (api *API) importDatabase(c *gin.Context) {
	log.Info("Starting import of database")

	file, header, err := c.Request.FormFile("database")
	if err != nil {
		log.Error("Failed to get uploaded file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded or invalid file"})
		return
	}
	defer func() {
		_ = file.Close()
	}()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".db") {
		log.Error("Invalid file extension: %s", header.Filename)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .db files are allowed"})
		return
	}

	tempImport := "import_temp.db"
	_ = os.Remove(tempImport)

	tempFile, err := os.Create(tempImport)
	if err != nil {
		log.Error("Failed to create temporary file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temporary file"})
		return
	}
	_, err = io.Copy(tempFile, file)

	defer func(tempFile *os.File) {
		_ = tempFile.Close()
	}(tempFile)

	if err != nil {
		log.Error("Failed to copy uploaded file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process uploaded file"})
		return
	}

	if err := validateSQLiteFile(tempImport); err != nil {
		log.Error("SQLite file validation failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SQLite database file: " + err.Error()})
		return
	}

	testDB, err := sql.Open("sqlite", tempImport)
	if err != nil {
		log.Error("Failed to open uploaded database: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid database file"})
		return
	}

	defer func(testDB *sql.DB) {
		_ = testDB.Close()
	}(testDB)

	if err := testDB.Ping(); err != nil {
		log.Error("Failed to ping uploaded database: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Corrupted or invalid database file"})
		return
	}

	var integrityResult string
	if err := testDB.QueryRow("PRAGMA integrity_check").Scan(&integrityResult); err != nil {
		log.Error("Failed to run integrity check: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Database integrity check failed"})
		return
	}
	if integrityResult != "ok" {
		log.Error("Database integrity check failed: %s", integrityResult)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Database integrity check failed: " + integrityResult})
		return
	}

	sqlDB, err := api.DBConn.DB()
	if err != nil {
		log.Error("Failed to get underlying sql.DB: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close current database"})
		return
	}

	defer func(sqlDB *sql.DB) {
		_ = sqlDB.Close()
	}(sqlDB)

	currentDBPath := filepath.Join("data", "database.db")
	backupPath := currentDBPath + ".backup." + time.Now().UTC().Format("2006-01-02_15:04:05")

	if err := copyFile(currentDBPath, backupPath); err != nil {
		log.Error("Failed to create backup: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create backup of current database"})
		return
	}

	if err := copyFile(tempImport, currentDBPath); err != nil {
		log.Error("Failed to replace database: %v", err)
		_ = copyFile(backupPath, currentDBPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import database, restored from backup"})
		return
	}

	gormConf := &gorm.Config{TranslateError: true}
	newDB, err := gorm.Open(sqlite.Open(currentDBPath), gormConf)
	if err != nil {
		log.Error("Failed to open imported database with GORM: %v", err)
		_ = copyFile(backupPath, currentDBPath)
		api.DBConn, _ = gorm.Open(sqlite.Open(currentDBPath), gormConf)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open imported database, restored from backup"})
		return
	}

	*api.DBConn = *newDB

	log.Info("Database imported successfully from %s", header.Filename)
	api.DNSServer.AuditService.CreateAudit(&audit.Entry{
		Topic:   audit.TopicDatabase,
		Message: "Database was imported",
	})

	c.JSON(http.StatusOK, gin.H{
		"message":        "Database imported successfully",
		"backup_created": backupPath,
	})
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = sourceFile.Close()
	}()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = destFile.Close()
	}()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}
