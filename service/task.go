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
	// "bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/henrylee2cn/mahonia"
	"github.com/keywordAnlyz/worddog"
	"github.com/ysqi/com"

	"github.com/keywordAnlyz/kaweb/models"
)

var suportFileType = []string{".txt", ".rar", ".doc"}

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

	if ext := filepath.Ext(header.Filename); com.IsSliceContainsStr(suportFileType, ext) == false {
		return task, fmt.Errorf("上传待解析文件格式%q不支持，目前仅支持%v", ext, suportFileType)
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
	beego.Error(err)
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
	beego.Debug(task.Status, models.Status_Running)
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

//获取任务
func (t *TaskService) GetTaskWords(taskId int, topCount int) ([]models.TaskWord, error) {
	o := orm.NewOrm()
	list := []models.TaskWord{}
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
	return list, nil
}

//获取任务
func (t *TaskService) FilterTaskWords(taskId int, keywords []string, minFre int) ([]models.TaskWord, error) {

	//粗暴处理，获取全部后进行过滤
	list, err := t.GetTaskWords(taskId, -1)
	if err != nil {
		return nil, err
	}
	if len(keywords) == 0 && minFre == 0 {
		return list, nil
	}
	beego.Debug(keywords, len(keywords))
	beego.Debug(minFre)
	words := list[:0]
	for _, v := range list {
		if minFre > 0 && v.Fre >= minFre {
			words = append(words, v)
		} else if com.IsSliceContainsStr(keywords, v.Word.Text) {
			words = append(words, v)
		}
	}
	return words, nil
}

//获得单个词汇信息，包含该认为下的词汇统计
func (t *TaskService) GetTaskSingleWords(taskId int, wordId int) (models.Word, []models.TaskWord, error) {

	o := orm.NewOrm()
	w := models.Word{Id: wordId}
	err := o.QueryTable(&w).Filter("Id", wordId).One(&w)
	if err != nil {
		return w, nil, err
	}

	list := []models.TaskWord{}
	qs := o.QueryTable(&models.TaskWord{})
	_, err = qs.Filter("TaskId", taskId).Filter("WordId", wordId).All(&list)
	return w, list, err
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
func (t *TaskService) HightlightFile(task models.Task, wordId int, filename string) (string, models.Word, error) {

	//先读取文件
	dir := t.getTaskNewFilePath(task)
	bytes, err := ioutil.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return "", models.Word{}, err
	}

	//不需要进行高亮，直接显示即可
	if wordId == 0 {
		return string(bytes), models.Word{}, nil
	}

	word, taskWords, err := t.GetTaskSingleWords(task.Id, wordId)
	if err != nil {
		return "", word, err
	}

	w := &worddog.Word{
		Text:      word.Text,
		Pos:       word.Pos,
		Positions: []worddog.Position{},
	}

	for _, v := range taskWords {
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
		return string(bytes), word, nil
	}

	html := worddog.HighlightDefault(bytes, w)
	return html, word, nil
}

func (t *TaskService) segment(task models.Task, file string) (int, error) {
	beego.Debug(worddog.Config.MinFre)
	words, err := worddog.SegmentFile(file)
	if err != nil {
		return 0, err
	}

	ws, err := t.SaveWords(words)
	if err != nil {
		return 0, err
	}

	//存储任务提取的词汇
	fileName := filepath.Base(file)
	o := orm.NewOrm()

	//删除原记录
	_, err = o.QueryTable(&models.TaskWord{}).Filter("TaskId", task.Id).Delete()
	if err != nil {
		return 0, err
	}

	for _, v := range words {
		postions := ""
		for _, p := range v.Positions {
			postions += fmt.Sprintf("(%d|%d),", p.Start, p.End)
		}
		_, err := o.Insert(&models.TaskWord{
			TaskId:   task.Id,
			WordId:   ws[v.Text].Id,
			FileName: fileName,
			Fre:      v.Frequency(),
			Postion:  postions,
		})
		if err != nil {
			return 0, err
		}
	}
	return len(words), nil
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

//解析原始文件
func (t *TaskService) AnyFile(task models.Task) error {
	//1.如果是压缩包，需解压
	//2.文件格式转换为 utf-8

	srcDir := t.getTaskSrcFilePath(task)

	toDir := t.getTaskNewFilePath(task)
	if err := os.MkdirAll(toDir, 0777); err != nil {
		return fmt.Errorf("创建文件存储目录是失败,%s", err)
	}

	err := filepath.Walk(srcDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		ext := filepath.Ext(f.Name())

		if ext == ".txt" {
			return convert2UTF8(path, filepath.Join(toDir, f.Name()))
		} else if ext == ".rar" {
			//解压文件
		}
		return nil
	})
	return err
}

var taskDoing chan int

func startTaskDo() {
	taskDoing = make(chan int, 50)
	go func() {
		for {
			select {
			case taskId := <-taskDoing:
				execTask(taskId)
			}
		}
	}()
}

//任务处理
func execTask(taskId int) {

	defer func() {
		//如果异常错误，则需要重新进入队列处理
		if err := recover(); err != nil {
			taskDoing <- taskId
		}
	}()

	s := TaskService{}

	o := orm.NewOrm()

	//更新状态
	task, err := s.GetTask(taskId)
	if err != nil {
		if err == orm.ErrNoRows {
			return
		}
		//重新获取
		taskDoing <- taskId
		//s.UpdateState(taskId, models.StopedWithError, "获取任务失败,"+err.Error())
		return
	}

	log := models.TaskLog{}

	if task.Status == models.Status_Running {
		o.Insert(log.Waring(taskId, "任务当前是%q状态,不能重新运行", task.Status))
		return
	}

	o.Insert(log.Notice(taskId, "开始任务处理"))
	err = s.UpdateState(taskId, models.Status_Running)
	if err != nil {
		o.Insert(log.Error(taskId, "更新任务状态失败,%s。终止任务", err))
		return
	}
	var isErrorStop bool
	defer func() {
		status := models.Status_Completed
		if isErrorStop {
			status = models.Status_StopedWithError
		}
		//更新任务状态
		err = s.UpdateState(taskId, status)
		if err != nil {
			o.Insert(log.Error(taskId, "更新任务状态失败,%s。终止任务", err))
		}
		o.Insert(log.Notice(taskId, "本次任务处理结束"))
	}()

	//开始解析文件
	o.Insert(log.Notice(taskId, "正在处理原始文件"))
	err = s.AnyFile(task)
	if err != nil {
		isErrorStop = true
		o.Insert(log.Error(taskId, "处理原始文件,%s。终止任务", err))
		return
	}

	//开始提取词汇信息
	//1.获取所有文件
	//2.并发提前

	o.Insert(log.Notice(taskId, "正在提取词汇"))

	fs := []string{}
	toDir := s.getTaskNewFilePath(task)
	err = filepath.Walk(toDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if f.IsDir() {
			return nil
		}
		if filepath.Ext(f.Name()) == ".txt" {
			fs = append(fs, path)
		}
		return nil
	})

	if err != nil {
		isErrorStop = true
		o.Insert(log.Error(taskId, "提取词汇失败,%s。终止任务", err))
		return
	}

	wordCount := 0 //提取词汇数
	if len(fs) > 0 {

		//粗暴方式更新Key
		gs := GlobalService{}
		minF := gs.GetItemValue("MINFRE")
		if minF != "" {
			worddog.Config.MinFre = com.StrTo(minF).MustInt()
		}

		//并发处理
		ws := sync.WaitGroup{}
		ws.Add(len(fs))

		for _, f := range fs {
			go func(f string) {
				c, err := s.segment(task, f)
				if err != nil {
					isErrorStop = true
					o.Insert(log.Error(taskId, "提取词汇失败,%s。终止任务", err))
					return
				}
				wordCount += c
				ws.Done()
			}(f)
		}
		ws.Wait()
	}
	o.Insert(log.Notice(taskId, "共提取%d个词汇", wordCount))

}

func convert2UTF8(srcFile, toFile string) error {
	beego.Debug(srcFile, toFile)
	f, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer f.Close()

	//默认按GBK编码处理
	rd := mahonia.NewDecoder("GB2312").NewReader(f)
	bytes, err := ioutil.ReadAll(rd)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(toFile, bytes, 0777)
}

func init() {
	startTaskDo()
}
