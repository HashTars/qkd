package pkg

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

func Download(c *gin.Context) {

	// 获取文件的UUID和Minio路径（这里假设你能从请求参数中获取）
	uuid := c.Query("uuid")

	// 查询数据库 根据uuid获取文件真实名称，存储路径
	fileStore, err := GetFileStoreByUUID(uuid)
	if err != nil {
		NotFound404(c, uuid)
		return
	}
	bucketName := fileStore.Bucket
	fileRealName := fileStore.FileName + fileStore.Ext
	relPathFileName := fileStore.RelPath + "/" + fileRealName
	// 获取 Minio 中的文件信息
	objInfo, err := MinioHelperIns.minioClient.StatObject(c.Request.Context(), bucketName, relPathFileName, minio.StatObjectOptions{})
	if err != nil {
		Fail500(c, "Error getting object info from Minio")
		return
	}

	// 获取文件大小
	fileSize := objInfo.Size

	// 设置分段下载的范围
	start, end, err := getRange(c.Request, fileSize)
	if err != nil {
		Fail500(c, "Invalid Range")
		return
	}

	// 设置响应头
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+fileRealName)
	c.Header("Content-Length", strconv.FormatInt(end-start+1, 10))
	c.Header("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(fileSize, 10))

	// 设置响应头，告诉浏览器该文件要下载
	//c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileRealName))

	log.Printf("Download file %s < %s", uuid, relPathFileName)
	// 获取 Minio 中的文件内容
	reader, err := MinioHelperIns.minioClient.GetObject(c.Request.Context(), bucketName, relPathFileName, minio.GetObjectOptions{})
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting content from Minio")
		return
	}
	defer reader.Close()

	// 分段复制文件内容到响应中
	_, err = reader.Seek(start, io.SeekStart)
	if err != nil {
		Fail500(c, "Error seeking content in file")
		return
	}

	_, err = io.CopyN(c.Writer, reader, end-start+1)
	if err != nil {
		Fail500(c, "Error streaming content to response")
		return
	}

	// // 直接将 Minio 中的文件内容返回到响应中
	// _, err = io.Copy(c.Writer, reader)
	// if err != nil {
	// 	c.String(http.StatusInternalServerError, "Error streaming content to response")
	// 	return
	// }
}

func Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		Fail500(c, err.Error())
		return
	}
	defer file.Close()

	// 获取文件名
	fileBaseName := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	// 获取扩展 带 "."
	ext := filepath.Ext(header.Filename)

	uuid, err := GenerateUUID()
	if err != nil {
		Fail500(c, err.Error())
		return
	}
	bucketName := c.PostForm("bucketName")
	if bucketName == "" {
		Fail400(c, "BucketName can not be empty")
		return
	}
	realFileName := fileBaseName + ext

	dateDir := GetCurrentDate()
	relPathFileName := dateDir + "/" + realFileName

	log.Printf("Upload file: %s > %s \n", header.Filename, relPathFileName)

	contentType := header.Header.Get("Content-Type")

	_, err = MinioHelperIns.minioClient.PutObject(context.Background(), bucketName, relPathFileName, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error uploading the file to Minio"})
		return
	}
	fileStore := FileStore{
		UUID:     uuid,
		Bucket:   bucketName,
		RelPath:  dateDir,
		FileName: fileBaseName,
		Ext:      ext,
	}
	InsertFileStore(fileStore)
	Success(c, uuid)
}

func getRange(req *http.Request, fileSize int64) (int64, int64, error) {
	rangeHeader := req.Header.Get("Range")
	if rangeHeader == "" {
		return 0, fileSize - 1, nil
	}

	// 处理 Range 头，获取分段下载的范围
	var start, end int64
	_, err := sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
	if err != nil {
		return 0, 0, err
	}

	// 处理结尾不指定的情况
	if end == 0 {
		end = fileSize - 1
	}

	// 防止范围超出文件大小
	if end > fileSize-1 {
		end = fileSize - 1
	}

	return start, end, nil
}

// sscanf 函数用于解析 Range 头中的范围
func sscanf(s, format string, a ...interface{}) (int, error) {
	return fmt.Sscanf(s, format, a...)
}

func NotFound404(ctx *gin.Context, uuid string) {
	ctx.JSON(http.StatusNotFound, gin.H{
		"message": uuid,
	})
	ctx.Next()
}

func Fail500(ctx *gin.Context, msg string) {
	ctx.JSON(http.StatusInternalServerError, gin.H{
		"message": msg,
	})
	ctx.Abort()
}

func Fail400(ctx *gin.Context, msg string) {
	ctx.JSON(http.StatusBadRequest, gin.H{
		"message": msg,
	})
	ctx.Abort()
}

func Success(ctx *gin.Context, uuid string) {
	ctx.JSON(http.StatusOK, gin.H{
		"uuid": uuid,
	})
	ctx.Next()
}
