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

import "github.com/keywordAnlyz/kaweb/service"

type GlobalController struct {
	BaseController
}

func (g *GlobalController) List() {

	s := service.GlobalService{}
	list, err := s.GetBaseConfigs()
	if err != nil {
		g.Flash.Error("获取全局配置失败:\r\n%s", err.Error())
	} else {
		g.Data["list"] = list
	}
	g.TplName = "global.html"
}

func (g *GlobalController) UpdateItem() {

	id, _ := g.GetInt("id", 0)
	value := g.GetString("value")

	s := service.GlobalService{}
	err := s.UpdateItemValue(id, value)

	if err != nil {
		g.Flash.Error("更新配置项失败,%s", err)
	} else {
		g.Flash.Success("更新配置项成功")
	}
	//跳转前必须保持Flash
	g.SetFlashTarget("/global/list.html")
	g.StoreFlash()
	g.Redirect("/global/list.html", 302)
}
