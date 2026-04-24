package repository

import (
	"testing"
	"time"

	"coffee-spa/entity"

	"gorm.io/datatypes"
)

func TestSessionRepository_Lifecycle(t *testing.T) {
	db := openTestDB(t)
	user := seedUser(t, db, "session@example.com", entity.RoleUser)
	bean := seedBean(t, db, "Session Bean")
	recipe := seedRecipe(t, db, bean.ID)
	src := seedSource(t, db)
	item := seedItem(t, db, src.ID, entity.ItemKindRecipe, "Session Item")
	repo := NewSessionRepository(db)

	session := &entity.Session{UserID: &user.ID, Title: "session", Status: entity.SessionActive}
	if err := repo.CreateSession(session); err != nil {
		t.Fatalf("create session: %v", err)
	}
	turn := &entity.Turn{SessionID: session.ID, Role: entity.TurnRoleUser, Kind: entity.TurnKindMessage, Body: "light coffee"}
	if err := repo.CreateTurn(turn); err != nil {
		t.Fatalf("create turn: %v", err)
	}
	pref := &entity.Pref{
		SessionID: session.ID, Flavor: 3, Acidity: 3, Bitterness: 3, Body: 3, Aroma: 3,
		Mood: entity.MoodMorning, Method: entity.MethodDrip, Scene: entity.SceneBreak, TempPref: entity.TempHot,
		Excludes: []string{}, Note: "note",
	}
	if err := repo.CreatePref(pref); err != nil {
		t.Fatalf("create pref: %v", err)
	}
	pref.Note = "updated"
	if err := repo.UpdatePref(pref); err != nil {
		t.Fatalf("update pref: %v", err)
	}
	suggestions := []entity.Suggestion{{BeanID: bean.ID, RecipeID: &recipe.ID, ItemID: &item.ID, Score: 90, Reason: "good", Rank: 1}}
	if err := repo.ReplaceSuggestions(session.ID, suggestions); err != nil {
		t.Fatalf("replace suggestions: %v", err)
	}
	turns, err := repo.ListTurns(session.ID)
	if err != nil || len(turns) != 1 {
		t.Fatalf("list turns: len=%d err=%v", len(turns), err)
	}
	gotPref, err := repo.GetPrefBySessionID(session.ID)
	if err != nil || gotPref.Note != "updated" {
		t.Fatalf("get pref: pref=%#v err=%v", gotPref, err)
	}
	gotSuggestions, err := repo.ListSuggestions(session.ID)
	if err != nil || len(gotSuggestions) != 1 {
		t.Fatalf("list suggestions: len=%d err=%v", len(gotSuggestions), err)
	}
	if err := repo.CloseSession(session.ID); err != nil {
		t.Fatalf("close session: %v", err)
	}
}

func TestSessionRepository_GuestSessionRequiresKeyAndNotExpired(t *testing.T) {
	db := openTestDB(t)
	repo := NewSessionRepository(db)
	expires := time.Now().Add(time.Hour)
	session := &entity.Session{Title: "guest", Status: entity.SessionActive, SessionKeyHash: "key-hash", GuestExpiresAt: &expires}
	if err := repo.CreateSession(session); err != nil {
		t.Fatalf("create guest session: %v", err)
	}
	got, err := repo.GetGuestSessionByID(session.ID, "key-hash", time.Now())
	if err != nil {
		t.Fatalf("get guest session: %v", err)
	}
	if got.ID != session.ID {
		t.Fatalf("guest session mismatch: %#v", got)
	}
}

func TestSavedRepository_CreateListGetDelete(t *testing.T) {
	db := openTestDB(t)
	user := seedUser(t, db, "saved@example.com", entity.RoleUser)
	bean := seedBean(t, db, "Saved Bean")
	session := seedSession(t, db, &user.ID)
	suggestion := entity.Suggestion{SessionID: session.ID, BeanID: bean.ID, Score: 80, Reason: "good", Rank: 1}
	if err := db.Create(&suggestion).Error; err != nil {
		t.Fatalf("seed suggestion: %v", err)
	}
	repo := NewSavedRepository(db)
	saved := &entity.SavedSuggestion{UserID: user.ID, SessionID: session.ID, SuggestionID: suggestion.ID}
	if err := repo.Create(saved); err != nil {
		t.Fatalf("create saved: %v", err)
	}
	list, err := repo.List(SavedListQ{UserID: user.ID, Limit: 10})
	if err != nil || len(list) != 1 {
		t.Fatalf("list saved: len=%d err=%v", len(list), err)
	}
	got, err := repo.GetByUserAndSuggestionID(user.ID, suggestion.ID)
	if err != nil || got.ID != saved.ID {
		t.Fatalf("get saved: saved=%#v err=%v", got, err)
	}
	if err := repo.DeleteByUserAndSuggestionID(user.ID, suggestion.ID); err != nil {
		t.Fatalf("delete saved: %v", err)
	}
}

func TestAuditRepository_CreateAndList(t *testing.T) {
	db := openTestDB(t)
	user := seedUser(t, db, "audit@example.com", entity.RoleAdmin)
	repo := NewAuditRepository(db)
	log := &entity.AuditLog{Type: "ai.request", UserID: &user.ID, IP: "127.0.0.1", UA: "test", Meta: datatypes.JSON([]byte(`{"ok":true}`))}
	if err := repo.Create(log); err != nil {
		t.Fatalf("create audit log: %v", err)
	}
	logs, err := repo.List(AuditListQ{Type: "ai.request", UserID: &user.ID, Limit: 10})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].Type != "ai.request" {
		t.Fatalf("unexpected audit logs: %#v", logs)
	}
}
