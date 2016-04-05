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

import "time"

type WordFrom int

const (
	WordFrom_Task    WordFrom = iota //词汇来自 任务解析
	WordFrom_Cust                    //词汇来自用户自定义
	WordFrom_CustDel                 //删除
)

var wordFroms = [...]string{
	"任务",
	"自定义",
	"已删除",
}

func (w WordFrom) String() string { return wordFroms[w] }

//词汇信息
type Word struct {
	Id         int
	Text       string    //词汇名
	Pos        string    //词汇属性
	Fre        int       //词汇频次
	From       WordFrom  //词汇来源
	CreateTime time.Time `orm:"auto_now_add;type(datetime)"` //词汇创建时间
}

func (w *Word) TableName() string {
	return "words"
}

//任务所属词汇位置信息
type TaskWord struct {
	Id       int
	TaskId   int    //任务ID
	WordId   int    //词汇ID
	Word     Word   `orm:"-"`
	Fre      int    //词汇出现次数
	Postion  string //词汇位置信息，{1,2},{2,3}
	FileName string //词汇所在文件名
}

func (w *Word) TaskWord() string {
	return "taskWords"
}

//用于汇总
type SumWord struct {
	*Word
	TaskWords []*TaskWord
}

//获取
func (s *SumWord) SumFre() int {
	sum := 0
	for _, v := range s.TaskWords {
		sum += v.Fre
	}
	return sum
}

//按文件汇总频次
func (s *SumWord) SumFreInFiles() map[string]int {

	fs := map[string]int{}
	for _, v := range s.TaskWords {
		fs[v.FileName] = fs[v.FileName] + v.Fre
	}
	return fs
}
