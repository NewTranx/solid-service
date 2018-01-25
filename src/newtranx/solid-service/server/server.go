package server

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"log"
	"path/filepath"
	"os"
	"io"
	"mime/multipart"
	"github.com/fsnotify/fsnotify"
	"context"
	"io/ioutil"
	"net/url"
)

const (
	SrcPath       = "src"
	OutputPath    = "out"
	ErrPath       = "err"
	FormFileField = "data"
)

type ServiceEndpoint struct {
	Host     string
	Port     int
	WorkPath string
	Cleanup  bool
	srcPath  string
	outPath  string
	errPath  string
}

func (s *ServiceEndpoint) Start() {
	s.srcPath = s.WorkPath + "/" + SrcPath
	s.outPath = s.WorkPath + "/" + OutputPath
	s.errPath = s.WorkPath + "/" + ErrPath

	router := gin.Default()

	v1 := router.Group("v1")

	{
		v1.POST("/convert/upload", s.handleUpload)
	}

	router.Run(fmt.Sprintf("%s:%d", s.Host, s.Port))
}

func (s *ServiceEndpoint) handleUpload(c *gin.Context) {
	id, err := uuid.NewV4()
	checkErr(err)
	if s.Cleanup {
		defer s.cleanup(id)
	}
	ctx, cancelCtx := newRequestContext(c, id)
	defer cancelCtx()
	log.Printf("<%s> begin", id.String())
	uploaded, err := c.FormFile(FormFileField)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	log.Printf("<%s> filename=%s, size=%d", id.String(), uploaded.Filename, uploaded.Size)
	fileExt := filepath.Ext(uploaded.Filename)[1:] // remove the dot
	baseName := uploaded.Filename[:len(uploaded.Filename)-(len(fileExt)+1)]
	if fileExt != "pdf" {
		c.String(http.StatusBadRequest, "not pdf")
		return
	}
	oneTimeFilePath := s.srcPath + "/" + id.String() + ".pdf"
	err = saveUploadedFile(uploaded, oneTimeFilePath)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("<%s> saved to %s", id.String(), oneTimeFilePath)
	err = waitForConversionDone(oneTimeFilePath)
	select {
	case <-ctx.Done():
		return
	default:
	}
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("<%s> conversion finished", id.String())
	errLogPath := s.errPath + "/" + id.String() + ".log"
	if _, err = os.Stat(errLogPath); os.IsNotExist(err) {
		resultFileName := baseName + ".docx"
		c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(resultFileName))
		c.File(s.outPath + "/" + id.String() + ".docx")
	} else {
		errFile, err := os.Open(errLogPath)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		defer errFile.Close()
		msg, _ := ioutil.ReadAll(errFile)
		if msg != nil {
			c.String(http.StatusInternalServerError, string(msg))
		} else {
			c.String(http.StatusInternalServerError, "unknown")
		}
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func saveUploadedFile(uploaded *multipart.FileHeader, dst string) error {
	srcFile, err := uploaded.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	defer dstFile.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	return nil
}

func waitForConversionDone(oneTimeFilePath string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	err = watcher.Add(oneTimeFilePath)
	if err != nil {
		return err
	}
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				return nil
			}
		case err := <-watcher.Errors:
			return err
		}
	}
}

func newRequestContext(c *gin.Context, id *uuid.UUID) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	notify := c.Writer.(http.CloseNotifier).CloseNotify()
	go func() {
		select {
		case <-ctx.Done():
		case <-notify:
			log.Printf("<%s> canceled", id.String())
			cancel()
		}
	}()
	return ctx, cancel
}

func (s *ServiceEndpoint) cleanup(id *uuid.UUID) {
	log.Printf("<%s> cleanup", id.String())
	outFilePath := s.outPath + "/" + id.String() + ".docx"
	os.Remove(outFilePath)
}
