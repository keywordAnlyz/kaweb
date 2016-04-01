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

package models

import (
	"fmt"
	"time"
)

// 任务状态
type TaskStatus int

const (
	Status_UnStart         TaskStatus = 1 + iota // 未开始
	Status_Stoped                                //已停止
	Status_StopedWithError                       //异常停止
	Status_Running                               //运行中
	Status_Completed                             //已完成
)

var taskstatus = [...]string{
	"未开始",
	"已停止",
	"异常停止",
	"运行中",
	"已完成",
}

func (t TaskStatus) String() string { return taskstatus[t-1] }

//任务
type Task struct {
	Id            int        //ID
	Name          string     //任务名
	CreateTime    time.Time  //创建时间
	Status        TaskStatus //状态
	CompletedTime time.Time  `orm:"null"` //完成时间
	FilePath      string     //文件存储位置
	Text1         string     `orm:"null"` //预留字段
	Text2         string     `orm:"null"` //预留字段
}

func (t *TaskStatus) TableName() string {
	return "task"
}

//消息级别
type MsgLevel int

const (
	MsgLevel_NOTICE MsgLevel = 1 + iota
	MsgLevel_WARNING
	MsgLevel_ERROR
)

var msglevels = [...]string{
	"通知",
	"警告",
	"错误",
}

func (m MsgLevel) String() string { return msglevels[m-1] }

type TaskLog struct {
	Id         int
	TaskId     int
	CreateTime time.Time `orm:"auto_now_add;type(datetime)"` //创建时间
	MsgLevel   MsgLevel  //日志级别
	Msg        string    //日志信息
}

func (t *TaskLog) TableName() string {
	return "tasklog"
}

func (t *TaskLog) Error(taskId int, msg string, args ...interface{}) *TaskLog {
	return &TaskLog{
		TaskId:     taskId,
		CreateTime: time.Now(),
		MsgLevel:   MsgLevel_ERROR,
		Msg:        fmt.Sprintf(msg, args...),
	}
}
func (t *TaskLog) Waring(taskId int, msg string, args ...interface{}) *TaskLog {
	return &TaskLog{
		TaskId:     taskId,
		CreateTime: time.Now(),
		MsgLevel:   MsgLevel_WARNING,
		Msg:        fmt.Sprintf(msg, args...),
	}
}
func (t *TaskLog) Notice(taskId int, msg string, args ...interface{}) *TaskLog {
	return &TaskLog{
		TaskId:     taskId,
		CreateTime: time.Now(),
		MsgLevel:   MsgLevel_NOTICE,
		Msg:        fmt.Sprintf(msg, args...),
	}
}

//新建任务对象
func NewTask(name string) Task {
	return Task{
		Name:       name,
		CreateTime: time.Now(),
		Status:     Status_UnStart,
	}
}
