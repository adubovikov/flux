package compiler

import (
	"errors"
	"fmt"

	"github.com/influxdata/flux"

	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func Compile(f *semantic.FunctionExpression, functionType semantic.Type, builtins Scope) (Func, error) {
	if functionType.Kind() != semantic.Function {
		return nil, errors.New("type must be a function kind")
	}
	f = f.Copy().(*semantic.FunctionExpression)
	declarations := externDeclarations(builtins)
	extern := &semantic.Extern{
		Declarations: declarations,
		Block:        &semantic.ExternBlock{Node: f},
	}

	semantic.Infer(extern)

	pt, err := extern.PolyType()
	if err != nil {
		return nil, err
	}
	if err := pt.Unify(functionType.PolyType()); err != nil {
		return nil, err
	}
	typ, mono := pt.Type()
	if !mono {
		return nil, errors.New("cannot compile polymorphic function")
	}

	root, err := compile(f.Block.Body, builtins)
	if err != nil {
		return nil, err
	}
	return compiledFn{
		root:   root,
		fnType: functionType,
	}, nil
}

func compile(n semantic.Node, builtIns Scope) (Evaluator, error) {
	switch n := n.(type) {
	case *semantic.BlockStatement:
		body := make([]Evaluator, len(n.Body))
		for i, s := range n.Body {
			node, err := compile(s, builtIns)
			if err != nil {
				return nil, err
			}
			body[i] = node
		}
		return &blockEvaluator{
			t:    n.ReturnStatement().Argument.Type(),
			body: body,
		}, nil
	case *semantic.ExpressionStatement:
		return nil, errors.New("statement does nothing, sideffects are not supported by the compiler")
	case *semantic.ReturnStatement:
		node, err := compile(n.Argument, builtIns)
		if err != nil {
			return nil, err
		}
		return returnEvaluator{
			Evaluator: node,
		}, nil
	case *semantic.NativeVariableDeclaration:
		node, err := compile(n.Init, builtIns)
		if err != nil {
			return nil, err
		}
		return &declarationEvaluator{
			t:    n.Init.Type(),
			id:   n.Identifier.Name,
			init: node,
		}, nil
	case *semantic.ObjectExpression:
		properties := make(map[string]Evaluator, len(n.Properties))
		propertyTypes := make(map[string]semantic.Type, len(n.Properties))
		for _, p := range n.Properties {
			node, err := compile(p.Value, builtIns)
			if err != nil {
				return nil, err
			}
			properties[p.Key.Name] = node
			propertyTypes[p.Key.Name] = node.Type()
		}
		return &objEvaluator{
			t:          semantic.NewObjectType(propertyTypes),
			properties: properties,
		}, nil
	case *semantic.IdentifierExpression:
		if v, ok := builtIns[n.Name]; ok {
			//Resolve any built in identifiers now
			return &valueEvaluator{
				value: v,
			}, nil
		}
		return &identifierEvaluator{
			t:    n.Type(),
			name: n.Name,
		}, nil
	case *semantic.MemberExpression:
		object, err := compile(n.Object, builtIns)
		if err != nil {
			return nil, err
		}
		return &memberEvaluator{
			t:        n.Type(),
			object:   object,
			property: n.Property,
		}, nil
	case *semantic.BooleanLiteral:
		return &booleanEvaluator{
			t: n.Type(),
			b: n.Value,
		}, nil
	case *semantic.IntegerLiteral:
		return &integerEvaluator{
			t: n.Type(),
			i: n.Value,
		}, nil
	case *semantic.FloatLiteral:
		return &floatEvaluator{
			t: n.Type(),
			f: n.Value,
		}, nil
	case *semantic.StringLiteral:
		return &stringEvaluator{
			t: n.Type(),
			s: n.Value,
		}, nil
	case *semantic.RegexpLiteral:
		return &regexpEvaluator{
			t: n.Type(),
			r: n.Value,
		}, nil
	case *semantic.DateTimeLiteral:
		return &timeEvaluator{
			t:    n.Type(),
			time: values.ConvertTime(n.Value),
		}, nil
	case *semantic.UnaryExpression:
		node, err := compile(n.Argument, builtIns)
		if err != nil {
			return nil, err
		}
		return &unaryEvaluator{
			t:    n.Type(),
			node: node,
		}, nil
	case *semantic.LogicalExpression:
		l, err := compile(n.Left, builtIns)
		if err != nil {
			return nil, err
		}
		r, err := compile(n.Right, builtIns)
		if err != nil {
			return nil, err
		}
		return &logicalEvaluator{
			t:        n.Type(),
			operator: n.Operator,
			left:     l,
			right:    r,
		}, nil
	case *semantic.BinaryExpression:
		l, err := compile(n.Left, builtIns)
		if err != nil {
			return nil, err
		}
		lt := l.Type()
		r, err := compile(n.Right, builtIns)
		if err != nil {
			return nil, err
		}
		rt := r.Type()
		f, err := values.LookupBinaryFunction(values.BinaryFuncSignature{
			Operator: n.Operator,
			Left:     lt,
			Right:    rt,
		})
		if err != nil {
			return nil, err
		}
		return &binaryEvaluator{
			t:     n.Type(),
			left:  l,
			right: r,
			f:     f,
		}, nil
	case *semantic.CallExpression:
		callee, err := compile(n.Callee, builtIns)
		if err != nil {
			return nil, err
		}
		args, err := compile(n.Arguments, builtIns)
		if err != nil {
			return nil, err
		}
		return &callEvaluator{
			t:      n.Type(),
			callee: callee,
			args:   args,
		}, nil
	case *semantic.FunctionExpression:
		body, err := compile(n.Body, builtIns)
		if err != nil {
			return nil, err
		}
		params := make([]functionParam, len(n.Params))
		for i, param := range n.Params {
			params[i] = functionParam{
				Key:  param.Key.Name,
				Type: param.Type(),
			}
			if param.Default != nil {
				d, err := compile(param.Default, builtIns)
				if err != nil {
					return nil, err
				}
				params[i].Default = d
			}
		}
		return &functionEvaluator{
			t:      n.Type(),
			params: params,
			body:   body,
		}, nil
	default:
		return nil, fmt.Errorf("unknown semantic node of type %T", n)
	}
}

// CompilationCache caches compilation results based on the type of the function.
type CompilationCache struct {
	fn       *semantic.FunctionExpression
	scope    Scope
	compiled map[semantic.Type]funcErr
}

func NewCompilationCache(fn *semantic.FunctionExpression, scope Scope) *CompilationCache {
	return &CompilationCache{
		fn:       fn,
		scope:    scope,
		compiled: make(map[semantic.Type]Func),
	}
}

// Compile returnes a compiled function bsaed on the provided type.
// The result will be cached for subsequent calls.
func (c *CompilationCache) Compile(fnType semantic.Type) (Func, error) {
	f, ok := c.compiled[fnType]
	if ok {
		return f.F, f.Err
	}
	f, err := c.compile(fnType)
	c.compiled[fnType] = funcErr{
		F:   f,
		Err: err,
	}
	return f, err
}

type funcErr struct {
	F   Func
	Err error
}

// compile recursively searches for a matching child node that has compiled the function.
// If the compilation has not been performed previously its result is cached and returned.
func (c *CompilationCache) compile(fnType semantic.Type) (Func, error) {
	Compile(c.fn, types, c.scope)
}

// Utility function for compiling an `fn` parameter for rename or drop/keep. In addition
// to the function expression, it takes two types to verify the result against:
// a single argument type, and a single return type.
func CompileFnParam(fn *semantic.FunctionExpression, paramType, returnType semantic.Type) (Func, string, error) {
	scope, decls := flux.BuiltIns()
	compileCache := NewCompilationCache(fn, scope, decls)
	if len(fn.Params) != 1 {
		return nil, "", fmt.Errorf("function should only have a single parameter, got %d", len(fn.Params))
	}
	paramName := fn.Params[0].Key.Name

	compiled, err := compileCache.Compile(map[string]semantic.Type{
		paramName: paramType,
	})
	if err != nil {
		return nil, "", err
	}

	if compiled.Type() != returnType {
		return nil, "", fmt.Errorf("provided function does not evaluate to type %s", returnType.Kind())
	}

	return compiled, paramName, nil
}

// externDeclarations produces a list of external declarations from a scope
func externDeclarations(scope Scope) []*semantic.ExternalVariableDeclaration {
	declarations := make([]*semantic.ExternalVariableDeclaration, len(scope))
	for k, v := range scope {
		declarations = append(declarations, &semantic.ExternalVariableDeclaration{
			Identifier: &semantic.Identifier{Name: k},
			ExternType: v.Type(),
		})
	}
	return declarations
}
