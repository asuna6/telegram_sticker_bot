package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
	tele "gopkg.in/telebot.v3"
)

func main() {
	pref := tele.Settings{
		Token:  "6341723545:AAHoxWlv_ME_c02zk84A3PcNBbZoO8RA_AA",
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
		// Poller: webhook,
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("%s 马猴烧酒启动", bot.Me.Username)

	commands := []tele.Command{{
		Text:        "get_id",
		Description: "发送此命令获取用户和群组的id",
	}}

	_ = bot.SetCommands(commands)

	bot.Handle("/get_id", func(c tele.Context) error {
		if c.Chat().Type == tele.ChatGroup || c.Chat().Type == tele.ChatSuperGroup {
			var msg = fmt.Sprintf("用户id:%d\n群组id:%d", c.Message().Sender.ID, c.Message().Chat.ID)
			return c.Send(msg, tele.ModeHTML)
		} else {
			var msg = fmt.Sprintf("用户id:%d\n对话id:%d", c.Message().Sender.ID, c.Message().Chat.ID)
			return c.Send(msg, tele.ModeHTML)
		}
	})

	bot.Handle(tele.OnSticker, func(c tele.Context) error {
		// 把sticker转换为合适的格式发送给用户
		sticker := c.Message().Sticker
		stickerSet := sticker.SetName
		baseDir, _ := os.Getwd()
		dir := fmt.Sprintf("%s/%s", baseDir, stickerSet)
		// 创建文件夹
		_, err := os.Stat(dir)
		if err != nil {
			if !os.IsNotExist(err) {
				return c.Send(err.Error(), tele.ModeHTML)
			}
			if os.IsNotExist(err) {
				err = os.Mkdir(dir, os.ModePerm)
				if err != nil {
					return c.Send(err.Error(), tele.ModeHTML)
				}
			}
		}

		if sticker.Animated {
			if sticker.Type != tele.StickerRegular {
				return c.Send(fmt.Sprintf("不支持的sticker类型:%s", sticker.Type), tele.ModeHTML)
			}

			tgsFileName := fmt.Sprintf("%s/%s.tgs", dir, sticker.UniqueID)
			gifFileName := strings.ReplaceAll(tgsFileName, ".tgs", ".tgs.gif")

			_, err = os.Stat(gifFileName)
			if err != nil {
				if !os.IsNotExist(err) {
					return c.Send(err.Error(), tele.ModeHTML)
				}

				err := bot.Download(sticker.MediaFile(), tgsFileName)
				if err != nil {
					return c.Send(err.Error(), tele.ModeHTML)
				}

				if runtime.GOOS != "linux" {
					return c.Send("请在linux中执行", tele.ModeHTML)
				}
				// 使用第三方的docker镜像转换tgs为gif, 参考 https://github.com/ed-asriyan/lottie-converter
				cmd := sh.Command("docker", "run", "--rm", "-v", fmt.Sprintf("%s:/source", dir), "edasriyan/lottie-to-gif")
				cmd.ShowCMD = true
				output, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Println(fmt.Sprintf("执行命令失败,错误信息:%s\n输出内容:%s", err.Error(), string(output)))
					return c.Send(string(output), tele.ModeHTML)
				}
			}

			imageFile := tele.Document{File: tele.FromDisk(gifFileName), FileName: fmt.Sprintf("%s.gif", sticker.UniqueID), Caption: "MP4格式的动图,下载后请自行转换格式为gif"}
			err = c.Send(&imageFile)
			if err != nil {
				return c.Send(err.Error(), tele.ModeHTML)
			} else {
				return nil
			}
		} else if sticker.Video {
			if sticker.Type != tele.StickerRegular {
				return c.Send(fmt.Sprintf("不支持的sticker类型:%s", sticker.Type), tele.ModeHTML)
			}

			webmFileName := fmt.Sprintf("%s/%s.webm", dir, sticker.UniqueID)
			mp4FileName := strings.ReplaceAll(webmFileName, ".webm", ".mp4")
			_, err = os.Stat(mp4FileName)
			if err != nil {
				if !os.IsNotExist(err) {
					return c.Send(err.Error(), tele.ModeHTML)
				}

				err := bot.Download(sticker.MediaFile(), webmFileName)
				if err != nil {
					return c.Send(err.Error(), tele.ModeHTML)
				}

				// windows下请手动下载ffmpeg并且放到环境变量目录中
				// 从go 1.19以后不能把可执行文件放在当前目录
				// 否则会产生一个报错类似于: exec: "ffmpeg": cannot run executable found relative to current directory
				// linux下可以自动下载 ffmpeg
				// 如果转换为gif格式会损失图像质量, 效果很不好
				cmd := sh.Command("ffmpeg", "-i", webmFileName, "-vf", "pad=ceil(iw/2)*2:ceil(ih/2)*2", mp4FileName)
				cmd.ShowCMD = true
				output, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Println(fmt.Sprintf("执行命令失败,错误信息:%s\n输出内容:%s", err.Error(), string(output)))
					return c.Send(string(output), tele.ModeHTML)
				}
			}

			mp4File := tele.Video{File: tele.FromDisk(mp4FileName), FileName: fmt.Sprintf("%s.gif", sticker.UniqueID), Caption: "MP4格式的动图,下载后请自行转换格式为gif"}
			err = c.Send(&mp4File)
			if err != nil {
				return c.Send(err.Error(), tele.ModeHTML)
			} else {
				return nil
			}
		} else {
			if sticker.Type == tele.StickerRegular {
				webpFileName := fmt.Sprintf("%s/%s.webp", dir, sticker.UniqueID)
				err := bot.Download(sticker.MediaFile(), webpFileName)
				if err != nil {
					return c.Send(err.Error(), tele.ModeHTML)
				}

				// 原始图片是webp格式, 因为有很多webp图片是透明背景, 转换为jpg格式会有黑色背景, 转换为png也有奇怪的问题, 用户可以下载后自行转换和处理
				imageFile := tele.Photo{File: tele.FromDisk(webpFileName)}
				err = c.Send(&imageFile)
				if err != nil {
					return c.Send(err.Error(), tele.ModeHTML)
				} else {
					return nil
				}
			}
			return c.Send("待完成", tele.ModeHTML)
		}
	})

	bot.Start()
}
