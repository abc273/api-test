package controller

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

const (
	alipayGatewayProduction = "https://openapi.alipay.com/gateway.do"
	alipayGatewaySandbox    = "https://openapi-sandbox.dl.alipaydev.com/gateway.do"

	alipayMethodTradePagePay = "alipay.trade.page.pay"
	alipayMethodTradeWapPay  = "alipay.trade.wap.pay"
	alipayMethodTradeQuery   = "alipay.trade.query"

	alipayPageProductCode = "FAST_INSTANT_TRADE_PAY"
	alipayWapProductCode  = "QUICK_WAP_WAY"

	alipayTradeStatusSuccess  = "TRADE_SUCCESS"
	alipayTradeStatusFinished = "TRADE_FINISHED"
	alipayTradeStatusClosed   = "TRADE_CLOSED"
	alipayTradeStatusWaiting  = "WAIT_BUYER_PAY"
)

type alipayClient struct {
	appID      string
	gatewayURL string
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

type alipayTradePayBizContent struct {
	OutTradeNo     string `json:"out_trade_no"`
	TotalAmount    string `json:"total_amount"`
	Subject        string `json:"subject"`
	ProductCode    string `json:"product_code"`
	Body           string `json:"body,omitempty"`
	TimeoutExpress string `json:"timeout_express,omitempty"`
}

type alipayTradeQueryBizContent struct {
	OutTradeNo string `json:"out_trade_no"`
}

type alipayTradeQueryEnvelope struct {
	Response alipayTradeQueryResponse `json:"alipay_trade_query_response"`
}

type alipayTradeQueryResponse struct {
	Code        string `json:"code"`
	Msg         string `json:"msg"`
	SubCode     string `json:"sub_code"`
	SubMsg      string `json:"sub_msg"`
	TradeStatus string `json:"trade_status"`
	OutTradeNo  string `json:"out_trade_no"`
	TradeNo     string `json:"trade_no"`
	TotalAmount string `json:"total_amount"`
}

func RequestAlipayPay(c *gin.Context, req *EpayRequest) {
	if !isAlipayTopUpEnabled() {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "当前管理员未配置支付宝支付"})
		return
	}
	if req.Amount < getMinTopup() {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", getMinTopup())})
		return
	}

	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}

	payMoney := getPayMoney(req.Amount, group)
	if payMoney < 0.01 {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}

	tradeNo := fmt.Sprintf("ALIUSR%dNO%s%d", id, common.GetRandomString(6), time.Now().Unix())
	client, err := getAlipayClient()
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 client 初始化失败 user_id=%d trade_no=%s error=%q", id, tradeNo, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "支付宝配置无效"})
		return
	}

	paymentURL, err := client.buildTradePayURL(req.Amount, payMoney, tradeNo, alipayTopUpNotifyURL(), alipayTopUpReturnHandlerURL(), c)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 创建充值支付链接失败 user_id=%d trade_no=%s amount=%d error=%q", id, tradeNo, req.Amount, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	amount := req.Amount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dAmount := decimal.NewFromInt(int64(amount))
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		amount = dAmount.Div(dQuotaPerUnit).IntPart()
	}

	topUp := &model.TopUp{
		UserId:          id,
		Amount:          amount,
		Money:           payMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodAlipayDirect,
		PaymentProvider: model.PaymentProviderAlipay,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 创建充值订单失败 user_id=%d trade_no=%s amount=%d error=%q", id, tradeNo, req.Amount, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}

	logger.LogInfo(c.Request.Context(), fmt.Sprintf("支付宝 充值订单创建成功 user_id=%d trade_no=%s amount=%d money=%.2f", id, tradeNo, req.Amount, payMoney))
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    nil,
		"url":     paymentURL,
	})
}

func SubscriptionRequestAlipayPay(c *gin.Context, req *SubscriptionEpayPayRequest) {
	if !isAlipayTopUpEnabled() {
		common.ApiErrorMsg(c, "当前管理员未配置支付宝支付")
		return
	}

	plan, err := model.GetSubscriptionPlanById(req.PlanId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !plan.Enabled {
		common.ApiErrorMsg(c, "套餐未启用")
		return
	}
	if plan.PriceAmount < 0.01 {
		common.ApiErrorMsg(c, "套餐金额过低")
		return
	}

	userID := c.GetInt("id")
	if plan.MaxPurchasePerUser > 0 {
		count, err := model.CountUserSubscriptionsByPlan(userID, plan.Id)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		if count >= int64(plan.MaxPurchasePerUser) {
			common.ApiErrorMsg(c, "已达到该套餐购买上限")
			return
		}
	}

	tradeNo := fmt.Sprintf("SUBALIUSR%dNO%s%d", userID, common.GetRandomString(6), time.Now().Unix())
	client, err := getAlipayClient()
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 client 初始化失败 subscription user_id=%d trade_no=%s error=%q", userID, tradeNo, err.Error()))
		common.ApiErrorMsg(c, "支付宝配置无效")
		return
	}

	paymentURL, err := client.buildSubscriptionPayURL(plan, tradeNo, alipaySubscriptionNotifyURL(), alipaySubscriptionReturnHandlerURL(), c)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 创建订阅支付链接失败 user_id=%d trade_no=%s plan_id=%d error=%q", userID, tradeNo, plan.Id, err.Error()))
		common.ApiErrorMsg(c, "拉起支付失败")
		return
	}

	order := &model.SubscriptionOrder{
		UserId:          userID,
		PlanId:          plan.Id,
		Money:           plan.PriceAmount,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodAlipayDirect,
		PaymentProvider: model.PaymentProviderAlipay,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := order.Insert(); err != nil {
		common.ApiErrorMsg(c, "创建订单失败")
		return
	}

	logger.LogInfo(c.Request.Context(), fmt.Sprintf("支付宝 订阅订单创建成功 user_id=%d trade_no=%s plan_id=%d money=%.2f", userID, tradeNo, plan.Id, plan.PriceAmount))
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    nil,
		"url":     paymentURL,
	})
}

func AlipayNotify(c *gin.Context) {
	if !isAlipayWebhookEnabled() {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("支付宝 webhook 被拒绝 reason=webhook_disabled path=%q client_ip=%s", c.Request.RequestURI, c.ClientIP()))
		writeAlipayWebhookResponse(c, false)
		return
	}

	client, err := getAlipayClient()
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 webhook client 初始化失败 path=%q client_ip=%s error=%q", c.Request.RequestURI, c.ClientIP(), err.Error()))
		writeAlipayWebhookResponse(c, false)
		return
	}

	params, err := collectAlipayParams(c)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 webhook 参数解析失败 path=%q client_ip=%s error=%q", c.Request.RequestURI, c.ClientIP(), err.Error()))
		writeAlipayWebhookResponse(c, false)
		return
	}
	logger.LogInfo(c.Request.Context(), fmt.Sprintf("支付宝 webhook 收到请求 path=%q client_ip=%s method=%s params=%q", c.Request.RequestURI, c.ClientIP(), c.Request.Method, common.GetJsonString(params)))

	if err := client.verify(params); err != nil {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("支付宝 webhook 验签失败 path=%q client_ip=%s error=%q", c.Request.RequestURI, c.ClientIP(), err.Error()))
		writeAlipayWebhookResponse(c, false)
		return
	}

	tradeNo := strings.TrimSpace(params["out_trade_no"])
	tradeStatus := strings.TrimSpace(params["trade_status"])
	if tradeNo == "" {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("支付宝 webhook 缺少 out_trade_no path=%q client_ip=%s", c.Request.RequestURI, c.ClientIP()))
		writeAlipayWebhookResponse(c, false)
		return
	}

	if err := handleAlipayTopUpStatus(c, tradeNo, tradeStatus, params); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 充值 webhook 处理失败 trade_no=%s trade_status=%s client_ip=%s error=%q", tradeNo, tradeStatus, c.ClientIP(), err.Error()))
		writeAlipayWebhookResponse(c, false)
		return
	}

	writeAlipayWebhookResponse(c, true)
}

func SubscriptionAlipayNotify(c *gin.Context) {
	if !isAlipayWebhookEnabled() {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("支付宝订阅 webhook 被拒绝 reason=webhook_disabled path=%q client_ip=%s", c.Request.RequestURI, c.ClientIP()))
		writeAlipayWebhookResponse(c, false)
		return
	}

	client, err := getAlipayClient()
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝订阅 webhook client 初始化失败 path=%q client_ip=%s error=%q", c.Request.RequestURI, c.ClientIP(), err.Error()))
		writeAlipayWebhookResponse(c, false)
		return
	}

	params, err := collectAlipayParams(c)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝订阅 webhook 参数解析失败 path=%q client_ip=%s error=%q", c.Request.RequestURI, c.ClientIP(), err.Error()))
		writeAlipayWebhookResponse(c, false)
		return
	}
	logger.LogInfo(c.Request.Context(), fmt.Sprintf("支付宝订阅 webhook 收到请求 path=%q client_ip=%s method=%s params=%q", c.Request.RequestURI, c.ClientIP(), c.Request.Method, common.GetJsonString(params)))

	if err := client.verify(params); err != nil {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("支付宝订阅 webhook 验签失败 path=%q client_ip=%s error=%q", c.Request.RequestURI, c.ClientIP(), err.Error()))
		writeAlipayWebhookResponse(c, false)
		return
	}

	tradeNo := strings.TrimSpace(params["out_trade_no"])
	tradeStatus := strings.TrimSpace(params["trade_status"])
	if tradeNo == "" {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("支付宝订阅 webhook 缺少 out_trade_no path=%q client_ip=%s", c.Request.RequestURI, c.ClientIP()))
		writeAlipayWebhookResponse(c, false)
		return
	}

	if err := handleAlipaySubscriptionStatus(c, tradeNo, tradeStatus, params); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝订阅 webhook 处理失败 trade_no=%s trade_status=%s client_ip=%s error=%q", tradeNo, tradeStatus, c.ClientIP(), err.Error()))
		writeAlipayWebhookResponse(c, false)
		return
	}

	writeAlipayWebhookResponse(c, true)
}

func AlipayReturn(c *gin.Context) {
	status := syncAlipayTopUpReturn(c)
	c.Redirect(http.StatusFound, alipayTopUpFinalReturnURL(status))
}

func SubscriptionAlipayReturn(c *gin.Context) {
	status := syncAlipaySubscriptionReturn(c)
	c.Redirect(http.StatusFound, alipaySubscriptionFinalReturnURL(status))
}

func handleAlipayTopUpStatus(c *gin.Context, tradeNo string, tradeStatus string, params map[string]string) error {
	switch tradeStatus {
	case alipayTradeStatusSuccess, alipayTradeStatusFinished:
		LockOrder(tradeNo)
		defer UnlockOrder(tradeNo)
		if err := model.RechargeAlipay(tradeNo, c.ClientIP()); err != nil {
			return err
		}
		logger.LogInfo(c.Request.Context(), fmt.Sprintf("支付宝 充值成功 trade_no=%s trade_status=%s client_ip=%s", tradeNo, tradeStatus, c.ClientIP()))
		return nil
	case alipayTradeStatusClosed:
		err := model.UpdatePendingTopUpStatus(tradeNo, model.PaymentProviderAlipay, common.TopUpStatusExpired)
		if err != nil &&
			!errors.Is(err, model.ErrTopUpStatusInvalid) &&
			!errors.Is(err, model.ErrTopUpNotFound) {
			return err
		}
		logger.LogInfo(c.Request.Context(), fmt.Sprintf("支付宝 充值订单关闭 trade_no=%s trade_status=%s client_ip=%s params=%q", tradeNo, tradeStatus, c.ClientIP(), common.GetJsonString(params)))
		return nil
	default:
		logger.LogInfo(c.Request.Context(), fmt.Sprintf("支付宝 webhook 忽略充值状态 trade_no=%s trade_status=%s client_ip=%s params=%q", tradeNo, tradeStatus, c.ClientIP(), common.GetJsonString(params)))
		return nil
	}
}

func handleAlipaySubscriptionStatus(c *gin.Context, tradeNo string, tradeStatus string, params map[string]string) error {
	switch tradeStatus {
	case alipayTradeStatusSuccess, alipayTradeStatusFinished:
		LockOrder(tradeNo)
		defer UnlockOrder(tradeNo)
		if err := model.CompleteSubscriptionOrder(tradeNo, common.GetJsonString(params), model.PaymentProviderAlipay, model.PaymentMethodAlipayDirect); err != nil {
			return err
		}
		logger.LogInfo(c.Request.Context(), fmt.Sprintf("支付宝 订阅支付成功 trade_no=%s trade_status=%s client_ip=%s", tradeNo, tradeStatus, c.ClientIP()))
		return nil
	case alipayTradeStatusClosed:
		if err := model.ExpireSubscriptionOrder(tradeNo, model.PaymentProviderAlipay); err != nil &&
			!errors.Is(err, model.ErrSubscriptionOrderStatusInvalid) &&
			!errors.Is(err, model.ErrSubscriptionOrderNotFound) {
			return err
		}
		logger.LogInfo(c.Request.Context(), fmt.Sprintf("支付宝 订阅订单关闭 trade_no=%s trade_status=%s client_ip=%s params=%q", tradeNo, tradeStatus, c.ClientIP(), common.GetJsonString(params)))
		return nil
	default:
		logger.LogInfo(c.Request.Context(), fmt.Sprintf("支付宝 webhook 忽略订阅状态 trade_no=%s trade_status=%s client_ip=%s params=%q", tradeNo, tradeStatus, c.ClientIP(), common.GetJsonString(params)))
		return nil
	}
}

func syncAlipayTopUpReturn(c *gin.Context) string {
	client, err := getAlipayClient()
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 return client 初始化失败 client_ip=%s error=%q", c.ClientIP(), err.Error()))
		return "fail"
	}

	params, err := collectAlipayParams(c)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 return 参数解析失败 client_ip=%s error=%q", c.ClientIP(), err.Error()))
		return "fail"
	}
	if err := client.verify(params); err != nil {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("支付宝 return 验签失败 client_ip=%s error=%q", c.ClientIP(), err.Error()))
		return "fail"
	}

	tradeNo := strings.TrimSpace(params["out_trade_no"])
	if tradeNo == "" {
		return "fail"
	}

	LockOrder(tradeNo)
	defer UnlockOrder(tradeNo)

	topUp := model.GetTopUpByTradeNo(tradeNo)
	if topUp == nil || topUp.PaymentProvider != model.PaymentProviderAlipay {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("支付宝 return 本地充值订单不存在或支付网关不匹配 trade_no=%s client_ip=%s", tradeNo, c.ClientIP()))
		return "fail"
	}
	if topUp.Status == common.TopUpStatusSuccess {
		return "success"
	}
	if topUp.Status == common.TopUpStatusExpired || topUp.Status == common.TopUpStatusFailed {
		return "fail"
	}

	queryResp, err := client.queryTradeStatus(c.Request.Context(), tradeNo)
	if err != nil {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("支付宝 return 查询充值状态失败 trade_no=%s client_ip=%s error=%q", tradeNo, c.ClientIP(), err.Error()))
		return "pending"
	}

	switch queryResp.TradeStatus {
	case alipayTradeStatusSuccess, alipayTradeStatusFinished:
		if err := model.RechargeAlipay(tradeNo, c.ClientIP()); err != nil {
			logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 return 补充完成充值失败 trade_no=%s client_ip=%s error=%q", tradeNo, c.ClientIP(), err.Error()))
			return "pending"
		}
		return "success"
	case alipayTradeStatusClosed:
		if err := model.UpdatePendingTopUpStatus(tradeNo, model.PaymentProviderAlipay, common.TopUpStatusExpired); err != nil &&
			!errors.Is(err, model.ErrTopUpStatusInvalid) &&
			!errors.Is(err, model.ErrTopUpNotFound) {
			logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 return 标记充值订单关闭失败 trade_no=%s client_ip=%s error=%q", tradeNo, c.ClientIP(), err.Error()))
			return "pending"
		}
		return "fail"
	default:
		return "pending"
	}
}

func syncAlipaySubscriptionReturn(c *gin.Context) string {
	client, err := getAlipayClient()
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝订阅 return client 初始化失败 client_ip=%s error=%q", c.ClientIP(), err.Error()))
		return "fail"
	}

	params, err := collectAlipayParams(c)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝订阅 return 参数解析失败 client_ip=%s error=%q", c.ClientIP(), err.Error()))
		return "fail"
	}
	if err := client.verify(params); err != nil {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("支付宝订阅 return 验签失败 client_ip=%s error=%q", c.ClientIP(), err.Error()))
		return "fail"
	}

	tradeNo := strings.TrimSpace(params["out_trade_no"])
	if tradeNo == "" {
		return "fail"
	}

	LockOrder(tradeNo)
	defer UnlockOrder(tradeNo)

	order := model.GetSubscriptionOrderByTradeNo(tradeNo)
	if order == nil || order.PaymentProvider != model.PaymentProviderAlipay {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("支付宝订阅 return 本地订单不存在或支付网关不匹配 trade_no=%s client_ip=%s", tradeNo, c.ClientIP()))
		return "fail"
	}
	if order.Status == common.TopUpStatusSuccess {
		return "success"
	}
	if order.Status == common.TopUpStatusExpired || order.Status == common.TopUpStatusFailed {
		return "fail"
	}

	queryResp, err := client.queryTradeStatus(c.Request.Context(), tradeNo)
	if err != nil {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("支付宝订阅 return 查询订单状态失败 trade_no=%s client_ip=%s error=%q", tradeNo, c.ClientIP(), err.Error()))
		return "pending"
	}

	switch queryResp.TradeStatus {
	case alipayTradeStatusSuccess, alipayTradeStatusFinished:
		if err := model.CompleteSubscriptionOrder(tradeNo, common.GetJsonString(params), model.PaymentProviderAlipay, model.PaymentMethodAlipayDirect); err != nil {
			logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝订阅 return 完成订单失败 trade_no=%s client_ip=%s error=%q", tradeNo, c.ClientIP(), err.Error()))
			return "pending"
		}
		return "success"
	case alipayTradeStatusClosed:
		if err := model.ExpireSubscriptionOrder(tradeNo, model.PaymentProviderAlipay); err != nil &&
			!errors.Is(err, model.ErrSubscriptionOrderStatusInvalid) &&
			!errors.Is(err, model.ErrSubscriptionOrderNotFound) {
			logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝订阅 return 标记订单关闭失败 trade_no=%s client_ip=%s error=%q", tradeNo, c.ClientIP(), err.Error()))
			return "pending"
		}
		return "fail"
	default:
		return "pending"
	}
}

func writeAlipayWebhookResponse(c *gin.Context, success bool) {
	if success {
		_, _ = c.Writer.Write([]byte("success"))
		return
	}
	_, _ = c.Writer.Write([]byte("fail"))
}

func collectAlipayParams(c *gin.Context) (map[string]string, error) {
	if err := c.Request.ParseForm(); err != nil {
		return nil, err
	}
	params := make(map[string]string, len(c.Request.Form))
	for key := range c.Request.Form {
		params[key] = c.Request.Form.Get(key)
	}
	return params, nil
}

func getAlipayClient() (*alipayClient, error) {
	privateKey, err := parseAlipayPrivateKey(setting.AlipayPrivateKey)
	if err != nil {
		return nil, err
	}
	publicKey, err := parseAlipayPublicKey(setting.AlipayPublicKey)
	if err != nil {
		return nil, err
	}

	gatewayURL := alipayGatewayProduction
	if setting.AlipaySandbox {
		gatewayURL = alipayGatewaySandbox
	}

	return &alipayClient{
		appID:      strings.TrimSpace(setting.AlipayAppId),
		gatewayURL: gatewayURL,
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func (client *alipayClient) buildTradePayURL(amount int64, payMoney float64, tradeNo string, notifyURL string, returnURL string, c *gin.Context) (string, error) {
	productCode := alipayPageProductCode
	methodName := alipayMethodTradePagePay
	if isMobileBrowser(c) {
		productCode = alipayWapProductCode
		methodName = alipayMethodTradeWapPay
	}

	bizContent := alipayTradePayBizContent{
		OutTradeNo:     tradeNo,
		TotalAmount:    formatPaymentAmount(payMoney),
		Subject:        fmt.Sprintf("%s 账户充值", common.SystemName),
		ProductCode:    productCode,
		Body:           fmt.Sprintf("Top up %d balance units", amount),
		TimeoutExpress: "15m",
	}
	return client.buildSignedURL(methodName, bizContent, notifyURL, returnURL)
}

func (client *alipayClient) buildSubscriptionPayURL(plan *model.SubscriptionPlan, tradeNo string, notifyURL string, returnURL string, c *gin.Context) (string, error) {
	productCode := alipayPageProductCode
	methodName := alipayMethodTradePagePay
	if isMobileBrowser(c) {
		productCode = alipayWapProductCode
		methodName = alipayMethodTradeWapPay
	}

	bizContent := alipayTradePayBizContent{
		OutTradeNo:     tradeNo,
		TotalAmount:    formatPaymentAmount(plan.PriceAmount),
		Subject:        fmt.Sprintf("%s - %s", common.SystemName, plan.Title),
		ProductCode:    productCode,
		Body:           fmt.Sprintf("Subscription purchase: %s", plan.Title),
		TimeoutExpress: "15m",
	}
	return client.buildSignedURL(methodName, bizContent, notifyURL, returnURL)
}

func (client *alipayClient) buildSignedURL(methodName string, bizContent any, notifyURL string, returnURL string) (string, error) {
	params, err := client.buildRequestValues(methodName, bizContent, notifyURL, returnURL)
	if err != nil {
		return "", err
	}
	return client.gatewayURL + "?" + params.Encode(), nil
}

func (client *alipayClient) buildRequestValues(methodName string, bizContent any, notifyURL string, returnURL string) (url.Values, error) {
	bizBytes, err := common.Marshal(bizContent)
	if err != nil {
		return nil, err
	}

	params := map[string]string{
		"app_id":      client.appID,
		"method":      methodName,
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(bizBytes),
	}
	if notifyURL != "" {
		params["notify_url"] = notifyURL
	}
	if returnURL != "" {
		params["return_url"] = returnURL
	}

	signature, err := client.sign(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = signature

	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	return values, nil
}

func (client *alipayClient) queryTradeStatus(ctx context.Context, tradeNo string) (*alipayTradeQueryResponse, error) {
	values, err := client.buildRequestValues(alipayMethodTradeQuery, alipayTradeQueryBizContent{
		OutTradeNo: tradeNo,
	}, "", "")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, client.gatewayURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected alipay status code: %d", resp.StatusCode)
	}

	var envelope alipayTradeQueryEnvelope
	if err := common.Unmarshal(bodyBytes, &envelope); err != nil {
		return nil, err
	}
	if envelope.Response.Code != "10000" {
		return nil, fmt.Errorf("alipay trade query failed: code=%s msg=%s sub_code=%s sub_msg=%s", envelope.Response.Code, envelope.Response.Msg, envelope.Response.SubCode, envelope.Response.SubMsg)
	}
	return &envelope.Response, nil
}

func (client *alipayClient) sign(params map[string]string) (string, error) {
	signSource := buildAlipaySignSource(params)
	hash := sha256.Sum256([]byte(signSource))
	signature, err := rsa.SignPKCS1v15(rand.Reader, client.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func (client *alipayClient) verify(params map[string]string) error {
	signature := strings.TrimSpace(params["sign"])
	if signature == "" {
		return errors.New("sign is empty")
	}

	signBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}

	signSource := buildAlipaySignSource(params)
	hash := sha256.Sum256([]byte(signSource))
	return rsa.VerifyPKCS1v15(client.publicKey, crypto.SHA256, hash[:], signBytes)
}

func buildAlipaySignSource(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for key, value := range params {
		if key == "sign" || key == "sign_type" || value == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+params[key])
	}
	return strings.Join(parts, "&")
}

func parseAlipayPrivateKey(raw string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(normalizePEM(raw, "PRIVATE KEY")))
	if block == nil {
		return nil, errors.New("invalid alipay private key")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("alipay private key is not RSA")
	}
	return key, nil
}

func parseAlipayPublicKey(raw string) (*rsa.PublicKey, error) {
	trimmed := strings.TrimSpace(raw)
	if strings.Contains(trimmed, "BEGIN CERTIFICATE") {
		block, _ := pem.Decode([]byte(trimmed))
		if block == nil {
			return nil, errors.New("invalid alipay certificate")
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		key, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("alipay certificate public key is not RSA")
		}
		return key, nil
	}

	block, _ := pem.Decode([]byte(normalizePEM(trimmed, "PUBLIC KEY")))
	if block == nil {
		return nil, errors.New("invalid alipay public key")
	}
	if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return key, nil
	}
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("alipay public key is not RSA")
	}
	return key, nil
}

func normalizePEM(raw string, label string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if strings.Contains(trimmed, "-----BEGIN ") {
		return trimmed
	}
	return fmt.Sprintf("-----BEGIN %s-----\n%s\n-----END %s-----", label, trimmed, label)
}

func isMobileBrowser(c *gin.Context) bool {
	if c.GetHeader("Sec-CH-UA-Mobile") == "?1" {
		return true
	}
	userAgent := strings.ToLower(c.GetHeader("User-Agent"))
	return strings.Contains(userAgent, "mobile") ||
		strings.Contains(userAgent, "android") ||
		strings.Contains(userAgent, "iphone") ||
		strings.Contains(userAgent, "ipad")
}

func formatPaymentAmount(amount float64) string {
	return strconv.FormatFloat(amount, 'f', 2, 64)
}

func joinBaseURL(base string, path string) string {
	return strings.TrimRight(base, "/") + path
}

func alipayTopUpNotifyURL() string {
	return joinBaseURL(service.GetCallbackAddress(), "/api/alipay/notify")
}

func alipayTopUpReturnHandlerURL() string {
	return joinBaseURL(system_setting.ServerAddress, "/api/alipay/return")
}

func alipaySubscriptionNotifyURL() string {
	return joinBaseURL(service.GetCallbackAddress(), "/api/subscription/alipay/notify")
}

func alipaySubscriptionReturnHandlerURL() string {
	return joinBaseURL(system_setting.ServerAddress, "/api/subscription/alipay/return")
}

func alipayTopUpFinalReturnURL(status string) string {
	if strings.TrimSpace(setting.AlipayReturnUrl) != "" {
		return strings.TrimSpace(setting.AlipayReturnUrl)
	}
	return defaultWalletReturnURL(status)
}

func alipaySubscriptionFinalReturnURL(status string) string {
	if strings.TrimSpace(setting.AlipaySubscriptionReturnUrl) != "" {
		return strings.TrimSpace(setting.AlipaySubscriptionReturnUrl)
	}
	return defaultWalletReturnURL(status)
}

func defaultWalletReturnURL(status string) string {
	base := strings.TrimRight(system_setting.ServerAddress, "/")
	if system_setting.GetThemeSettings().Frontend == "default" {
		return base + "/wallet?show_history=true"
	}
	return base + "/console/topup?pay=" + status
}
