/*
* @Author: ysqi
* @Date:   2016-03-31 20:53:05
* @Last Modified by:   ysqi
* @Last Modified time: 2016-04-02 02:27:33
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
	}
	t.Data["task"] = task

	logs, err := s.GetTaskLogs(task.Id)
	if err != nil {
		t.Flash.Error("获取任务日志失败,%s", err)
	}
	t.Data["tasklogs"] = logs

	fiels, err := s.GetTaskFiles(task.Id)
	if err != nil {
		t.Flash.Error("获取任务下文件失败,%s", err)
	}
	t.Data["taskfiles"] = fiels

	words, err := s.GetTaskWords(task.Id, 30)
	if err != nil {
		t.Flash.Error("获取文件词汇失败,%s", err)
	}

	t.Data["words"] = words

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

func (t *TaskController) ShowWords() {
	t.TplName = "task_words.html"

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
	words, err := s.GetTaskWords(id, -1)
	if err != nil {
		t.Flash.Error("获取文件词汇失败,%s", err)
	}
	t.Data["words"] = words
	t.Data["id"] = id
	t.Data["task"], _ = s.GetTask(id)
}

func (t *TaskController) List() {

	t.TplName = "task_list.html"

	s := service.TaskService{}

	list, err := s.GetTaskList(50)

	if err != nil {
		t.Flash.Error("获取任务列表失败,%s", err)
	}
	t.Data["tasks"] = list
}

func (t *TaskController) WordDetail() {
	t.TplName = "task_word_detail.html"

	taskId := com.StrTo(t.Ctx.Input.Param(":taskId")).MustInt()
	wordId := com.StrTo(t.Ctx.Input.Param(":wordId")).MustInt()

	//任务信息
	s := service.TaskService{}
	task, err := s.GetTask(taskId)
	if err != nil {
		t.Flash.Error("获取任务信息失败,%s", err)
	}
	t.Data["task"] = task

	//本词汇信息
	word, taskwords, err := s.GetTaskSingleWords(taskId, wordId)
	if err != nil {
		t.Flash.Error("获取词汇信息失败,%s", err)
	}

	sum := 0
	for _, v := range taskwords {
		sum += v.Fre
	}

	t.Data["word"] = word
	t.Data["taskWords"] = taskwords
	t.Data["sumFre"] = sum
}
