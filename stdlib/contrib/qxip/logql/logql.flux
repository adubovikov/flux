// Package query provides functions meant to simplify common logql queries.
//
// The primary function in this package is `logql.query_range()`
//
// ## Metadata
// introduced: 0.187.1
//
package logql

import "csv"
import "date"
import "experimental"
import "experimental/http/requests"

// query queries data from a specified LogQL query within given time bounds,
// filters data by query, timerange, and optional limit expressions.
//
// ## Parameters
// - url: LogQL/qryn API.
// - limit: Query limit. Default is 100.
// - query: LogQL query to execute.
// - start: Earliest time to include in results. Default is `-1h`.
//
//   Results include points that match the specified start time.
//   Use a relative duration, absolute time, or integer (Unix timestamp in nanoseconds).
//   For example, `-1h`, `2019-08-28T22:00:00.801064Z`, or `1545136086801064`.
//
// - end: Latest time to include in results. Default is `now()`.
//
//   Results exclude points that match the specified stop time.
//   Use a relative duration, absolute time, or integer (Unix timestamp in nanoseconds).
//   For example, `now()`, `2019-08-28T22:00:00.801064Z`, or `1545136086801064`.
//
// - step: Query stepping. Default is 10.
// - orgid: Optional Organization Id for partitioning.
//
// ## Examples
// ### Query specific fields in a measurement from LogQL/qryn
// ```no_run
// import "contrib/qxip/logql"
//
// logql.query_range(
//     url: "http://qryn:3100",
//     path: "/loki/api/v1/query_range",
//     start: -1h,
//     end: now(),
//     query: "{job=\"dummy-server\"}",
//     limit: "100",
// )
// ```
//
// ## Metadata
// tags: inputs
//
query_range = (
    url="http://127.0.0.1:3100",
    path="/loki/api/v1/query_range",
    query="",
    limit="100",
    step="10",
    start=-1h,
    end=now(),
    orgid="",
) =>
   {
    dstart = date.sub(from: start, d: 0d)
    dend = date.sub(from: end, d: 0d)
    response = requests.get(
      url: url + path,
      params: ["query": [query], "limit": [limit], "start": [string(v: uint(v: dstart))], "end": [string(v: uint(v: dend))], "step": [step], "csv": ["1"]],
      headers: if orgid != "" then ["X-Scope-OrgID": orgid] else ["X-Scope-OrgID": "0"],
      body: bytes(v: query)
    )
    return csv.from(csv: string(v: response.body), mode: "raw")
   }
