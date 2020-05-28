package openmock

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
)

func (ms MocksArray) DoActions(ctx Context) {
	for _, m := range ms {
		if conditionMatch := m.DoActions(ctx); conditionMatch {
			return
		}
	}
}

func (m *Mock) DoActions(ctx Context) (conditionMatch bool) {
	ctx.Values = m.Values
	ctx.currentMock = m
	if !ctx.MatchCondition(m.Expect.Condition) {
		return false
	}
	logger := newOmLogger(ctx)
	logger.Info("doing actions")
	for _, actionDispatcher := range m.Actions {
		actualAction := GetActualAction(actionDispatcher)
		if err := actualAction.Perform(ctx); err != nil {
			logger.WithFields(logrus.Fields{
				"err":    err,
				"action": fmt.Sprintf("%T", actualAction),
			}).Errorf("failed to do action")
			return true
		}
	}
	return true
}

func (a ActionSendHTTP) Perform(ctx Context) error {
	bodyStr, err := ctx.Render(a.Body)
	if err != nil {
		return err
	}

	urlStr, err := ctx.Render(a.URL)
	if err != nil {
		return err
	}

	request := gorequest.New().
		SetLogger(newOmLogger(ctx)).
		CustomMethod(a.Method, urlStr)

	a.Headers, err = renderHeaders(ctx, a.Headers)
	if err != nil {
		return err
	}

	for k, v := range a.Headers {
		request.Set(k, v)
	}

	_, _, errs := request.Send(bodyStr).End()
	if len(errs) != 0 {
		return errs[0]
	}
	return nil
}

func (a ActionReplyHTTP) Perform(ctx Context) (err error) {
	ec := ctx.HTTPContext
	contentType := echo.MIMEApplicationJSON // default to JSON
	if ct, ok := a.Headers[echo.HeaderContentType]; ok {
		contentType = ct
	}

	a.Headers, err = renderHeaders(ctx, a.Headers)
	if err != nil {
		return err
	}

	for k, v := range a.Headers {
		ec.Response().Header().Set(k, v)
	}

	msg, err := ctx.Render(a.Body)
	if err != nil {
		newOmLogger(ctx).WithField("err", err).Error("failed to render template for http body")
		return err
	}

	// finalize the HTTP response so that further actions make our response wait
	msgLen := fmt.Sprintf("%d", len(msg))
	ec.Response().Header().Set("Content-Length", msgLen)
	ec.Response().Header().Set("Content-Type", contentType)
	ec.Response().WriteHeader(a.StatusCode)

	_, err = ec.Response().Write([]byte(msg))
	if err != nil {
		return err
	}

	ec.Response().Flush()
	conn, _, err := ec.Response().Hijack()
	if err != nil {
		return nil
	}
	conn.Close()

	return nil
}

func (a ActionReplyGRPC) Perform(ctx Context) error {
	ec := ctx.GRPCContext
	msg, err := ctx.Render(a.Payload)
	if err != nil {
		return err
	}

	ec.Response().Header().Set("Content-Type", "application/grpc")
	ec.Response().Header().Set("Trailer", "grpc-status, grpc-message")
	ec.Response().Header().Set("grpc-status", "0")
	ec.Response().Header().Set("grpc-message", "OK")

	a.Headers, err = renderHeaders(ctx, a.Headers)
	if err != nil {
		return err
	}
	for k, v := range a.Headers {
		ec.Response().Header().Set(k, v)
	}

	hdr, data, err := ctx.om.convertJSONToH2Response(ctx, msg)
	if err != nil {
		return err
	}

	_, err = ec.Response().Write(hdr)
	if err != nil {
		return err
	}

	_, err = ec.Response().Write(data)
	if err != nil {
		return err
	}

	ec.Response().Flush()
	return nil
}

func (a ActionRedis) Perform(ctx Context) error {
	for _, cmd := range a {
		_, err := ctx.Render(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a ActionSleep) Perform(ctx Context) error {
	time.Sleep(a.Duration)
	return nil
}

func (a ActionPublishKafka) Perform(ctx Context) error {
	logger := newOmLogger(ctx)
	msg := a.Payload
	msg, err := ctx.Render(msg)
	if err != nil {
		logger.WithField("err", err).Error("failed to render template for kafka payload")
		return err
	}
	err = ctx.om.kafkaClient.sendMessage(a.Topic, []byte(msg))
	if err != nil {
		logger.WithField("err", err).Error("failed to publish to kafka")
	}
	return err
}

func (a ActionPublishAMQP) Perform(ctx Context) error {
	msg, err := ctx.Render(a.Payload)
	if err != nil {
		newOmLogger(ctx).WithField("err", err).Error("failed to render template for amqp")
		return err
	}
	publishToAMQP(
		ctx.om.AMQPURL,
		a.Exchange,
		a.RoutingKey,
		msg,
	)
	return nil
}

func renderHeaders(ctx Context, headers map[string]string) (map[string]string, error) {
	ret := make(map[string]string)
	for k, v := range headers {
		msg, err := ctx.Render(v)
		if err != nil {
			newOmLogger(ctx).WithField("err", err).Error("failed to render template for http headers")
			return nil, err
		}
		ret[k] = msg
	}
	return ret, nil
}
