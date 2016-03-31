/*
* @Author: ysqi
* @Date:   2016-03-31 20:53:05
* @Last Modified by:   ysqi
* @Last Modified time: 2016-03-31 23:04:34
 */

package controllers

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/utils"

	// "github.com/keywordAnlyz/kaweb/service"
)

var suportFileType = []string{".txt", ".rar", ".doc"}

type TaskController struct {
	BaseController
}

//上传文件
func (t *TaskController) Upload() {
	t.TplName = "task_add.html"

	t.Data["taskname"] = "Task_" + time.Now().Format("200601021504")
	if t.Ctx.Input.IsPost() == false {
		return
	}

	name := strings.Trim(t.GetString("taskname"), "")

	if name == "" {
		t.Flash.Error("任务名称不能为空")
	}
	if len(name) < 4 {
		t.Data["taskname"] = name
		t.Flash.Error("任务名称必需有4个长度")
		return
	}

	file, hearder, err := t.GetFile("srcfile")
	if err != nil {
		t.Flash.Error("获取待解析文件错误,%s", err)
		return
	}
	defer file.Close()

	ext := filepath.Ext(hearder.Filename)
	if utils.InSlice(ext, suportFileType) == false {
		t.Flash.Error("上传待解析文件格式%q不支持，目前仅支持%v", ext, suportFileType)
		return
	}

	tofile := filepath.Join(beego.AppPath, "data/upload/", time.Now().Format("20060102"), strconv.FormatInt(time.Now().UnixNano(), 10)+ext)

	//创建文件夹
	os.Mkdir(filepath.Dir(tofile), 0777)

	f, err := os.OpenFile(tofile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		t.Flash.Error("存储待解析文件失败,%s", err)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	t.Flash.Success("创建任务成功")
}
