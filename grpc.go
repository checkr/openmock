package openmock

import (
	"github.com/sirupsen/logrus"
)

func (om *OpenMock) startGRPC() {
	// - Where is the proto definition at?
	//   + Server needs this to start
	//   + Client also needs this
	// - see https://github.com/tokopedia/gripmock
	//	 + should come up with scheme for storing
	//     state of the call in Context...
	logrus.Infof("GRPC Mock Not Implemented")

	// Check if the call matches the 'Expect.GRPC.Method / Service'
	// if so, then setup context and then mock(s).DoActions(context)
	//   the context holds the state of the call, such as arguments
	//   to match against the mock.Expect.Condition

	// Also, implement the right actions based on the mock
}
