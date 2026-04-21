package cmd

import (
	"fmt"
	"strconv"

	"github.com/chatwoot/chatwoot-cli/internal/output"
	"github.com/chatwoot/chatwoot-cli/internal/sdk"
)

type InboxCmd struct {
	List   InboxListCmd   `cmd:"" default:"1" help:"List inboxes."`
	View   InboxViewCmd   `cmd:"" help:"View an inbox."`
	Create InboxCreateCmd `cmd:"" help:"Create an inbox."`
}

type InboxListCmd struct{}

func (c *InboxListCmd) Run(app *App) error {
	resp, err := app.Client.Inboxes().List()
	if err != nil {
		return err
	}

	if app.Printer.Format == "json" && !app.Printer.Quiet {
		app.Printer.PrintJSON(resp)
		return nil
	}

	if len(resp.Payload) == 0 {
		fmt.Println("No inboxes found.")
		return nil
	}

	headers := []string{"ID", "Name", "Channel Type"}
	rows := make([][]string, 0, len(resp.Payload))
	for _, inbox := range resp.Payload {
		rows = append(rows, []string{
			strconv.Itoa(inbox.ID),
			inbox.Name,
			inbox.ChannelType,
		})
	}

	app.Printer.PrintTable(headers, rows)
	return nil
}

type InboxViewCmd struct {
	ID int `arg:"" help:"Inbox ID."`
}

func (c *InboxViewCmd) Run(app *App) error {
	inbox, err := app.Client.Inboxes().Get(c.ID)
	if err != nil {
		return err
	}

	if app.Printer.Format == "json" && !app.Printer.Quiet {
		app.Printer.PrintJSON(inbox)
		return nil
	}

	app.Printer.PrintDetail([]output.KeyValue{
		{Key: "ID", Value: strconv.Itoa(inbox.ID)},
		{Key: "Name", Value: inbox.Name},
		{Key: "Channel Type", Value: inbox.ChannelType},
		{Key: "Greeting", Value: inbox.GreetingMessage},
	})

	return nil
}

type InboxCreateCmd struct {
	Name       string `required:"" help:"Inbox name."`
	Channel    string `default:"api" enum:"api" help:"Channel type. Currently only 'api' is supported."`
	WebhookURL string `help:"Webhook URL for the API channel (optional)."`
}

func (c *InboxCreateCmd) Run(app *App) error {
	inbox, err := app.Client.Inboxes().Create(sdk.CreateInboxRequest{
		Name: c.Name,
		Channel: sdk.InboxChannel{
			Type:       c.Channel,
			WebhookURL: c.WebhookURL,
		},
	})
	if err != nil {
		return err
	}

	if app.Printer.Format == "json" && !app.Printer.Quiet {
		app.Printer.PrintJSON(inbox)
		return nil
	}

	if app.Printer.Quiet {
		fmt.Println(inbox.ID)
		return nil
	}

	rows := []output.KeyValue{
		{Key: "ID", Value: strconv.Itoa(inbox.ID)},
		{Key: "Name", Value: inbox.Name},
		{Key: "Channel Type", Value: inbox.ChannelType},
	}
	if inbox.InboxIdentifier != "" {
		rows = append(rows, output.KeyValue{Key: "Inbox Identifier", Value: inbox.InboxIdentifier})
	}
	if inbox.WebhookURL != "" {
		rows = append(rows, output.KeyValue{Key: "Webhook URL", Value: inbox.WebhookURL})
	}
	app.Printer.PrintDetail(rows)
	return nil
}
