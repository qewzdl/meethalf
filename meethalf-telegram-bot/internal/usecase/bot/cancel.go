package bot

import (
	"context"
	"errors"

	"meethalf-telegram-bot/internal/domain"
)

func (s *service) cancelMessage(ctx context.Context, msg domain.IncomingMessage) (string, *domain.InlineKeyboard, error) {
	text, cancelErr := s.cancelAction(ctx, msg)
	if s == nil {
		return text, nil, cancelErr
	}

	profile, status, statusErr := s.resolveProfileStatus(ctx, msg.User.ID)
	if s.helpText != "" {
		text = text + "\n" + s.helpText
	}

	role := s.adminRoleForProfile(msg.User, profile, status)
	return text, s.startInlineKeyboardByStatus(status, role), errors.Join(cancelErr, statusErr)
}

func (s *service) cancelAction(ctx context.Context, msg domain.IncomingMessage) (string, error) {
	if msg.User.ID == 0 {
		return s.actionCanceledText(), errors.New("user id is missing")
	}

	var (
		cancelErr     error
		canceledDraft bool
		draftMode     domain.ProfileDraftMode
	)

	if s == nil || s.drafts == nil {
		cancelErr = errors.Join(cancelErr, errors.New("profile draft repository is not configured"))
	} else {
		draft, found, err := s.drafts.Get(ctx, msg.User.ID)
		if err != nil {
			cancelErr = errors.Join(cancelErr, err)
		} else if found {
			draftMode = s.draftMode(draft)
			if err := s.drafts.Delete(ctx, msg.User.ID); err != nil {
				cancelErr = errors.Join(cancelErr, err)
			} else {
				canceledDraft = true
			}
		}
	}

	if s != nil && s.deleteConfirmations != nil {
		if err := s.deleteConfirmations.Delete(ctx, msg.User.ID); err != nil {
			cancelErr = errors.Join(cancelErr, err)
		}
	}

	if s != nil && s.adminActions != nil {
		if err := s.adminActions.Delete(ctx, msg.User.ID); err != nil {
			cancelErr = errors.Join(cancelErr, err)
		}
	}

	if canceledDraft {
		if draftMode == domain.ProfileDraftModeEdit {
			return s.profileEditCanceledText(), cancelErr
		}
		return s.profileSetupCanceledText(), cancelErr
	}

	return s.actionCanceledText(), cancelErr
}
