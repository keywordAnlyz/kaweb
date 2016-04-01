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
	// "strings"

	"github.com/ysqi/com"

	"github.com/keywordAnlyz/kaweb/models"
	"github.com/keywordAnlyz/kaweb/service"
)

type FileViewController struct {
	BaseController
}

func (f *FileViewController) Get() {
	f.TplName = "file_view.html"

	taskId := com.StrTo(f.Ctx.Input.Param(":taskId")).MustInt()
	wordId := com.StrTo(f.Ctx.Input.Param(":wordId")).MustInt()
	fileName := f.GetString("fileName")

	//任务信息
	s := service.TaskService{}
	task, err := s.GetTask(taskId)
	if err != nil {
		f.Flash.Error("获取任务信息失败,%s", err)
	}
	f.Data["task"] = task

	//本词汇信息
	html, word, err := s.HightlightFile(task, wordId, fileName)
	if err != nil {
		f.Flash.Error("处理失败,%s", err)
	}

	if word == nil {
		word = &models.Word{}
	}

	f.Data["word"] = word
	f.Data["fileName"] = fileName
	f.Data["html"] = html

}
