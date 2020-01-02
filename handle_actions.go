package openmock

import (
	"fmt"
	"time"

	"github.com/labstack/echo"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
)

// DoActions performs all actions for a given MocksArray
func (ms MocksArray) DoActions(c Context) error {
	for _, m := range ms {
		if err := m.DoActions(c); err != nil {
			return nil
		}
	}
	return nil
}

// DoActions checks if a given condition matches and performs each action defined by m.Actions
func (m *Mock) DoActions(c Context) error {
	c.Values = m.Values
	if !c.MatchCondition(m.Expect.Condition) {
		return nil
	}

	for _, actionDispatcher := range m.Actions {
		actualAction := getActualAction(actionDispatcher)
		actualAction.Perform(c)
	}

	return nil
}

// Perform defines the action to perform for the ActionSendHTTP action
func (a ActionSendHTTP) Perform(context Context) {
	go func(a ActionSendHTTP) {
		if a.Sleep != 0 {
			time.Sleep(a.Sleep)
		}

		bodyStr, err := context.Render(a.Body)
		handleActionErr(a, err)

		urlStr, err := context.Render(a.URL)
		handleActionErr(a, err)

		request := gorequest.New().
			SetDebug(true).
			CustomMethod(a.Method, urlStr)

		for k, v := range a.Headers {
			request.Set(k, v)
		}

		_, _, errs := request.Send(bodyStr).End()
		if len(errs) != 0 {
			if errs[0] != nil {
				handleActionErr(a, errs[0])
			}
		}
	}(a)
}

// Perform defines the action to perform for the ActionReplyHTTP action
func (a ActionReplyHTTP) Perform(context Context) {
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
	}

	err = ec.Blob(a.StatusCode, contentType, []byte(msg))
	handleActionErr(a, err)
}

// Perform defines the action to perform for the ActionRedis action
func (a ActionRedis) Perform(context Context) {
	for _, cmd := range a {
		_, err := context.Render(cmd)
		handleActionErr(a, err)
	}
}

// Perform defines the action to perform for the ActionSleep action
func (a ActionSleep) Perform(context Context) {
	time.Sleep(a.Duration)
}

// Perform defines the action to perform for the ActionPublishKafka action
func (a ActionPublishKafka) Perform(context Context) {
	msg := a.Payload
	msg, err := context.Render(msg)
	if err != nil {
		logrus.WithField("err", err).Error("failed to render template for kafka payload")
	}
	err = context.om.kafkaClient.sendMessage(a.Topic, []byte(msg))
	if err != nil {
		logrus.WithField("err", err).Error("failed to publish to kafka")
	}
}

// Perform defines the action to perform for the ActionPublishAMQP action
func (a ActionPublishAMQP) Perform(context Context) {
	msg, err := context.Render(a.Payload)
	if err != nil {
		logrus.WithField("err", err).Error("failed to render template for amqp")
	}
	publishToAMQP(
		context.om.AMQPURL,
		a.Exchange,
		a.RoutingKey,
		msg,
	)
}

func handleActionErr(a Action, err error) {
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"action": fmt.Sprintf("%T", a),
		}).Errorf("failed to do action")
	}
}
