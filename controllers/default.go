package controllers

import (
	"github.com/astaxie/beego"

	"github.com/keywordAnlyz/kaweb/service"
)

type BaseController struct {
	beego.Controller
	Flash *beego.FlashData
}

// Prepare runs before request function execution.
func (b *BaseController) Prepare() {

	b.Data["Site"] = service.SiteInfo

	//Read Flash
	b.Flash = beego.ReadFromRequest(&b.Controller)
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
