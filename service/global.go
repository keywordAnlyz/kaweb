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
	"errors"

	"github.com/Unknwon/com"
	"github.com/astaxie/beego/orm"

	"github.com/keywordAnlyz/kaweb/models"
)

type GlobalService struct{}

var baseGlobalItems []models.KWGlobal
var defaultGlobalItems = []models.KWGlobal{
	{Cate: models.BaseItem, ItemDisplay: "最低频次", Item: "MINFRE", Value: "10", Desc: "词汇解析最低频次要求，低于该频次将忽略。"},
}

// 获取最小频次设置
func (g *GlobalService) GetMinFre() (int, error) {

	o := orm.NewOrm()
	item := models.KWGlobal{}
	err := o.QueryTable(item).Filter("Item", "MINFRE").One(&item)
	if err != nil {
		return 0, err
	}
	if item.Value == "" {
		return 0, nil
	}

	return com.StrTo(item.Value).MustInt(), nil
}

//获取全局基础配置信息
func (g *GlobalService) GetBaseConfigs() ([]models.KWGlobal, error) {

	if len(baseGlobalItems) != 0 {
		return baseGlobalItems, nil
	}

	o := orm.NewOrm()
	item := models.KWGlobal{}
	list := []models.KWGlobal{}
	_, err := o.QueryTable(item).Filter("Cate", models.BaseItem).All(&list)
	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		if err := g.initBaseConfig(); err != nil {
			return nil, err
		}
		baseGlobalItems = defaultGlobalItems[:]
	} else {
		//默认项检查
		for _, v := range defaultGlobalItems {
			find := false
			for _, v2 := range list {
				if v2.Item == v.Item {
					find = true
					break
				}
			}
			if !find {
				list = append(list, v)
			}
		}
		baseGlobalItems = list
	}

	return baseGlobalItems, nil
}

//更新配置项
func (g *GlobalService) UpdateItemValue(id int, value string) error {

	if id <= 0 || value == "" {
		return errors.New("数据非法，Id为空或者Value为空")
	}

	item := models.KWGlobal{Id: id, Value: value}
	counts, err := orm.NewOrm().Update(&item, "Value")
	if err != nil {
		return err
	}
	if counts == 0 {
		return errors.New("数据库不存在此配置项")
	}

	//更新成功后，更新缓存
	for i, v := range baseGlobalItems {
		if v.Id == id {
			baseGlobalItems[i].Value = value
		}
	}
	return nil
}

//初始化默认值
func (g *GlobalService) initBaseConfig() error {
	if len(defaultGlobalItems) == 0 {
		return nil
	}
	_, err := orm.NewOrm().InsertMulti(len(defaultGlobalItems), defaultGlobalItems)
	return err
}
