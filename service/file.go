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
	"archive/zip"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/astaxie/beego/orm"
	"github.com/henrylee2cn/mahonia"
	"github.com/keywordAnlyz/worddog"
	"github.com/ysqi/com"

	"github.com/keywordAnlyz/kaweb/models"
)

var suportFileType = []string{".txt", ".zip", ".doc", ".rar", ".docx"}

//解压 ZIP 文件
func CompressZip(filename string, toDir string) ([]string, error) {

	cf, err := zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}
	defer cf.Close()

	savePaths := []string{}

	for _, f := range cf.File {

		name := mahonia.NewDecoder("GB2312").ConvertString(f.Name)
		//检查不通过，继续处理下一个文件
		if checkTaskFile(name) != nil {
			continue
		}
		//忽略
		if filepath.Ext(name) == ".rar" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		//默认按GBK编码处理
		rd := mahonia.NewDecoder("GB2312").NewReader(rc)
		bytes, err := ioutil.ReadAll(rd)
		if err != nil {
			return nil, err
		}
		saveto := filepath.Join(toDir, filepath.Base(name))
		err = ioutil.WriteFile(saveto, bytes, 0777)
		if err != nil {
			return nil, err
		}
		savePaths = append(savePaths, saveto)
	}

	return savePaths, nil

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
	fileCount, err := parsSrcFile(task)
	if err != nil {
		isErrorStop = true
		o.Insert(log.Error(taskId, "处理原始文件,%s。终止任务", err))
		return
	} else {
		o.Insert(log.Notice(taskId, "提取原始文件%d个", fileCount))
	}

	//开始提取词汇信息
	//1.获取所有文件
	//2.并发提前

	o.Insert(log.Notice(taskId, "开始提取词汇"))

	fs := []string{}
	toDir := s.getTaskNewFilePath(task)
	err = filepath.Walk(toDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if f.IsDir() {
			return nil
		}
		if ext := filepath.Ext(f.Name()); ext == ".txt" {
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

		//删除原记录
		_, err = o.QueryTable(&models.TaskWord{}).Filter("TaskId", task.Id).Delete()
		if err != nil {
			isErrorStop = true
			return
		}

		//并发处理
		// ws := sync.WaitGroup{}
		// ws.Add(len(fs))

		for _, f := range fs {
			// go func(f string) {
			// defer ws.Done()
			c, err := segmentWord(task, f)
			if err != nil {
				isErrorStop = true
				o.Insert(log.Error(taskId, "从%s提取词汇失败,%s。忽略", filepath.Base(f), err))
				continue
			}
			wordCount += c
			// }(f)
		}
		// ws.Wait()
	}
	o.Insert(log.Notice(taskId, "共提取%d个词汇", wordCount))

}

func init() {
	startTaskDo()
}

func checkTaskFile(filename string) error {
	if ext := filepath.Ext(filename); com.IsSliceContainsStr(suportFileType, ext) == false {
		return fmt.Errorf("上传待解析文件格式%q不支持，目前仅支持%v", ext, suportFileType)
	}
	return nil
}

func convert2UTF8(srcFile, toFile string) error {
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

//解析原始文件
func parsSrcFile(task models.Task) (int, error) {
	//1.如果是压缩包，需解压
	//2.文件格式转换为 utf-8
	var err error
	t := TaskService{}
	srcDir := t.getTaskSrcFilePath(task)
	toDir := t.getTaskNewFilePath(task)
	if err = os.MkdirAll(toDir, 0777); err != nil {
		return 0, fmt.Errorf("创建文件存储目录是失败,%s", err)
	}

	count := 0

	if runtime.GOOS == "windows" {

		//通过exe 进行文件转换
		cmd := exec.Command("./bin/Convert2txt.exe", srcDir, toDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return 0, err
		}

		if result := string(output); result != "OK" {
			return 0, fmt.Errorf("解析原始文件失败,%s", result)
		}
		//从文件夹下提取文件数
		err = filepath.Walk(toDir, func(path string, f os.FileInfo, err error) error {
			count++
			return nil
		})
		return count, err
	}
	err = filepath.Walk(srcDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		ext := filepath.Ext(f.Name())

		if ext == ".txt" {
			count++
			return convert2UTF8(path, filepath.Join(toDir, f.Name()))
		} else if ext == ".zip" { //解压处理
			if fs, err := CompressZip(path, toDir); err != nil {
				return err
			} else {
				count += len(fs)
			}
		}
		return nil
	})
	return count, err

}

//提取词汇
func segmentWord(task models.Task, file string) (int, error) {
	words, err := worddog.SegmentFile(file)
	if err != nil {
		return 0, err
	}

	t := TaskService{}
	ws, err := t.SaveWords(words)
	if err != nil {
		return 0, err
	}

	//存储任务提取的词汇
	fileName := filepath.Base(file)
	o := orm.NewOrm()

	taskWords := []*models.TaskWord{}
	for _, v := range words {
		postions := ""
		for _, p := range v.Positions {
			postions += fmt.Sprintf("(%d|%d),", p.Start, p.End)
		}
		taskWords = append(taskWords, &models.TaskWord{
			TaskId:   task.Id,
			WordId:   ws[v.Text].Id,
			FileName: fileName,
			Fre:      v.Frequency(),
			Postion:  postions,
		})
	}
	//批量存储
	counts, err := o.InsertMulti(100, taskWords)
	if err != nil {
		return 0, err
	}
	if int(counts) != len(taskWords) {
		return 0, fmt.Errorf("词汇存储不正确，提取文件%q词汇%d个，实际上存储数为%d", fileName, len(taskWords), counts)
	}
	return len(words), nil
}
