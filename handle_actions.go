package openmock

import (
	"fmt"
	"time"

	"github.com/labstack/echo"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
)

// DoActions do actions based on the context
func (ms MocksArray) DoActions(c Context) error {
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

func (m *Mock) DoActions(c Context) error {
	for _, a := range m.Actions {
		actualAction := a.GetActualAction()
		if err := actualAction.Perform(c); err != nil {
			logrus.WithFields(logrus.Fields{
				"err":    err,
				"action": fmt.Sprintf("%T", actualAction),
			}).Errorf("failed to do action")
		}
	}
	return nil
}

func (self ActionSendHTTP) Perform(context Context) error {
	bodyStr, err := context.Render(self.Body)
	if err != nil {
		return err
	}

	urlStr, err := context.Render(self.URL)
	if err != nil {
		return err
	}

	request := gorequest.New().
		SetDebug(true).
		CustomMethod(self.Method, urlStr)

	for k, v := range self.Headers {
		request.Set(k, v)
	}

	_, _, errs := request.Send(bodyStr).End()
	if len(errs) != 0 {
		return errs[0]
	}
	return nil
}

func (self ActionReplyHTTP) Perform(context Context) error {
	ec := context.HTTPContext
	contentType := echo.MIMEApplicationJSON // default to JSON
	if ct, ok := self.Headers[echo.HeaderContentType]; ok {
		contentType = ct
	}
	for k, v := range self.Headers {
		ec.Response().Header().Set(k, v)
	}
	msg, err := context.Render(self.Body)
	if err != nil {
		logrus.WithField("err", err).Error("failed to render template for http")
		return err
	}
	return ec.Blob(self.StatusCode, contentType, []byte(msg))
}

func (self ActionRedis) Perform(context Context) error {
	for _, cmd := range self {
		_, err := context.Render(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self ActionSleep) Perform(context Context) error {
	time.Sleep(self.Duration)
	return nil
}

func (self ActionPublishKafka) Perform(context Context) error {
	msg := self.Payload
	msg, err := context.Render(msg)
	if err != nil {
		logrus.WithField("err", err).Error("failed to render template for kafka payload")
		return err
	}
	err = context.om.kafkaClient.sendMessage(self.Topic, []byte(msg))
	if err != nil {
		logrus.WithField("err", err).Error("failed to publish to kafka")
	}
	return err
}

func (self ActionPublishAMQP) Perform(context Context) error {
	msg, err := context.Render(self.Payload)
	if err != nil {
		logrus.WithField("err", err).Error("failed to render template for amqp")
		return err
	}
	publishToAMQP(
		context.om.AMQPURL,
		self.Exchange,
		self.RoutingKey,
		msg,
	)
	return nil
}
