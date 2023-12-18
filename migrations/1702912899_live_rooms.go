package migrations

import (
	"fmt"

	"github.com/docker/go-units"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
)

func init() {
	migrations.Register(func(db dbx.Builder) (err error) {
		defer err0.Then(&err, nil, nil)
		dao := daos.New(db)
		lrooms := &models.Collection{
			BaseModel: models.BaseModel{Id: "live_rooms"},
			Name:      "live_rooms",
			System:    true,
			Type:      models.CollectionTypeBase,
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Id:     "name",
					Name:   "name",
					System: true,
					Type:   schema.FieldTypeText,
				},
				&schema.SchemaField{
					Id:     "running",
					Name:   "running",
					System: true,
					Type:   schema.FieldTypeBool,
				},
				&schema.SchemaField{
					Id:     "push",
					Name:   "push",
					System: true,
					Type:   schema.FieldTypeBool,
				},
				&schema.SchemaField{
					Id:       "video",
					Name:     "video",
					System:   true,
					Required: true,
					Type:     schema.FieldTypeFile,
					Options: schema.FileOptions{
						MaxSelect: 1,
						MaxSize:   1 * units.PiB,
						MimeTypes: []string{"video/mp4"},
					},
				},
				&schema.SchemaField{
					Id:       "rtmps",
					Name:     "rtmps",
					System:   true,
					Required: true,
					Type:     schema.FieldTypeText,
				},
				&schema.SchemaField{
					Id:      "offset",
					Name:    "offset",
					System:  true,
					Type:    schema.FieldTypeNumber,
					Options: schema.NumberOptions{NoDecimal: true},
				},
				&schema.SchemaField{
					Id:      "duration",
					Name:    "duration",
					System:  true,
					Type:    schema.FieldTypeNumber,
					Options: schema.NumberOptions{NoDecimal: true},
				},
			),
		}
		try.To(dao.SaveCollection(lrooms))
		return nil
	}, func(db dbx.Builder) error {
		return fmt.Errorf("add table live_rooms rollback todo")
	})
}
