package valmock

import "coffee-spa/usecase"

type Source struct{ Err error }

func (m Source) Create(in usecase.CreateSourceIn) error { return m.Err }
func (m Source) Get(id uint) error                      { return m.Err }
func (m Source) List(limit int, offset int) error       { return m.Err }
