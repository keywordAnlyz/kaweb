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
	"os"
	"path/filepath"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/mattn/go-sqlite3"

	"github.com/keywordAnlyz/kaweb/models"
)

var SiteInfo models.Site

func init() {

	SiteInfo = models.Site{}
	SiteInfo.Name = beego.AppConfig.DefaultString("site:name", "关键字分析系统")

	initDB()
}

func initDB() {

	dbPath := filepath.Join(beego.AppPath, "data", "keywordanylya.db")

	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0777)
	}

	err := orm.RegisterDataBase("default", "sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	orm.RegisterModel(&models.KWGlobal{}, &models.Task{}, &models.TaskLog{}, &models.Word{}, &models.TaskWord{})

	if beego.BConfig.RunMode == beego.DEV {
		orm.Debug = true
	}
	//初始化表结构
	orm.RunSyncdb("default", false, true)
}
