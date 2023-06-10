/*
 *
 * Copyright 2022 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package v2 provides the S2Av2 transport credentials used by a gRPC
// application.
package v2

import (
	"context"
	"crypto/tls"
	"errors"
<<<<<<< HEAD
	"net"
	"os"
=======
	"flag"
	"net"
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/s2a-go/fallback"
	"github.com/google/s2a-go/internal/handshaker/service"
	"github.com/google/s2a-go/internal/tokenmanager"
	"github.com/google/s2a-go/internal/v2/tlsconfigstore"
<<<<<<< HEAD
	"github.com/google/s2a-go/stream"
=======
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	commonpbv1 "github.com/google/s2a-go/internal/proto/common_go_proto"
	s2av2pb "github.com/google/s2a-go/internal/proto/v2/s2a_go_proto"
)

const (
	s2aSecurityProtocol = "tls"
<<<<<<< HEAD
	defaultS2ATimeout   = 3 * time.Second
)

// An environment variable, which sets the timeout enforced on the connection to the S2A service for handshake.
const s2aTimeoutEnv = "S2A_TIMEOUT"
=======
)

var S2ATimeout = flag.Duration("s2a_timeout", 3*time.Second, "Timeout enforced on the connection to the S2A service for handshake.")
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276

type s2av2TransportCreds struct {
	info         *credentials.ProtocolInfo
	isClient     bool
	serverName   string
	s2av2Address string
	tokenManager *tokenmanager.AccessTokenManager
	// localIdentity should only be used by the client.
	localIdentity *commonpbv1.Identity
	// localIdentities should only be used by the server.
<<<<<<< HEAD
	localIdentities           []*commonpbv1.Identity
	verificationMode          s2av2pb.ValidatePeerCertificateChainReq_VerificationMode
	fallbackClientHandshake   fallback.ClientHandshake
	getS2AStream              func(ctx context.Context, s2av2Address string) (stream.S2AStream, error)
	serverAuthorizationPolicy []byte
=======
	localIdentities         []*commonpbv1.Identity
	verificationMode        s2av2pb.ValidatePeerCertificateChainReq_VerificationMode
	fallbackClientHandshake fallback.ClientHandshake
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
}

// NewClientCreds returns a client-side transport credentials object that uses
// the S2Av2 to establish a secure connection with a server.
<<<<<<< HEAD
func NewClientCreds(s2av2Address string, localIdentity *commonpbv1.Identity, verificationMode s2av2pb.ValidatePeerCertificateChainReq_VerificationMode, fallbackClientHandshakeFunc fallback.ClientHandshake, getS2AStream func(ctx context.Context, s2av2Address string) (stream.S2AStream, error), serverAuthorizationPolicy []byte) (credentials.TransportCredentials, error) {
=======
func NewClientCreds(s2av2Address string, localIdentity *commonpbv1.Identity, verificationMode s2av2pb.ValidatePeerCertificateChainReq_VerificationMode, fallbackClientHandshakeFunc fallback.ClientHandshake) (credentials.TransportCredentials, error) {
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
	// Create an AccessTokenManager instance to use to authenticate to S2Av2.
	accessTokenManager, err := tokenmanager.NewSingleTokenAccessTokenManager()

	creds := &s2av2TransportCreds{
		info: &credentials.ProtocolInfo{
			SecurityProtocol: s2aSecurityProtocol,
		},
<<<<<<< HEAD
		isClient:                  true,
		serverName:                "",
		s2av2Address:              s2av2Address,
		localIdentity:             localIdentity,
		verificationMode:          verificationMode,
		fallbackClientHandshake:   fallbackClientHandshakeFunc,
		getS2AStream:              getS2AStream,
		serverAuthorizationPolicy: serverAuthorizationPolicy,
=======
		isClient:                true,
		serverName:              "",
		s2av2Address:            s2av2Address,
		localIdentity:           localIdentity,
		verificationMode:        verificationMode,
		fallbackClientHandshake: fallbackClientHandshakeFunc,
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
	}
	if err != nil {
		creds.tokenManager = nil
	} else {
		creds.tokenManager = &accessTokenManager
	}
	if grpclog.V(1) {
		grpclog.Info("Created client S2Av2 transport credentials.")
	}
	return creds, nil
}

// NewServerCreds returns a server-side transport credentials object that uses
// the S2Av2 to establish a secure connection with a client.
<<<<<<< HEAD
func NewServerCreds(s2av2Address string, localIdentities []*commonpbv1.Identity, verificationMode s2av2pb.ValidatePeerCertificateChainReq_VerificationMode, getS2AStream func(ctx context.Context, s2av2Address string) (stream.S2AStream, error)) (credentials.TransportCredentials, error) {
=======
func NewServerCreds(s2av2Address string, localIdentities []*commonpbv1.Identity, verificationMode s2av2pb.ValidatePeerCertificateChainReq_VerificationMode) (credentials.TransportCredentials, error) {
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
	// Create an AccessTokenManager instance to use to authenticate to S2Av2.
	accessTokenManager, err := tokenmanager.NewSingleTokenAccessTokenManager()
	creds := &s2av2TransportCreds{
		info: &credentials.ProtocolInfo{
			SecurityProtocol: s2aSecurityProtocol,
		},
		isClient:         false,
		s2av2Address:     s2av2Address,
		localIdentities:  localIdentities,
		verificationMode: verificationMode,
<<<<<<< HEAD
		getS2AStream:     getS2AStream,
=======
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
	}
	if err != nil {
		creds.tokenManager = nil
	} else {
		creds.tokenManager = &accessTokenManager
	}
	if grpclog.V(1) {
		grpclog.Info("Created server S2Av2 transport credentials.")
	}
	return creds, nil
}

// ClientHandshake performs a client-side mTLS handshake using the S2Av2.
func (c *s2av2TransportCreds) ClientHandshake(ctx context.Context, serverAuthority string, rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	if !c.isClient {
		return nil, nil, errors.New("client handshake called using server transport credentials")
	}
	// Remove the port from serverAuthority.
	serverName := removeServerNamePort(serverAuthority)
<<<<<<< HEAD
	timeoutCtx, cancel := context.WithTimeout(ctx, GetS2ATimeout())
	defer cancel()
	s2AStream, err := createStream(timeoutCtx, c.s2av2Address, c.getS2AStream)
=======
	timeoutCtx, cancel := context.WithTimeout(ctx, *S2ATimeout)
	defer cancel()
	cstream, err := createStream(timeoutCtx, c.s2av2Address)
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
	if err != nil {
		grpclog.Infof("Failed to connect to S2Av2: %v", err)
		if c.fallbackClientHandshake != nil {
			return c.fallbackClientHandshake(ctx, serverAuthority, rawConn, err)
		}
		return nil, nil, err
	}
<<<<<<< HEAD
	defer s2AStream.CloseSend()
=======
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
	if grpclog.V(1) {
		grpclog.Infof("Connected to S2Av2.")
	}
	var config *tls.Config

	var tokenManager tokenmanager.AccessTokenManager
	if c.tokenManager == nil {
		tokenManager = nil
	} else {
		tokenManager = *c.tokenManager
	}

	if c.serverName == "" {
<<<<<<< HEAD
		config, err = tlsconfigstore.GetTLSConfigurationForClient(serverName, s2AStream, tokenManager, c.localIdentity, c.verificationMode, c.serverAuthorizationPolicy)
=======
		config, err = tlsconfigstore.GetTLSConfigurationForClient(serverName, cstream, tokenManager, c.localIdentity, c.verificationMode)
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
		if err != nil {
			grpclog.Info("Failed to get client TLS config from S2Av2: %v", err)
			if c.fallbackClientHandshake != nil {
				return c.fallbackClientHandshake(ctx, serverAuthority, rawConn, err)
			}
			return nil, nil, err
		}
	} else {
<<<<<<< HEAD
		config, err = tlsconfigstore.GetTLSConfigurationForClient(c.serverName, s2AStream, tokenManager, c.localIdentity, c.verificationMode, c.serverAuthorizationPolicy)
=======
		config, err = tlsconfigstore.GetTLSConfigurationForClient(c.serverName, cstream, tokenManager, c.localIdentity, c.verificationMode)
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
		if err != nil {
			grpclog.Info("Failed to get client TLS config from S2Av2: %v", err)
			if c.fallbackClientHandshake != nil {
				return c.fallbackClientHandshake(ctx, serverAuthority, rawConn, err)
			}
			return nil, nil, err
		}
	}
	if grpclog.V(1) {
		grpclog.Infof("Got client TLS config from S2Av2.")
	}
	creds := credentials.NewTLS(config)

	conn, authInfo, err := creds.ClientHandshake(ctx, serverName, rawConn)
	if err != nil {
		grpclog.Infof("Failed to do client handshake using S2Av2: %v", err)
		if c.fallbackClientHandshake != nil {
			return c.fallbackClientHandshake(ctx, serverAuthority, rawConn, err)
		}
		return nil, nil, err
	}
	grpclog.Infof("Successfully done client handshake using S2Av2 to: %s", serverName)

	return conn, authInfo, err
}

// ServerHandshake performs a server-side mTLS handshake using the S2Av2.
func (c *s2av2TransportCreds) ServerHandshake(rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	if c.isClient {
		return nil, nil, errors.New("server handshake called using client transport credentials")
	}
<<<<<<< HEAD
	ctx, cancel := context.WithTimeout(context.Background(), GetS2ATimeout())
	defer cancel()
	s2AStream, err := createStream(ctx, c.s2av2Address, c.getS2AStream)
=======
	ctx, cancel := context.WithTimeout(context.Background(), *S2ATimeout)
	defer cancel()
	cstream, err := createStream(ctx, c.s2av2Address)
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
	if err != nil {
		grpclog.Infof("Failed to connect to S2Av2: %v", err)
		return nil, nil, err
	}
<<<<<<< HEAD
	defer s2AStream.CloseSend()
=======
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
	if grpclog.V(1) {
		grpclog.Infof("Connected to S2Av2.")
	}

	var tokenManager tokenmanager.AccessTokenManager
	if c.tokenManager == nil {
		tokenManager = nil
	} else {
		tokenManager = *c.tokenManager
	}

<<<<<<< HEAD
	config, err := tlsconfigstore.GetTLSConfigurationForServer(s2AStream, tokenManager, c.localIdentities, c.verificationMode)
=======
	config, err := tlsconfigstore.GetTLSConfigurationForServer(cstream, tokenManager, c.localIdentities, c.verificationMode)
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
	if err != nil {
		grpclog.Infof("Failed to get server TLS config from S2Av2: %v", err)
		return nil, nil, err
	}
	if grpclog.V(1) {
		grpclog.Infof("Got server TLS config from S2Av2.")
	}
	creds := credentials.NewTLS(config)
	return creds.ServerHandshake(rawConn)
}

// Info returns protocol info of s2av2TransportCreds.
func (c *s2av2TransportCreds) Info() credentials.ProtocolInfo {
	return *c.info
}

// Clone makes a deep copy of s2av2TransportCreds.
func (c *s2av2TransportCreds) Clone() credentials.TransportCredentials {
	info := *c.info
	serverName := c.serverName
	fallbackClientHandshake := c.fallbackClientHandshake

	s2av2Address := c.s2av2Address
	var tokenManager tokenmanager.AccessTokenManager
	if c.tokenManager == nil {
		tokenManager = nil
	} else {
		tokenManager = *c.tokenManager
	}
	verificationMode := c.verificationMode
	var localIdentity *commonpbv1.Identity
	if c.localIdentity != nil {
		localIdentity = proto.Clone(c.localIdentity).(*commonpbv1.Identity)
	}
	var localIdentities []*commonpbv1.Identity
	if c.localIdentities != nil {
		localIdentities = make([]*commonpbv1.Identity, len(c.localIdentities))
		for i, localIdentity := range c.localIdentities {
			localIdentities[i] = proto.Clone(localIdentity).(*commonpbv1.Identity)
		}
	}
	creds := &s2av2TransportCreds{
		info:                    &info,
		isClient:                c.isClient,
		serverName:              serverName,
		fallbackClientHandshake: fallbackClientHandshake,
		s2av2Address:            s2av2Address,
		localIdentity:           localIdentity,
		localIdentities:         localIdentities,
		verificationMode:        verificationMode,
	}
	if c.tokenManager == nil {
		creds.tokenManager = nil
	} else {
		creds.tokenManager = &tokenManager
	}
	return creds
}

// NewClientTLSConfig returns a tls.Config instance that uses S2Av2 to establish a TLS connection as
// a client. The tls.Config MUST only be used to establish a single TLS connection.
func NewClientTLSConfig(
	ctx context.Context,
	s2av2Address string,
	tokenManager tokenmanager.AccessTokenManager,
	verificationMode s2av2pb.ValidatePeerCertificateChainReq_VerificationMode,
<<<<<<< HEAD
	serverName string,
	serverAuthorizationPolicy []byte) (*tls.Config, error) {
	s2AStream, err := createStream(ctx, s2av2Address, nil)
=======
	serverName string) (*tls.Config, error) {
	cstream, err := createStream(ctx, s2av2Address)
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
	if err != nil {
		grpclog.Infof("Failed to connect to S2Av2: %v", err)
		return nil, err
	}

<<<<<<< HEAD
	return tlsconfigstore.GetTLSConfigurationForClient(removeServerNamePort(serverName), s2AStream, tokenManager, nil, verificationMode, serverAuthorizationPolicy)
=======
	return tlsconfigstore.GetTLSConfigurationForClient(removeServerNamePort(serverName), cstream, tokenManager, nil, verificationMode)
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
}

// OverrideServerName sets the ServerName in the s2av2TransportCreds protocol
// info. The ServerName MUST be a hostname.
func (c *s2av2TransportCreds) OverrideServerName(serverNameOverride string) error {
	serverName := removeServerNamePort(serverNameOverride)
	c.info.ServerName = serverName
	c.serverName = serverName
	return nil
}

// Remove the trailing port from server name.
func removeServerNamePort(serverName string) string {
	name, _, err := net.SplitHostPort(serverName)
	if err != nil {
		name = serverName
	}
	return name
}

<<<<<<< HEAD
type s2AGrpcStream struct {
	stream s2av2pb.S2AService_SetUpSessionClient
}

func (x s2AGrpcStream) Send(m *s2av2pb.SessionReq) error {
	return x.stream.Send(m)
}

func (x s2AGrpcStream) Recv() (*s2av2pb.SessionResp, error) {
	return x.stream.Recv()
}

func (x s2AGrpcStream) CloseSend() error {
	return x.stream.CloseSend()
}

func createStream(ctx context.Context, s2av2Address string, getS2AStream func(ctx context.Context, s2av2Address string) (stream.S2AStream, error)) (stream.S2AStream, error) {
	if getS2AStream != nil {
		return getS2AStream(ctx, s2av2Address)
	}
=======
func createStream(ctx context.Context, s2av2Address string) (s2av2pb.S2AService_SetUpSessionClient, error) {
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
	// TODO(rmehta19): Consider whether to close the connection to S2Av2.
	conn, err := service.Dial(s2av2Address)
	if err != nil {
		return nil, err
	}
	client := s2av2pb.NewS2AServiceClient(conn)
<<<<<<< HEAD
	gRPCStream, err := client.SetUpSession(ctx, []grpc.CallOption{}...)
	if err != nil {
		return nil, err
	}
	return &s2AGrpcStream{
		stream: gRPCStream,
	}, nil
}

// GetS2ATimeout returns the timeout enforced on the connection to the S2A service for handshake.
func GetS2ATimeout() time.Duration {
	timeout, err := time.ParseDuration(os.Getenv(s2aTimeoutEnv))
	if err != nil {
		return defaultS2ATimeout
	}
	return timeout
=======
	return client.SetUpSession(ctx, []grpc.CallOption{}...)
>>>>>>> 7efbb82b89cd2e7053d7227badb0fe4320485276
}
