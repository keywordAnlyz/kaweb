/*
* @Author: ysqi
* @Date:   2016-03-31 20:53:05
* @Last Modified by:   ysqi
* @Last Modified time: 2016-03-31 23:04:34
 */

package controllers

import (
	"fmt"
	"strings"
	"time"

	"github.com/ysqi/com"

	"github.com/keywordAnlyz/kaweb/service"
)

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

	file, header, err := t.GetFile("srcfile")
	if err != nil {
		t.Flash.Error("获取待解析文件错误,%s", err)
		return
	}
	defer file.Close()

	s := service.TaskService{}
	task, err := s.NewTask(name, file, header)
	if err != nil {
		t.Flash.Error("创建任务失败,%s", err)
		return
	}
	t.Flash.Success("创建任务成功,任务名称：%s,<a href='/task/%d/detail.html'>点击查看</a>", task.Name, task.Id)
}

func (t *TaskController) Detail() {
	t.TplName = "task_detail.html"

	strId := t.Ctx.Input.Param(":id")
	if strId == "" {
		t.Flash.Error("缺失任务ID，无法获取任务具体信息")
		return
	}
	id := com.StrTo(strId).MustInt()
	if id <= 0 {
		t.Flash.Error("任务ID[%s]非法", strId)
		return
	}

	s := service.TaskService{}
	task, err := s.GetTask(id)
	if err != nil {
		t.Flash.Error("获取任务信息失败,%s", err)
		return
	}
	t.Data["task"] = task

	logs, err := s.GetTaskLogs(task.Id)
	if err != nil {
		t.Flash.Error("获取任务日志失败,%s", err)
		return
	}
	t.Data["tasklogs"] = logs

	fiels, err := s.GetTaskFiles(task.Id)
	if err != nil {
		t.Flash.Error("获取任务下文件失败,%s", err)
	}
	t.Data["taskfiles"] = fiels

}

func (t *TaskController) StartTask() {

	strId := t.Ctx.Input.Param(":id")
	if strId == "" {
		t.Flash.Error("缺失任务ID，无法获取任务具体信息")
		return
	}
	id := com.StrTo(strId).MustInt()
	if id <= 0 {
		t.Flash.Error("任务ID[%s]非法", strId)
		return
	}

	s := service.TaskService{}
	err := s.StartTask(id)
	if err != nil {
		t.Flash.Error("尝试重启任务失败,%s", err)
	} else {
		t.Flash.Success("重启任务成功,刷新查看状态")
	}
	url := fmt.Sprintf("/task/%d/detail.html", id)
	t.SetFlashTarget(url)
	t.StoreFlash()
	t.Redirect(url, 302)
}
