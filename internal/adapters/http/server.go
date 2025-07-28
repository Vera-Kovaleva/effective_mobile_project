package http

import (
	"context"
	"log"
	"time"

	"ef_project/internal/domain"
	oapi "ef_project/internal/generated/oapi"
	"ef_project/internal/infra/pointer"
)

var _ oapi.StrictServerInterface = (*Server)(nil)

type Server struct {
	subscriptions domain.SubscriptionInterface
}

func NewServer(
	subscriptions domain.SubscriptionInterface,
) *Server {
	return &Server{
		subscriptions: subscriptions,
	}
}

func (s *Server) DeleteSubscriptions(ctx context.Context, request oapi.DeleteSubscriptionsRequestObject) (oapi.DeleteSubscriptionsResponseObject, error) {
	err := s.subscriptions.Delete(ctx, request.Params.Id, request.Params.Name)
	if err != nil {
		return oapi.DeleteSubscriptions400JSONResponse{
			Message: "",
		}, nil
	}
	return oapi.DeleteSubscriptions200JSONResponse{
		Message: "",
	}, nil
}

func (s *Server) GetSubscriptions(ctx context.Context, request oapi.GetSubscriptionsRequestObject) (oapi.GetSubscriptionsResponseObject, error) {
	var respponse oapi.GetSubscriptions200JSONResponse

	subscriptionsByUserID, err := s.subscriptions.ReadAllByUserID(ctx, *request.Params.Id)
	if err != nil {
		return oapi.GetSubscriptions400JSONResponse{
			Message: "",
		}, nil
	}

	for _, curSubscription := range subscriptionsByUserID {
		respponse = append(respponse, oapi.Subscription{
			Cost:      curSubscription.Cost,
			DateStart: curSubscription.StartDate.Format("01-2006"),
			DateEnd:   pointer.Ref(curSubscription.EndDate.Format("01-2006")),
			Id:        curSubscription.UserID,
			Name:      curSubscription.Name,
		})
	}

	return respponse, nil
}

func (s *Server) GetSubscriptionsTotalCost(ctx context.Context, request oapi.GetSubscriptionsTotalCostRequestObject) (oapi.GetSubscriptionsTotalCostResponseObject, error) {
	startDate, err := time.Parse("01-2006", request.Params.StartDate)
	if err != nil {
		return oapi.GetSubscriptionsTotalCost400JSONResponse{
			Message: "Неверный формат даты начала",
		}, nil
	}
	endDate, err := time.Parse("01-2006", request.Params.EndDate)
	if err != nil {
		return oapi.GetSubscriptionsTotalCost400JSONResponse{
			Message: "Неверный формат даты окончания",
		}, nil
	}
	totalCost, err := s.subscriptions.TotalSubscriptionsCost(ctx, *request.Params.Id, *request.Params.Name, startDate, pointer.Ref(endDate))

	return oapi.GetSubscriptionsTotalCost200JSONResponse{
		TotalCost: totalCost,
	}, nil
}

func (s *Server) PostSubscriptions(ctx context.Context, request oapi.PostSubscriptionsRequestObject) (oapi.PostSubscriptionsResponseObject, error) {
	startDate, err := time.Parse("01-2006", request.Body.DateStart)
	if err != nil {
		return oapi.PostSubscriptions400JSONResponse{
			Message: "Неверный формат даты начала",
		}, nil
	}
	var endDate *time.Time
	if request.Body.DateEnd != nil {
		parsedEndDate, err := time.Parse("01-2006", *request.Body.DateEnd)
		if err != nil {
			return oapi.PostSubscriptions400JSONResponse{Message: "Неверный формат даты окончания"}, nil
		}
		endDate = &parsedEndDate
	}
	err = s.subscriptions.Create(ctx, domain.Subscription{
		Name:      request.Body.Name,
		Cost:      request.Body.Cost,
		UserID:    request.Body.Id,
		StartDate: startDate,
		EndDate:   endDate,
	})

	if err != nil {
		log.Printf("failed to create subscription: %v", err)
		return oapi.PostSubscriptions400JSONResponse{Message: "Ошибка создания подписки"}, nil
	}

	return oapi.PostSubscriptions201JSONResponse{
		Message: "",
	}, nil
}

func (s *Server) PutSubscriptions(ctx context.Context, request oapi.PutSubscriptionsRequestObject) (oapi.PutSubscriptionsResponseObject, error) {
	startDate, err := time.Parse("01-2006", request.Body.DateStart)
	if err != nil {
		return oapi.PutSubscriptions400JSONResponse{
			Message: "Неверный формат даты начала",
		}, nil
	}
	var endDate *time.Time
	if request.Body.DateEnd != nil {
		parsedEndDate, err := time.Parse("01-2006", *request.Body.DateEnd)
		if err != nil {
			return oapi.PutSubscriptions200JSONResponse{Message: "Неверный формат даты окончания"}, nil
		}
		endDate = &parsedEndDate
	}
	err = s.subscriptions.Update(ctx, domain.Subscription{
		Name:      request.Body.Name,
		Cost:      request.Body.Cost,
		UserID:    request.Body.Id,
		StartDate: startDate,
		EndDate:   endDate,
	})

	return oapi.PutSubscriptions200JSONResponse{
		Message: "",
	}, nil
}
