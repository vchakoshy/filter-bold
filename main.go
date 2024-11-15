package filterbold

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

// FilterBold Gorm Gin simple list api
// Usage :
//
//	fb := NewFilterBold(c).
//		Model(&models.UserBill{}).
//		Filters("user_id").
//		Find(&b)
type FilterBold struct {
	apiPrefix  string
	db         *gorm.DB
	c          *gin.Context
	filters    []string
	wheres     map[string]interface{}
	likes      []string
	wheresRaw  map[string]interface{}
	preloads   []string
	joins      []string
	fieldAlias map[string]string
	model      interface{}
	Result     interface{}
	Offset     int
	Limit      int
	NextURL    string
	order      string
	RowsCount  int64
}

func NewFilterBold(c *gin.Context, db *gorm.DB) *FilterBold {
	offset := cast.ToInt(c.DefaultQuery("offset", "0"))
	limit := cast.ToInt(c.DefaultQuery("limit", "10"))

	f := &FilterBold{
		apiPrefix:  GetOrDefault("API_PREFIX", "http://127.0.0.1:8080"),
		db:         db,
		Limit:      limit,
		Offset:     offset,
		wheres:     make(map[string]interface{}, 0),
		wheresRaw:  make(map[string]interface{}, 0),
		fieldAlias: make(map[string]string, 0),
		c:          c,
		order:      "id DESC",
	}
	f.OrderAuto()
	return f
}

func (f *FilterBold) Preloads(m ...string) *FilterBold {
	f.preloads = append(f.preloads, m...)

	return f
}

func (f *FilterBold) Where(k string, v interface{}) *FilterBold {
	f.wheres[k] = v
	return f
}

func (f *FilterBold) Like(k string) *FilterBold {
	f.likes = append(f.likes, k)
	return f
}

func (f *FilterBold) Likes(k ...string) *FilterBold {
	f.likes = append(f.likes, k...)
	return f
}

func (f *FilterBold) WhereRaw(k string, v interface{}) *FilterBold {
	f.wheresRaw[k] = v
	return f
}

func (f *FilterBold) Alias(k, v string) *FilterBold {
	f.fieldAlias[k] = v
	return f
}

func (f *FilterBold) Joins(m ...string) *FilterBold {
	f.joins = append(f.joins, m...)

	return f
}

func (f *FilterBold) Model(m interface{}) *FilterBold {
	f.model = m
	return f
}

func (f *FilterBold) OrderAuto() *FilterBold {
	o := f.c.DefaultQuery("order", "")
	if o != "" {
		f.Order(o)
	}
	return f
}

func (f *FilterBold) Order(o string) *FilterBold {
	f.order = o
	return f
}

func (f *FilterBold) Filters(i ...string) *FilterBold {
	f.filters = append(f.filters, i...)
	return f
}

func (f *FilterBold) Find(o interface{}) *FilterBold {
	q := f.db.Model(f.model)
	for _, filter := range f.filters {
		i := f.c.DefaultQuery(filter, "")
		if i != "" {
			qf := fmt.Sprintf("%s=?", filter)
			q = q.Where(qf, i)
		}
	}

	for k, v := range f.wheres {
		qf := fmt.Sprintf("%s=?", k)
		q = q.Where(qf, v)
	}

	for _, filter := range f.likes {
		i := f.c.DefaultQuery(filter, "")
		if i != "" {
			qf := fmt.Sprintf("%s like ?", filter)
			q = q.Where(qf, "%"+i+"%")
		}
	}

	for k, v := range f.wheresRaw {
		q = q.Where(k, v)
	}

	for k, filter := range f.fieldAlias {
		i := f.c.DefaultQuery(k, "")
		if i != "" {
			qf := fmt.Sprintf("%s=?", filter)
			q = q.Where(qf, i)
		}
	}

	for _, p := range f.preloads {
		q = q.Preload(p)
	}

	for _, p := range f.joins {
		q = q.Joins(p)
	}

	q.Count(&f.RowsCount)

	q.Order(f.order).
		Limit(f.Limit).
		Offset(f.Offset).
		Find(o)

	f.makeNextURL()

	return f
}

// ApplyAccessFilter
//
//	NewFilterBold(c, u.db).Model(&models.Task{}).
//	ApplyAccessFilter(func(db *gorm.DB) *gorm.DB {
//		role := session.GetUserRole(c)
//		uid := session.GetUserID(c)
//		if role == session.RoleUser {
//			return db.Where("user_id = ?", uid)
//		} else if role == session.RoleEditor {
//			return db.Where("department_id = ?", session.GetDepartmentID(c))
//		}
//		return db
//	}).
//	Preloads("Project", "Department", "User").
//	Likes("name").
//	Find(&l)
func (f *FilterBold) ApplyAccessFilter(whereFunc func(db *gorm.DB) *gorm.DB) *FilterBold {
	if whereFunc != nil {
		f.db = whereFunc(f.db)
	}

	return f
}

func (f *FilterBold) makeNextURL() *FilterBold {
	up := f.c.Request.URL
	uq := up.Query()
	uq.Set("offset", cast.ToString(f.Offset+f.Limit))

	for _, filter := range f.filters {
		i := f.c.DefaultQuery(filter, "")
		if i != "" {
			uq.Set(filter, i)
		}
	}
	up.RawQuery = uq.Encode()

	f.NextURL = f.apiPrefix + up.String()

	return f
}

func (f *FilterBold) GetMeta() gin.H {
	return gin.H{
		"next":        f.NextURL,
		"limit":       f.Limit,
		"total_count": f.RowsCount,
	}
}
