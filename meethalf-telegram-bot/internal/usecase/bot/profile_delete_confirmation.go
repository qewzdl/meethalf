package bot

import (
	"context"
	"errors"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) requestProfileDelete(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, error) {
	if s == nil || s.deleteConfirmations == nil {
		return l.message(msgProfileDeleteUnavailable), errors.New("profile delete confirmation repository is not configured")
	}

	if msg.User.ID == 0 {
		return l.message(msgProfileDeleteUnavailableChat), errors.New("user id is missing")
	}

	confirmation := domain.ProfileDeletionConfirmation{
		UserID:      msg.User.ID,
		ChatID:      msg.ChatID,
		RequestedAt: s.now(msg.ReceivedAt),
	}

	if err := s.deleteConfirmations.Save(ctx, confirmation); err != nil {
		return l.message(msgProfileDeletePrepareFailed), err
	}

	return s.profileDeleteConfirmText(l), nil
}

func (s *service) confirmProfileDelete(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, error) {
	if s == nil || s.deleteConfirmations == nil {
		return l.message(msgProfileDeleteUnavailable), errors.New("profile delete confirmation repository is not configured")
	}

	if msg.User.ID == 0 {
		return l.message(msgProfileDeleteUnavailableChat), errors.New("user id is missing")
	}

	confirmation, found, err := s.deleteConfirmations.Get(ctx, msg.User.ID)
	if err != nil {
		return l.message(msgProfileDeleteFailed), err
	}

	if !found || confirmation.ChatID != msg.ChatID {
		return s.profileDeleteExpiredText(l), nil
	}

	if err := s.deleteConfirmations.Delete(ctx, msg.User.ID); err != nil {
		return l.message(msgProfileDeleteFailed), err
	}

	return s.deleteProfile(ctx, msg, l)
}

func (s *service) cancelProfileDelete(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, error) {
	if s == nil || s.deleteConfirmations == nil {
		return l.message(msgProfileDeleteUnavailable), errors.New("profile delete confirmation repository is not configured")
	}

	if msg.User.ID == 0 {
		return l.message(msgProfileDeleteUnavailableChat), errors.New("user id is missing")
	}

	if err := s.deleteConfirmations.Delete(ctx, msg.User.ID); err != nil {
		return l.message(msgProfileDeleteCancelFailed), err
	}

	return s.profileDeleteCanceledText(l), nil
}
