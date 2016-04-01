package controllers

import (
	"time"

	"github.com/astaxie/beego"

	"github.com/keywordAnlyz/kaweb/service"
)

type BaseController struct {
	beego.Controller
	Flash *beego.FlashData
}

const flashTargetFlag = "targeturl"

// Prepare runs before request function execution.
func (b *BaseController) Prepare() {

	b.Data["Site"] = service.SiteInfo

	//Read Flash
	b.Flash = beego.ReadFromRequest(&b.Controller)
	//判断 Flash 是否是传递给本页面的
	if v := b.Flash.Data[flashTargetFlag]; v != "" && v != b.Ctx.Request.RequestURI {
		b.Flash = beego.NewFlash()
	} else {
		b.SetFlashTarget("delete")
	}

}

func (b *BaseController) Render() error {
	b.StoreFlash()
	return b.Controller.Render()
}

//存储Flash
func (b *BaseController) StoreFlash() {
	// save flash
	b.Flash.Store(&b.Controller)
}

//设置Flash传递地址
func (b *BaseController) SetFlashTarget(url string) {
	b.Flash.Data[flashTargetFlag] = url
}

func init() {
	//转换为本地时间
	beego.AddFuncMap("tolocal", func(t time.Time, layout string) string {
		return t.Local().Format(layout)
	})
}
