// Package query provides functions meant to simplify clickhouse http queries.
//
// The primary function in this package is `clickhouse.query()`
//
// ## Metadata
// introduced: 0.187.1
//
package clickhouse

import "csv"
import "experimental"
import "experimental/http/requests"

// query queries data from a specified ClickHouse query with the provided parameters
//
// ## Parameters
// - url: ClickHouse HTTP API.
// - query: ClickHouse query to execute.
// - limit: Query rows limit.
// - max_bytes: Query bytes limit.
// - format: Query format. Default CSVWithNames.
//
// ## Example
// ```no_run
// import "contrib/qxip/clickhouse"
//
// clickhouse.query(
//   url: "https://play@play.clickhouse.com",
//   query: "SELECT version()"
// )
// ```
//
// ## Metadata
// tags: inputs
//
query = (
    url="http://127.0.0.1:8123",
    query="SELECT 1",
    limit="100",
    cors="1",
    max_bytes="10000000",
    format="CSVWithNames",
) =>
   {
    response = requests.get(
      url: url,
      params: ["query": [query], "default_format": [format], "max_result_rows": [limit], "max_result_bytes": [max_bytes], "add_http_cors_header": [cors]],
      headers: ["X-ClickHouse-Format": string(v: format)],
    )
    return csv.from(csv: string(v: response.body), mode: "raw")
   }
