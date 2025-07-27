package service

import (
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/global"
	"DiTing-Go/logic"
	pkgResp "DiTing-Go/pkg/domain/vo/resp"
)

// CaptchaService 验证码发送
func CaptchaService(captchaReq req.CaptchaReq) (pkgResp.ResponseData, error) {
	if captchaReq.Phone == "" {
		global.Logger.Infof("手机号不能为空")
		return pkgResp.ErrorResponseData("手机号不能为空"), nil
	}

	//检查验证码是否未过期,未过期不生成
	if logic.CheckCaptchaExist(captchaReq.Phone) {
		global.Logger.Infof("手机号:%s,验证码未过期", captchaReq.Phone)
		return pkgResp.ErrorResponseData("发送频率过高,请稍后再试"), nil
	}

	// 生成验证码
	captcha, err := logic.GenerateCaptcha(captchaReq.Phone)
	if err != nil {
		global.Logger.Errorf("phone:%s ,生成验证码失败: %v", captchaReq.Phone, err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试"), err
	}

	// 发送验证码
	// TODO:空实现，实际应用中可以使用短信服务商的SDK发送验证码
	if err := logic.SendCaptcha(captchaReq.Phone, captcha); err != nil {
		global.Logger.Errorf("phone:%s ,发送验证码失败: %v", captchaReq.Phone, err)
		return pkgResp.ErrorResponseData("验证码发送失败，请稍后再试"), err
	}

	return pkgResp.SuccessResponseDataWithMsg("验证码发送成功"), nil
}
