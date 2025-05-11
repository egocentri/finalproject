package grpc

import (
    "context"
    "sync"

    proto "github.com/egocentri/finalproject/cmd/orchestrator/proto"
    "github.com/egocentri/finalproject/internal/config"
    "github.com/egocentri/finalproject/internal/models"
    "gorm.io/gorm"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type dispatcherServer struct {
    proto.UnimplementedDispatcherServer
    db     *gorm.DB
    cfg    *config.EnvConfig
    mu     sync.Mutex
    nextID uint32
}

func NewServer(db *gorm.DB, cfg *config.EnvConfig) proto.DispatcherServer {
    return &dispatcherServer{db: db, cfg: cfg}
}

func (s *dispatcherServer) GetTask(_ context.Context, _ *proto.Empty) (*proto.TaskResponse, error) {
var expr models.Expression
    if err := s.db.Where("result = ?", "").First(&expr).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, status.Error(codes.NotFound, "no tasks available")
        }
        // Внутренняя ошибка СУБД
        return nil, status.Errorf(codes.Internal, "database error: %v", err)
    }
    s.mu.Lock()
    s.nextID++
    id := s.nextID
    s.mu.Unlock()

    return &proto.TaskResponse{
        Task: &proto.Task{
            Id:            id,
            Expression:    expr.Expression,
            OperationTime: uint32(s.cfg.TimeEvaluation),
        },
    }, nil
}

func (s *dispatcherServer) PostTaskResult(_ context.Context, tr *proto.TaskResult) (*proto.Ack, error) {
    if err := s.db.Model(&models.Expression{}).
        Where("id = ?", tr.Id).
        Update("result", tr.Result).Error; err != nil {
        return &proto.Ack{Ok: false}, err
    }
    return &proto.Ack{Ok: true}, nil
}

