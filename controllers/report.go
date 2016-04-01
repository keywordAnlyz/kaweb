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

package controllers

import (
	"strings"

	"github.com/keywordAnlyz/kaweb/service"
)

type ReportController struct {
	BaseController
}

func (r *ReportController) Get() {
	r.TplName = "report.html"

	taskId, err := r.GetInt("taskId", 0)
	if err != nil {
		r.Flash.Error("任务ID非法")
		return
	}

	s := service.TaskService{}
	task, err := s.GetTask(taskId)
	if err != nil {
		r.Flash.Error("获取任务信息失败,%s", err)
	}
	r.Data["task"] = task

	ks := r.GetString("keywords")
	ks = strings.Replace(ks, " ", ",", -1)
	ks = strings.Replace(ks, "，", ",", -1)
	keywords := strings.Split(ks, ",")
	minFre, _ := r.GetInt("minFre", 0)

	words, err := s.FilterTaskWords(taskId, keywords, minFre)
	if err != nil {
		r.Flash.Error("获取待统计关键字失败,%s", err)
	} else if len(words) == 0 {
		r.Flash.Error("无查询结果数据")
	}
	r.Data["words"] = words
	r.Data["keywords"] = ks
	r.Data["minFre"] = minFre
}
