package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	sensorAPI "github.com/stackrox/rox/generated/internalapi/sensor"
	utils "github.com/stackrox/rox/pkg/net"

	"github.com/stackrox/rox/generated/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	gMockSensorPort = 1338
	gMaxMsgSize     = 12 * 1024 * 1024

	gDefaultRingSize = 32
)

type MockSensor struct {
	testName string
	logger   *log.Logger
	logFile  *os.File

	listener   net.Listener
	grpcServer *grpc.Server

	// every event will be forwarded to these channels, to allow
	// tests to look directly at the incoming data without
	// losing anything underneath
	processChannel    RingChan[*storage.ProcessSignal]
	lineageChannel    RingChan[*storage.ProcessSignal_LineageInfo]
	connectionChannel RingChan[*sensorAPI.NetworkConnection]
	endpointChannel   RingChan[*sensorAPI.NetworkEndpoint]
}

func NewMockSensor(test string) *MockSensor {
	return &MockSensor{
		testName: test,
	}
}

// LiveProcesses returns a channel that can be used to read live
// process events
func (m *MockSensor) LiveProcesses() <-chan *storage.ProcessSignal {
	return m.processChannel.Stream()
}

// LiveLineages returns a channel that can be used to read live
// process lineage events
func (m *MockSensor) LiveLineages() <-chan *storage.ProcessSignal_LineageInfo {
	return m.lineageChannel.Stream()
}

// LiveConnections returns a channel that can be used to read live
// connection events
func (m *MockSensor) LiveConnections() <-chan *sensorAPI.NetworkConnection {
	return m.connectionChannel.Stream()
}

// Liveendpoints returns a channel that can be used to read live
// endpoint events
func (m *MockSensor) LiveEndpoints() <-chan *sensorAPI.NetworkEndpoint {
	return m.endpointChannel.Stream()
}

// Start will initialize the gRPC server and begin serving
// The server itself runs in a separate thread.
func (m *MockSensor) Start() {
	var err error

	m.logFile, err = os.OpenFile(
		filepath.Join("sensor.log"),
		os.O_CREATE|os.O_WRONLY, 0644,
	)

	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	m.logger = log.New(m.logFile, "", log.LstdFlags)

	m.listener, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", gMockSensorPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	m.grpcServer = grpc.NewServer(
		grpc.MaxRecvMsgSize(gMaxMsgSize),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time: 40 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	sensorAPI.RegisterSignalServiceServer(m.grpcServer, m)
	sensorAPI.RegisterNetworkConnectionInfoServiceServer(m.grpcServer, m)

	m.processChannel = NewRingChan[*storage.ProcessSignal](gDefaultRingSize)
	m.lineageChannel = NewRingChan[*storage.ProcessSignal_LineageInfo](gDefaultRingSize)
	m.connectionChannel = NewRingChan[*sensorAPI.NetworkConnection](gDefaultRingSize)
	m.endpointChannel = NewRingChan[*sensorAPI.NetworkEndpoint](gDefaultRingSize)

	go func() {
		if err := m.grpcServer.Serve(m.listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
}

// Stop will shut down the gRPC server and clear the internal store of
// all events
func (m *MockSensor) Stop() {
	m.grpcServer.Stop()
	m.listener.Close()
	m.logFile.Close()
	m.logger = nil

	m.processChannel.Stop()
	m.lineageChannel.Stop()
	m.connectionChannel.Stop()
	m.endpointChannel.Stop()
}

// PushSignals conforms to the Sensor API. It is here that process signals and
// process lineage information is handled and stored/sent to the relevant channel
func (m *MockSensor) PushSignals(stream sensorAPI.SignalService_PushSignalsServer) error {
	for {
		signal, err := stream.Recv()
		if err != nil {
			return err
		}

		if signal != nil && signal.GetSignal() != nil && signal.GetSignal().GetProcessSignal() != nil {
			processSignal := signal.GetSignal().GetProcessSignal()

			if strings.HasPrefix(processSignal.GetExecFilePath(), "/proc/self") {
				//
				// There exists a potential race condition for the driver
				// to capture very early container process events.
				//
				// This is known in falco, and somewhat documented here:
				//     https://github.com/falcosecurity/falco/blob/555bf9971cdb79318917949a5e5f9bab5293b5e2/rules/falco_rules.yaml#L1961
				//
				// It is also filtered in sensor here:
				//    https://github.com/stackrox/stackrox/blob/4d3fb539547d1935a35040e4a4e8c258a53a92e4/sensor/common/signal/signal_service.go#L90
				//
				// Further details can be found here https://issues.redhat.com/browse/ROX-11544
				//
				m.logger.Printf("runtime-process: %s %s:%s:%d:%d:%d:%s\n",
					processSignal.GetContainerId(),
					processSignal.GetName(),
					processSignal.GetExecFilePath(),
					processSignal.GetUid(),
					processSignal.GetGid(),
					processSignal.GetPid(),
					processSignal.GetArgs())
				continue
			}

			m.processChannel.Write(processSignal)

			for _, lineage := range processSignal.GetLineageInfo() {
				m.lineageChannel.Write(lineage)
			}
		}
	}
}

// PushNetworkConnectionInfo conforms to the Sensor API. It is here that networking
// events (connections and endpoints) are handled and stored/sent to the relevant channel
func (m *MockSensor) PushNetworkConnectionInfo(stream sensorAPI.NetworkConnectionInfoService_PushNetworkConnectionInfoServer) error {
	for {
		signal, err := stream.Recv()
		if err != nil {
			return err
		}

		networkConnInfo := signal.GetInfo()
		connections := networkConnInfo.GetUpdatedConnections()
		endpoints := networkConnInfo.GetUpdatedEndpoints()

		for _, endpoint := range endpoints {
			m.endpointChannel.Write(endpoint)
		}

		for _, connection := range connections {
			m.connectionChannel.Write(connection)
		}
	}
}

// translateAddress is a helper function for converting binary representations
// of network addresses (in the signals) to usable forms for testing
func (m *MockSensor) translateAddress(addr *sensorAPI.NetworkAddress) string {
	ipPortPair := utils.NetworkPeerID{
		Address: utils.IPFromBytes(addr.GetAddressData()),
		Port:    uint16(addr.GetPort()),
	}
	return ipPortPair.String()
}
