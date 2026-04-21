package skills

import (
	"github.com/mssantosdev/norn/internal/docs"
	"github.com/mssantosdev/norn/internal/norn"
)

func Save(root string, doc norn.Document) error   { return docs.Save(root, doc) }
func Load(root, id string) (norn.Document, error) { return docs.Load(root, id) }
func List(root string) ([]norn.Document, error)   { return docs.List(root) }
func Delete(root, id string) error                { return docs.Delete(root, id) }
