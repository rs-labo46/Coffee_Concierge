package valmock

import (
	"coffee-spa/entity"
	"coffee-spa/usecase"
)

type Auth struct{ Err error }

func (m Auth) Signup(email string, pw string) error { return m.Err }
func (m Auth) Login(email string, pw string) error  { return m.Err }
func (m Auth) Email(email string) error             { return m.Err }
func (m Auth) NewPw(pw string) error                { return m.Err }
func (m Auth) Token(token string) error             { return m.Err }

type Search struct{ Err error }

func (m Search) StartSession(in usecase.StartSessionIn) error { return m.Err }
func (m Search) SetPref(in usecase.SetPrefIn) error           { return m.Err }
func (m Search) AddTurn(in usecase.AddTurnIn) error           { return m.Err }
func (m Search) PatchPref(in usecase.PatchPrefIn) error       { return m.Err }
func (m Search) GetSession(in usecase.GetSessionIn) error     { return m.Err }
func (m Search) ListHistory(in usecase.ListHistoryIn) error   { return m.Err }
func (m Search) CloseSession(in usecase.CloseSessionIn) error { return m.Err }

type Saved struct{ Err error }

func (m Saved) Save(in usecase.SaveSuggestionIn) error { return m.Err }
func (m Saved) List(in usecase.ListSavedIn) error      { return m.Err }
func (m Saved) Delete(in usecase.DeleteSavedIn) error  { return m.Err }

type Item struct{ Err error }

func (m Item) Create(in usecase.CreateItemIn) error { return m.Err }
func (m Item) Get(id uint) error                    { return m.Err }
func (m Item) List(q entity.ItemQ) error            { return m.Err }
func (m Item) Top(limit int) error                  { return m.Err }

type Bean struct{ Err error }

func (m Bean) Create(in usecase.CreateBeanIn) error { return m.Err }
func (m Bean) Update(in usecase.UpdateBeanIn) error { return m.Err }
func (m Bean) Get(id uint) error                    { return m.Err }
func (m Bean) List(in usecase.BeanListIn) error     { return m.Err }

type Recipe struct{ Err error }

func (m Recipe) Create(in usecase.CreateRecipeIn) error { return m.Err }
func (m Recipe) Update(in usecase.UpdateRecipeIn) error { return m.Err }
func (m Recipe) Get(id uint) error                      { return m.Err }
func (m Recipe) List(in usecase.RecipeListIn) error     { return m.Err }

type Audit struct{ Err error }

func (m Audit) List(in usecase.AuditListIn) error { return m.Err }
