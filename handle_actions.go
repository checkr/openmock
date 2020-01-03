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
	for _, actionDispatcher := range m.Actions {
		actualAction := getActualAction(actionDispatcher)
		if err := actualAction.Perform(c); err != nil {
			logrus.WithFields(logrus.Fields{
				"err":    err,
				"action": fmt.Sprintf("%T", actualAction),
			}).Errorf("failed to do action")
		}
	}
	return nil
}

func (a ActionSendHTTP) Perform(context Context) error {
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
