package api_test

import (
	"context"

	"github.com/TheJubadze/RateLimiter/proto/pb"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	loginCapacity    = 11
	passwordCapacity = 111
	ipCapacity       = 1111
)

var _ = ginkgo.Describe("GrpcServer Integration Tests", func() {
	var (
		client pb.RateLimiterClient
		conn   *grpc.ClientConn
	)

	ginkgo.BeforeEach(func() {
		var err error
		gRPCAddr := "rate-limiter:8081"
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		conn, err = grpc.NewClient(gRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		client = pb.NewRateLimiterClient(conn)
	})

	ginkgo.AfterEach(func() {
		_ = conn.Close()
	})

	ginkgo.Context("Authorize", func() {
		ginkgo.It("should authorize", func() {
			req := &pb.AuthorizeRequest{Ip: "192.168.1.1"}

			resp, err := client.Authorize(context.Background(), req)

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(resp.Authorized).To(gomega.BeTrue())
			gomega.Expect(resp.Message).To(gomega.Equal("Authorized"))
		})

		ginkgo.It("should authorize whitelisted IP", func() {
			r := &pb.AddToWhitelistRequest{Ip: "192.168.1.1/24"}
			_, _ = client.AddToWhitelist(context.Background(), r)
			req := &pb.AuthorizeRequest{Ip: "192.168.1.1"}

			resp, err := client.Authorize(context.Background(), req)

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(resp.Authorized).To(gomega.BeTrue())
			gomega.Expect(resp.Message).To(gomega.Equal("Authorized: IP is whitelisted"))
			rr := &pb.RemoveFromWhitelistRequest{Ip: "192.168.1.1/24"}
			_, _ = client.RemoveFromWhitelist(context.Background(), rr)
		})

		ginkgo.It("should not authorize blacklisted IP", func() {
			r := &pb.AddToBlacklistRequest{Ip: "192.168.1.1/24"}
			_, _ = client.AddToBlacklist(context.Background(), r)
			req := &pb.AuthorizeRequest{Ip: "192.168.1.2"}

			resp, err := client.Authorize(context.Background(), req)

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(resp.Authorized).To(gomega.BeFalse())
			gomega.Expect(resp.Message).To(gomega.Equal("Unauthorized: IP is blacklisted"))
			rr := &pb.RemoveFromBlacklistRequest{Ip: "192.168.1.1/24"}
			_, _ = client.RemoveFromBlacklist(context.Background(), rr)
		})

		ginkgo.It("should not authorize when login rate limit is exceeded", func() {
			req := &pb.AuthorizeRequest{Ip: "192.168.1.3", Login: "test_login"}
			var resp *pb.AuthorizeResponse
			var err error

			for i := 0; i < loginCapacity; i++ {
				resp, err = client.Authorize(context.Background(), req)
			}

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(resp.Authorized).To(gomega.BeFalse())
			gomega.Expect(resp.Message).To(gomega.Equal("Login rate limit exceeded"))
		})

		ginkgo.It("should not authorize when password rate limit is exceeded", func() {
			req := &pb.AuthorizeRequest{Ip: "192.168.1.4", Password: "test_password"}
			var resp *pb.AuthorizeResponse
			var err error

			for i := 0; i < passwordCapacity; i++ {
				resp, err = client.Authorize(context.Background(), req)
			}

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(resp.Authorized).To(gomega.BeFalse())
			gomega.Expect(resp.Message).To(gomega.Equal("Password rate limit exceeded"))
		})

		ginkgo.It("should not authorize when IP rate limit is exceeded", func() {
			req := &pb.AuthorizeRequest{Ip: "192.168.1.5"}
			var resp *pb.AuthorizeResponse
			var err error

			for i := 0; i < ipCapacity; i++ {
				resp, err = client.Authorize(context.Background(), req)
			}

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(resp.Authorized).To(gomega.BeFalse())
			gomega.Expect(resp.Message).To(gomega.Equal("IP rate limit exceeded"))
		})
	})
})
