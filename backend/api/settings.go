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
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func (api *API) registerSettingsRoutes() {
	api.routes.POST("/settings", api.updateSettings)

	api.routes.GET("/settings", api.getSettings)
	switch api.Config.DB.DbType {
	case "sqlite":
		api.routes.POST("/sqlite/import", api.importSQLiteDatabase)
		api.routes.GET("/sqlite/export", api.exportSQLiteDatabase)
	case "postgres":
		api.routes.POST("/postgres/import", api.importPostgresDatabase)
		api.routes.GET("/postgres/export", api.exportPostgresDatabase)
	}
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

	api.Config.UpdateSettings(updatedSettings)
	settingsJson, _ := json.MarshalIndent(updatedSettings, "", "  ")
	log.Debug("%s", string(settingsJson))

	api.DNSServer.Audits.CreateAudit(&audit.Entry{
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

func (api *API) exportSQLiteDatabase(c *gin.Context) {
	if api.Config.DB.DbType != "sqlite" {
		log.Error("SQLite database is not active")
		c.JSON(http.StatusNotImplemented, gin.H{"error": "SQLite database is not active"})
		return
	}

	log.Debug("Starting export of database")

	// Temporary filename for export the database into
	tempExport := "export_temp.db"

	// remove in case it already exists, otherwise VACUUM INTO will fail
	_ = os.Remove(tempExport)

	// Create a new connection to a temp file and vacuum into it
	if err := api.DBManager.Conn.Exec(fmt.Sprintf("VACUUM INTO '%s';", tempExport)).Error; err != nil {
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

	api.DNSServer.Audits.CreateAudit(&audit.Entry{
		Topic:   audit.TopicDatabase,
		Message: "Database was exported",
	})
}

func validateSQLiteFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open file: %v", err)
	}
	go func() {
		_ = file.Close()
	}()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat file: %v", err)
	}

	if stat.Size() < 50 {
		return fmt.Errorf("file too small to be a valid SQLite database")
	}

	header := make([]byte, 16)
	_, err = file.Read(header)
	if err != nil {
		return fmt.Errorf("cannot read file header: %v", err)
	}

	expectedHeader := "SQLite format 3\x00"
	if string(header) != expectedHeader {
		return fmt.Errorf("invalid SQLite header - file may be corrupted or not a SQLite database")
	}

	return nil
}

func (api *API) importSQLiteDatabase(c *gin.Context) {
	if api.Config.DB.DbType != "sqlite" {
		log.Error("SQLite database is not active")
		c.JSON(http.StatusNotImplemented, gin.H{"error": "SQLite database is not active"})
	}
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

	defer func(tempfile *os.File) {
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

	sqlDB, err := api.DBManager.Conn.DB()
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

	newDB, err := gorm.Open(sqlite.Open(currentDBPath), &gorm.Config{})
	if err != nil {
		log.Error("Failed to open imported database with GORM: %v", err)
		_ = copyFile(backupPath, currentDBPath)
		api.DBManager.Conn, _ = gorm.Open(sqlite.Open(currentDBPath), &gorm.Config{})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open imported database, restored from backup"})
		return
	}

	api.DBManager.Conn = newDB

	log.Info("Database imported successfully from %s", header.Filename)
	api.DNSServer.Audits.CreateAudit(&audit.Entry{
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

func (api *API) exportPostgresDatabase(c *gin.Context) {
	dbConfig := api.Config.DB
	if dbConfig.DbType != "postgres" {
		log.Error("Postgres database is not active")
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Postgres database is not active"})
		return
	}

	log.Debug("Starting export of database")

	_, ok := api.DBManager.Conn.Dialector.(*postgres.Dialector)
	if !ok {
		log.Error("Failed to get postgres dialector")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get postgres dialector"})
		return
	}

	// temp file for dump
	tempFile, err := os.CreateTemp("", "db_export_*.dump")
	if err != nil {
		log.Error("Failed to create temp file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp file"})
		return
	}
	defer func() {
		err := tempFile.Close()
		if err != nil {
			log.Error("Failed to close temp file: %v", err)
		}
		err = os.Remove(tempFile.Name())
		if err != nil {
			log.Error("Failed to remove temp file: %v", err)
		}
	}()

	// run pg_dump into temp file
	cmd := exec.Command(
		"pg_dump", "-Fc",
		"-f", tempFile.Name(),
		"-h", *dbConfig.Host,
		"-U", *dbConfig.User,
		"-d", *dbConfig.Database,
		"-p", strconv.Itoa(*dbConfig.Port), // you missed passing the port number!
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+*dbConfig.Pass)
	if err := cmd.Run(); err != nil {
		log.Error("Postgres dump failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Postgres dump failed"})
		return
	}

	// set headers
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=database.dump")
	c.Header("Cache-Control", "no-cache")
	info, err := tempFile.Stat()
	if err != nil {
		log.Error("Failed to stat temp file: %v", err)
	} else {
		c.Header("Content-Length", fmt.Sprintf("%d", info.Size()))
	}

	// stream temp file to client
	_, err = tempFile.Seek(0, 0)
	if err != nil {
		log.Error("Failed to reset temp file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset temp file"})
		return
	}
	if _, err := io.Copy(c.Writer, tempFile); err != nil {
		log.Error("Failed to stream dump to client: %v", err)
		return
	}

	api.DNSServer.Audits.CreateAudit(&audit.Entry{
		Topic:   audit.TopicDatabase,
		Message: "Database was exported",
	})

}

func (api *API) importPostgresDatabase(c *gin.Context) {
	dbConfig := api.Config.DB

	if dbConfig.DbType != "postgres" {
		log.Error("Postgres database is not active")
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Postgres database is not active"})
		return
	}

	log.Debug("Starting import of database")

	_, ok := api.DBManager.Conn.Dialector.(*postgres.Dialector)
	if !ok {
		log.Error("Failed to get postgres dialector")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get postgres dialector"})
		return
	}

	file, _, err := c.Request.FormFile("database")
	if err != nil {
		log.Error("Failed to get uploaded file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded or invalid file"})
		return
	}
	defer func() {
		_ = file.Close()
	}()

	tempFile, err := os.CreateTemp("", "db_import_*.dump")
	if err != nil {
		log.Error("Failed to create temp file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp file"})
		return
	}
	defer func() {
		err := tempFile.Close()
		if err != nil {
			log.Error("Failed to close temp file: %v", err)
		}
		err = os.Remove(tempFile.Name())
		if err != nil {
			log.Error("Failed to remove temp file: %v", err)
		}
	}()

	if _, err := io.Copy(tempFile, file); err != nil {
		log.Error("Failed to save uploaded file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
		return
	}

	checkCmd := exec.Command("pg_restore", "-l", tempFile.Name())
	if output, err := checkCmd.CombinedOutput(); err != nil {
		log.Error("Postgres dump validation failed: %v, output: %s", err, string(output))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Database integrity check failed",
			"details": string(output),
		})
		return
	}

	backupPath := filepath.Join("data", "database.dump") + ".backup." + time.Now().UTC().Format("2006-01-02_15:04:05")
	backupFile, err := os.Create(backupPath)
	if err != nil {
		log.Error("Failed to create backup temp file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create backup"})
		return
	}
	defer func() {
		err := backupFile.Close()
		if err != nil {
			log.Error("Failed to close backup file: %v", err)
		}
		err = os.Remove(backupFile.Name())
		if err != nil {
			log.Error("Failed to remove backup file: %v", err)
		}
	}()

	backupCmd := exec.Command(
		"pg_dump", "-Fc",
		"-f", backupFile.Name(),
		"-h", *dbConfig.Host,
		"-U", *dbConfig.User,
		"-d", *dbConfig.Database,
		"-p", strconv.Itoa(*dbConfig.Port), // you missed passing the port number!
	)
	backupCmd.Env = append(os.Environ(), "PGPASSWORD="+*dbConfig.Pass)
	if err := backupCmd.Run(); err != nil {
		log.Error("Failed to backup current DB: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to backup current database"})
		return
	}

	restoreTemplate := []string{
		"-Fc",
		"-c", "--if-exists",
		"-h", *dbConfig.Host,
		"-U", *dbConfig.User,
		"-d", *dbConfig.Database,
		"-p", strconv.Itoa(*dbConfig.Port),
	}
	// restore uploaded DB
	restoreArgs := append(restoreTemplate, tempFile.Name())
	restoreCmd := exec.Command(
		"pg_restore", restoreArgs...,
	)
	restoreCmd.Env = append(os.Environ(), "PGPASSWORD="+*dbConfig.Pass)
	if output, err := restoreCmd.CombinedOutput(); err != nil {
		log.Error("Postgres restore failed: %v, output: %s", err, string(output))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Postgres restore failed",
			"details": string(output),
		})
		// attempt restore from backup
		restoreBackupArgs := append(restoreTemplate, tempFile.Name())
		restoreCmd = exec.Command(
			"pg_restore", restoreBackupArgs...,
		)
		restoreCmd.Env = append(os.Environ(), "PGPASSWORD="+*dbConfig.Pass)
		err = restoreCmd.Run()
		if err != nil {
			log.Error("Postgres restore failed: %v, output: %s", err, string(output))
			os.Exit(1)
		}
		return
	}

	api.DNSServer.Audits.CreateAudit(&audit.Entry{
		Topic:   audit.TopicDatabase,
		Message: "Database was imported",
	})

	c.JSON(http.StatusOK, gin.H{
		"message":        "Database imported successfully",
		"backup_created": backupFile.Name(),
	})
}
