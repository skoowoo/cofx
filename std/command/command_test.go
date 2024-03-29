package command

import (
	"context"
	"os"
	"testing"

	"github.com/skoowoo/cofx/functiondriver/go/spec"
	"github.com/skoowoo/cofx/pkg/sqlite"
	"github.com/skoowoo/cofx/service/resource"
	"github.com/skoowoo/cofx/service/resource/db"
	"github.com/skoowoo/cofx/service/resource/labels"
	"github.com/stretchr/testify/assert"
)

func TestCommandFunction(t *testing.T) {
	ls := labels.Labels{
		"flow_id":   "1234567890",
		"node_seq":  "1000",
		"node_name": "command",
	}
	mdb, err := sqlite.NewMemDB()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	tbl, err := mdb.CreateTable(context.Background(), db.StatementCreateOutputParsingTable)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	mf, ep, _ := New()
	assert.Equal(t, "go", mf.Driver)
	bundle := spec.EntrypointBundle{
		Version: "latest",
		Resources: resource.Resources{
			Logwriter:    os.Stdout,
			Labels:       ls,
			OutputParser: &tbl,
		},
	}
	returns, err := ep(context.Background(), bundle, map[string]string{
		"cmd":            "echo hello cofx ... && sleep 2 && echo hello world !!!",
		"split":          "",
		"extract_fields": "0,1",
		"query_columns":  "c0,c1",
		"query_where":    "c1 = 'world'",
	})
	assert.Len(t, returns, 1)
	assert.Equal(t, "hello world", returns["outcome_0"])
	assert.NoError(t, err)

	rows, err := bundle.Resources.OutputParser.Query(context.Background(), []string{"c0", "c1"}, "")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Len(t, rows, 0)
}
