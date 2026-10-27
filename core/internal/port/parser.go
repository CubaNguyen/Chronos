// internal/port/parser.go
package port

import "context"

// 1. Định nghĩa Enum (Dùng string cho dễ đọc log thay vì dùng int32 của protobuf)
type ClassKind string

const (
	ClassKindUnspecified ClassKind = "UNSPECIFIED"
	ClassKindClass       ClassKind = "CLASS"
	ClassKindStruct      ClassKind = "STRUCT"
	ClassKindInterface   ClassKind = "INTERFACE"
	ClassKindEnum        ClassKind = "ENUM"
	ClassKindTrait       ClassKind = "TRAIT"
)

type Visibility string

const (
	VisibilityUnspecified Visibility = "UNSPECIFIED"
	VisibilityPublic      Visibility = "PUBLIC"
	VisibilityPrivate     Visibility = "PRIVATE"
	VisibilityProtected   Visibility = "PROTECTED"
	VisibilityPackage     Visibility = "PACKAGE"
)

// 2. Định nghĩa các Domain Struct (Không hề có tag json hay protobuf nào)
type Ast struct {
	File      FileInfo
	Classes   []Class
	Functions []Function
	Fields    []Field
	Calls     []Call
	Imports   []Import
}

type FileInfo struct {
	Path      string
	Language  string
	Hash      string
	SizeBytes int64
	LineCount int32
}

type Class struct {
	Name       string
	Kind       ClassKind
	Visibility Visibility
	Extends    *string // Dùng con trỏ (*string) vì trong proto là optional
	Implements []string
	LineStart  int32
	LineEnd    int32
	FilePath   string
}

type Function struct {
	Name       string
	ClassName  *string // Dùng con trỏ vì có thể null (free function)
	Params     []Param
	ReturnType *string
	Visibility Visibility
	LineStart  int32
	LineEnd    int32
	FilePath   string
}

type Param struct {
	Name     string
	TypeName string
}

type Field struct {
	Name       string
	TypeName   string
	Visibility Visibility
	ClassName  *string
	Line       int32
	FilePath   string
}

type Call struct {
	Caller   string
	Callee   string
	Line     int32
	FilePath string
}

type Import struct {
	ModulePath string
	Alias      *string
	Line       int32
	IsExternal bool
	FilePath   string
}

type Parser interface {
	ParseFile(ctx context.Context, path string, source string) (*Ast, error)
}
