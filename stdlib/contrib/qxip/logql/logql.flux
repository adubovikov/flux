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
// - url: InfluxDB bucket name.
// - limit: Query limit.
// - query: LogQL query to execute.
// - start: Earliest time to include in results.
//
//   Results include points that match the specified start time.
//   Use a relative duration, absolute time, or integer (Unix timestamp in seconds).
//   For example, `-1h`, `2019-08-28T22:00:00Z`, or `1567029600`.
//   Durations are relative to `now()`.
//
// - stop: Latest time to include in results. Default is `now()`.
//
//   Results exclude points that match the specified stop time.
//   Use a relative duration, absolute time, or integer (Unix timestamp in seconds).For example, `-1h`, `2019-08-28T22:00:00Z`, or `1567029600`.
//   Durations are relative to `now()`.
//
// ## Examples
// ### Query specific fields in a measurement from InfluxDB
// ```no_run
// import "contrib/qxip/logql"
//
// logql.query_range(
//     url: "http://qryn:3100",
//     start: -1h,
//     stop: now(),
//     query: "{\"job\"=\"dummy-server\"}",
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
    start,
    stop=now(),
) =>
    response = requests.get(
      url: url + "/loki/api/v1/query_range?query=" + query + "&limit=" + limit + "&start=" + start + "&end=" +end + "&step=0&csv=1",
      body: bytes(v: query)
    )
    csv.from(csv: string(v: response.body), mode: "raw")
