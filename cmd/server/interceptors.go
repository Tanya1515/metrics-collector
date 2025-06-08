package main

import (
	"context"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// InterceptorTrustedIP - function for checking, if client IP-address is trusted.
func (App *Application) StreamInterceptorTrustedIP(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	md, ok := metadata.FromIncomingContext(ss.Context())
	if ok {
		agentIP := md.Get("X-Real-IP")
		if agentIP[0] != "" {
			_, cidr, err := net.ParseCIDR(App.TrustedSubnet)
			if err != nil {
				App.Logger.Errorln("Error during parsing CIDR")
				return err
			}

			trustedIP := cidr.Contains(net.ParseIP(agentIP[0]))
			if !trustedIP {
				App.Logger.Errorln("Untrusted IP-adress: access denied ", agentIP[0])
				return err
			}
		}
	}

	return handler(srv, ss)
}

// InterceptorLogger - function for logging information about processing GRPC request.
func (App *Application) StreamInterceptorLogger(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {

	start := time.Now()

	err = handler(srv, ss)

	duration := time.Since(start)
	if err != nil {
		App.Logger.Errorln(
			"Method", info.FullMethod,
			"Error while processing GRPC request: ", err,
			"Duration", duration,
		)
	} else {
		App.Logger.Infoln(
			"Method", strings.Split(info.FullMethod, "/")[2],
			"Duration", duration,
			"ReponseStatus", "OK",
		)
	}

	return
}

func (App *Application) InterceptorLogger(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	start := time.Now()

	resp, err = handler(ctx, req)

	duration := time.Since(start)
	if err != nil {
		App.Logger.Errorln(
			"Method", info.FullMethod,
			"Error while processing GRPC request: ", err,
			"Duration", duration,
		)
	} else {
		App.Logger.Infoln(
			"Method", strings.Split(info.FullMethod, "/")[2],
			"Duration", duration,
			"ReponseStatus", "OK",
		)
	}

	return
}

func (App *Application) InterceptorTrustedIP(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		agentIP := md.Get("X-Real-IP")
		if agentIP[0] != "" {
			_, cidr, err := net.ParseCIDR(App.TrustedSubnet)
			if err != nil {
				App.Logger.Errorln("Error during parsing CIDR")
				return "", err
			}

			trustedIP := cidr.Contains(net.ParseIP(agentIP[0]))
			if !trustedIP {
				App.Logger.Errorln("Untrusted IP-adress: access denied ", agentIP[0])
				return "", err
			}
		}
	}

	return handler(ctx, req)
}
