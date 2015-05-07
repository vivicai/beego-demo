package controllers

import (
	"beego-demo/models"
	"crypto/md5"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	"io"
	"os"
	"path/filepath"
	"time"
)

type UserController struct {
	beego.Controller
}

func (this *UserController) Register() {
	form := models.RegisterForm{}
	if err := this.ParseForm(&form); err != nil {
		beego.Debug("ParseRegsiterForm:", err)
		this.Data["json"] = models.NewErrorInfo(ErrInputData)
		this.ServeJson()
		return
	}
	beego.Debug("ParseRegsiterForm:", &form)

	valid := validation.Validation{}
	ok, err := valid.Valid(&form)
	if err != nil {
		beego.Debug("ValidRegsiterForm:", err)
		this.Data["json"] = models.NewErrorInfo(ErrInputData)
		this.ServeJson()
		return
	}
	if !ok {
		beego.Debug("ValidRegsiterForm errors:")
		for _, err := range valid.Errors {
			beego.Debug(err.Key, err.Message)
		}
		this.Data["json"] = models.NewErrorInfo(ErrInputData)
		this.ServeJson()
		return
	}

	regDate := time.Now()
	user, err := models.NewUser(&form, regDate)
	if err != nil {
		beego.Debug("NewUser:", err)
		this.Data["json"] = models.NewErrorInfo(ErrSystem)
		this.ServeJson()
		return
	}
	beego.Debug("NewUser:", user)

	if code, err := user.Insert(); err != nil {
		beego.Debug("InsertUser:", err)
		if code == 100 {
			this.Data["json"] = models.NewErrorInfo(ErrDupUser)
		} else {
			this.Data["json"] = models.NewErrorInfo(ErrDatabase)
		}
		this.ServeJson()
		return
	}

	go models.IncTotalUserCount(regDate)

	this.Data["json"] = models.NewNormalInfo("Succes")
	this.ServeJson()
}

func (this *UserController) Login() {
	form := models.LoginForm{}
	if err := this.ParseForm(&form); err != nil {
		beego.Debug("ParseLoginForm:", err)
		this.Data["json"] = models.NewErrorInfo(ErrInputData)
		this.ServeJson()
		return
	}
	beego.Debug("ParseLoginForm:", &form)

	valid := validation.Validation{}
	ok, err := valid.Valid(&form)
	if err != nil {
		beego.Debug("ValidLoginForm:", err)
		this.Data["json"] = models.NewErrorInfo(ErrInputData)
		this.ServeJson()
		return
	}
	if !ok {
		beego.Debug("ValidLoginForm errors:")
		for _, err := range valid.Errors {
			beego.Debug(err.Key, err.Message)
		}
		this.Data["json"] = models.NewErrorInfo(ErrInputData)
		this.ServeJson()
		return
	}

	user := models.User{}
	if code, err := user.FindById(form.Phone); err != nil {
		beego.Debug("FindUserById:", err)
		if code == 100 {
			this.Data["json"] = models.NewErrorInfo(ErrNoUser)
		} else {
			this.Data["json"] = models.NewErrorInfo(ErrDatabase)
		}
		this.ServeJson()
		return
	}
	beego.Debug("UserInfo:", &user)

	if ok, err := user.CheckPass(form.Password); err != nil {
		beego.Debug("CheckUserPass:", err)
		this.Data["json"] = models.NewErrorInfo(ErrSystem)
		this.ServeJson()
		return
	} else if !ok {
		this.Data["json"] = models.NewErrorInfo(ErrPass)
		this.ServeJson()
		return
	}
	user.ClearPass()

	this.SetSession("user_id", form.Phone)

	this.Data["json"] = &models.LoginInfo{Code: 0, UserInfo: &user}
	this.ServeJson()
}

func (this *UserController) Logout() {
	form := models.LogoutForm{}
	if err := this.ParseForm(&form); err != nil {
		beego.Debug("ParseLogoutForm:", err)
		this.Data["json"] = models.NewErrorInfo(ErrInputData)
		this.ServeJson()
		return
	}
	beego.Debug("ParseLogoutForm:", &form)

	valid := validation.Validation{}
	ok, err := valid.Valid(&form)
	if err != nil {
		beego.Debug("ValidLogoutForm:", err)
		this.Data["json"] = models.NewErrorInfo(ErrInputData)
		this.ServeJson()
		return
	}
	if !ok {
		beego.Debug("ValidLogoutForm errors:")
		for _, err := range valid.Errors {
			beego.Debug(err.Key, err.Message)
		}
		this.Data["json"] = models.NewErrorInfo(ErrInputData)
		this.ServeJson()
		return
	}

	if this.GetSession("user_id") != form.Phone {
		this.Data["json"] = models.NewErrorInfo(ErrInvalidUser)
		this.ServeJson()
		return
	}

	this.DelSession("user_id")

	this.Data["json"] = models.NewNormalInfo("Succes")
	this.ServeJson()
}

func (this *UserController) Passwd() {
	form := models.PasswdForm{}
	if err := this.ParseForm(&form); err != nil {
		beego.Debug("ParsePasswdForm:", err)
		this.Data["json"] = models.NewErrorInfo(ErrInputData)
		this.ServeJson()
		return
	}
	beego.Debug("ParsePasswdForm:", &form)

	valid := validation.Validation{}
	ok, err := valid.Valid(&form)
	if err != nil {
		beego.Debug("ValidPasswdForm:", err)
		this.Data["json"] = models.NewErrorInfo(ErrInputData)
		this.ServeJson()
		return
	}
	if !ok {
		beego.Debug("ValidPasswdForm errors:")
		for _, err := range valid.Errors {
			beego.Debug(err.Key, err.Message)
		}
		this.Data["json"] = models.NewErrorInfo(ErrInputData)
		this.ServeJson()
		return
	}

	if this.GetSession("user_id") != form.Phone {
		this.Data["json"] = models.NewErrorInfo(ErrInvalidUser)
		this.ServeJson()
		return
	}

	code, err := models.ChangePass(form.Phone, form.OldPass, form.NewPass)
	if err != nil {
		beego.Debug("ChangeUserPass:", err)
		if code == 100 {
			this.Data["json"] = models.NewErrorInfo(ErrNoUserPass)
		} else if code == -1 {
			this.Data["json"] = models.NewErrorInfo(ErrDatabase)
		} else {
			this.Data["json"] = models.NewErrorInfo(ErrSystem)
		}
		this.ServeJson()
		return
	}

	this.Data["json"] = models.NewNormalInfo("Succes")
	this.ServeJson()
}

func (this *UserController) Uploads() {
	form := models.UploadsForm{}
	if err := this.ParseForm(&form); err != nil {
		beego.Debug("ParseUploadsForm:", err)
		this.Data["json"] = models.NewErrorInfo(ErrInputData)
		this.ServeJson()
		return
	}
	beego.Debug("ParseUploadsForm:", &form)

	valid := validation.Validation{}
	ok, err := valid.Valid(&form)
	if err != nil {
		beego.Debug("ValidUploadsForm:", err)
		this.Data["json"] = models.NewErrorInfo(ErrInputData)
		this.ServeJson()
		return
	}
	if !ok {
		beego.Debug("ValidUploadsForm errors:")
		for _, err := range valid.Errors {
			beego.Debug(err.Key, err.Message)
		}
		this.Data["json"] = models.NewErrorInfo(ErrInputData)
		this.ServeJson()
		return
	}

	if this.GetSession("user_id") != form.Phone {
		this.Data["json"] = models.NewErrorInfo(ErrInvalidUser)
		this.ServeJson()
		return
	}

	files := this.Ctx.Request.MultipartForm.File["photos"]
	for i, _ := range files {
		src, err := files[i].Open()
		if err != nil {
			beego.Debug("Open MultipartForm File:", err)
			this.Data["json"] = models.NewErrorInfo(ErrOpenFile)
			this.ServeJson()
			return
		}
		defer src.Close()

		hash := md5.New()
		if _, err := io.Copy(hash, src); err != nil {
			beego.Debug("Copy File to Hash:", err)
			this.Data["json"] = models.NewErrorInfo(ErrWriteFile)
			this.ServeJson()
			return
		}
		hex := fmt.Sprintf("%x", hash.Sum(nil))

		dst, err := os.Create(beego.AppConfig.String("apppath") +
			"static/" + hex + filepath.Ext(files[i].Filename))
		if err != nil {
			beego.Debug("Create File:", err)
			this.Data["json"] = models.NewErrorInfo(ErrWriteFile)
			this.ServeJson()
		}
		defer dst.Close()

		src.Seek(0, 0)
		if _, err := io.Copy(dst, src); err != nil {
			beego.Debug("Copy File:", err)
			this.Data["json"] = models.NewErrorInfo(ErrWriteFile)
			this.ServeJson()
			return
		}
	}

	this.Data["json"] = models.NewNormalInfo("Succes")
	this.ServeJson()
}
