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

	"github.com/ysqi/com"

	"github.com/keywordAnlyz/kaweb/service"
)

type WordController struct {
	BaseController
}

func (w *WordController) List() {
	w.TplName = "word.html"

	words, err := service.GetAllCustomerDict()
	if err != nil {
		w.Flash.Error("获取自定义字典列表失败,%s", err)
		return
	}
	w.Data["words"] = words
}

func (w *WordController) Add() {
	w.TplName = "word.html"

	input := w.GetString("words")
	w.Data["inputWords"] = input

	words := strings.Split(input, "\n")
	if len(words) == 0 {
		w.Flash.Error("请输入字典像，一行一个字典")
		return
	}

	err := service.AddCustomerDict(words...)
	if err != nil {
		w.Flash.Error("新增自定义字典失败，%s", err)
		return
	}
	w.Flash.Success("新增自定义字典成功")

	url := "/word/list.html"
	w.SetFlashTarget(url)
	w.StoreFlash()
	w.Redirect(url, 302)
}

func (w *WordController) Delete() {
	w.TplName = "word.html"

	wordId := com.StrTo(w.Ctx.Input.Param(":wordId")).MustInt()

	err := service.DeleteCustomerWord(wordId)
	if err != nil {
		w.Flash.Error("删除自定义字典失败,%s", err)
		return
	}
	w.Flash.Success("删除自定义字典成功")
	url := "/word/list.html"
	w.SetFlashTarget(url)
	w.StoreFlash()
	w.Redirect(url, 302)
}
