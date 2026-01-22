package bot

import (
	"context"
	"errors"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) requestProfileDelete(ctx context.Context, msg domain.IncomingMessage) (string, error) {
	if s == nil || s.deleteConfirmations == nil {
		return "Profile delete is not available right now.", errors.New("profile delete confirmation repository is not configured")
	}

	if msg.User.ID == 0 {
		return "Profile delete is not available for this chat.", errors.New("user id is missing")
	}

	confirmation := domain.ProfileDeletionConfirmation{
		UserID:      msg.User.ID,
		ChatID:      msg.ChatID,
		RequestedAt: s.now(msg.ReceivedAt),
	}

	if err := s.deleteConfirmations.Save(ctx, confirmation); err != nil {
		return "Unable to prepare profile deletion. Please try again later.", err
	}

	return s.profileDeleteConfirmText(), nil
}

func (s *service) confirmProfileDelete(ctx context.Context, msg domain.IncomingMessage) (string, error) {
	if s == nil || s.deleteConfirmations == nil {
		return "Profile delete is not available right now.", errors.New("profile delete confirmation repository is not configured")
	}

	if msg.User.ID == 0 {
		return "Profile delete is not available for this chat.", errors.New("user id is missing")
	}

	confirmation, found, err := s.deleteConfirmations.Get(ctx, msg.User.ID)
	if err != nil {
		return "Unable to delete profile. Please try again later.", err
	}

	if !found || confirmation.ChatID != msg.ChatID {
		return s.profileDeleteExpiredText(), nil
	}

	if err := s.deleteConfirmations.Delete(ctx, msg.User.ID); err != nil {
		return "Unable to delete profile. Please try again later.", err
	}

	return s.deleteProfile(ctx, msg)
}

func (s *service) cancelProfileDelete(ctx context.Context, msg domain.IncomingMessage) (string, error) {
	if s == nil || s.deleteConfirmations == nil {
		return "Profile delete is not available right now.", errors.New("profile delete confirmation repository is not configured")
	}

	if msg.User.ID == 0 {
		return "Profile delete is not available for this chat.", errors.New("user id is missing")
	}

	if err := s.deleteConfirmations.Delete(ctx, msg.User.ID); err != nil {
		return "Unable to cancel profile deletion. Please try again later.", err
	}

	return s.profileDeleteCanceledText(), nil
}
