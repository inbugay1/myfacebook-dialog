package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/inbugay1/httprouter"
	"myfacebook-dialog/internal/apiv1"
	"myfacebook-dialog/internal/myfacebookapiclient"
	"myfacebook-dialog/internal/repository"
)

type SendDialog struct {
	DialogRepository    repository.DialogRepository
	MyfacebookAPIClient *myfacebookapiclient.Client
}

type sendDialogRequest struct {
	Text string `json:"text"`
}

func (h *SendDialog) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	var sendDialogReq sendDialogRequest
	if err := json.NewDecoder(request.Body).Decode(&sendDialogReq); err != nil {
		return apiv1.NewServerError(fmt.Errorf("send dialog handler, cannot decode request body: %w", err))
	}

	defer request.Body.Close()

	if sendDialogReq.Text == "" {
		return apiv1.NewInvalidRequestErrorMissingRequiredParameter("text")
	}

	ctx := request.Context()

	senderID := ctx.Value("user_id").(string)
	receiverID := httprouter.RouteParam(ctx, "user_id") // todo validate

	dialogMsg := repository.DialogMessage{
		From: senderID,
		To:   receiverID,
		Text: sendDialogReq.Text,
	}

	err := h.MyfacebookAPIClient.SendDialogMessage(ctx, dialogMsg.From, dialogMsg.To, dialogMsg.Text)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("send dialog handler, failed to send dialog message via myfacebook api client: %w", err))
	}

	err = h.DialogRepository.Add(ctx, dialogMsg)
	if err != nil {
		return apiv1.NewServerError(fmt.Errorf("send dialog handler, failed to add dialog message to repository: %w", err))
	}

	responseWriter.Header().Set("Content-Type", "application/json; utf-8")
	responseWriter.WriteHeader(http.StatusOK)

	return nil
}
