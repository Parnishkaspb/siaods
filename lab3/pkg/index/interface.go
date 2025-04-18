package index

import "lab3/pkg/file"

type Index interface {
	Create(filecsv file.DataImpl) bool
}
