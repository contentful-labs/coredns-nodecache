// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"

	"go.uber.org/mock/mockgen/model"
)

func getTypeSpecTypeParams(ts *ast.TypeSpec) []*ast.Field {
	if ts == nil || ts.TypeParams == nil {
		return nil
	}
	return ts.TypeParams.List
}

func (p *fileParser) parseGenericType(pkg string, typ ast.Expr, tps map[string]model.Type) (model.Type, error) {
	switch v := typ.(type) {
	case *ast.IndexExpr:
		m, err := p.parseType(pkg, v.X, tps)
		if err != nil {
			return nil, err
		}
		nm, ok := m.(*model.NamedType)
		if !ok {
			return m, nil
		}
		t, err := p.parseType(pkg, v.Index, tps)
		if err != nil {
			return nil, err
		}
		nm.TypeParams = &model.TypeParametersType{TypeParameters: []model.Type{t}}
		return m, nil
	case *ast.IndexListExpr:
		m, err := p.parseType(pkg, v.X, tps)
		if err != nil {
			return nil, err
		}
		nm, ok := m.(*model.NamedType)
		if !ok {
			return m, nil
		}
		var ts []model.Type
		for _, expr := range v.Indices {
			t, err := p.parseType(pkg, expr, tps)
			if err != nil {
				return nil, err
			}
			ts = append(ts, t)
		}
		nm.TypeParams = &model.TypeParametersType{TypeParameters: ts}
		return m, nil
	}
	return nil, nil
}

func (p *fileParser) parseGenericMethod(field *ast.Field, it *namedInterface, iface *model.Interface, pkg string, tps map[string]model.Type) ([]*model.Method, error) {
	var indices []ast.Expr
	var typ ast.Expr
	switch v := field.Type.(type) {
	case *ast.IndexExpr:
		indices = []ast.Expr{v.Index}
		typ = v.X
	case *ast.IndexListExpr:
		indices = v.Indices
		typ = v.X
	case *ast.UnaryExpr:
		if v.Op == token.TILDE {
			return nil, errConstraintInterface
		}
		return nil, fmt.Errorf("~T may only appear as constraint for %T", field.Type)
	case *ast.BinaryExpr:
		if v.Op == token.OR {
			return nil, errConstraintInterface
		}
		return nil, fmt.Errorf("A|B may only appear as constraint for %T", field.Type)
	default:
		return nil, fmt.Errorf("don't know how to mock method of type %T", field.Type)
	}

	nf := &ast.Field{
		Doc:     field.Comment,
		Names:   field.Names,
		Type:    typ,
		Tag:     field.Tag,
		Comment: field.Comment,
	}

	it.embeddedInstTypeParams = indices

	return p.parseMethod(nf, it, iface, pkg, tps)
}

var errConstraintInterface = errors.New("interface contains constraints")
