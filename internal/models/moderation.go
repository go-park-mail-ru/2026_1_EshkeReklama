package models

// ModerationQueueItem — объявление в очереди модерации с метаданными тарифа.
type ModerationQueueItem struct {
	Ad                 *Ad
	PriorityModeration bool
}
