// Copyright 2016 Author ysqi. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// @Author: ysqi
// @Email: devysq@gmail.com or 460857340@qq.com

package service

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/keywordAnlyz/worddog"
	"github.com/ysqi/com"

	"github.com/keywordAnlyz/kaweb/models"
)

type TaskService struct {
}

func (t *TaskService) createTaskSaveDir(task *models.Task) error {
	//循环判断目录是否存在

	var count int
	var name, dir string
	name = task.Name
	for {
		dir = filepath.Join(beego.AppPath, "/data/upload/", time.Now().Format("200601"), name)
		if _, err := os.Stat(task.FilePath); os.IsNotExist(err) {
			break
		}
		count++
		name = fmt.Sprintf("%s(%d)", task.Name, count)
	}
	//创建文件夹
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}
	task.Name = name
	task.FilePath = dir
	return nil

}
func (t *TaskService) getTaskSrcFilePath(task models.Task) string {
	return filepath.Join(task.FilePath, "src")
}
func (t *TaskService) getTaskNewFilePath(task models.Task) string {
	return filepath.Join(task.FilePath, "new")
}

func (t *TaskService) NewTask(name string, file multipart.File, header *multipart.FileHeader) (models.Task, error) {
	task := models.NewTask(name)

	if name == "" {
		return task, errors.New("任务名称不能为空")
	}
	if file == nil {
		return task, errors.New("文件为空")
	}

	if err := checkTaskFile(header.Filename); err != nil {
		return task, err
	}
	if err := t.createTaskSaveDir(&task); err != nil {
		return task, fmt.Errorf("创建文件存储目录是失败,%s", err)
	}

	//创建原始文件存储位置
	savePath := t.getTaskSrcFilePath(task)
	if err := os.MkdirAll(savePath, 0777); err != nil {
		return task, fmt.Errorf("创建文件存储目录是失败,%s", err)
	}

	//存储文件
	f, err := os.OpenFile(filepath.Join(savePath, header.Filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return task, fmt.Errorf("存储待解析文件失败,%s", err)
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		return task, fmt.Errorf("存储待解析文件失败,%s", err)
	}

	//记录信息
	o := orm.NewOrm()
	_, err = o.Insert(&task)
	if err != nil {
		return task, fmt.Errorf("存储任务到数据库失败,%s", err)
	}
	taskDoing <- task.Id
	return task, nil
}

//根据ID获取任务信息
func (t *TaskService) GetTask(taskId int) (models.Task, error) {

	o := orm.NewOrm()

	task := models.Task{Id: taskId}
	err := o.Read(&task)
	return task, err
}

//更新任务状态
func (t *TaskService) UpdateState(taskId int, status models.TaskStatus, msg ...string) error {
	o := orm.NewOrm()
	task := models.Task{Id: taskId, Status: status}
	fields := []string{"Status"}

	if length := len(msg); length == 1 {
		fields = append(fields, "Text1")
		task.Text1 = msg[0]
	} else if length > 1 {
		fields = append(fields, "Text1", "Text2")
		task.Text1 = msg[0]
		task.Text2 = msg[1]
	}
	_, err := o.Update(&task, fields...)
	return err
}

func (t *TaskService) StartTask(taskId int) error {
	if taskId <= 0 {
		return errors.New("任务ID非法")
	}

	task, err := t.GetTask(taskId)
	if err != nil {
		return err
	}
	if task.Status == models.Status_Running {
		return fmt.Errorf("任务当前是%q状态,不能重新运行", task.Status)
	}

	taskDoing <- taskId
	return nil
}

// 获取任务日志
func (t *TaskService) GetTaskLogs(taskId int) ([]models.TaskLog, error) {
	o := orm.NewOrm()

	list := []models.TaskLog{}
	_, err := o.QueryTable(&models.TaskLog{}).Filter("TaskId", taskId).OrderBy("-CreateTime").All(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

//获取任务下文件列表
func (t *TaskService) GetTaskFiles(taskId int) ([]os.FileInfo, error) {
	task, err := t.GetTask(taskId)
	if err != nil {
		return nil, err
	}

	dir := t.getTaskSrcFilePath(task)
	files := []os.FileInfo{}
	err = filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		files = append(files, f)
		return nil
	})
	return files, err
}

//获取任务下词汇
func (t *TaskService) GetTaskWords(taskId int, topCount int) ([]*models.SumWord, error) {
	o := orm.NewOrm()
	list := []*models.TaskWord{}
	qs := o.QueryTable(&models.TaskWord{})

	_, err := qs.Filter("TaskId", taskId).OrderBy("-Fre").Limit(topCount).All(&list)
	if err != nil {
		return nil, err
	}
	for i, v := range list {
		w := models.Word{Id: v.WordId}
		o.QueryTable(&w).Filter("Id", w.Id).One(&w)
		list[i].Word = w
	}
	return t.GroupTaskWordsByWordId(list), nil
}

//获取任务
func (t *TaskService) FilterTaskWords(taskId int, keywords []string, minFre int) ([]*models.SumWord, error) {

	//粗暴处理，获取全部后进行过滤
	list, err := t.GetTaskWords(taskId, -1)
	if err != nil {
		return nil, err
	}

	if len(keywords) == 0 && minFre == 0 {
		return list, nil
	}
	if minFre == 0 && strings.Join(keywords, "") == "" {
		return list, nil
	}

	words := list[:0]
	for _, v := range list {
		if minFre > 0 && v.SumFre() >= minFre {
			words = append(words, v)
		} else if com.IsSliceContainsStr(keywords, v.Word.Text) {
			words = append(words, v)
		}
	}
	return words, nil
}

//获得单个词汇信息，包含该认为下的词汇统计
func (t *TaskService) GetTaskSingleWords(taskId int, wordId int) (*models.SumWord, error) {

	o := orm.NewOrm()
	w := &models.Word{Id: wordId}
	err := o.QueryTable(&w).Filter("Id", wordId).One(w)
	if err != nil {
		return nil, err
	}

	list := []*models.TaskWord{}
	qs := o.QueryTable(&models.TaskWord{})
	_, err = qs.Filter("TaskId", taskId).Filter("WordId", wordId).All(&list)
	if len(list) == 0 {
		return &models.SumWord{Word: w}, nil
	}
	list[0].Word = *w
	return t.GroupTaskWordsByWordId(list)[0], err
}

func (t *TaskService) GroupTaskWordsByWordId(words []*models.TaskWord) []*models.SumWord {
	ws := map[int]*models.SumWord{}
	for _, v := range words {
		if w, ok := ws[v.WordId]; ok {
			w.TaskWords = append(w.TaskWords, v)
		} else {
			ws[v.WordId] = &models.SumWord{
				Word:      &v.Word,
				TaskWords: []*models.TaskWord{v},
			}
		}
	}

	list := make([]*models.SumWord, len(ws))
	i := 0
	for _, v := range ws {
		list[i] = v
		i++
	}

	//需要根据总频次排序，从高到底排序
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[i].SumFre() < list[j].SumFre() {
				list[i], list[j] = list[j], list[i]
			}
		}
	}

	return list

}

//获取任务列表
func (t *TaskService) GetTaskList(topCount int) ([]models.Task, error) {
	o := orm.NewOrm()
	list := []models.Task{}

	_, err := o.QueryTable(&models.Task{}).OrderBy("-CreateTime").Limit(topCount).All(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

//高亮文件指定词汇
func (t *TaskService) HightlightFile(task models.Task, wordId int, filename string) (string, *models.Word, error) {

	//先读取文件
	dir := t.getTaskNewFilePath(task)
	bytes, err := ioutil.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return "", nil, err
	}

	//不需要进行高亮，直接显示即可
	if wordId == 0 {
		return string(bytes), nil, nil
	}

	sumWord, err := t.GetTaskSingleWords(task.Id, wordId)
	if err != nil {
		return "", nil, err
	}

	w := &worddog.Word{
		Text:      sumWord.Word.Text,
		Pos:       sumWord.Word.Pos,
		Positions: []worddog.Position{},
	}

	for _, v := range sumWord.TaskWords {

		if v.FileName != filename {
			continue
		}
		if v.Postion == "" {
			continue
		}

		items := strings.Split(v.Postion, ",")
		for _, p := range items {
			if len(p) <= 3 {
				continue
			}
			if p[0] != '(' || p[len(p)-1] != ')' || strings.ContainsAny(p, "|") == false {
				continue
			}
			xy := strings.Split(p[1:len(p)-1], "|")

			w.Positions = append(w.Positions,
				worddog.Position{
					Start: com.StrTo(xy[0]).MustInt(),
					End:   com.StrTo(xy[1]).MustInt(),
				})
		}
	}
	if len(w.Positions) == 0 {
		return string(bytes), sumWord.Word, nil
	}

	html := worddog.HighlightDefault(bytes, w)
	return html, sumWord.Word, nil
}

//保存获取读取词汇基本信息
func (t *TaskService) SaveWords(words map[string]*worddog.Word) (map[string]*models.Word, error) {
	o := orm.NewOrm()
	ws := make(map[string]*models.Word, len(words))
	for _, v := range words {
		//存储词汇
		w := &models.Word{
			Text: v.Text,
			Pos:  v.Pos,
		}
		if _, id, err := o.ReadOrCreate(w, "Text"); err != nil {
			return nil, err
		} else {
			w.Id = int(id)
		}
		ws[w.Text] = w
	}
	return ws, nil
}
