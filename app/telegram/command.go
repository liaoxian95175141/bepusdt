package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/cast"
	"github.com/v03413/bepusdt/app/config"
	"github.com/v03413/bepusdt/app/model"
	"github.com/v03413/bepusdt/app/rate"
	"time"
)

const cmdGetId = "id"
const cmdStart = "start"
const cmdState = "state"
const cmdWallet = "wallet"
const cmdOrder = "order"

const replayAddressText = "🚚 请发送一个合法的钱包地址"

func cmdGetIdHandle(_msg *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(_msg.Chat.ID, "您的ID: "+fmt.Sprintf("`%v`(点击复制)", _msg.Chat.ID))
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyToMessageID = _msg.MessageID
	_, _ = botApi.Send(msg)
}

func cmdStartHandle() {
	var msg = tgbotapi.NewMessage(0, "请点击钱包地址按照提示进行操作")
	var was []model.WalletAddress
	var inlineBtn [][]tgbotapi.InlineKeyboardButton
	if model.DB.Find(&was).Error == nil {
		for _, wa := range was {
			var _address = fmt.Sprintf("[✅已启用] %s", wa.Address)
			if wa.Status == model.StatusDisable {
				_address = fmt.Sprintf("[❌已禁用] %s", wa.Address)
			}

			inlineBtn = append(inlineBtn, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(_address, fmt.Sprintf("%s|%v", cbAddress, wa.Id))))
		}
	}

	inlineBtn = append(inlineBtn, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("👛 添加新的钱包地址", cbAddressAdd)))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(inlineBtn...)

	SendMsg(msg)
}

func cmdStateHandle() {
	var rows []model.TradeOrders
	model.DB.Where("created_at > ?", time.Now().Format(time.DateOnly)).Find(&rows)
	var succ uint64
	var money float64
	for _, o := range rows {
		if o.Status == model.OrderStatusSuccess {
			succ++
			money += o.Money
		}
	}

	var text = "```" + `
🎁今日成功数量：%d
💎今日订单总数：%d
💰今日成功收款：%.2f
🌟扫块成功数据：%s
-----------------------
🪧基准汇率(TRX)：%v
🪧基准汇率(USDT)：%v
✅订单汇率(TRX)：%v
✅订单汇率(USDT)：%v
-----------------------
` + "```" + `
>基准汇率：来源于交易所的原始数据。
>订单汇率：订单创建过程中实际使用的汇率。
>扫块成功数据：如果该值过低，说明您的服务器与区块链网络连接不稳定，请尝试更换区块节点。
`
	var msg = tgbotapi.NewMessage(0, fmt.Sprintf(text,
		succ,
		len(rows),
		money,
		config.GetBlockScanSuccRate(),
		cast.ToString(rate.GetOkxTrxRawRate()),
		cast.ToString(rate.GetOkxUsdtRawRate()),
		cast.ToString(rate.GetTrxCalcRate(config.DefaultTrxCnyRate)),
		cast.ToString(rate.GetUsdtCalcRate(config.DefaultUsdtCnyRate)),
	))
	msg.ParseMode = tgbotapi.ModeMarkdownV2

	SendMsg(msg)
}

func cmdWalletHandle() {
	var msg = tgbotapi.NewMessage(0, "请选择需要查询的钱包地址")
	var was []model.WalletAddress
	var inlineBtn [][]tgbotapi.InlineKeyboardButton
	if model.DB.Find(&was).Error == nil {
		for _, wa := range was {
			var _address = fmt.Sprintf("[✅已启用] %s", wa.Address)
			if wa.Status == model.StatusDisable {
				_address = fmt.Sprintf("[❌已禁用] %s", wa.Address)
			}

			inlineBtn = append(inlineBtn, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(_address, fmt.Sprintf("%s|%v", cbWallet, wa.Address))))
		}
	}

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(inlineBtn...)

	SendMsg(msg)
}

func cmdOrderHandle() {
	var msg = tgbotapi.NewMessage(0, "*下面是最近的8个订单，点击可查看详细信息*\n```\n🟢 收款成功 🔴 交易过期 🟡 等待支付\n```")
	msg.ParseMode = tgbotapi.ModeMarkdown

	var orders []model.TradeOrders
	var inlineBtn [][]tgbotapi.InlineKeyboardButton
	if model.DB.Order("id desc").Limit(8).Find(&orders).Error == nil {
		for _, order := range orders {
			var _state = "🟢"
			if order.Status == model.OrderStatusExpired {
				_state = "🔴"
			}
			if order.Status == model.OrderStatusWaiting {
				_state = "🟡"
			}

			inlineBtn = append(inlineBtn, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s 💰%.2f", _state, order.OrderId, order.Money),
				fmt.Sprintf("%s|%v", cbOrderDetail, order.TradeId),
			)))
		}
	}

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(inlineBtn...)

	SendMsg(msg)
}
