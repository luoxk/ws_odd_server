package models

import bbinWails "bbinWails/src"

var (
	tugou bbinWails.ITuGou
	bc    bbinWails.ISport
)

type Bot bbinWails.BetClient

func init() {
	if tugou == nil {
		tugou = bbinWails.NewTuGouYl()
	}
	tugou.SetHost("https://www.lok88.com/")
	tugou.SetProxy("socks5://127.0.0.1:17890")
	if !tugou.CheckLogin() {
		resp := tugou.Login(bbinWails.TuGouAccountInfo{
			Username:    "yingfeng001",
			Password:    "xx123456",
			CaptchaCode: "",
			RealName:    "",
		})
		if !resp.Success {
			panic(resp.Message)
		}
	}
	if bc == nil {
		bc = bbinWails.NewSportPandaYb("", "", "", tugou)
	}

	tugou.SetSport(bc)
	err := tugou.GetSport().Login(nil)
	if err != nil {
		panic(err)
	}
}

func GetBotInstance() bbinWails.ISport {
	if bc == nil || !bc.LoginCheckIn() {
		return nil
	}
	return bc
}
