package usecase

import (
	"encoding/json"
	"strconv"

	"coffee-spa/entity"
	"coffee-spa/repository"
)

// Beanの作成・更新・取得・一覧を扱う。
type BeanUC interface {
	Create(actor entity.Actor, in CreateBeanIn) (entity.Bean, error)
	Update(actor entity.Actor, in UpdateBeanIn) (entity.Bean, error)
	Get(id uint) (entity.Bean, error)
	List(in BeanListIn) ([]entity.Bean, error)
}

// Bean一覧入力。
type BeanListIn struct {
	Q      string
	Roast  entity.Roast
	Active *bool
	Limit  int
	Offset int
}

// Bean作成入力。
type CreateBeanIn struct {
	Name       string
	Roast      entity.Roast
	Origin     string
	Flavor     int
	Acidity    int
	Bitterness int
	Body       int
	Aroma      int
	Desc       string
	BuyURL     string
	Active     bool
}

// Bean更新入力。
type UpdateBeanIn struct {
	ID         uint
	Name       string
	Roast      entity.Roast
	Origin     string
	Flavor     int
	Acidity    int
	Bitterness int
	Body       int
	Aroma      int
	Desc       string
	BuyURL     string
	Active     bool
}

type BeanVal interface {
	Create(in CreateBeanIn) error
	Update(in UpdateBeanIn) error
	Get(id uint) error
	List(in BeanListIn) error
}

type beanUsecase struct {
	beans  repository.BeanRepository
	audits repository.AuditRepository
	val    BeanVal
}

func NewBeanUsecase(
	beans repository.BeanRepository,
	audits repository.AuditRepository,
	val BeanVal,
) BeanUC {
	return &beanUsecase{
		beans:  beans,
		audits: audits,
		val:    val,
	}
}

// Beanを新規作成する。
// adminのみ許可する。
func (u *beanUsecase) Create(actor entity.Actor, in CreateBeanIn) (entity.Bean, error) {
	if actor.Role != entity.RoleAdmin {
		return entity.Bean{}, repository.ErrForbidden
	}

	if err := u.val.Create(in); err != nil {
		return entity.Bean{}, err
	}

	bean := &entity.Bean{
		Name:       in.Name,
		Roast:      in.Roast,
		Origin:     in.Origin,
		Flavor:     in.Flavor,
		Acidity:    in.Acidity,
		Bitterness: in.Bitterness,
		Body:       in.Body,
		Aroma:      in.Aroma,
		Desc:       in.Desc,
		BuyURL:     in.BuyURL,
		Active:     in.Active,
	}

	if err := u.beans.Create(bean); err != nil {
		return entity.Bean{}, err
	}

	u.writeAudit(
		"admin.beans.create",
		&actor.UserID,
		map[string]string{
			"bean_id": uintToStr(bean.ID),
			"name":    bean.Name,
			"roast":   string(bean.Roast),
			"active":  boolToStr(bean.Active),
		},
	)

	return *bean, nil
}

// Beanを更新する。
// adminのみ許可する。
func (u *beanUsecase) Update(actor entity.Actor, in UpdateBeanIn) (entity.Bean, error) {
	if actor.Role != entity.RoleAdmin {
		return entity.Bean{}, repository.ErrForbidden
	}

	if err := u.val.Update(in); err != nil {
		return entity.Bean{}, err
	}

	bean, err := u.beans.GetByID(in.ID)
	if err != nil {
		return entity.Bean{}, err
	}

	bean.Name = in.Name
	bean.Roast = in.Roast
	bean.Origin = in.Origin
	bean.Flavor = in.Flavor
	bean.Acidity = in.Acidity
	bean.Bitterness = in.Bitterness
	bean.Body = in.Body
	bean.Aroma = in.Aroma
	bean.Desc = in.Desc
	bean.BuyURL = in.BuyURL
	bean.Active = in.Active

	if err := u.beans.Update(bean); err != nil {
		return entity.Bean{}, err
	}

	u.writeAudit(
		"admin.beans.update",
		&actor.UserID,
		map[string]string{
			"bean_id": uintToStr(bean.ID),
			"name":    bean.Name,
			"roast":   string(bean.Roast),
			"active":  boolToStr(bean.Active),
		},
	)

	return *bean, nil
}

// Beanを1件取得する。
func (u *beanUsecase) Get(id uint) (entity.Bean, error) {
	if err := u.val.Get(id); err != nil {
		return entity.Bean{}, err
	}

	bean, err := u.beans.GetByID(id)
	if err != nil {
		return entity.Bean{}, err
	}

	return *bean, nil
}

// Bean一覧を返す。
func (u *beanUsecase) List(in BeanListIn) ([]entity.Bean, error) {
	if err := u.val.List(in); err != nil {
		return nil, err
	}

	out, err := u.beans.List(repository.BeanListQ{
		Q:      in.Q,
		Roast:  in.Roast,
		Active: in.Active,
		Limit:  in.Limit,
		Offset: in.Offset,
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (u *beanUsecase) writeAudit(
	typ string,
	userID *uint,
	meta map[string]string,
) {
	if u.audits == nil {
		return
	}

	raw, err := json.Marshal(meta)
	if err != nil {
		raw = []byte(`{}`)
	}

	_ = u.audits.Create(&entity.AuditLog{
		Type:   typ,
		UserID: userID,
		IP:     "",
		UA:     "",
		Meta:   raw,
	})
}

func boolToStr(v bool) string {
	return strconv.FormatBool(v)
}
