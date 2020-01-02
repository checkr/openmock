package openmock

import (
	"fmt"
	"time"

	"github.com/labstack/echo"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
)

func (ms MocksArray) DoActions(c Context) error {
	for _, m := range ms {
		if err := m.DoActions(c); err != nil {
			return nil
		}
	}
	return nil
}

func (m *Mock) DoActions(c Context) error {
	c.Values = m.Values
	if !c.MatchCondition(m.Expect.Condition) {
		return nil
	}
	var replyAction Action

	for _, actionDispatcher := range m.Actions {
		actualAction := getActualAction(actionDispatcher)
		_, isReplyHTTP := actualAction.(ActionReplyHTTP)
		_, isSendHTTP := actualAction.(ActionSendHTTP)

		if isReplyHTTP {
			if replyAction != nil {
				logrus.Fatalf("More than 1 reply_http defined for http behavior")
			}

			replyAction = actualAction
		} else if isSendHTTP {
			go func(actualAction Action, c Context) {
				performAction(actualAction, c)
			}(actualAction, c)
		} else {
			performAction(actualAction, c)
		}
	}

	if replyAction != nil {
		performAction(replyAction, c)
	}

	return nil
}

func performAction(action Action, c Context) {
	err := action.Perform(c)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"action": fmt.Sprintf("%T", action),
		}).Errorf("failed to do action")
	}
}

func (a ActionSendHTTP) Perform(context Context) error {
	if a.Sleep != 0 {
		time.Sleep(a.Sleep)
	}

	bodyStr, err := context.Render(a.Body)
	if err != nil {
		return err
	}

	urlStr, err := context.Render(a.URL)
	if err != nil {
		return err
	}

	request := gorequest.New().
		SetDebug(true).
		CustomMethod(a.Method, urlStr)

	for k, v := range a.Headers {
		request.Set(k, v)
	}

	_, _, errs := request.Send(bodyStr).End()
	if len(errs) != 0 {
		return errs[0]
	}
	return nil
}

func (a ActionReplyHTTP) Perform(context Context) error {
	ec := context.HTTPContext
	contentType := echo.MIMEApplicationJSON // default to JSON
	if ct, ok := a.Headers[echo.HeaderContentType]; ok {
		contentType = ct
	}
	for k, v := range a.Headers {
		ec.Response().Header().Set(k, v)
	}

	msg, err := context.Render(a.Body)
	if err != nil {
		logrus.WithField("err", err).Error("failed to render template for http")
		return err
	}
	return ec.Blob(a.StatusCode, contentType, []byte(msg))
}

func (a ActionRedis) Perform(context Context) error {
	for _, cmd := range a {
		_, err := context.Render(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a ActionSleep) Perform(context Context) error {
	time.Sleep(a.Duration)
	return nil
}

func (a ActionPublishKafka) Perform(context Context) error {
	msg := a.Payload
	msg, err := context.Render(msg)
	if err != nil {
		logrus.WithField("err", err).Error("failed to render template for kafka payload")
		return err
	}
	err = context.om.kafkaClient.sendMessage(a.Topic, []byte(msg))
	if err != nil {
		logrus.WithField("err", err).Error("failed to publish to kafka")
	}
	return err
}

func (a ActionPublishAMQP) Perform(context Context) error {
	msg, err := context.Render(a.Payload)
	if err != nil {
		logrus.WithField("err", err).Error("failed to render template for amqp")
		return err
	}
	publishToAMQP(
		context.om.AMQPURL,
		a.Exchange,
		a.RoutingKey,
		msg,
	)
	return nil
}
