package bot

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"meethalf-telegram-bot/internal/domain"
)

const (
	adminAdBroadcastDelay  = 30 * time.Millisecond
	adminAdButtonPrefix    = "button:"
	adminAdButtonPrefixAlt = "btn:"
	adminAdButtonPrefixRu  = "кнопка:"
	adminAdButtonsPerRow   = 2
	adminAdButtonEmoji     = "🔗"
)

type adBroadcastJob struct {
	AdminID     int64
	AdminChatID int64
	Text        string
	PhotoIDs    []string
	Buttons     []domain.AdButton
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

	if s == nil || s.adminActions == nil {
		return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin action repository is not configured")
	}

	action, found, err := s.adminActions.Get(ctx, msg.User.ID)
	if err != nil {
		return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}
	wasButtonAction := found && action.Action == domain.AdminActionPostAdButton

	text, photoIDs, buttons, payloadErr := s.adminAdPayload(msg)
	if payloadErr != nil {
		_, hasPhoto, buttonCount := adminAdDraftMeta(action)
		return s.adminAdUsageText(l), s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0), payloadErr
	}
	if text != "" && len([]rune(text)) > maxAdTextLength {
		_, hasPhoto, buttonCount := adminAdDraftMeta(action)
		return s.adminAdTooLongText(l, maxAdTextLength), s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0), nil
	}
	if !found || !isAdminAdAction(action) {
		if text == "" && len(photoIDs) == 0 && len(buttons) == 0 {
			return s.startAdminAd(ctx, msg, role, l)
		}
		action = domain.AdminActionState{
			UserID:      msg.User.ID,
			ChatID:      msg.ChatID,
			Action:      domain.AdminActionPostAd,
			RequestedAt: s.now(msg.ReceivedAt),
		}
	} else {
		action.Action = domain.AdminActionPostAd
		if action.ChatID == 0 {
			action.ChatID = msg.ChatID
		}
	}

	updated := s.applyAdminAdDraftUpdate(&action, text, photoIDs, buttons)
	if updated || !found || wasButtonAction {
		action.RequestedAt = s.now(msg.ReceivedAt)
		if err := s.adminActions.Save(ctx, action); err != nil {
			return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), err
		}
	}

	hasText, hasPhoto, buttonCount := adminAdDraftMeta(action)
	if !hasText && !hasPhoto {
		return s.adminAdUsageText(l), s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0), nil
	}

	return s.adminAdDraftStatusText(l, hasText, hasPhoto, buttonCount), s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0), nil
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

	return s.adminAdUsageText(l), s.adminAdInlineKeyboard(l, false, false), nil
}

func (s *service) adminAdPayload(msg domain.IncomingMessage) (string, []string, []domain.AdButton, error) {
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

	if text == "" {
		return text, photoIDs, nil, nil
	}

	cleanText, buttons, err := extractAdButtons(text)
	if err != nil {
		return "", photoIDs, nil, err
	}

	return cleanText, photoIDs, buttons, nil
}

func (s *service) adminAdPreviewMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(l), nil, roleErr
	}
	if role != adminRoleAdmin {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg, l)
		return text, keyboard, errors.Join(roleErr, err)
	}

	action, found, err := s.loadAdminAdAction(ctx, msg.User.ID)
	if err != nil {
		return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}
	if !found {
		return s.adminAdUsageText(l), s.adminAdInlineKeyboard(l, false, false), nil
	}
	if action.Action == domain.AdminActionPostAdButton {
		action.Action = domain.AdminActionPostAd
		action.RequestedAt = s.now(msg.ReceivedAt)
		if err := s.adminActions.Save(ctx, action); err != nil {
			return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), err
		}
	}
	if action.Action == domain.AdminActionPostAdButton {
		action.Action = domain.AdminActionPostAd
		action.RequestedAt = s.now(msg.ReceivedAt)
		if err := s.adminActions.Save(ctx, action); err != nil {
			return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), err
		}
	}

	hasText, hasPhoto, buttonCount := adminAdDraftMeta(action)
	if !hasText && !hasPhoto {
		return s.adminAdUsageText(l), s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0), nil
	}

	photoIDs := []string{}
	if strings.TrimSpace(action.AdPhotoID) != "" {
		photoIDs = []string{action.AdPhotoID}
	}
	if err := s.sendAdToUser(ctx, msg.ChatID, action.AdText, photoIDs, action.AdButtons); err != nil {
		return s.adminAdFailedText(l), s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0), err
	}

	return s.adminAdPreviewSentText(l), s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0), nil
}

func (s *service) adminAdSendMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
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

	action, found, err := s.loadAdminAdAction(ctx, msg.User.ID)
	if err != nil {
		return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}
	if !found {
		return s.adminAdUsageText(l), s.adminAdInlineKeyboard(l, false, false), nil
	}

	hasText, hasPhoto, buttonCount := adminAdDraftMeta(action)
	if !hasText && !hasPhoto {
		return s.adminAdUsageText(l), s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0), nil
	}
	if len([]rune(action.AdText)) > maxAdTextLength {
		return s.adminAdTooLongText(l, maxAdTextLength), s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0), nil
	}

	if _, err := s.admin.CreateAd(ctx, action.AdText, action.AdPhotoID, action.AdButtons); err != nil {
		errorText := s.adminAdFailedText(l)
		var status statusError
		if errors.As(err, &status) {
			if status.StatusCode() == http.StatusBadRequest {
				errorText = s.adminAdUsageText(l)
			}
		}
		return errorText, s.adminMenuInlineKeyboard(l, role), err
	}

	photoIDs := []string{}
	if strings.TrimSpace(action.AdPhotoID) != "" {
		photoIDs = []string{action.AdPhotoID}
	}
	job := adBroadcastJob{
		AdminID:     msg.User.ID,
		AdminChatID: msg.ChatID,
		Text:        action.AdText,
		PhotoIDs:    photoIDs,
		Buttons:     action.AdButtons,
		RequestedAt: s.now(msg.ReceivedAt),
	}
	if !s.enqueueAdBroadcast(job) {
		return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), errors.New("ad broadcast queue is full")
	}

	_ = s.clearAdminAction(ctx, msg.User.ID)
	s.registerAdminActionCleanup(msg)
	return s.adminAdQueuedText(l), s.adminMenuInlineKeyboard(l, role), nil
}

func (s *service) adminAdAddButtonMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(l), nil, roleErr
	}
	if role != adminRoleAdmin {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg, l)
		return text, keyboard, errors.Join(roleErr, err)
	}
	if s == nil || s.adminActions == nil {
		return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin action repository is not configured")
	}

	action, found, err := s.loadAdminAdAction(ctx, msg.User.ID)
	if err != nil {
		return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}
	if !found {
		action = domain.AdminActionState{
			UserID:      msg.User.ID,
			ChatID:      msg.ChatID,
			Action:      domain.AdminActionPostAdButton,
			RequestedAt: s.now(msg.ReceivedAt),
		}
	} else {
		action.Action = domain.AdminActionPostAdButton
		if action.ChatID == 0 {
			action.ChatID = msg.ChatID
		}
		action.RequestedAt = s.now(msg.ReceivedAt)
	}
	if err := s.adminActions.Save(ctx, action); err != nil {
		return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}

	_, hasPhoto, buttonCount := adminAdDraftMeta(action)
	return s.adminAdButtonUsageText(l), s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0), nil
}

func (s *service) adminAdClearButtonsMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(l), nil, roleErr
	}
	if role != adminRoleAdmin {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg, l)
		return text, keyboard, errors.Join(roleErr, err)
	}
	if s == nil || s.adminActions == nil {
		return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin action repository is not configured")
	}

	action, found, err := s.loadAdminAdAction(ctx, msg.User.ID)
	if err != nil {
		return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}
	if !found {
		return s.adminAdUsageText(l), s.adminAdInlineKeyboard(l, false, false), nil
	}

	action.AdButtons = nil
	action.Action = domain.AdminActionPostAd
	action.RequestedAt = s.now(msg.ReceivedAt)
	if err := s.adminActions.Save(ctx, action); err != nil {
		return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}

	hasText, hasPhoto, buttonCount := adminAdDraftMeta(action)
	return s.adminAdDraftStatusText(l, hasText, hasPhoto, buttonCount), s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0), nil
}

func (s *service) adminAdRemovePhotoMessage(ctx context.Context, msg domain.IncomingMessage, l localizer) (string, *domain.InlineKeyboard, error) {
	role, roleErr := s.resolveAdminRole(ctx, msg.User)
	if roleErr != nil && isBannedError(roleErr) {
		return s.userBannedText(l), nil, roleErr
	}
	if role != adminRoleAdmin {
		text, keyboard, err := s.adminAccessDeniedMessage(ctx, msg, l)
		return text, keyboard, errors.Join(roleErr, err)
	}
	if s == nil || s.adminActions == nil {
		return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), errors.New("admin action repository is not configured")
	}

	action, found, err := s.loadAdminAdAction(ctx, msg.User.ID)
	if err != nil {
		return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}
	if !found {
		return s.adminAdUsageText(l), s.adminAdInlineKeyboard(l, false, false), nil
	}

	action.AdPhotoID = ""
	action.Action = domain.AdminActionPostAd
	action.RequestedAt = s.now(msg.ReceivedAt)
	if err := s.adminActions.Save(ctx, action); err != nil {
		return s.adminAdFailedText(l), s.adminMenuInlineKeyboard(l, role), err
	}

	hasText, hasPhoto, buttonCount := adminAdDraftMeta(action)
	return s.adminAdDraftStatusText(l, hasText, hasPhoto, buttonCount), s.adminAdInlineKeyboard(l, hasPhoto, buttonCount > 0), nil
}

func (s *service) loadAdminAdAction(ctx context.Context, userID int64) (domain.AdminActionState, bool, error) {
	if s == nil || s.adminActions == nil || userID == 0 {
		return domain.AdminActionState{}, false, errors.New("admin action repository is not configured")
	}

	action, found, err := s.adminActions.Get(ctx, userID)
	if err != nil {
		return domain.AdminActionState{}, false, err
	}
	if !found || !isAdminAdAction(action) {
		return domain.AdminActionState{}, false, nil
	}

	return action, true, nil
}

func isAdminAdAction(action domain.AdminActionState) bool {
	switch action.Action {
	case domain.AdminActionPostAd, domain.AdminActionPostAdButton:
		return true
	default:
		return false
	}
}

func adminAdDraftMeta(action domain.AdminActionState) (bool, bool, int) {
	hasText := strings.TrimSpace(action.AdText) != ""
	hasPhoto := strings.TrimSpace(action.AdPhotoID) != ""
	buttonCount := len(action.AdButtons)
	return hasText, hasPhoto, buttonCount
}

func (s *service) applyAdminAdDraftUpdate(action *domain.AdminActionState, text string, photoIDs []string, buttons []domain.AdButton) bool {
	if action == nil {
		return false
	}

	changed := false
	if strings.TrimSpace(text) != "" {
		action.AdText = strings.TrimSpace(text)
		changed = true
	}

	for _, id := range photoIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		action.AdPhotoID = id
		changed = true
		break
	}

	if len(buttons) > 0 {
		action.AdButtons = append(action.AdButtons, buttons...)
		changed = true
	}

	return changed
}

func extractAdButtons(text string) (string, []domain.AdButton, error) {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return strings.TrimSpace(text), nil, nil
	}

	outLines := make([]string, 0, len(lines))
	buttons := make([]domain.AdButton, 0)
	for _, line := range lines {
		button, isButton, err := parseAdButtonLine(line)
		if err != nil {
			return "", nil, err
		}
		if isButton {
			if button.Text != "" && button.URL != "" {
				buttons = append(buttons, button)
			}
			continue
		}
		outLines = append(outLines, line)
	}

	cleanText := strings.TrimSpace(strings.Join(outLines, "\n"))
	return cleanText, buttons, nil
}

func parseAdButtonLine(line string) (domain.AdButton, bool, error) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return domain.AdButton{}, false, nil
	}

	prefix, ok := adButtonPrefix(trimmed)
	if !ok {
		return domain.AdButton{}, false, nil
	}

	payload := strings.TrimSpace(trimmed[len(prefix):])
	if payload == "" {
		return domain.AdButton{}, true, errors.New("ad button payload is empty")
	}

	parts := strings.SplitN(payload, "|", 2)
	if len(parts) != 2 {
		return domain.AdButton{}, true, errors.New("ad button must contain label and url")
	}

	label := strings.TrimSpace(parts[0])
	link := strings.TrimSpace(parts[1])
	if label == "" || link == "" {
		return domain.AdButton{}, true, errors.New("ad button must contain label and url")
	}

	return domain.AdButton{Text: label, URL: link}, true, nil
}

func parseAdButtonInput(value string) (domain.AdButton, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return domain.AdButton{}, errors.New("ad button input is empty")
	}

	button, isButton, err := parseAdButtonLine(value)
	if err != nil {
		return domain.AdButton{}, err
	}
	if isButton {
		return button, nil
	}

	parts := strings.SplitN(value, "|", 2)
	if len(parts) != 2 {
		return domain.AdButton{}, errors.New("ad button must contain label and url")
	}
	label := strings.TrimSpace(parts[0])
	link := strings.TrimSpace(parts[1])
	if label == "" || link == "" {
		return domain.AdButton{}, errors.New("ad button must contain label and url")
	}

	return domain.AdButton{Text: label, URL: link}, nil
}

func adButtonPrefix(line string) (string, bool) {
	lower := strings.ToLower(strings.TrimSpace(line))
	switch {
	case strings.HasPrefix(lower, adminAdButtonPrefix):
		return adminAdButtonPrefix, true
	case strings.HasPrefix(lower, adminAdButtonPrefixAlt):
		return adminAdButtonPrefixAlt, true
	case strings.HasPrefix(lower, adminAdButtonPrefixRu):
		return adminAdButtonPrefixRu, true
	default:
		return "", false
	}
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
			if err := s.sendAdToUser(ctx, user.UserID, job.Text, job.PhotoIDs, job.Buttons); err != nil {
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

func (s *service) sendAdToUser(ctx context.Context, chatID int64, text string, photoIDs []string, buttons []domain.AdButton) error {
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

	keyboard := s.adInlineKeyboard(buttons)
	if photoID != "" {
		photoMessage := domain.OutgoingMessage{
			ChatID:   chatID,
			Text:     text,
			PhotoIDs: []string{photoID},
		}
		sendText := text != "" && len([]rune(text)) > maxAdCaptionLength
		if text != "" && len([]rune(text)) > maxAdCaptionLength {
			photoMessage.Text = ""
		}
		if !sendText && keyboard != nil {
			photoMessage.InlineKeyboard = keyboard
		}
		if _, err := s.broadcastSender.Send(ctx, photoMessage); err != nil {
			return err
		}
		if text == "" || len([]rune(text)) <= maxAdCaptionLength {
			return nil
		}

		_, err := s.broadcastSender.Send(ctx, domain.OutgoingMessage{
			ChatID:         chatID,
			Text:           text,
			InlineKeyboard: keyboard,
		})
		return err
	}

	if text == "" {
		return nil
	}
	_, err := s.broadcastSender.Send(ctx, domain.OutgoingMessage{
		ChatID:         chatID,
		Text:           text,
		InlineKeyboard: keyboard,
	})
	return err
}

func (s *service) adInlineKeyboard(buttons []domain.AdButton) *domain.InlineKeyboard {
	if len(buttons) == 0 {
		return nil
	}

	perRow := adminAdButtonsPerRow
	if perRow <= 0 {
		perRow = 1
	}

	rows := make([][]domain.InlineButton, 0, (len(buttons)+perRow-1)/perRow)
	row := make([]domain.InlineButton, 0, perRow)
	for _, button := range buttons {
		text := strings.TrimSpace(button.Text)
		link := strings.TrimSpace(button.URL)
		if text == "" || link == "" {
			continue
		}
		text = decorateAdminAdButtonText(text)
		row = append(row, domain.InlineButton{Text: text, URL: link})
		if len(row) >= perRow {
			rows = append(rows, row)
			row = make([]domain.InlineButton, 0, perRow)
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return nil
	}

	return &domain.InlineKeyboard{Buttons: rows}
}

func decorateAdminAdButtonText(text string) string {
	if text == "" {
		return text
	}
	if strings.HasPrefix(text, adminAdButtonEmoji+" ") || text == adminAdButtonEmoji {
		return text
	}
	return adminAdButtonEmoji + " " + text
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
