package turn_server

import (
	"github.com/pion/turn/v2"
	"github.com/sirupsen/logrus"
	"net"
)

var s *turn.Server

func StartTurnServer(publicIP, port, realm string) {
	logrus.Infof("Starting turn server on %s:%s", publicIP, port)
	if port == "" {
		port = "3478"
	}
	if len(publicIP) == 0 {
		logrus.WithField("public-ip", publicIP).Panicf("'public-ip' is required")
	}
	// Create a UDP listener to pass into pion/turn
	// pion/turn itself doesn't allocate any UDP sockets, but lets the user pass them in
	// this allows us to add logging, storage or modify inbound/outbound traffic
	udpListener, err := net.ListenPacket("udp4", "0.0.0.0:"+port)
	if err != nil {
		logrus.Panicf("Failed to create TURN server listener: %s", err)
	}
	s, err = turn.NewServer(turn.ServerConfig{
		Realm: realm,
		// Set AuthHandler callback
		// This is called everytime a user tries to authenticate with the TURN server
		// Return the key for that user, or false when no user is found
		AuthHandler: turn.Auth,
		// PacketConnConfigs is a list of UDP Listeners and the configuration around them
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: udpListener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorPortRange{
					RelayAddress: net.ParseIP(publicIP), // Claim that we are listening on IP passed by user (This should be your Public IP)
					Address:      "0.0.0.0",             // But actually be listening on every interface
					MinPort:      50000,
					MaxPort:      55000,
				},
			},
		},
	})
	if err != nil {
		logrus.Panic(err)
	}
}

func AddTmpAuth(username, realm, password string, limit int64) error {
	return turn.AddTmpAuth(username, realm, password, limit)
}

func DelTmpAuth(username string) error {
	return turn.DelTmpAuth(username)
}

func CloseTurnServer() error {
	logrus.Infof("Stopping turn server")
	if s == nil {
		return nil
	}
	if err := s.Close(); err != nil {
		logrus.Panicf("err occour when shutting down the turn server:%s", err.Error())
		return err
	}
	logrus.Infof("Stopped turn server")
	return nil
}
