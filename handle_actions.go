package openmock

import (
	"time"

	"github.com/fatih/structs"
	"github.com/labstack/echo"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
)

// DoActions do actions based on the context
func (ms MocksArray) DoActions(c *Context) error {
	for _, m := range ms {
		if !c.MatchCondition(m.Expect.Condition) {
			continue
		}
		if err := m.DoActions(c); err != nil {
			return nil
		}
	}
	return nil
}

// DoActions runs all the actions
func (m *Mock) DoActions(c *Context) error {
	for _, a := range m.Actions {
		if err := m.doAction(c, a); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mock) doAction(c *Context, a Action) (err error) {
	var action string

	defer func() {
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err":    err,
				"action": action,
			}).Errorf("failed to do action")
		}
	}()

	if err := m.doActionSleep(c, a); err != nil {
		action = "sleep"
		return err
	}
	if err := m.doActionRedis(c, a); err != nil {
		action = "redis"
		return err
	}
	if err := m.doActionPublishKafka(c, a); err != nil {
		action = "publish_kafka"
		return err
	}
	if err := m.doActionPublishAMQP(c, a); err != nil {
		action = "publish_amqp"
		return err
	}
	if err := m.doActionSendHTTP(c, a); err != nil {
		action = "send_http"
		return err
	}

	if err := m.doActionReplyHTTP(c, a); err != nil {
		action = "reply_http"
		return err
	}

	return nil
}

func (m *Mock) doActionSendHTTP(c *Context, a Action) error {
	if structs.IsZero(a.ActionSendHTTP) {
		return nil
	}

	return a.ActionSendHTTP.Perform(*c)
}

func (actionSendHTTP ActionSendHTTP) Perform(context Context) error {
	bodyStr, err := context.Render(actionSendHTTP.Body)
	if err != nil {
		return err
	}

	urlStr, err := context.Render(actionSendHTTP.URL)
	if err != nil {
		return err
	}

	request := gorequest.New().
		SetDebug(true).
		CustomMethod(actionSendHTTP.Method, urlStr)

	for k, v := range actionSendHTTP.Headers {
		request.Set(k, v)
	}

	_, _, errs := request.Send(bodyStr).End()
	if len(errs) != 0 {
		return errs[0]
	}
	return nil
}

func (m *Mock) doActionReplyHTTP(c *Context, a Action) error {
	if structs.IsZero(a.ActionReplyHTTP) {
		return nil
	}
	return a.ActionReplyHTTP.Perform(*c)
}

func (actionReplyHTTP ActionReplyHTTP) Perform(context Context) error {
	ec := context.HTTPContext
	contentType := echo.MIMEApplicationJSON // default to JSON
	if ct, ok := actionReplyHTTP.Headers[echo.HeaderContentType]; ok {
		contentType = ct
	}
	for k, v := range actionReplyHTTP.Headers {
		ec.Response().Header().Set(k, v)
	}
	msg, err := context.Render(actionReplyHTTP.Body)
	if err != nil {
		logrus.WithField("err", err).Error("failed to render template for http")
		return err
	}
	return ec.Blob(actionReplyHTTP.StatusCode, contentType, []byte(msg))
}

func (m *Mock) doActionRedis(c *Context, a Action) error {
	if len(a.ActionRedis) == 0 {
		return nil
	}
	return a.ActionRedis.Perform(*c)
}

func (actionRedis ActionRedis) Perform(context Context) error {
	for _, cmd := range actionRedis {
		_, err := context.Render(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Mock) doActionSleep(c *Context, a Action) error {
	if structs.IsZero(a.ActionSleep) {
		return nil
	}
	return a.ActionSleep.Perform(*c)
}

func (actionSleep ActionSleep) Perform(context Context) error {
	time.Sleep(actionSleep.Duration)
	return nil
}

func (m *Mock) doActionPublishKafka(c *Context, a Action) error {
	if structs.IsZero(a.ActionPublishKafka) {
		return nil
	}

	return a.ActionPublishKafka.Perform(*c)
}

func (actionPublishKafka ActionPublishKafka) Perform(context Context) error {
	msg := actionPublishKafka.Payload
	msg, err := context.Render(msg)
	if err != nil {
		logrus.WithField("err", err).Error("failed to render template for kafka payload")
		return err
	}
	err = context.om.kafkaClient.sendMessage(actionPublishKafka.Topic, []byte(msg))
	if err != nil {
		logrus.WithField("err", err).Error("failed to publish to kafka")
	}
	return err
}

func (m *Mock) doActionPublishAMQP(c *Context, a Action) error {
	if structs.IsZero(a.ActionPublishAMQP) {
		return nil
	}
	return a.ActionPublishAMQP.Perform(*c)
}

func (actionPublishAMQP ActionPublishAMQP) Perform(context Context) error {
	msg, err := context.Render(actionPublishAMQP.Payload)
	if err != nil {
		logrus.WithField("err", err).Error("failed to render template for amqp")
		return err
	}
	publishToAMQP(
		context.om.AMQPURL,
		actionPublishAMQP.Exchange,
		actionPublishAMQP.RoutingKey,
		msg,
	)
	return nil
}
