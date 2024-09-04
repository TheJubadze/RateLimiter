package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/TheJubadze/RateLimiter/proto/pb"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	configFile string
	grpcAddr   string
)

var rootCmd = &cobra.Command{
	Use:   "rate-limiter-cli",
	Short: "CLI for Rate Limiter Service\nUsage Example:\n  rate-limiter-cli --config \"config.yaml\" --grpc-addr \"localhost:8081\" add-wl --ip=192.168.1.1/24",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "/etc/rate-limiter/config.yaml", "Path to configuration file")
	rootCmd.PersistentFlags().StringVar(&grpcAddr, "grpc-addr", "localhost:50051", "Address of the gRPC server")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func executeGRPCCommand(ip string, grpcFunc func(client pb.RateLimiterClient, ctx context.Context) (string, error)) {
	if ip == "" {
		fmt.Println("IP must be provided")
		return
	}
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)

	client := pb.NewRateLimiterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	message, err := grpcFunc(client, ctx)
	if err != nil {
		log.Printf("command execution failed: %v", err)
	}
	fmt.Println(message)
}

var addToWhitelistCmd = &cobra.Command{
	Use:   "add-wl",
	Short: "Add an IP to the whitelist",
	Run: func(cmd *cobra.Command, _ []string) {
		ip, _ := cmd.Flags().GetString("ip")
		executeGRPCCommand(ip, func(client pb.RateLimiterClient, ctx context.Context) (string, error) {
			response, err := client.AddToWhitelist(ctx, &pb.AddToWhitelistRequest{Ip: ip})
			if err != nil {
				return "", err
			}
			return response.Message, nil
		})
	},
}

var addToBlacklistCmd = &cobra.Command{
	Use:   "add-bl",
	Short: "Add an IP to the blacklist",
	Run: func(cmd *cobra.Command, _ []string) {
		ip, _ := cmd.Flags().GetString("ip")
		executeGRPCCommand(ip, func(client pb.RateLimiterClient, ctx context.Context) (string, error) {
			response, err := client.AddToBlacklist(ctx, &pb.AddToBlacklistRequest{Ip: ip})
			if err != nil {
				return "", err
			}
			return response.Message, nil
		})
	},
}

var removeFromWhitelistCmd = &cobra.Command{
	Use:   "rm-wl",
	Short: "Remove an IP from the whitelist",
	Run: func(cmd *cobra.Command, _ []string) {
		ip, _ := cmd.Flags().GetString("ip")
		executeGRPCCommand(ip, func(client pb.RateLimiterClient, ctx context.Context) (string, error) {
			response, err := client.RemoveFromWhitelist(ctx, &pb.RemoveFromWhitelistRequest{Ip: ip})
			if err != nil {
				return "", err
			}
			return response.Message, nil
		})
	},
}

var removeFromBlacklistCmd = &cobra.Command{
	Use:   "rm-bl",
	Short: "Remove an IP from the blacklist",
	Run: func(cmd *cobra.Command, _ []string) {
		ip, _ := cmd.Flags().GetString("ip")
		executeGRPCCommand(ip, func(client pb.RateLimiterClient, ctx context.Context) (string, error) {
			response, err := client.RemoveFromBlacklist(ctx, &pb.RemoveFromBlacklistRequest{Ip: ip})
			if err != nil {
				return "", err
			}
			return response.Message, nil
		})
	},
}

func init() {
	rootCmd.AddCommand(addToWhitelistCmd)
	addToWhitelistCmd.Flags().String("ip", "", "IP to add to the whitelist")

	rootCmd.AddCommand(addToBlacklistCmd)
	addToBlacklistCmd.Flags().String("ip", "", "IP to add to the blacklist")

	rootCmd.AddCommand(removeFromWhitelistCmd)
	removeFromWhitelistCmd.Flags().String("ip", "", "IP to remove from the whitelist")

	rootCmd.AddCommand(removeFromBlacklistCmd)
	removeFromBlacklistCmd.Flags().String("ip", "", "IP to remove from the blacklist")
}
