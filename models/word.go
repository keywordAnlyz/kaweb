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

//词汇信息
type Word struct {
	Id   int
	Text string //词汇名
	Pos  string //词汇属性
}

func (w *Word) TableName() string {
	return "words"
}

//任务所属词汇位置信息
type TaskWord struct {
	Id       int
	TaskId   int    //任务ID
	WordId   int    //词汇ID
	Fre      int    //词汇出现次数
	Postion  string //词汇位置信息，{1,2},{2,3}
	FileName string //词汇所在文件名
}

func (w *Word) TaskWord() string {
	return "taskWords"
}
