package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"ipfs_upload_with_golang_http/models"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type image struct {
	Name string            `json:"name"`
	Cid  map[string]string `json:"cid"`
	Size int64             `json:"size"`
}

func main() {
	models.InitSQL()
	// Upload route
	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		// 限制文件上传大小为 10 MB
		_ = r.ParseMultipartForm(10 << 20)

		// 从 handler 获取 filename, size and headers
		file, header, err := r.FormFile("file")
		if err != nil {
			fmt.Println("Error Retrieving the File")
			fmt.Println(err)
			return
		}

		// 创建空文件
		dst, err := os.Create(header.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 复制文件流到创建的空文件
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 关闭文件流
		file.Close()
		dst.Close()

		// 上传图片到 ipfs
		body, err := postFile(header.Filename, "http://192.168.1.245:9094/add")
		if err != nil {
			return
		}

		// image type
		nameArr := strings.Split(header.Filename, ".")
		imageType := nameArr[len(nameArr)-1]

		// 转换数据类型 && 赋值
		var n image
		err = json.Unmarshal(body, &n)
		if err != nil {
			log.Println("unmarshal error", err)
			return
		}
		var filePath, cid string
		for p, id := range n.Cid {
			filePath = p
			cid = id
		}
		fff := models.File{
			Cid:  cid,
			Name: n.Name,
			Path: filePath,
			Size: n.Size,
			Type: imageType,
		}

		// 移除服务器缓存的图片
		_ = os.Remove(header.Filename)

		// 如果 cid已存在则直接返回数据
		if fff.Count(cid) {
			log.Println(200, n)
			return
		}

		// 执行sql
		_, _, err = fff.Add(fff)
		if err != nil {
			log.Println("sql add error", err)
			return
		}
	})
	//Listen on port 8080
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		return
	}

}

// ipfs上传文件请求
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
