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

type GlobalCate string

const (
	BaseItem GlobalCate = "BASE" //基础配置
)

//全局配置信息
type KWGlobal struct {
	Id          int
	Cate        GlobalCate //配置分类
	ItemDisplay string     //配置显示名
	Item        string     //配置名
	Value       string     //配置值
	Desc        string     //配置项说明
}

func (k *KWGlobal) TableName() string {
	return "globalinfo"
}
