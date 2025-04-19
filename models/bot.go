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
			Username:    "luoxk123",
			Password:    "1Q2w3e4r",
			CaptchaCode: "",
			RealName:    "",
		})
		if !resp.Success {
			panic(resp.Message)
		}
	}
	if bc == nil {
		bc = bbinWails.NewSportYongle2("", "", "", tugou)
	}

	tugou.SetSport(bc)
	tugou.GetSport().Login(nil)
}

func GetBotInstance() bbinWails.BetClient {
	if bc == nil || !bc.LoginCheckIn() {
		return nil
	}
	return bc
}
