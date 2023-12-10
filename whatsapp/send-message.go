package whatsapp

import (
	"context"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func SendMessage(ctx context.Context, wac *whatsmeow.Client, user string, msg string) error {
	_, err := wac.SendMessage(ctx, types.JID{
		User:   "6282269305789",
		Server: types.DefaultUserServer,
	}, &waProto.Message{
		Conversation: proto.String(msg),
	})
	// Check if an error occurred
	if err != nil {
		return err
	}
	return nil
}
