package openmock

import (
	"time"

	"github.com/fatih/structs"
	"github.com/labstack/echo"
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

func (m *Mock) doAction(c *Context, a Action) error {
	if err := m.doActionSleep(c, a); err != nil {
		return err
	}
	if err := m.doActionPublishKafka(c, a); err != nil {
		return err
	}
	if err := m.doActionPublishAMQP(c, a); err != nil {
		return err
	}

	// lastly, do reply_http
	return m.doActionReplyHTTP(c, a)
}

func (m *Mock) doActionReplyHTTP(c *Context, a Action) error {
	if structs.IsZero(a.ActionReplyHTTP) {
		return nil
	}
	h := a.ActionReplyHTTP
	ec := c.HTTPContext
	contentType := echo.MIMEApplicationJSON // default to JSON
	if ct, ok := h.Headers[echo.HeaderContentType]; ok {
		contentType = ct
	}
	for k, v := range h.Headers {
		ec.Response().Header().Set(k, v)
	}
	msg, err := c.Render(h.Body)
	if err != nil {
		logrus.WithField("err", err).Error("failed to render template for http")
		return err
	}
	return ec.Blob(h.StatusCode, contentType, []byte(msg))
}

func (m *Mock) doActionSleep(c *Context, a Action) error {
	if structs.IsZero(a.ActionSleep) {
		return nil
	}
	time.Sleep(a.ActionSleep.Duration)
	return nil
}

func (m *Mock) doActionPublishKafka(c *Context, a Action) error {
	if structs.IsZero(a.ActionPublishKafka) {
		return nil
	}

	k := a.ActionPublishKafka
	msg := k.Payload
	msg, err := c.Render(msg)
	if err != nil {
		logrus.WithField("err", err).Error("failed to render template for kafka payload")
		return err
	}
	err = c.om.kafkaClient.sendMessage(k.Topic, []byte(msg))
	if err != nil {
		logrus.WithField("err", err).Error("failed to publish to kafka")
	}
	return err
}

func (m *Mock) doActionPublishAMQP(c *Context, a Action) error {
	if structs.IsZero(a.ActionPublishAMQP) {
		return nil
	}
	w := a.ActionPublishAMQP
	msg, err := c.Render(w.Payload)
	if err != nil {
		logrus.WithField("err", err).Error("failed to render template for amqp")
		return err
	}
	publishToAMQP(
		c.om.AMQPURL,
		w.Exchange,
		w.RoutingKey,
		msg,
	)
	return nil
}
