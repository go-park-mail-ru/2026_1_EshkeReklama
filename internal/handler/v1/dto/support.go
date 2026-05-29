package dto

import (
	"time"

	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

type SupportThreadResponse struct {
	ID         int `json:"id"`
	CampaignID int `json:"campaign_id"`
}

type GetOrCreateSupportThreadResponse struct {
	Thread *SupportThreadResponse `json:"thread"`
}

type SupportMessageResponse struct {
	ID        int       `json:"id"`
	ThreadID  int       `json:"thread_id"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type ListSupportMessagesResponse struct {
	Messages []*SupportMessageResponse `json:"messages"`
}

type CreateSupportMessageRequest struct {
	Text string `json:"text" validate:"required,max=2000"`
}

type CreateSupportMessageResponse struct {
	Message *SupportMessageResponse `json:"message"`
}

type AdminSupportThreadResponse struct {
	ID           int                     `json:"id"`
	CampaignID   int                     `json:"campaign_id"`
	AdvertiserID int                     `json:"advertiser_id"`
	UpdatedAt    time.Time               `json:"updated_at"`
	LastMessage  *SupportMessageResponse `json:"last_message,omitempty"`
}

type ListAdminSupportThreadsResponse struct {
	Threads []*AdminSupportThreadResponse `json:"threads"`
}

func (r *CreateSupportMessageRequest) ToInput(threadID int) *serviceinput.CreateSupportMessage {
	return &serviceinput.CreateSupportMessage{
		ThreadID: threadID,
		Text:     r.Text,
	}
}

func ToSupportThreadResponse(thread *models.SupportThread) *SupportThreadResponse {
	if thread == nil {
		return nil
	}
	return &SupportThreadResponse{
		ID:         thread.ID,
		CampaignID: thread.CampaignID,
	}
}

func ToSupportMessageResponse(msg *models.SupportMessage) *SupportMessageResponse {
	if msg == nil {
		return nil
	}
	return &SupportMessageResponse{
		ID:        msg.ID,
		ThreadID:  msg.ThreadID,
		Author:    string(msg.AuthorType),
		Text:      msg.Text,
		CreatedAt: msg.CreatedAt,
	}
}

func ToSupportMessagesResponse(messages []*models.SupportMessage) []*SupportMessageResponse {
	out := make([]*SupportMessageResponse, 0, len(messages))
	for _, msg := range messages {
		out = append(out, ToSupportMessageResponse(msg))
	}
	return out
}

func ToAdminSupportThreadsResponse(threads []*models.SupportThreadSummary) []*AdminSupportThreadResponse {
	out := make([]*AdminSupportThreadResponse, 0, len(threads))
	for _, thread := range threads {
		if thread == nil || thread.Thread == nil {
			continue
		}
		out = append(out, &AdminSupportThreadResponse{
			ID:           thread.Thread.ID,
			CampaignID:   thread.Thread.CampaignID,
			AdvertiserID: thread.Thread.AdvertiserID,
			UpdatedAt:    thread.LastActivity,
			LastMessage:  ToSupportMessageResponse(thread.LastMessage),
		})
	}
	return out
}
