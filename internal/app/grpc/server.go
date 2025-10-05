package grpc

import (
	"context"

	"github.com/buharamanya/shortener/internal/app/auth"
	"github.com/buharamanya/shortener/internal/app/core"
	"github.com/buharamanya/shortener/internal/app/proto"
	"github.com/buharamanya/shortener/internal/app/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	proto.UnimplementedShortenerServer
	CoreService *core.ShortenerService // экспортируемое поле
}

func NewServer(coreService *core.ShortenerService) *Server {
	return &Server{
		CoreService: coreService,
	}
}

func (s *Server) ShortenURL(ctx context.Context, req *proto.ShortenURLRequest) (*proto.ShortenURLResponse, error) {
	userID, ok := ctx.Value(auth.UserIDContextKey).(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	shortURL, err := s.CoreService.ShortenURL(ctx, req.Url, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.ShortenURLResponse{
		Result: shortURL,
		Status: 201,
	}, nil
}

func (s *Server) ShortenURLJSON(ctx context.Context, req *proto.ShortenURLJSONRequest) (*proto.ShortenURLJSONResponse, error) {
	userID, ok := ctx.Value(auth.UserIDContextKey).(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	shortURL, err := s.CoreService.ShortenURL(ctx, req.Url, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.ShortenURLJSONResponse{
		Result: shortURL,
		Status: 201,
	}, nil
}

func (s *Server) ShortenURLBatch(ctx context.Context, req *proto.ShortenURLBatchRequest) (*proto.ShortenURLBatchResponse, error) {
	userID, ok := ctx.Value(auth.UserIDContextKey).(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	var batchItems []core.BatchRequestItem
	for _, item := range req.Items {
		batchItems = append(batchItems, core.BatchRequestItem{
			CorrelationID: item.CorrelationId,
			OriginalURL:   item.OriginalUrl,
		})
	}

	results, err := s.CoreService.ShortenURLBatch(ctx, batchItems, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var responseItems []*proto.BatchResponse
	for _, result := range results {
		responseItems = append(responseItems, &proto.BatchResponse{
			CorrelationId: result.CorrelationID,
			ShortUrl:      result.ShortURL,
		})
	}

	return &proto.ShortenURLBatchResponse{
		Items:  responseItems,
		Status: 201,
	}, nil
}

func (s *Server) RedirectByShortURL(ctx context.Context, req *proto.RedirectRequest) (*proto.RedirectResponse, error) {
	originalURL, err := s.CoreService.GetOriginalURL(ctx, req.ShortCode)
	if err != nil {
		if err == storage.ErrDeleted {
			return &proto.RedirectResponse{
				Status: 410,
			}, nil
		}
		return &proto.RedirectResponse{
			Status: 400,
		}, nil
	}

	return &proto.RedirectResponse{
		Location: originalURL,
		Status:   307,
	}, nil
}

func (s *Server) GetUserURLs(ctx context.Context, _ *emptypb.Empty) (*proto.UserURLsResponse, error) {
	userID, ok := ctx.Value(auth.UserIDContextKey).(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	records, err := s.CoreService.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(records) == 0 {
		return &proto.UserURLsResponse{
			Status: 204,
		}, nil
	}

	var items []*proto.UserURL
	for _, record := range records {
		items = append(items, &proto.UserURL{
			ShortUrl:    s.CoreService.BaseURL + "/" + record.ShortCode, // исправленная строка
			OriginalUrl: record.OriginalURL,
		})
	}

	return &proto.UserURLsResponse{
		Items:  items,
		Status: 200,
	}, nil
}

func (s *Server) DeleteUserURLs(ctx context.Context, req *proto.DeleteUserURLsRequest) (*emptypb.Empty, error) {
	userID, ok := ctx.Value(auth.UserIDContextKey).(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	go func() {
		_ = s.CoreService.DeleteUserURLs(context.Background(), req.ShortCodes, userID)
	}()

	return &emptypb.Empty{}, nil
}

func (s *Server) Ping(ctx context.Context, _ *emptypb.Empty) (*proto.PingResponse, error) {
	err := s.CoreService.Ping(ctx)
	if err != nil {
		return &proto.PingResponse{
			Success: false,
			Status:  500,
		}, nil
	}

	return &proto.PingResponse{
		Success: true,
		Status:  200,
	}, nil
}

func (s *Server) GetStats(ctx context.Context, _ *emptypb.Empty) (*proto.StatsResponse, error) {
	urlsCount, usersCount, err := s.CoreService.GetStats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.StatsResponse{
		Urls:   int32(urlsCount),
		Users:  int32(usersCount),
		Status: 200,
	}, nil
}
