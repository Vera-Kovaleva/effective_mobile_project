package http

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"ef_project/internal/domain"
	oapi "ef_project/internal/generated/oapi"
	"ef_project/internal/infra/log"
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

func (s *Server) GetSubscriptions(
	ctx context.Context,
	request oapi.GetSubscriptionsRequestObject,
) (oapi.GetSubscriptionsResponseObject, error) {
	slog.InfoContext(
		ctx,
		"Request to get last subscription.",
		slog.String("user_id", request.Params.Id.String()),
		log.RequestID(ctx),
	)

	subscription, err := s.subscriptions.GetLatest(ctx, *request.Params.Id)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"Subscriptions did not get. Failed to get subscription.",
			log.ErrorAttr(err),
			log.RequestID(ctx),
		)
		return oapi.GetSubscriptions400JSONResponse{
			Message: "Неверный запрос",
		}, nil
	}

	slog.InfoContext(ctx, "Subscription successfully deleted.", log.RequestID(ctx))

	return oapi.GetSubscriptions200JSONResponse{
		Id:        subscription.UserID,
		Name:      subscription.Name,
		Cost:      subscription.Cost,
		DateEnd:   pointer.Ref(subscription.EndDate.Format("01-2006")),
		DateStart: subscription.StartDate.Format("01-2006"),
	}, nil
}

func (s *Server) GetAll(
	ctx context.Context,
	request oapi.GetAllRequestObject,
) (oapi.GetAllResponseObject, error) {
	slog.InfoContext(
		ctx,
		"Request to get subscriptions by user id.",
		slog.String("user_id", request.Params.Id.String()),
		log.RequestID(ctx))

	var respponse oapi.GetAll200JSONResponse
	subscriptionsByUserID, err := s.subscriptions.ReadAllByUserID(ctx, *request.Params.Id)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"Subscriptions did not get. Failed to get subscription.",
			log.ErrorAttr(err),
			log.RequestID(ctx),
		)
		return oapi.GetAll400JSONResponse{
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

	slog.InfoContext(ctx, "Subscriptions successfully got.", log.RequestID(ctx))

	return respponse, nil
}

func (s *Server) DeleteSubscriptions(
	ctx context.Context,
	request oapi.DeleteSubscriptionsRequestObject,
) (oapi.DeleteSubscriptionsResponseObject, error) {
	slog.InfoContext(
		ctx,
		"Request to delete last subscription.",
		slog.String("user_id", request.Params.Id.String()),
		slog.String("service_name", request.Params.Name),
		log.RequestID(ctx),
	)

	err := s.subscriptions.Delete(ctx, request.Params.Id, request.Params.Name)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"Subscriptions did not delete. Failed to delete subscription.",
			log.ErrorAttr(err),
			log.RequestID(ctx),
		)
		return oapi.DeleteSubscriptions400JSONResponse{
			Message: "Неверный запрос",
		}, nil
	}

	slog.InfoContext(ctx, "Subscription successfully deleted.", log.RequestID(ctx))

	return oapi.DeleteSubscriptions200JSONResponse{
		Message: "Подписка удалена",
	}, nil
}

func (s *Server) GetSubscriptionsTotalCost(
	ctx context.Context,
	request oapi.GetSubscriptionsTotalCostRequestObject,
) (oapi.GetSubscriptionsTotalCostResponseObject, error) {
	startDate, err := time.Parse("01-2006", request.Params.StartDate)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid start date format.", log.ErrorAttr(err), log.RequestID(ctx))
		return oapi.GetSubscriptionsTotalCost400JSONResponse{
			Message: "Неверный формат даты начала",
		}, nil
	}
	endDate, err := time.Parse("01-2006", request.Params.EndDate)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid end date format.", log.ErrorAttr(err), log.RequestID(ctx))
		return oapi.GetSubscriptionsTotalCost400JSONResponse{
			Message: "Неверный формат даты окончания",
		}, nil
	}
	slog.InfoContext(
		ctx,
		"Request to calculate total cost.",
		slog.String("user_id", request.Params.Id.String()),
		slog.String("start_period_date", request.Params.StartDate),
		slog.String("end_period_date", request.Params.EndDate),
		log.RequestID(ctx),
	)

	totalCost, err := s.subscriptions.TotalSubscriptionsCost(
		ctx,
		*request.Params.Id,
		*request.Params.Name,
		startDate,
		pointer.Ref(endDate),
	)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"Total cost did not calculate. Failed to calculcate cost.",
			log.ErrorAttr(err),
			log.RequestID(ctx),
		)

		return oapi.GetSubscriptionsTotalCost400JSONResponse{
			Message: "Ошибка подсчета цен подписок",
		}, nil
	}

	slog.InfoContext(
		ctx,
		"Total cost successfully calculated.",
		slog.String("total_cost", strconv.Itoa(totalCost)),
		log.RequestID(ctx),
	)
	return oapi.GetSubscriptionsTotalCost200JSONResponse{
		TotalCost: totalCost,
	}, nil
}

func (s *Server) PostSubscriptions(
	ctx context.Context,
	request oapi.PostSubscriptionsRequestObject,
) (oapi.PostSubscriptionsResponseObject, error) {
	startDate, err := time.Parse("01-2006", request.Body.DateStart)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid start date format.", log.ErrorAttr(err), log.RequestID(ctx))
		return oapi.PostSubscriptions400JSONResponse{
			Message: "Неверный формат даты начала",
		}, nil
	}
	var endDate *time.Time
	if request.Body.DateEnd != nil {
		parsedEndDate, err := time.Parse("01-2006", *request.Body.DateEnd)
		if err != nil {
			slog.ErrorContext(
				ctx,
				"Invalid end date format.",
				log.ErrorAttr(err),
				log.RequestID(ctx),
			)
			return oapi.PostSubscriptions400JSONResponse{
				Message: "Неверный формат даты окончания",
			}, nil
		}
		endDate = &parsedEndDate
	} else {
		endDate = pointer.Ref(time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC))
	}
	slog.InfoContext(ctx, "Request to create subscription.",
		slog.String("user_id", request.Body.Id.String()),
		slog.String("service_name", request.Body.Name),
		slog.String("service_cost", strconv.Itoa(request.Body.Cost)),
		slog.String("start_period_date", request.Body.DateStart),
		slog.String("end_period_date", endDate.String()),
		log.RequestID(ctx),
	)
	err = s.subscriptions.Create(ctx, domain.Subscription{
		Name:      request.Body.Name,
		Cost:      request.Body.Cost,
		UserID:    request.Body.Id,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		slog.ErrorContext(
			ctx,
			"Subscriptions did not create. Failed to create subscription.",
			log.ErrorAttr(err),
			log.RequestID(ctx),
		)
		return oapi.PostSubscriptions400JSONResponse{Message: "Ошибка создания подписки"}, nil
	}

	slog.InfoContext(ctx, "Subscription successfully created.", log.RequestID(ctx))
	return oapi.PostSubscriptions201JSONResponse{
		Message: "Подписка создана",
	}, nil
}

func (s *Server) PutSubscriptions(
	ctx context.Context,
	request oapi.PutSubscriptionsRequestObject,
) (oapi.PutSubscriptionsResponseObject, error) {
	startDate, err := time.Parse("01-2006", request.Body.DateStart)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid start date format.", log.ErrorAttr(err), log.RequestID(ctx))
		return oapi.PutSubscriptions400JSONResponse{
			Message: "Неверный формат даты начала",
		}, nil
	}
	var endDate *time.Time
	if request.Body.DateEnd != nil {
		parsedEndDate, err := time.Parse("01-2006", *request.Body.DateEnd)
		if err != nil {
			slog.ErrorContext(
				ctx,
				"Invalid end date format.",
				log.ErrorAttr(err),
				log.RequestID(ctx),
			)
			return oapi.PutSubscriptions400JSONResponse{
				Message: "Неверный формат даты окончания",
			}, nil
		}
		endDate = &parsedEndDate
	} else {
		endDate = pointer.Ref(time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC))
	}
	slog.InfoContext(ctx, "Request to update subscription.",
		slog.String("user_id", request.Body.Id.String()),
		slog.String("service_name", request.Body.Name),
		slog.String("service_cost", strconv.Itoa(request.Body.Cost)),
		slog.String("start_period_date", request.Body.DateStart),
		slog.String("end_period_date", endDate.String()),
		log.RequestID(ctx),
	)

	err = s.subscriptions.Update(ctx, domain.Subscription{
		Name:      request.Body.Name,
		Cost:      request.Body.Cost,
		UserID:    request.Body.Id,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		slog.ErrorContext(
			ctx,
			"Subscriptions did not update. Failed to update subscription.",
			log.ErrorAttr(err),
			log.RequestID(ctx),
		)
		return oapi.PutSubscriptions400JSONResponse{Message: "Ошибка обновления подписки"}, nil
	}

	slog.InfoContext(ctx, "Subscription successfully updated.", log.RequestID(ctx))
	return oapi.PutSubscriptions200JSONResponse{
		Message: "Подписка обновлена",
	}, nil
}
