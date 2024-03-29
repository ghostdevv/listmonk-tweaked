package main

import (
	"fmt"
	"net/http"
	"net/textproto"

	"github.com/ghostdevv/listmonk-tweaked/models"
	"github.com/labstack/echo/v4"
)

func handleSendTxcMessage(c echo.Context) error {
	var (
		app = c.Get("app").(*App)
		m   models.TxcMessage
	)

	//* Validate Data

	if err := c.Bind(&m); err != nil {
		return err
	}

	//? Get Subscriber
	sub, err := app.core.GetSubscriber(m.SubscriberID, "", "")
	if err != nil {
		return err
	}

	if sub.Status == models.SubscriberStatusBlockListed {
		return echo.NewHTTPError(http.StatusBadRequest,
			app.i18n.Ts("globals.messages.errorCreating", "name", "email", "error", "subscriber is blocklisted"))
	}

	//? Get list
	if m.ListID == 0 {
		return echo.NewHTTPError(http.StatusBadRequest,
			app.i18n.Ts("globals.messages.notFound", "name", fmt.Sprintf("list %d", m.ListID)))
	}

	list, err := app.core.GetList(m.ListID, "")
	if err != nil {
		return err
	}

	//? Get Subscriber Lists
	var subLists []models.List

	if err := sub.Lists.Unmarshal(&subLists); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			app.i18n.Ts("globals.messages.errorFetching", "name", "lists", "error", err.Error()))
	}

	isSubscribed := false

	for _, list := range subLists {
		if list.ID == m.ListID && list.SubscriptionStatus != models.SubscriptionStatusUnsubscribed {
			isSubscribed = true
		}
	}

	if !isSubscribed {
		return echo.NewHTTPError(http.StatusBadRequest,
			app.i18n.Ts("globals.messages.errorCreating", "name", "email", "error", "subscriber not subscribed to list"))
	}

	//? Get Template
	tpl, err := app.manager.GetTpl(m.TemplateID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			app.i18n.Ts("globals.messages.notFound", "name", fmt.Sprintf("template %d", m.TemplateID)))
	}

	//* Build Message

	//? Find the default capaign template
	templates, err := app.core.GetTemplates(models.TemplateTypeCampaign, false)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.Ts("globals.messages.internalError"))
	}

	var defaultTemplate models.Template

	for _, template := range templates {
		if template.IsDefault {
			defaultTemplate = template
			break
		}
	}

	unsubscribeURL := fmt.Sprintf(app.constants.UnsubURL, list.UUID, sub.UUID)

	txcRenderData := models.TxcRenderData{
		UnsubscribeURL: unsubscribeURL,
		Subscriber:     sub,
		Campaign: models.TxcRenderDataCampaign{
			UUID:      list.UUID,
			Name:      list.Name,
			Subject:   tpl.Subject,
			FromEmail: app.constants.FromEmail,
		},
	}

	//? If the list is private don't show the name
	if list.Type == models.ListTypePrivate {
		txcRenderData.Campaign.Name = "list"
	}

	//? Render message template
	if err := m.Render(&defaultTemplate, tpl, app.manager.GenericTemplateFuncs(), txcRenderData); err != nil {
		app.log.Printf("error rendering template (%s): %v", tpl.Subject, err)
		return err
	}

	//? Prepare the final message.
	msg := models.Message{}
	msg.Subscriber = sub
	msg.To = []string{sub.Email}
	msg.From = app.constants.FromEmail
	msg.Subject = tpl.Subject
	msg.ContentType = ""
	msg.Messenger = "email"
	msg.Body = m.Body

	msg.Headers = make(textproto.MIMEHeader, 2)
	msg.Headers.Add("List-Unsubscribe-Post", "List-Unsubscribe=One-Click")
	msg.Headers.Add("List-Unsubscribe", unsubscribeURL)

	//? Send Message
	if err := app.manager.PushMessage(msg); err != nil {
		app.log.Printf("error sending message (%s): %v", msg.Subject, err)
		return err
	}

	return c.JSON(http.StatusOK, okResp{true})
}
