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
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/keywordAnlyz/worddog"

	"github.com/keywordAnlyz/kaweb/models"
)

func init() {

	err := ReloadDict()
	if err != nil {
		beego.Warn("加载字典异常", err)
	}

}

var allwords = map[int]*models.Word{}

//加载词汇到内存
func LoadWordToCache() error {
	list := []*models.Word{}
	o := orm.NewOrm()
	_, err := o.QueryTable(&models.Word{}).All(&list)
	if err != nil {
		return err
	}
	for _, v := range list {
		allwords[v.Id] = v
	}
	return nil
}

func ReloadDict() error {
	//加载数据库定义的字典

	if len(allwords) == 0 {
		err := LoadWordToCache()
		if err != nil {
			return err
		}
	}

	if len(allwords) == 0 {
		return nil
	}

	for _, v := range allwords {
		if v.From == models.WordFrom_Cust {
			err := LoadDict(v)
			if err != nil {
				beego.Warn("加载自定义字典失败", err)
			}
		}
	}
	return nil
}

func LoadDict(word *models.Word) error {
	return worddog.Segmenter.LoadSignleDict(word.Text, word.Fre, word.Pos)

}

//获取所有自定义字典项
func GetAllCustomerDict() ([]*models.Word, error) {
	list := []*models.Word{}
	o := orm.NewOrm()
	_, err := o.QueryTable(&models.Word{}).Filter("From", models.WordFrom_Cust).All(&list)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func DeleteCustomerWord(wordId int) error {
	w, err := GetWordInfo(wordId)
	if err != nil {
		return err
	}
	if w.From != models.WordFrom_Cust {
		return fmt.Errorf("该字典属于%q,不是自定义字典不允许删除", w.From)
	}
	//不是删除，而是更新
	w.From = models.WordFrom_CustDel
	w.Fre = 10000000
	o := orm.NewOrm()
	_, err = o.Update(w, "From", "Fre")
	if err != nil {
		return err
	}
	//更新字典库
	return worddog.Segmenter.RemoteDict(w.Text)

}

//添加或更新自定义字典
func AddCustomerDict(words ...string) error {
	if len(words) == 0 {
		return errors.New("需要添加的新自定义字典项不能为空")
	}

	var needInserts []*models.Word
	var needUpdates []*models.Word
	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) == 0 {
			continue
		}
		w, err := GetWordInfoByName(word)
		if err != nil {
			return err
		}
		//已存在则更新
		if w != nil {
			if w.From != models.WordFrom_Cust {
				w.From = models.WordFrom_Cust
				needUpdates = append(needUpdates, w)
			}
			continue
		}

		//否则需要新写入
		needInserts = append(needInserts, &models.Word{
			Text: word,
			From: models.WordFrom_Cust,
			Pos:  "n",
			Fre:  10000000,
		})
	}
	if len(needUpdates) == 0 && len(needInserts) == 0 {
		return nil
	}

	o := orm.NewOrm()
	if len(needUpdates) > 0 {
		for _, v := range needUpdates {
			if _, err := o.Update(v, "From"); err != nil {
				return err
			}
			LoadDict(v)
		}
	}
	if len(needInserts) > 0 {
		_, err := o.InsertMulti(100, needInserts)
		if err != nil {
			return err
		}

		//存储
		for _, v := range needInserts {
			w, err := GetWordInfoByName(v.Text)
			if err != nil {
				continue //可以忽略该异常
			}
			allwords[w.Id] = w
			LoadDict(v)
		}
	}
	return nil
}

func GetWordInfo(wordId int) (*models.Word, error) {
	if v, ok := allwords[wordId]; ok {
		return v, nil
	}
	o := orm.NewOrm()
	w := &models.Word{Id: wordId}
	err := o.QueryTable(w).Filter("Id", w.Id).One(w)
	if err != nil {
		return nil, err
	}
	allwords[w.Id] = w
	return w, nil
}

func GetWordInfoByName(wordText string) (*models.Word, error) {
	wordText = strings.TrimSpace(wordText)
	for _, v := range allwords {
		if v.Text == wordText {
			return v, nil
		}
	}
	//从DB中查找
	o := orm.NewOrm()
	w := &models.Word{}
	err := o.QueryTable(w).Filter("Text", wordText).One(w)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	allwords[w.Id] = w
	return w, nil

}

//保存获取读取词汇基本信息
func SaveWords(words map[string]*worddog.Word) (map[string]*models.Word, error) {
	o := orm.NewOrm()
	needCreate := []*models.Word{}
	ws := make(map[string]*models.Word, len(words))
	for _, v := range words {

		w, err := GetWordInfoByName(v.Text)
		if err != nil {
			return nil, err
		}
		if w == nil {
			needCreate = append(needCreate, &models.Word{
				Text: v.Text,
				Pos:  v.Pos,
			})
			continue
		}
		//存储
		ws[w.Text] = w
	}
	//批量存储
	if len(needCreate) > 0 {
		_, err := o.InsertMulti(100, needCreate)
		if err != nil {
			return nil, err
		}

		//存储
		for _, v := range needCreate {
			w, err := GetWordInfoByName(v.Text)
			if err != nil {
				continue //可以忽略该异常
			}
			ws[w.Text] = w
			allwords[w.Id] = w
		}
	}
	return ws, nil
}

func GroupTaskWordsByWordId(words []*models.TaskWord) []*models.SumWord {
	ws := map[int]*models.SumWord{}
	for _, v := range words {
		if w, ok := ws[v.WordId]; ok {
			w.TaskWords = append(w.TaskWords, v)
		} else {
			ws[v.WordId] = &models.SumWord{
				Word:      &v.Word,
				TaskWords: []*models.TaskWord{v},
			}
		}
	}

	list := make([]*models.SumWord, len(ws))
	i := 0
	for _, v := range ws {
		list[i] = v
		i++
	}

	//需要根据总频次排序，从高到底排序
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[i].SumFre() < list[j].SumFre() {
				list[i], list[j] = list[j], list[i]
			}
		}
	}

	return list

}
