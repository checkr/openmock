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

func (actionRedis ActionRedis) Perform(context Context) error {
	for _, cmd := range actionRedis {
		_, err := context.Render(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (actionSleep ActionSleep) Perform(context Context) error {
	time.Sleep(actionSleep.Duration)
	return nil
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
