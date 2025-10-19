//! Хранилище пользовательского состояния

package userstates

import (
	"ai_tg_search/struct_types/newstypes"
	"sync"
)

// Структуры
// Хранение сессии новостей пользователя
type UserNewsSession struct {
	News         *newstypes.NewsResponse
	CurrentIndex int
}

// Диалог пользователя с ботом
var (
	UserState   = make(map[int64]string)
	UserStateMu sync.RWMutex
)

// Дата пользователя
var (
	UserData   = make(map[int64]string)
	UserDataMu sync.RWMutex
)

// История пользователя
var (
	UserHistory = make(map[int64]*UserNewsSession)
	UserSearch  = make(map[int64]*UserNewsSession)
	UserNewsMu  sync.RWMutex
)
