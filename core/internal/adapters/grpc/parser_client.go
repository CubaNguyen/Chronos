package grpc

import (
	"context"
	"fmt"

	"github.com/cuba3/chronos/core/internal/port"
	pb "github.com/cuba3/chronos/core/pkg/contracts/parserv1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// parserClient là struct thực thi interface port.Parser
type parserClient struct {
	// client là thằng được protoc sinh ra, chuyên gọi mạng
	client pb.ParserServiceClient
}

// NewParserClient là hàm khởi tạo.
// Nó nhận vào "đường dây" (conn) và trả về một Interface (port.Parser)
func NewParserClient(conn *grpc.ClientConn) port.Parser {
	return &parserClient{
		client: pb.NewParserServiceClient(conn),
	}
}

// ParseFile là hàm thực thi bản hợp đồng
func (p *parserClient) ParseFile(ctx context.Context, path string, source string) (*port.Ast, error) {
	// 1. Chuẩn bị Request theo đúng chuẩn Protobuf
	req := &pb.ParseFileRequest{
		Path:   path,
		Source: source,
	}

	// 2. Gọi sang server Rust qua mạng
	resp, err := p.client.ParseFile(ctx, req)
	if err != nil {
		// Xử lý lỗi hệ thống (mạng đứt, server chết, timeout...)
		// Chuyển lỗi gRPC thành lỗi chuẩn của Go để tầng Core dễ đọc
		st, ok := status.FromError(err)
		if ok {
			return nil, fmt.Errorf("lỗi gRPC từ Rust (code: %s): %s", st.Code(), st.Message())
		}
		return nil, fmt.Errorf("lỗi không xác định khi gọi Rust: %w", err)
	}

	// 3. Kiểm tra lỗi nghiệp vụ (Syntax Error do Rust trả về trong Response)
	if resp.Error != "" {
		// Đây không phải lỗi hệ thống, mà là code bạn gửi lên bị sai cú pháp
		return nil, fmt.Errorf("parse thất bại: %s", resp.Error)
	}

	// 4. BẢN DỊCH (Mapping): Chuyển đổi từ Protobuf (pb) sang Domain (port)
	// Đây là bước quan trọng nhất để bảo vệ tầng Core
	domainAst := mapPbAstToPort(resp.Ast)

	return domainAst, nil
}

// =========================================================================
// CÁC HÀM HELPER ĐỂ MAPPING (Chống ăn mòn Domain)
// =========================================================================

func mapPbAstToPort(pbAst *pb.Ast) *port.Ast {
	if pbAst == nil {
		return nil
	}

	ast := &port.Ast{
		File: port.FileInfo{
			Path:      pbAst.File.Path,
			Language:  pbAst.File.Language,
			Hash:      pbAst.File.Hash,
			SizeBytes: pbAst.File.SizeBytes,
			LineCount: pbAst.File.LineCount,
		},
	}

	// Map Classes
	for _, c := range pbAst.Classes {
		ast.Classes = append(ast.Classes, port.Class{
			Name:       c.Name,
			Kind:       mapClassKind(c.Kind),
			Visibility: mapVisibility(c.Visibility),
			Extends:    c.Extends, // Cả 2 đều là *string nên gán trực tiếp được
			Implements: c.Implements,
			LineStart:  c.LineStart,
			LineEnd:    c.LineEnd,
			FilePath:   c.FilePath,
		})
	}

	// Map Functions
	for _, f := range pbAst.Functions {
		var params []port.Param
		for _, p := range f.Params {
			params = append(params, port.Param{
				Name:     p.Name,
				TypeName: p.TypeName,
			})
		}
		ast.Functions = append(ast.Functions, port.Function{
			Name:       f.Name,
			ClassName:  f.ClassName,
			Params:     params,
			ReturnType: f.ReturnType,
			Visibility: mapVisibility(f.Visibility),
			LineStart:  f.LineStart,
			LineEnd:    f.LineEnd,
			FilePath:   f.FilePath,
		})
	}

	// Map Fields
	for _, f := range pbAst.Fields {
		ast.Fields = append(ast.Fields, port.Field{
			Name:       f.Name,
			TypeName:   f.TypeName,
			Visibility: mapVisibility(f.Visibility),
			ClassName:  f.ClassName,
			Line:       f.Line,
			FilePath:   f.FilePath,
		})
	}

	// Map Calls
	for _, c := range pbAst.Calls {
		ast.Calls = append(ast.Calls, port.Call{
			Caller:   c.Caller,
			Callee:   c.Callee,
			Line:     c.Line,
			FilePath: c.FilePath,
		})
	}

	// Map Imports
	for _, i := range pbAst.Imports {
		ast.Imports = append(ast.Imports, port.Import{
			ModulePath: i.ModulePath,
			Alias:      i.Alias,
			Line:       i.Line,
			IsExternal: i.IsExternal,
			FilePath:   i.FilePath,
		})
	}

	return ast
}

// Hàm map Enum từ int32 (Protobuf) sang string (Domain)
func mapClassKind(kind pb.ClassKind) port.ClassKind {
	switch kind {
	case pb.ClassKind_CLASS_KIND_CLASS:
		return port.ClassKindClass
	case pb.ClassKind_CLASS_KIND_STRUCT:
		return port.ClassKindStruct
	case pb.ClassKind_CLASS_KIND_INTERFACE:
		return port.ClassKindInterface
	case pb.ClassKind_CLASS_KIND_ENUM:
		return port.ClassKindEnum
	case pb.ClassKind_CLASS_KIND_TRAIT:
		return port.ClassKindTrait
	default:
		return port.ClassKindUnspecified
	}
}

func mapVisibility(v pb.Visibility) port.Visibility {
	switch v {
	case pb.Visibility_VISIBILITY_PUBLIC:
		return port.VisibilityPublic
	case pb.Visibility_VISIBILITY_PRIVATE:
		return port.VisibilityPrivate
	case pb.Visibility_VISIBILITY_PROTECTED:
		return port.VisibilityProtected
	case pb.Visibility_VISIBILITY_PACKAGE:
		return port.VisibilityPackage
	default:
		return port.VisibilityUnspecified
	}
}
