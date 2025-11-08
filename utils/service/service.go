package service

import (
	"StarHop/pb"
	"StarHop/utils/logger"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc/peer"
)

func GenerateCert() (tls.Certificate, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	notBefore := time.Now()
	notAfter := time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

	serialNumber, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	cert := tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  priv,
	}
	return cert, nil
}

func GetRandomListen() (net.Listener, string) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		logger.Error("failed to listen on port", err.Error())
	}
	_, port, err := net.SplitHostPort(lis.Addr().String())
	if err != nil {
		logger.Error("failed to parse listener address: ", err.Error())
	}
	return lis, port
}

// 通过Stream获取客户端IP地址
func GetClientIPFromStream(stream pb.Stream) string {
	var (
		ctx   context.Context
		label string
	)

	switch s := stream.(type) {
	case pb.HopTunnel_StreamServer:
		ctx = s.Context()
		label = "server"
	case pb.HopTunnel_StreamClient:
		ctx = s.Context()
		label = "client"
	default:
		logger.Debug("unsupported stream concrete type: ", fmt.Sprintf("%T", stream))
		return ""
	}

	p, ok := peer.FromContext(ctx)
	if !ok {
		logger.Debug("peer not found in", label, "stream context")
		return ""
	}

	host, _, err := net.SplitHostPort(p.Addr.String())
	if err != nil {
		logger.Warn("split host/port failed("+label+"): ", err.Error(), " raw=", p.Addr.String())
		return p.Addr.String()
	}
	return strings.Trim(host, "[]")
}
