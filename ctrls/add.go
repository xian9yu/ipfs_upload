package ctrls

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"ipfs_upload/models"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var url = "http://192.168.1.245:9094/add"

type image struct {
	Name string            `json:"name"`
	Cid  map[string]string `json:"cid"`
	Size int64             `json:"size"`
}

func Add(c *gin.Context) {
	// 上传文件
	header, err := c.FormFile("file")
	if err != nil {
		c.JSON(200, gin.H{
			"upload file error": err,
		})
		c.Abort()
		return
	}

	// 限制文件上传大小
	if float64(header.Size)/1024/1024 > 10 {
		c.JSON(200, gin.H{
			"file size error": "上传文件不能大于10M",
		})
		c.Abort()
		return
	}

	nameArr := strings.Split(header.Filename, ".")
	suffixName := nameArr[len(nameArr)-1]

	// 验证 是否支持该格式上传,懒人写法
	suffixArr := []string{"jpg", "JPG", "png", "PNG", "jpeg", "JPEG", "gif", "GIF"}

	var imageType string
	for i := 0; i < len(suffixArr); i++ {
		if suffixArr[i] == suffixName {
			imageType = "." + suffixArr[i]
		}
	}

	if imageType == "" {
		c.JSON(200, gin.H{
			"error": "暂时不支持该格式图片上传",
		})
		c.Abort()
		return
	}

	// 图片缓存在服务器
	//fileName := "./file/" + nameArr[0] + indexStr
	fileName := nameArr[0] + imageType
	err = c.SaveUploadedFile(header, fileName)
	if err != nil {
		c.JSON(200, gin.H{
			"save file error": err,
		})
		c.Abort()
		return
	}

	// 上传图片到 ipfs
	body, err := postFile(fileName, url)
	if err != nil {
		c.JSON(200, gin.H{
			"post file error": err,
		})
	}

	// 移除服务器缓存的图片
	_ = os.Remove(fileName)

	// 转换数据类型 && 赋值
	var n image
	err = json.Unmarshal(body, &n)
	if err != nil {
		c.JSON(200, gin.H{
			"unmarshal error": err,
		})
		c.Abort()
		return
	}
	var filePath, cid string
	for p, id := range n.Cid {
		filePath = p
		cid = id
	}
	f := models.File{
		Cid:  cid,
		Name: n.Name,
		Path: filePath,
		Size: n.Size,
		Type: imageType,
	}

	// 如果 cid已存在则直接返回数据
	if f.Count(cid) {
		c.JSON(200, n)
		c.Abort()
		return
	}

	// 执行sql
	_, _, err = f.Add(f)
	if err != nil {
		c.JSON(200, gin.H{
			"sql add error": err,
		})
		c.Abort()
		return
	}

	c.JSON(200, n)
}

// 上传文件请求
func postFile(filePath string, url string) ([]byte, error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	part1, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part1, file)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "multipart/form-data")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
