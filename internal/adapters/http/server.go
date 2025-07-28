package http

import (
	"context"
	"log/slog"
	"strconv"
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

func (s *Server) DeleteSubscriptions(
	ctx context.Context,
	request oapi.DeleteSubscriptionsRequestObject,
) (oapi.DeleteSubscriptionsResponseObject, error) {
	slog.InfoContext(
		ctx,
		"Request delete last subscription.",
		slog.String("user_id", request.Params.Id.String()),
		slog.String("service_name", request.Params.Name),
	)
	err := s.subscriptions.Delete(ctx, request.Params.Id, request.Params.Name)

	slog.InfoContext(ctx, "Deleted last subscription")

	if err != nil {
		slog.InfoContext(ctx, "Subscriptions did not delete")
		slog.ErrorContext(ctx, "Failed to delete subscription", "error", err.Error())
		return oapi.DeleteSubscriptions400JSONResponse{
			Message: "Неверный запрос",
		}, nil
	}
	return oapi.DeleteSubscriptions200JSONResponse{
		Message: "Подписка удалена",
	}, nil
}

func (s *Server) GetSubscriptions(
	ctx context.Context,
	request oapi.GetSubscriptionsRequestObject,
) (oapi.GetSubscriptionsResponseObject, error) {
	slog.InfoContext(
		ctx,
		"Request get subscriptions by user id.",
		slog.String("user_id", request.Params.Id.String()),
	)

	var respponse oapi.GetSubscriptions200JSONResponse
	subscriptionsByUserID, err := s.subscriptions.ReadAllByUserID(ctx, *request.Params.Id)
	if err != nil {
		slog.InfoContext(ctx, "Subscriptions did not get")
		slog.ErrorContext(ctx, "Failed to get subscription", "error", err.Error())
		return oapi.GetSubscriptions400JSONResponse{
			Message: "Неверный запрос",
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

	slog.InfoContext(ctx, "Got subscriptions", "subscriptions:", respponse)

	return respponse, nil
}

func (s *Server) GetSubscriptionsTotalCost(
	ctx context.Context,
	request oapi.GetSubscriptionsTotalCostRequestObject,
) (oapi.GetSubscriptionsTotalCostResponseObject, error) {
	slog.InfoContext(
		ctx,
		"Request calculate total cost.",
		slog.String("user_id", request.Params.Id.String()),
		slog.String("start_period_date", request.Params.StartDate),
		slog.String("end_period_date", request.Params.EndDate),
	)

	startDate, err := time.Parse("01-2006", request.Params.StartDate)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid start date format", "err", err)
		return oapi.GetSubscriptionsTotalCost400JSONResponse{
			Message: "Неверный формат даты начала",
		}, nil
	}
	endDate, err := time.Parse("01-2006", request.Params.EndDate)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid end date format", "err", err)
		return oapi.GetSubscriptionsTotalCost400JSONResponse{
			Message: "Неверный формат даты окончания",
		}, nil
	}
	totalCost, err := s.subscriptions.TotalSubscriptionsCost(
		ctx,
		*request.Params.Id,
		*request.Params.Name,
		startDate,
		pointer.Ref(endDate),
	)
	if err != nil {
		slog.InfoContext(ctx, "Total cost not calculated")
		slog.ErrorContext(ctx, "Failed to get total cost subscription", "error", err.Error())
		return oapi.GetSubscriptionsTotalCost400JSONResponse{
			Message: "Ошибка подсчета цен подписок",
		}, nil
	}

	slog.InfoContext(ctx, "Total cost calculated", "totalCost", totalCost)
	return oapi.GetSubscriptionsTotalCost200JSONResponse{
		TotalCost: totalCost,
	}, nil
}

func (s *Server) PostSubscriptions(
	ctx context.Context,
	request oapi.PostSubscriptionsRequestObject,
) (oapi.PostSubscriptionsResponseObject, error) {
	slog.InfoContext(ctx, "Request create subscription.",
		slog.String("user_id", request.Body.Id.String()),
		slog.String("start_period_date", request.Body.Name),
		slog.String("start_period_date", request.Body.Name),
		slog.String("start_period_date", strconv.Itoa(request.Body.Cost)),
		slog.String("start_period_date", request.Body.DateStart),
		slog.String("end_period_date", *request.Body.DateEnd))

	startDate, err := time.Parse("01-2006", request.Body.DateStart)
	if err != nil {
		slog.InfoContext(ctx, "Subscription not created, invalid start date format")
		slog.ErrorContext(ctx, "Invalid start date format", "err", err)
		return oapi.PostSubscriptions400JSONResponse{
			Message: "Неверный формат даты начала",
		}, nil
	}
	var endDate *time.Time
	if request.Body.DateEnd != nil {
		parsedEndDate, err := time.Parse("01-2006", *request.Body.DateEnd)
		if err != nil {
			slog.InfoContext(ctx, "Subscription not created, invalid end date format")
			slog.ErrorContext(ctx, "Invalid end date format", "err", err)
			return oapi.PostSubscriptions400JSONResponse{
				Message: "Неверный формат даты окончания",
			}, nil
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
		slog.InfoContext(ctx, "Subscription not created")
		slog.ErrorContext(ctx, "Failed to create subscription", "error", err.Error())
		return oapi.PostSubscriptions400JSONResponse{Message: "Ошибка создания подписки"}, nil
	}

	slog.InfoContext(ctx, "Subscription created")
	return oapi.PostSubscriptions201JSONResponse{
		Message: "Подписка создана",
	}, nil
}

func (s *Server) PutSubscriptions(
	ctx context.Context,
	request oapi.PutSubscriptionsRequestObject,
) (oapi.PutSubscriptionsResponseObject, error) {
	slog.InfoContext(ctx, "Request update subscription.",
		slog.String("user_id", request.Body.Id.String()),
		slog.String("start_period_date", request.Body.Name),
		slog.String("start_period_date", request.Body.Name),
		slog.String("start_period_date", strconv.Itoa(request.Body.Cost)),
		slog.String("start_period_date", request.Body.DateStart),
		slog.String("end_period_date", *request.Body.DateEnd))

	startDate, err := time.Parse("01-2006", request.Body.DateStart)
	if err != nil {
		slog.InfoContext(ctx, "Subscription not updated, invalid start date format")
		slog.ErrorContext(ctx, "Invalid start date format", "err", err)
		return oapi.PutSubscriptions400JSONResponse{
			Message: "Неверный формат даты начала",
		}, nil
	}
	var endDate *time.Time
	if request.Body.DateEnd != nil {
		parsedEndDate, err := time.Parse("01-2006", *request.Body.DateEnd)
		if err != nil {
			slog.InfoContext(ctx, "Subscription not updated, invalid end date format")
			slog.ErrorContext(ctx, "Invalid end date format", "err", err)
			return oapi.PutSubscriptions400JSONResponse{
				Message: "Неверный формат даты окончания",
			}, nil
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
	if err != nil {
		slog.InfoContext(ctx, "Subscription not updated")
		slog.ErrorContext(ctx, "Failed to update subscription", "error", err.Error())
		return oapi.PutSubscriptions400JSONResponse{Message: "Ошибка обновления подписки"}, nil
	}

	slog.InfoContext(ctx, "Subscription updated")
	return oapi.PutSubscriptions200JSONResponse{
		Message: "Подписка обновлена",
	}, nil
}
