package cmd

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type MessageCmd struct {
	List MessageListCmd `cmd:"" help:"List messages in a conversation."`
	Send MessageSendCmd `cmd:"" help:"Send a reply or private note."`
}

type MessageListCmd struct {
	ConversationID int `arg:"" help:"Conversation ID."`
	Before         int `help:"Messages before this message ID."`
}

func (c *MessageListCmd) Run(app *App) error {
	resp, err := app.Client.Messages(c.ConversationID).List(c.Before)
	if err != nil {
		return err
	}

	if app.Printer.Format == "json" && !app.Printer.Quiet {
		app.Printer.PrintJSON(resp)
		return nil
	}

	messages := resp.Payload
	if len(messages) == 0 {
		fmt.Println("No messages found.")
		return nil
	}

	headers := []string{"ID", "Type", "Sender", "Content", "Time"}
	rows := make([][]string, 0, len(messages))
	for _, msg := range messages {
		sender := ""
		if msg.Sender != nil {
			sender = msg.Sender.Name
		}
		msgType := messageTypeName(msg.MessageType)
		if msg.Private {
			msgType = "note"
		}
		content := truncate(strings.ReplaceAll(msg.Content, "\n", " "), 60)
		ts := formatTimestamp(msg.CreatedAt)

		rows = append(rows, []string{
			strconv.Itoa(msg.ID),
			msgType,
			sender,
			content,
			ts,
		})
	}

	app.Printer.PrintTable(headers, rows)
	return nil
}

type MessageSendCmd struct {
	ConversationID int    `arg:"" help:"Conversation ID."`
	Content        string `arg:"" optional:"" help:"Message content. Use '-' or omit to read from stdin."`
	Private        bool   `help:"Send as a private note instead of a public reply."`
}

func (c *MessageSendCmd) Run(app *App) error {
	content, err := resolveMessageContent(c.Content)
	if err != nil {
		return err
	}

	msg, err := app.Client.Messages(c.ConversationID).Create(content, c.Private)
	if err != nil {
		return err
	}

	if app.Printer.Format == "json" && !app.Printer.Quiet {
		app.Printer.PrintJSON(msg)
		return nil
	}

	if app.Printer.Quiet {
		fmt.Println(msg.ID)
		return nil
	}

	kind := "reply"
	if c.Private {
		kind = "note"
	}
	fmt.Printf("Sent %s (message ID: %d)\n", kind, msg.ID)
	return nil
}

func resolveMessageContent(arg string) (string, error) {
	if arg != "" && arg != "-" {
		return arg, nil
	}

	stat, _ := os.Stdin.Stat()
	if stat.Mode()&os.ModeCharDevice != 0 {
		return "", fmt.Errorf("no content provided: pass as argument or pipe via stdin")
	}

	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}

	content := strings.TrimRight(string(b), "\n")
	if content == "" {
		return "", fmt.Errorf("stdin was empty")
	}
	return content, nil
}

func messageTypeName(t int) string {
	switch t {
	case 0:
		return "incoming"
	case 1:
		return "outgoing"
	case 2:
		return "activity"
	default:
		return strconv.Itoa(t)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
