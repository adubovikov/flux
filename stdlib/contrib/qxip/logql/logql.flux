// Package query provides functions meant to simplify common logql queries.
//
// The primary function in this package is `logql.query_range()`
//
// ## Metadata
// introduced: 0.187.1
//
package logql

import "csv"
import "experimental"
import "experimental/http/requests"

// query queries data from a specified LogQL query within given time bounds,
// filters data by query, timerange, and optional limit expressions.
//
// ## Parameters
// - url: LogQL/qryn API.
// - limit: Query limit.
// - query: LogQL query to execute.
// - start: Earliest time to include in results. Default is `-1h`.
//
//   Results include points that match the specified start time.
//   Use a relative duration, absolute time, or integer (Unix timestamp in nanoseconds).
//   For example, `-1h`, `2019-08-28T22:00:00.801064Z`, or `1545136086801064`.
//   Timestamps are expressed as `uint()`. For example: `uint(v: -1h  )`
//
// - stop: Latest time to include in results. Default is `now()`.
//
//   Results exclude points that match the specified stop time.
//   Use a relative duration, absolute time, or integer (Unix timestamp in nanoseconds).
//   For example, `now()`, `2019-08-28T22:00:00.801064Z`, or `1545136086801064`.
//   Timestamps are expressed as `uint()`. For example: `uint(v: now()  )`
//
// ## Examples
// ### Query specific fields in a measurement from LogQL/qryn
// ```no_run
// import "contrib/qxip/logql"
//
// logql.query_range(
//     url: "http://qryn:3100",
//     start: uint(v: -1h  ),
//     stop: uint(v: now() ),
//     query: "{\"job\"=\"dummy-server\"}",
//     limit: 100, 
// )
// ```
//
// ## Metadata
// tags: inputs
//
query_range = (
    url="http://127.0.0.1:3100",
    query,
    limit=100,
    start=uint(v: -1h ),
    stop=uint(v: now() ),
) =>
    response = requests.get(
      url: url + "/loki/api/v1/query_range?query=" + query + "&limit=" + limit + "&start=" + start + "&end=" + stop + "&step=0&csv=1",
      body: bytes(v: query)
    )
    csv.from(csv: string(v: response.body), mode: "raw")
