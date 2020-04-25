package openmock

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/labstack/echo"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
)

func (ms MocksArray) DoActions(ctx Context) error {
	for _, m := range ms {
		if err := m.DoActions(ctx); err != nil {
			return nil
		}
	}
	return nil
}

func (m *Mock) DoActions(ctx Context) error {
	ctx.Values = m.Values
	ctx.currentMock = m
	if !ctx.MatchCondition(m.Expect.Condition) {
		return nil
	}
	logger := newOmLogger(ctx)
	logger.Info("doing actions")
	for _, actionDispatcher := range m.Actions {
		actualAction := getActualAction(actionDispatcher)
		if err := actualAction.Perform(ctx); err != nil {
			logger.WithFields(logrus.Fields{
				"err":    err,
				"action": fmt.Sprintf("%T", actualAction),
			}).Errorf("failed to do action")
		}
	}
	return nil
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

func (a ActionReplyGRPC) Perform(context Context) error {
	ec := context.GRPCContext

	msg, err := context.Render(a.Payload)
	if err != nil {
		logrus.WithField("err", err).Error("failed to render template for grpc")
		return err
	}

	ec.Response().Header().Set("Content-Type", "application/grpc")
	ec.Response().Header().Set("Trailer", "grpc-status, grpc-message")
	//ec.Response().WriteHeader(200)

	responseStruct := ServiceMethodResponseMap[context.GRPCService][context.GRPCMethod]
	jsonpb.UnmarshalString(msg, responseStruct)
	b, err := proto.Marshal(responseStruct)
	hdr, data := msgHeader(b)

	// length-prefixed message, see https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md
	_, err = ec.Response().Write(hdr)
	_, err = ec.Response().Write(data)

	ec.Response().Header().Set("grpc-status", "0")
	ec.Response().Header().Set("grpc-message", "OK")


	if err != nil {
		return err
	}

	ec.Response().Flush()

	if err != nil {
		return nil
	}

	return nil
}

const (
	payloadLen = 1
	sizeLen    = 4
	headerLen  = payloadLen + sizeLen
)

func msgHeader(data []byte) (hdr []byte, payload []byte) {
	hdr = make([]byte, headerLen)

	hdr[0] = byte(0)

	// Write length of payload into buf
	binary.BigEndian.PutUint32(hdr[payloadLen:], uint32(len(data)))
	return hdr, data
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
