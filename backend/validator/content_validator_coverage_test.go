package validator_test

import (
	"testing"
	"time"

	"coffee-spa/entity"
	"coffee-spa/usecase"
	"coffee-spa/validator"
)

func TestSourceItemBeanRecipeValidators_Coverage(t *testing.T) {
	sourceVal := validator.NewSourceValidator()
	itemVal := validator.NewItemValidator()
	beanVal := validator.NewBeanValidator()
	recipeVal := validator.NewRecipeValidator()

	tests := []struct {
		name string
		fn   func() error
		want bool
	}{
		{name: "V-SRC-ADD-01 source create ok", fn: func() error {
			return sourceVal.Create(usecase.CreateSourceIn{Name: "shop", SiteURL: "https://example.com"})
		}, want: false},
		{name: "V-SRC-ADD-02 source bad url", fn: func() error { return sourceVal.Create(usecase.CreateSourceIn{Name: "shop", SiteURL: "bad"}) }, want: true},
		{name: "V-SRC-ADD-03 source list bad limit", fn: func() error { return sourceVal.List(0, 0) }, want: true},
		{name: "V-ITEM-ADD-01 item create ok", fn: func() error { return itemVal.Create(validCreateItemIn()) }, want: false},
		{name: "V-ITEM-ADD-02 item bad kind", fn: func() error { in := validCreateItemIn(); in.Kind = entity.ItemKind("bad"); return itemVal.Create(in) }, want: true},
		{name: "V-ITEM-ADD-03 item bad source id", fn: func() error { in := validCreateItemIn(); in.SourceID = 0; return itemVal.Create(in) }, want: true},
		{name: "V-BEAN-ADD-01 bean create ok", fn: func() error { return beanVal.Create(validCreateBeanIn()) }, want: false},
		{name: "V-BEAN-ADD-02 bean bad score", fn: func() error { in := validCreateBeanIn(); in.Flavor = 6; return beanVal.Create(in) }, want: true},
		{name: "V-BEAN-ADD-03 bean bad roast", fn: func() error { in := validCreateBeanIn(); in.Roast = entity.Roast("bad"); return beanVal.Create(in) }, want: true},
		{name: "V-REC-ADD-01 recipe create ok", fn: func() error { return recipeVal.Create(validCreateRecipeIn()) }, want: false},
		{name: "V-REC-ADD-02 recipe bad temp", fn: func() error { in := validCreateRecipeIn(); in.Temp = 101; return recipeVal.Create(in) }, want: true},
		{name: "V-REC-ADD-03 recipe bad method", fn: func() error {
			in := validCreateRecipeIn()
			in.Method = entity.Method("bad")
			return recipeVal.Create(in)
		}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if (err != nil) != tt.want {
				t.Fatalf("error presence = %v, want %v, err=%v", err != nil, tt.want, err)
			}
		})
	}
}

func validCreateItemIn() usecase.CreateItemIn {
	return usecase.CreateItemIn{Title: "title", Summary: "summary", URL: "https://example.com/item", ImageURL: "https://example.com/image.jpg", Kind: entity.ItemKindNews, SourceID: 1, PublishedAt: time.Now()}
}

func validCreateBeanIn() usecase.CreateBeanIn {
	return usecase.CreateBeanIn{Name: "bean", Roast: entity.RoastMedium, Origin: "Ethiopia", Flavor: 3, Acidity: 3, Bitterness: 3, Body: 3, Aroma: 3, Desc: "desc", BuyURL: "https://example.com/bean", Active: true}
}

func validCreateRecipeIn() usecase.CreateRecipeIn {
	return usecase.CreateRecipeIn{BeanID: 1, Name: "drip", Method: entity.MethodDrip, TempPref: entity.TempHot, Grind: "medium", Ratio: "1:15", Temp: 90, TimeSec: 180, Steps: []string{"pour"}, Desc: "desc", Active: true}
}
