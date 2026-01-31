package bot

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"meethalf-telegram-bot/internal/domain"
)

const adminAdBroadcastDelay = 30 * time.Millisecond

type adBroadcastJob struct {
	AdminID     int64
	AdminChatID int64
	Text        string
	PhotoIDs    []string
	RequestedAt time.Time
}

type adBroadcastResult struct {
	Total   int
	Sent    int
	Failed  int
	Skipped int
	Err     error
}

func (s *service) adminAdMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(l), nil, roleErr
	}
	if role != adminRoleAdmin {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg, l)
		return text, keyboard, errors.Join(roleErr, err)
	}

	if s == nil || s.admin == nil {
		return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin service is not configured")
	}

	text, photoIDs := s.adminAdPayload(msg)
	if text == "" && len(photoIDs) == 0 {
		return s.startAdminAd(ctx, msg, role, l)
	}

	if len([]rune(text)) > maxAdTextLength {
		return s.adminAdTooLongText(l, maxAdTextLength), s.adminAdInlineKeyboard(l), nil
	}

	photoID := ""
	if len(photoIDs) > 0 {
		photoID = photoIDs[0]
	}

	if _, err := s.admin.CreateAd(ctx, text, photoID); err != nil {
		errorText := s.adminAdFailedText(l)
		var status statusError
		if errors.As(err, &status) {
			if status.StatusCode() == http.StatusBadRequest {
				errorText = s.adminAdUsageText(l)
			}
		}
		return errorText, s.adminMenuInlineKeyboard(l, role), err
	}

	job := adBroadcastJob{
		AdminID:     msg.User.ID,
		AdminChatID: msg.ChatID,
		Text:        text,
		PhotoIDs:    photoIDs,
		RequestedAt: s.now(msg.ReceivedAt),
	}
	if !s.enqueueAdBroadcast(job) {
		return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), errors.New("ad broadcast queue is full")
	}

	_ = s.clearAdminAction(ctx, msg.User.ID)
	s.registerAdminActionCleanup(msg)
	return s.adminAdQueuedText(l), s.adminMenuInlineKeyboard(l, role), nil
}

func (s *service) startAdminAd(ctx context.Context, msg domain.IncomingMessage, role adminRole, l localizer) (string, *domain.InlineKeyboard, error) {
	if s == nil || s.adminActions == nil {
		return s.adminAdUsageText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin action repository is not configured")
	}

	action := domain.AdminActionState{
		UserID:      msg.User.ID,
		ChatID:      msg.ChatID,
		Action:      domain.AdminActionPostAd,
		RequestedAt: s.now(msg.ReceivedAt),
	}
	if err := s.adminActions.Save(ctx, action); err != nil {
		return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}

	return s.adminAdUsageText(l), s.adminAdInlineKeyboard(l), nil
}

func (s *service) adminAdPayload(msg domain.IncomingMessage) (string, []string) {
	text := strings.TrimSpace(msg.Arguments)
	if text == "" {
		text = strings.TrimSpace(msg.Text)
	}
	if text == msg.Command || text == "/"+msg.Command {
		text = ""
	}

	photoIDs := make([]string, 0, len(msg.PhotoIDs))
	for _, id := range msg.PhotoIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		photoIDs = append(photoIDs, id)
	}

	return text, photoIDs
}

func (s *service) startAdBroadcastWorker() {
	if s == nil || s.broadcastSender == nil || s.admin == nil {
		return
	}
	if s.adBroadcastQueue != nil {
		return
	}
	if adminAdBroadcastQueueSize <= 0 {
		return
	}

	s.adBroadcastQueue = make(chan adBroadcastJob, adminAdBroadcastQueueSize)
	go s.adBroadcastWorker()
}

func (s *service) enqueueAdBroadcast(job adBroadcastJob) bool {
	if s == nil || s.adBroadcastQueue == nil {
		return false
	}

	select {
	case s.adBroadcastQueue <- job:
		return true
	default:
		return false
	}
}

func (s *service) adBroadcastWorker() {
	for job := range s.adBroadcastQueue {
		s.processAdBroadcast(job)
	}
}

func (s *service) processAdBroadcast(job adBroadcastJob) {
	ctx := context.Background()
	result := s.broadcastAd(ctx, job)
	s.sendAdBroadcastSummary(ctx, job, result)
}

func (s *service) broadcastAd(ctx context.Context, job adBroadcastJob) adBroadcastResult {
	result := adBroadcastResult{}
	if s == nil || s.admin == nil || s.broadcastSender == nil {
		result.Err = errors.New("ad broadcast is not configured")
		return result
	}

	offset := 0
	for {
		list, err := s.admin.ListUsers(ctx, adminAdBroadcastPageSize, offset, false, false, false, false)
		if err != nil {
			result.Err = err
			return result
		}
		if len(list.Users) == 0 {
			break
		}

		for _, user := range list.Users {
			if user.UserID <= 0 || user.IsBanned {
				result.Skipped++
				continue
			}
			result.Total++
			if err := s.sendAdToUser(ctx, user.UserID, job.Text, job.PhotoIDs); err != nil {
				result.Failed++
			} else {
				result.Sent++
			}
			if adminAdBroadcastDelay > 0 {
				time.Sleep(adminAdBroadcastDelay)
			}
		}

		offset += list.Limit
		if list.Limit <= 0 {
			break
		}
		if list.Total > 0 && offset >= list.Total {
			break
		}
	}

	return result
}

func (s *service) sendAdToUser(ctx context.Context, chatID int64, text string, photoIDs []string) error {
	if s == nil || s.broadcastSender == nil || chatID == 0 {
		return nil
	}

	text = strings.TrimSpace(text)
	photoID := ""
	for _, id := range photoIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		photoID = id
		break
	}

	if photoID != "" {
		photoMessage := domain.OutgoingMessage{
			ChatID:   chatID,
			Text:     text,
			PhotoIDs: []string{photoID},
		}
		if text != "" && len([]rune(text)) > maxAdCaptionLength {
			photoMessage.Text = ""
		}
		if _, err := s.broadcastSender.Send(ctx, photoMessage); err != nil {
			return err
		}
		if text == "" || len([]rune(text)) <= maxAdCaptionLength {
			return nil
		}

		_, err := s.broadcastSender.Send(ctx, domain.OutgoingMessage{
			ChatID: chatID,
			Text:   text,
		})
		return err
	}

	if text == "" {
		return nil
	}
	_, err := s.broadcastSender.Send(ctx, domain.OutgoingMessage{
		ChatID: chatID,
		Text:   text,
	})
	return err
}

func (s *service) sendAdBroadcastSummary(ctx context.Context, job adBroadcastJob, result adBroadcastResult) {
	if s == nil || s.broadcastSender == nil || job.AdminChatID == 0 {
		return
	}

	l := s.localizerForUser(ctx, job.AdminID, domain.DefaultLanguage)
	summary := s.adminAdSummaryText(l, result.Total, result.Sent, result.Failed, result.Skipped)
	if result.Err != nil {
		summary = s.adminAdSummaryFailedText(l, result.Total, result.Sent, result.Failed, result.Skipped)
	}

	_, _ = s.broadcastSender.Send(ctx, domain.OutgoingMessage{
		ChatID: job.AdminChatID,
		Text:   summary,
	})
}
