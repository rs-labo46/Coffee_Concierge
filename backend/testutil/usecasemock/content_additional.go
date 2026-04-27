package usecasemock

import (
	"coffee-spa/entity"
	"coffee-spa/usecase"
)

type Source struct {
	CreateFn func(entity.Actor, usecase.CreateSourceIn) (entity.Source, error)
	GetFn    func(uint) (entity.Source, error)
	ListFn   func(int, int) ([]entity.Source, error)
}

func (m *Source) Create(a entity.Actor, in usecase.CreateSourceIn) (entity.Source, error) {
	if m.CreateFn == nil {
		return entity.Source{ID: 1, Name: in.Name, SiteURL: in.SiteURL}, nil
	}
	return m.CreateFn(a, in)
}
func (m *Source) Get(id uint) (entity.Source, error) {
	if m.GetFn == nil {
		return entity.Source{ID: id, Name: "source"}, nil
	}
	return m.GetFn(id)
}
func (m *Source) List(limit int, offset int) ([]entity.Source, error) {
	if m.ListFn == nil {
		return []entity.Source{{ID: 1, Name: "source"}}, nil
	}
	return m.ListFn(limit, offset)
}

type Item struct {
	CreateFn func(entity.Actor, usecase.CreateItemIn) (entity.Item, error)
	GetFn    func(uint) (entity.Item, error)
	ListFn   func(entity.ItemQ) ([]entity.Item, error)
	TopFn    func(int) (entity.TopItems, error)
}

func (m *Item) Create(a entity.Actor, in usecase.CreateItemIn) (entity.Item, error) {
	if m.CreateFn == nil {
		return entity.Item{ID: 1, Title: in.Title, Kind: in.Kind, SourceID: in.SourceID}, nil
	}
	return m.CreateFn(a, in)
}
func (m *Item) Get(id uint) (entity.Item, error) {
	if m.GetFn == nil {
		return entity.Item{ID: id, Title: "item"}, nil
	}
	return m.GetFn(id)
}
func (m *Item) List(q entity.ItemQ) ([]entity.Item, error) {
	if m.ListFn == nil {
		return []entity.Item{{ID: 1, Title: "item", Kind: q.Kind}}, nil
	}
	return m.ListFn(q)
}
func (m *Item) Top(limit int) (entity.TopItems, error) {
	if m.TopFn == nil {
		return entity.TopItems{News: []entity.Item{{ID: 1, Kind: entity.ItemKindNews}}}, nil
	}
	return m.TopFn(limit)
}

type Bean struct {
	CreateFn func(entity.Actor, usecase.CreateBeanIn) (entity.Bean, error)
	UpdateFn func(entity.Actor, usecase.UpdateBeanIn) (entity.Bean, error)
	GetFn    func(uint) (entity.Bean, error)
	ListFn   func(usecase.BeanListIn) ([]entity.Bean, error)
}

func (m *Bean) Create(a entity.Actor, in usecase.CreateBeanIn) (entity.Bean, error) {
	if m.CreateFn == nil {
		return entity.Bean{ID: 1, Name: in.Name, Roast: in.Roast, Active: in.Active}, nil
	}
	return m.CreateFn(a, in)
}
func (m *Bean) Update(a entity.Actor, in usecase.UpdateBeanIn) (entity.Bean, error) {
	if m.UpdateFn == nil {
		return entity.Bean{ID: in.ID, Name: in.Name, Roast: in.Roast, Active: in.Active}, nil
	}
	return m.UpdateFn(a, in)
}
func (m *Bean) Get(id uint) (entity.Bean, error) {
	if m.GetFn == nil {
		return entity.Bean{ID: id, Name: "bean"}, nil
	}
	return m.GetFn(id)
}
func (m *Bean) List(in usecase.BeanListIn) ([]entity.Bean, error) {
	if m.ListFn == nil {
		return []entity.Bean{{ID: 1, Name: "bean"}}, nil
	}
	return m.ListFn(in)
}

type Recipe struct {
	CreateFn func(entity.Actor, usecase.CreateRecipeIn) (entity.Recipe, error)
	UpdateFn func(entity.Actor, usecase.UpdateRecipeIn) (entity.Recipe, error)
	GetFn    func(uint) (entity.Recipe, error)
	ListFn   func(usecase.RecipeListIn) ([]entity.Recipe, error)
}

func (m *Recipe) Create(a entity.Actor, in usecase.CreateRecipeIn) (entity.Recipe, error) {
	if m.CreateFn == nil {
		return entity.Recipe{ID: 1, BeanID: in.BeanID, Name: in.Name, Method: in.Method, TempPref: in.TempPref}, nil
	}
	return m.CreateFn(a, in)
}
func (m *Recipe) Update(a entity.Actor, in usecase.UpdateRecipeIn) (entity.Recipe, error) {
	if m.UpdateFn == nil {
		return entity.Recipe{ID: in.ID, BeanID: in.BeanID, Name: in.Name}, nil
	}
	return m.UpdateFn(a, in)
}
func (m *Recipe) Get(id uint) (entity.Recipe, error) {
	if m.GetFn == nil {
		return entity.Recipe{ID: id, Name: "recipe"}, nil
	}
	return m.GetFn(id)
}
func (m *Recipe) List(in usecase.RecipeListIn) ([]entity.Recipe, error) {
	if m.ListFn == nil {
		return []entity.Recipe{{ID: 1, Name: "recipe"}}, nil
	}
	return m.ListFn(in)
}

type Audit struct {
	ListFn func(entity.Actor, usecase.AuditListIn) ([]entity.AuditLog, error)
}

func (m *Audit) List(a entity.Actor, in usecase.AuditListIn) ([]entity.AuditLog, error) {
	if m.ListFn == nil {
		return []entity.AuditLog{{ID: 1, Type: in.Type}}, nil
	}
	return m.ListFn(a, in)
}
