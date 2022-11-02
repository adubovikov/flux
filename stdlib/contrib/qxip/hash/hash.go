package hash

import (
        "context"
        "fmt"

        "github.com/influxdata/flux/runtime"
        "github.com/influxdata/flux/semantic"
        "github.com/influxdata/flux/values"
        "github.com/influxdata/flux/internal/errors"
        "github.com/influxdata/flux/codes"
        "github.com/cespare/xxhash/v2"
)

var hashFuncName = "hash"

var SpecialFns map[string]values.Function

func init() {
  SpecialFns = map[string]values.Function{
     "test": values.NewFunction(
        "test",
        runtime.MustLookupBuiltinType("hash", "test"),
        func(ctx context.Context, args values.Object) (values.Value, error) {
          v, ok := args.Get("v")
          if !ok {
            return nil, errors.New(codes.Invalid, "missing argument v")
          }
          if !v.IsNull() && v.Type().Nature() == semantic.String {
            value := xxhash.Sum64([]byte(v.Str())) // v.Str()
            return values.NewString(fmt.Sprint(value)), nil
          }
          return nil, errors.Newf(codes.Invalid, "cannot hash value %v", v)
        },
        false,
     ),
  }

  runtime.RegisterPackageValue("hash", "test", SpecialFns["test"])

}
